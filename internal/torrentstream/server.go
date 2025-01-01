package torrentstream

import (
	"bufio"
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"net"
	"strconv"
	"strings"
)

type (
	// serverManager manages the streaming server
	serverManager struct {
		repository *Repository
	}
)

// ref: torrserver
func dnsResolve() {
	addrs, _ := net.LookupHost("www.google.com")
	if len(addrs) == 0 {
		//fmt.Println("Check dns failed", addrs, err)

		fn := func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", "1.1.1.1:53")
		}

		net.DefaultResolver = &net.Resolver{
			Dial: fn,
		}

		addrs, _ = net.LookupHost("www.google.com")
		//fmt.Println("Check cloudflare dns", addrs, err)
	} else {
		//fmt.Println("Check dns OK", addrs, err)
	}
}

// newServerManager is called once during the lifetime of the application.
func newServerManager(repository *Repository) *serverManager {
	ret := &serverManager{
		repository: repository,
	}

	dnsResolve()

	return ret
}

func (s *serverManager) serve(c *fiber.Ctx) error {
	s.repository.logger.Trace().Msg("torrentstream: Stream endpoint hit [server]")

	if s.repository.client.currentFile.IsAbsent() || s.repository.client.currentTorrent.IsAbsent() {
		s.repository.logger.Error().Msg("torrentstream: No torrent to stream [server]")
		return c.Status(fiber.StatusNotFound).SendString("No torrent to stream")
	}

	file := s.repository.client.currentFile.MustGet()
	fileSize := file.FileInfo().Length

	// Parse range header
	rangeHeader := c.Get("Range")
	var start, end int64 = 0, fileSize - 1

	if rangeHeader != "" {
		// Parse the Range header value
		segments := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
		if len(segments) == 2 {
			var err error
			if segments[0] != "" {
				start, err = strconv.ParseInt(segments[0], 10, 64)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).SendString("Invalid range start")
				}
			}
			if segments[1] != "" {
				end, err = strconv.ParseInt(segments[1], 10, 64)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).SendString("Invalid range end")
				}
			}

			if start >= fileSize || end >= fileSize || start > end {
				return c.Status(fiber.StatusRequestedRangeNotSatisfiable).SendString("Invalid range")
			}

			c.Status(fiber.StatusPartialContent)
			c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
		}
	}

	c.Set("Content-Length", fmt.Sprintf("%d", end-start+1))
	c.Set("Content-Type", "video/mp4")
	c.Set("Accept-Ranges", "bytes")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		reader := file.NewReader()
		defer reader.Close()

		reader.SetResponsive()
		//reader.SetReadahead(file.FileInfo().Length / 50)

		// Seek to start position if needed
		if start > 0 {
			_, err := reader.Seek(start, io.SeekStart)
			if err != nil {
				s.repository.logger.Error().Err(err).Msg("torrentstream: Failed to seek to position")
				return
			}
		}

		// Create a limited reader to handle the range
		limitedReader := io.LimitReader(reader, end-start+1)
		bufferedReader := bufio.NewReaderSize(limitedReader, 1024*1024) // 1MB buffer

		// Copy the data
		_, err := io.Copy(w, bufferedReader)
		if err != nil {
			//s.repository.logger.Error().Err(err).Msg("torrentstream: Error while streaming")
			return
		}

		w.Flush()
	})

	return nil
}
