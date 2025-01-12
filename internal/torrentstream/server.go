package torrentstream

import (
	"context"
	"github.com/anacrolix/torrent"
	"net"
	"net/http"
	"time"
)

type (
	// serverManager manages the streaming server
	serverManager struct {
		repository    *Repository
		lastUsed      time.Time // Used to track the last time the server was used
		serverRunning bool      // Whether the server is running
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

func (s *serverManager) initializeServer() {
	// no-op
}

func (s *serverManager) createServer() {
	// no-op
}

func (s *serverManager) startServer() {
	// no-op
}

// stopServer stops the streaming server.
func (s *serverManager) stopServer() {
	// no-op
}

func (s *serverManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
