package torrentstream

import (
	"context"
	"errors"
	"fmt"
	"github.com/samber/mo"
	"net"
	"net/http"
	"time"
)

type (
	// serverManager manages the streaming server
	serverManager struct {
		httpserver    mo.Option[*http.Server] // The server instance
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

func newServerManager(repository *Repository) *serverManager {
	ret := &serverManager{
		repository: repository,
		httpserver: mo.None[*http.Server](),
	}

	dnsResolve()

	http.HandleFunc("/stream", func(w http.ResponseWriter, _r *http.Request) {
		ret.lastUsed = time.Now()
		ret.repository.logger.Trace().Msg("torrentstream: Stream endpoint hit")
		w.Header().Set("Content-Type", "video/mp4")

		if ret.repository.client.currentFile.IsAbsent() {
			ret.repository.logger.Error().Msg("torrentstream: No torrent to stream")
			http.Error(w, "No torrent to stream", http.StatusNotFound)
			return
		}

		file := ret.repository.client.currentFile.MustGet()
		fr := file.NewReader()
		defer fr.Close()
		//fr.SetReadahead(file.FileInfo().Length / 100)
		fr.SetResponsive()

		http.ServeContent(
			w,
			_r,
			file.DisplayPath(),
			time.Now(),
			fr,
		)
	})

	http.HandleFunc("/ping", func(w http.ResponseWriter, _r *http.Request) {
		w.Write([]byte("pong"))
	})

	// DEVNOTE: Currently can't accurately track the last time the server was used
	// This risks stopping the server while it's being used
	// FIXME - Find a way to get the playback manager to refresh the lastUsed time
	//go func() {
	//	for {
	//		// Stop the server if it hasn't been used for 5 minutes
	//		if time.Since(ret.lastUsed) > 5*time.Minute && ret.serverRunning {
	//			ret.StopServer()
	//		}
	//		time.Sleep(10 * time.Minute)
	//		ret.repository.logger.Debug().Msg("torrentstream: Stream server health check")
	//	}
	//}()

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
}
