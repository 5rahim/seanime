package torrentstream

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/mo"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type (
	// serverManager manages the streaming server
	serverManager struct {
		repository    *Repository
		httpserver    mo.Option[*http.Server] // The server instance
		lastUsed      time.Time               // Used to track the last time the server was used
		serverRunning bool                    // Whether the server is running
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
		httpserver: mo.None[*http.Server](),
	}

	dnsResolve()

	http.HandleFunc("/stream/", ret.httpServe)

	http.HandleFunc("/ping", func(w http.ResponseWriter, _r *http.Request) {
		w.Write([]byte("pong"))
	})

	return ret
}

// initializeServer overrides the server with a new one, whether it exists or not.
// Unlike CreateServer, this will close the existing server if it exists.
// Useful when the settings are changed.
func (s *serverManager) initializeServer() {
	if s.repository.settings.IsAbsent() {
		s.repository.logger.Error().Msg("torrentstream: No settings found, cannot initialize the streaming server")
		return
	}

	existingServer, exists := s.httpserver.Get()
	if exists {
		err := existingServer.Close()
		if err != nil {
			s.repository.logger.Error().Err(err).Msg("torrentstream: Failed to close existing streaming server")
			return
		}
	}

	s.httpserver = mo.None[*http.Server]()
	s.serverRunning = false
	s.createServer()
}

// createServer creates the streaming server.
// If the server is already present, it won't create a new one.
func (s *serverManager) createServer() {
	if s.repository.settings.IsAbsent() {
		s.repository.logger.Error().Msg("torrentstream: No settings found, cannot create the server")
		return
	}

	if s.httpserver.IsPresent() {
		return
	}

	host := s.repository.settings.MustGet().StreamingServerHost
	port := s.repository.settings.MustGet().StreamingServerPort

	s.repository.logger.Info().Msgf("torrentstream: Creating streaming server on %s:%d", host, port)

	// Create the server
	// Default address is "0.0.0.0:43214"
	server := &http.Server{
		Addr: fmt.Sprintf("%s:%d", host, port),
	}

	s.httpserver = mo.Some(server)
}

// startServer starts the streaming server.
// If the server is already running, it won't start a new one.
// This is safe to call
func (s *serverManager) startServer() {
	server, exists := s.httpserver.Get()
	if !exists {
		s.repository.logger.Error().Msg("torrentstream: No streaming server found, cannot start the server")
		return
	}

	if s.serverRunning {
		return
	}

	s.repository.logger.Debug().Msg("torrentstream: Starting the streaming server")

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		s.repository.logger.Error().Err(err).Msg("torrentstream: Failed to start the streaming server")
		return
	}

	s.repository.logger.Info().Msgf("torrentstream: Streaming server started on %s", server.Addr)

	go func() {
		s.serverRunning = true
		defer func() {
			s.serverRunning = false
		}()
		if err := server.Serve(ln); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
		}
	}()
}

// stopServer stops the streaming server.
func (s *serverManager) stopServer() {
	server, exists := s.httpserver.Get()
	if !exists {
		return
	}

	if !s.serverRunning {
		return
	}

	if err := server.Close(); err != nil {
		s.repository.logger.Error().Err(err).Msg("torrentstream: Failed to stop the streaming server")
	}
	s.serverRunning = false

	s.repository.logger.Info().Msg("torrentstream: Streaming server stopped")

	// Do not forget to reinitialize the server
	s.initializeServer()
}

func (s *serverManager) httpServe(w http.ResponseWriter, r *http.Request) {
	s.lastUsed = time.Now()
	s.repository.logger.Trace().Msg("torrentstream: Stream endpoint hit [server]")

	if s.repository.client.currentFile.IsAbsent() || s.repository.client.currentTorrent.IsAbsent() {
		s.repository.logger.Error().Msg("torrentstream: No torrent to stream [server]")
		http.Error(w, "No torrent to stream", http.StatusNotFound)
		return
	}

	file := s.repository.client.currentFile.MustGet()
	tr := file.NewReader()
	defer func(tr torrent.Reader) {
		_ = tr.Close()
	}(tr)
	tr.SetResponsive()
	tr.SetReadahead(file.FileInfo().Length / 100)

	s.repository.logger.Trace().Str("file", file.DisplayPath()).Msg("torrentstream: Serving file content")
	w.Header().Set("Content-Type", "video/mp4")
	http.ServeContent(
		w,
		r,
		file.DisplayPath(),
		time.Now(),
		tr,
	)
	s.repository.logger.Trace().Msg("torrentstream: File content served")
}

func (s *serverManager) serve(c *fiber.Ctx) error {
	file := s.repository.client.currentFile.MustGet()
	fileSize := file.FileInfo().Length

	rangeHeader := c.Get("Range")
	var start, end int64 = 0, fileSize - 1

	if rangeHeader != "" {
		ranges := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
		if len(ranges) == 2 {
			if ranges[0] != "" {
				start, _ = strconv.ParseInt(ranges[0], 10, 64)
			}
			if ranges[1] != "" {
				end, _ = strconv.ParseInt(ranges[1], 10, 64)
			}
		}
		c.Status(fiber.StatusPartialContent)
		c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	}

	c.Set("Content-Length", fmt.Sprintf("%d", end-start+1))
	c.Set("Content-Type", "video/mp4")
	c.Set("Accept-Ranges", "bytes")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		reader := file.NewReader()
		defer reader.Close()

		if _, err := reader.Seek(start, io.SeekStart); err != nil {
			s.repository.logger.Error().Err(err).Msg("Failed to seek to position")
			return
		}

		buf := make([]byte, 32*1024) // 32 KB buffer
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				if _, writeErr := w.Write(buf[:n]); writeErr != nil {
					break
				}
				w.Flush() // Ensure timely delivery
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				s.repository.logger.Error().Err(err).Msg("Error during streaming")
				return
			}
		}
	})

	return nil
}

//func (s *serverManager) serve(c *fiber.Ctx) error {
//	s.repository.logger.Trace().Msg("torrentstream: Stream endpoint hit [server]")
//
//	if s.repository.client.currentFile.IsAbsent() || s.repository.client.currentTorrent.IsAbsent() {
//		s.repository.logger.Error().Msg("torrentstream: No torrent to stream [server]")
//		return c.Status(fiber.StatusNotFound).SendString("No torrent to stream")
//	}
//
//	file := s.repository.client.currentFile.MustGet()
//	fileSize := file.FileInfo().Length
//
//	// Parse range header
//	rangeHeader := c.Get("Range")
//	s.repository.logger.Trace().Str("range", rangeHeader).Msg("torrentstream: Range header")
//	var start, end int64 = 0, fileSize - 1
//
//	if rangeHeader != "" {
//		segments := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
//		if len(segments) == 2 {
//			var err error
//			if segments[0] != "" {
//				start, err = strconv.ParseInt(segments[0], 10, 64)
//				if err != nil {
//					return c.Status(fiber.StatusBadRequest).SendString("Invalid range start")
//				}
//			}
//			if segments[1] != "" {
//				end, err = strconv.ParseInt(segments[1], 10, 64)
//				if err != nil {
//					return c.Status(fiber.StatusBadRequest).SendString("Invalid range end")
//				}
//			}
//
//			if start >= fileSize || end >= fileSize || start > end {
//				return c.Status(fiber.StatusRequestedRangeNotSatisfiable).SendString("Invalid range")
//			}
//
//			c.Status(fiber.StatusPartialContent)
//			c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
//		}
//	}
//
//	s.repository.logger.Trace().Int64("start", start).Int64("end", end).Msg("torrentstream: Range values")
//
//	c.Set("Content-Length", fmt.Sprintf("%d", end-start+1))
//	c.Set("Content-Type", "video/mp4")
//	c.Set("Accept-Ranges", "bytes")
//	c.Set("Connection", "keep-alive")
//
//	const initialBufferSize = 32 * 1024
//
//	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
//		reader := file.NewReader()
//		defer reader.Close()
//
//		reader.SetResponsive()
//
//		// Dynamically adjust readahead based on requested range
//		rangeSize := end - start + 1
//		readaheadSize := int64(1024 * 1024) // default 1MB
//		if rangeSize < 1024*1024*5 {        // if requesting less than 5MB
//			readaheadSize = rangeSize / 2
//		} else if rangeSize > 1024*1024*50 { // if requesting more than 50MB
//			readaheadSize = 1024 * 1024 * 2 // increase to 2MB
//		}
//		reader.SetReadahead(readaheadSize)
//
//		if start > 0 {
//			_, err := reader.Seek(start, io.SeekStart)
//			if err != nil {
//				s.repository.logger.Error().Err(err).Msg("torrentstream: Failed to seek to position")
//				return
//			}
//		}
//
//		limitedReader := io.LimitReader(reader, end-start+1)
//
//		// Use a smaller buffer initially
//		bufferedReader := bufio.NewReaderSize(limitedReader, initialBufferSize)
//
//		// Copy with a smaller buffer size for the initial data
//		initialBuf := make([]byte, initialBufferSize)
//		n, err := bufferedReader.Read(initialBuf)
//		if err != nil && err != io.EOF {
//			s.repository.logger.Error().Err(err).Msg("torrentstream: Error reading initial data")
//			return
//		}
//
//		// Write the initial chunk immediately
//		if n > 0 {
//			if _, err := w.Write(initialBuf[:n]); err != nil {
//				s.repository.logger.Error().Err(err).Msg("torrentstream: Error writing initial data")
//				return
//			}
//			s.repository.logger.Trace().Int("size", n).Msg("torrentstream: Initial data written")
//			w.Flush()
//		}
//
//		// After initial data is sent, increase buffer size for better throughput
//		bufferedReader = bufio.NewReaderSize(limitedReader, 1024*1024) // 1MB for subsequent reads
//
//		// Continue with the rest of the data
//		_, err = io.Copy(w, bufferedReader)
//		if err != nil {
//			return
//		}
//
//		s.repository.logger.Trace().Msg("torrentstream: Stream completed")
//
//		w.Flush()
//	})
//
//	return nil
//}
