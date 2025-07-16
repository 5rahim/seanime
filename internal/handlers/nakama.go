package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/anime"
	"seanime/internal/nakama"
	"seanime/internal/util"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
)

// HandleNakamaWebSocket handles WebSocket connections for Nakama peers
//
//	@summary handles WebSocket connections for Nakama peers.
//	@desc This endpoint handles WebSocket connections from Nakama peers when this instance is acting as a host.
//	@route /api/v1/nakama/ws [GET]
func (h *Handler) HandleNakamaWebSocket(c echo.Context) error {
	// Use the standard library HTTP ResponseWriter and Request
	w := c.Response().Writer
	r := c.Request()

	// Let the Nakama manager handle the WebSocket connection
	h.App.NakamaManager.HandlePeerConnection(w, r)
	return nil
}

// HandleSendNakamaMessage
//
//	@summary sends a custom message through Nakama.
//	@desc This allows sending custom messages to connected peers or the host.
//	@route /api/v1/nakama/message [POST]
//	@returns nakama.MessageResponse
func (h *Handler) HandleSendNakamaMessage(c echo.Context) error {
	type body struct {
		MessageType string      `json:"messageType"`
		Payload     interface{} `json:"payload"`
		PeerID      string      `json:"peerId,omitempty"` // If specified, send to specific peer (host only)
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	var err error
	if b.PeerID != "" && h.App.Settings.GetNakama().IsHost {
		// Send to specific peer
		err = h.App.NakamaManager.SendMessageToPeer(b.PeerID, nakama.MessageType(b.MessageType), b.Payload)
	} else if h.App.Settings.GetNakama().IsHost {
		// Send to all peers
		err = h.App.NakamaManager.SendMessage(nakama.MessageType(b.MessageType), b.Payload)
	} else {
		// Send to host
		err = h.App.NakamaManager.SendMessageToHost(nakama.MessageType(b.MessageType), b.Payload)
	}

	if err != nil {
		return h.RespondWithError(c, err)
	}

	response := &nakama.MessageResponse{
		Success: true,
		Message: "Message sent successfully",
	}

	return h.RespondWithData(c, response)
}

// HandleGetNakamaAnimeLibrary
//
//	@summary shares the local anime collection with Nakama clients.
//	@desc This creates a new LibraryCollection struct and returns it.
//	@desc This is used to share the local anime collection with Nakama clients.
//	@route /api/v1/nakama/host/anime/library/collection [GET]
//	@returns nakama.NakamaAnimeLibrary
func (h *Handler) HandleGetNakamaAnimeLibrary(c echo.Context) error {
	if !h.App.Settings.GetNakama().HostShareLocalAnimeLibrary {
		return h.RespondWithError(c, errors.New("host is not sharing its anime library"))
	}

	animeCollection, err := h.App.GetAnimeCollection(false)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	if animeCollection == nil {
		return h.RespondWithData(c, &anime.LibraryCollection{})
	}

	lfs, _, err := db_bridge.GetLocalFiles(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	unsharedAnimeIds := h.App.Settings.GetNakama().HostUnsharedAnimeIds
	unsharedAnimeIdsMap := make(map[int]struct{})
	for _, id := range unsharedAnimeIds {
		unsharedAnimeIdsMap[id] = struct{}{}
	}
	if len(unsharedAnimeIds) > 0 {
		lfs = lo.Filter(lfs, func(lf *anime.LocalFile, _ int) bool {
			_, ok := unsharedAnimeIdsMap[lf.MediaId]
			return !ok
		})
	}

	libraryCollection, err := anime.NewLibraryCollection(c.Request().Context(), &anime.NewLibraryCollectionOptions{
		AnimeCollection:  animeCollection,
		Platform:         h.App.AnilistPlatform,
		LocalFiles:       lfs,
		MetadataProvider: h.App.MetadataProvider,
	})
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Hydrate total library size
	if libraryCollection != nil && libraryCollection.Stats != nil {
		libraryCollection.Stats.TotalSize = util.Bytes(h.App.TotalLibrarySize)
	}

	return h.RespondWithData(c, &nakama.NakamaAnimeLibrary{
		LocalFiles:      lfs,
		AnimeCollection: animeCollection,
	})
}

// HandleGetNakamaAnimeLibraryCollection
//
//	@summary shares the local anime collection with Nakama clients.
//	@desc This creates a new LibraryCollection struct and returns it.
//	@desc This is used to share the local anime collection with Nakama clients.
//	@route /api/v1/nakama/host/anime/library/collection [GET]
//	@returns anime.LibraryCollection
func (h *Handler) HandleGetNakamaAnimeLibraryCollection(c echo.Context) error {
	if !h.App.Settings.GetNakama().HostShareLocalAnimeLibrary {
		return h.RespondWithError(c, errors.New("host is not sharing its anime library"))
	}

	animeCollection, err := h.App.GetAnimeCollection(false)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	if animeCollection == nil {
		return h.RespondWithData(c, &anime.LibraryCollection{})
	}

	lfs, _, err := db_bridge.GetLocalFiles(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	unsharedAnimeIds := h.App.Settings.GetNakama().HostUnsharedAnimeIds
	unsharedAnimeIdsMap := make(map[int]struct{})
	for _, id := range unsharedAnimeIds {
		unsharedAnimeIdsMap[id] = struct{}{}
	}
	if len(unsharedAnimeIds) > 0 {
		lfs = lo.Filter(lfs, func(lf *anime.LocalFile, _ int) bool {
			_, ok := unsharedAnimeIdsMap[lf.MediaId]
			return !ok
		})
	}

	libraryCollection, err := anime.NewLibraryCollection(c.Request().Context(), &anime.NewLibraryCollectionOptions{
		AnimeCollection:  animeCollection,
		Platform:         h.App.AnilistPlatform,
		LocalFiles:       lfs,
		MetadataProvider: h.App.MetadataProvider,
	})
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Hydrate total library size
	if libraryCollection != nil && libraryCollection.Stats != nil {
		libraryCollection.Stats.TotalSize = util.Bytes(h.App.TotalLibrarySize)
	}

	return h.RespondWithData(c, libraryCollection)
}

// HandleGetNakamaAnimeLibraryFiles
//
//	@summary return the local files for the given AniList anime media id.
//	@desc This is used by the anime media entry pages to get all the data about the anime.
//	@route /api/v1/nakama/host/anime/library/files/{id} [POST]
//	@param id - int - true - "AniList anime media ID"
//	@returns []anime.LocalFile
func (h *Handler) HandleGetNakamaAnimeLibraryFiles(c echo.Context) error {
	if !h.App.Settings.GetNakama().HostShareLocalAnimeLibrary {
		return h.RespondWithError(c, errors.New("host is not sharing its anime library"))
	}

	mId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Get all the local files
	lfs, _, err := db_bridge.GetLocalFiles(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	retLfs := lo.Filter(lfs, func(lf *anime.LocalFile, _ int) bool {
		return lf.MediaId == mId
	})

	return h.RespondWithData(c, retLfs)
}

// HandleGetNakamaAnimeAllLibraryFiles
//
//	@summary return all the local files for the host.
//	@desc This is used to share the local anime collection with Nakama clients.
//	@route /api/v1/nakama/host/anime/library/files [POST]
//	@returns []anime.LocalFile
func (h *Handler) HandleGetNakamaAnimeAllLibraryFiles(c echo.Context) error {
	if !h.App.Settings.GetNakama().HostShareLocalAnimeLibrary {
		return h.RespondWithError(c, errors.New("host is not sharing its anime library"))
	}

	// Get all the local files
	lfs, _, err := db_bridge.GetLocalFiles(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	unsharedAnimeIds := h.App.Settings.GetNakama().HostUnsharedAnimeIds
	unsharedAnimeIdsMap := make(map[int]struct{})
	for _, id := range unsharedAnimeIds {
		unsharedAnimeIdsMap[id] = struct{}{}
	}
	if len(unsharedAnimeIds) > 0 {
		lfs = lo.Filter(lfs, func(lf *anime.LocalFile, _ int) bool {
			_, ok := unsharedAnimeIdsMap[lf.MediaId]
			return !ok
		})
	}

	return h.RespondWithData(c, lfs)
}

// HandleNakamaPlayVideo
//
//	@summary plays the media from the host.
//	@route /api/v1/nakama/play [POST]
//	@returns bool
func (h *Handler) HandleNakamaPlayVideo(c echo.Context) error {
	type body struct {
		Path         string `json:"path"`
		MediaId      int    `json:"mediaId"`
		AniDBEpisode string `json:"anidbEpisode"`
	}
	b := new(body)
	if err := c.Bind(b); err != nil {
		return h.RespondWithError(c, err)
	}

	if !h.App.NakamaManager.IsConnectedToHost() {
		return h.RespondWithError(c, errors.New("not connected to host"))
	}

	media, err := h.App.AnilistPlatform.GetAnime(c.Request().Context(), b.MediaId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	err = h.App.NakamaManager.PlayHostAnimeLibraryFile(b.Path, c.Request().Header.Get("User-Agent"), media, b.AniDBEpisode)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// Note: This is not used anymore. Each peer will independently stream the torrent.
// route /api/v1/nakama/host/torrentstream/stream
// Allows peers to stream the currently playing torrent.
func (h *Handler) HandleNakamaHostTorrentstreamServeStream(c echo.Context) error {
	h.App.TorrentstreamRepository.HTTPStreamHandler().ServeHTTP(c.Response().Writer, c.Request())
	return nil
}

var videoProxyClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		ForceAttemptHTTP2:   false, // Fixes issues on Linux
	},
	Timeout: 60 * time.Second,
}

// route /api/v1/nakama/host/debridstream/stream
// Allows peers to stream the currently playing torrent.
func (h *Handler) HandleNakamaHostDebridstreamServeStream(c echo.Context) error {
	streamUrl, ok := h.App.DebridClientRepository.GetStreamURL()
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "no stream url")
	}

	// Proxy the stream to the peer
	// The debrid stream URL directly comes from the debrid service
	req, err := http.NewRequest(c.Request().Method, streamUrl, c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create request")
	}

	// Copy original request headers to the proxied request
	for key, values := range c.Request().Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := videoProxyClient.Do(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to proxy request")
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Response().Header().Add(key, value)
		}
	}

	// Set the status code
	c.Response().WriteHeader(resp.StatusCode)

	// Stream the response body
	_, err = io.Copy(c.Response().Writer, resp.Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to stream response body")
	}
	return nil
}

// route /api/v1/nakama/host/debridstream/url
// Returns the debrid stream URL for direct access by peers to avoid host bandwidth usage
func (h *Handler) HandleNakamaHostGetDebridstreamURL(c echo.Context) error {
	streamUrl, ok := h.App.DebridClientRepository.GetStreamURL()
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "no stream url")
	}

	return h.RespondWithData(c, map[string]string{
		"streamUrl": streamUrl,
	})
}

// route /api/v1/nakama/host/anime/library/stream?path={base64_encoded_path}
func (h *Handler) HandleNakamaHostAnimeLibraryServeStream(c echo.Context) error {
	filepath := c.QueryParam("path")
	decodedPath, err := base64.StdEncoding.DecodeString(filepath)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid path")
	}

	h.App.Logger.Info().Msgf("nakama: Serving anime library file: %s", string(decodedPath))

	// Make sure file is in library
	isInLibrary := false
	libraryPaths := h.App.Settings.GetLibrary().GetLibraryPaths()
	for _, libraryPath := range libraryPaths {
		if util.IsFileUnderDir(string(decodedPath), libraryPath) {
			isInLibrary = true
			break
		}
	}

	if !isInLibrary {
		return echo.NewHTTPError(http.StatusNotFound, "file not in library")
	}

	return c.File(string(decodedPath))
}

// route /api/v1/nakama/stream
// Proxies stream requests to the host. It inserts the Nakama password in the headers.
// It checks if the password is valid.
// For debrid streams, it redirects directly to the debrid service to avoid host bandwidth usage.
func (h *Handler) HandleNakamaProxyStream(c echo.Context) error {

	streamType := c.QueryParam("type") // "file", "torrent", "debrid"
	if streamType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "type is required")
	}

	hostServerUrl := h.App.Settings.GetNakama().RemoteServerURL
	hostServerUrl = strings.TrimSuffix(hostServerUrl, "/")

	if streamType == "debrid" {
		// Get the debrid stream URL from the host
		urlEndpoint := hostServerUrl + "/api/v1/nakama/host/debridstream/url"

		req, err := http.NewRequest(http.MethodGet, urlEndpoint, nil)
		if err != nil {
			h.App.Logger.Error().Err(err).Str("url", urlEndpoint).Msg("nakama: Failed to create debrid URL request")
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create request")
		}

		// Add Nakama password for authentication
		req.Header.Set("X-Seanime-Nakama-Token", h.App.Settings.GetNakama().RemoteServerPassword)

		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			h.App.Logger.Error().Err(err).Str("url", urlEndpoint).Msg("nakama: Failed to get debrid stream URL")
			return echo.NewHTTPError(http.StatusBadGateway, "failed to get stream URL")
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			h.App.Logger.Warn().Int("status", resp.StatusCode).Str("url", urlEndpoint).Msg("nakama: Failed to get debrid stream URL")
			return echo.NewHTTPError(resp.StatusCode, "failed to get stream URL")
		}

		// Parse the response to get the stream URL
		type urlResponse struct {
			Data struct {
				StreamUrl string `json:"streamUrl"`
			} `json:"data"`
		}

		var urlResp urlResponse
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			h.App.Logger.Error().Err(err).Msg("nakama: Failed to read debrid URL response")
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to read response")
		}

		if err := json.Unmarshal(body, &urlResp); err != nil {
			h.App.Logger.Error().Err(err).Msg("nakama: Failed to parse debrid URL response")
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to parse response")
		}

		if urlResp.Data.StreamUrl == "" {
			h.App.Logger.Error().Msg("nakama: Empty debrid stream URL")
			return echo.NewHTTPError(http.StatusNotFound, "no stream URL available")
		}

		req, err = http.NewRequest(c.Request().Method, urlResp.Data.StreamUrl, c.Request().Body)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create request")
		}

		// Copy original request headers to the proxied request
		for key, values := range c.Request().Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		resp, err = videoProxyClient.Do(req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to proxy request")
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Response().Header().Add(key, value)
			}
		}

		// Set the status code
		c.Response().WriteHeader(resp.StatusCode)

		// Stream the response body
		_, err = io.Copy(c.Response().Writer, resp.Body)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to stream response body")
		}
		return nil
	}

	requestUrl := ""
	switch streamType {
	case "file":
		// Path should be base64 encoded
		filepath := c.QueryParam("path")
		if filepath == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "path is required")
		}
		requestUrl = hostServerUrl + "/api/v1/nakama/host/anime/library/stream?path=" + filepath
	case "torrent":
		requestUrl = hostServerUrl + "/api/v1/nakama/host/torrentstream/stream"
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "invalid type")
	}

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 2,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   true, // Disable keep-alive to prevent connection reuse issues
			ForceAttemptHTTP2:   false,
		},
		Timeout: 120 * time.Second,
	}

	if c.Request().Method == http.MethodHead {
		req, err := http.NewRequest(http.MethodHead, requestUrl, nil)
		if err != nil {
			h.App.Logger.Error().Err(err).Str("url", requestUrl).Msg("nakama: Failed to create HEAD request")
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create request")
		}

		// Add Nakama password for authentication
		req.Header.Set("X-Seanime-Nakama-Token", h.App.Settings.GetNakama().RemoteServerPassword)

		// Add User-Agent from original request
		if userAgent := c.Request().Header.Get("User-Agent"); userAgent != "" {
			req.Header.Set("User-Agent", userAgent)
		}

		resp, err := client.Do(req)
		if err != nil {
			h.App.Logger.Error().Err(err).Str("url", requestUrl).Msg("nakama: Failed to proxy HEAD request")
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to proxy request")
		}
		defer resp.Body.Close()

		// Log authentication failures
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			h.App.Logger.Warn().Int("status", resp.StatusCode).Str("url", requestUrl).Msg("nakama: Authentication failed - check password configuration")
		}

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Response().Header().Add(key, value)
			}
		}

		return c.NoContent(resp.StatusCode)
	}

	// Create request with timeout context
	ctx := c.Request().Context()
	req, err := http.NewRequestWithContext(ctx, c.Request().Method, requestUrl, c.Request().Body)
	if err != nil {
		h.App.Logger.Error().Err(err).Str("url", requestUrl).Msg("nakama: Failed to create request")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create request")
	}

	// Copy request headers but skip problematic ones
	for key, values := range c.Request().Header {
		// Skip headers that should not be forwarded or might cause errors
		if key == "Host" || key == "Content-Length" || key == "Connection" ||
			key == "Transfer-Encoding" || key == "Accept-Encoding" ||
			key == "Upgrade" || key == "Proxy-Connection" ||
			strings.HasPrefix(key, "Sec-") { // Skip WebSocket and security headers
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	req.Header.Set("Accept", "*/*")
	// req.Header.Set("Accept-Encoding", "identity") // Disable compression to avoid issues

	// Add Nakama password for authentication
	req.Header.Set("X-Seanime-Nakama-Token", h.App.Settings.GetNakama().RemoteServerPassword)

	h.App.Logger.Debug().Str("url", requestUrl).Str("method", c.Request().Method).Msg("nakama: Proxying request")

	// Add retry mechanism for intermittent network issues
	var resp *http.Response
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = client.Do(req)
		if err == nil {
			break
		}

		if attempt < maxRetries-1 {
			h.App.Logger.Warn().Err(err).Int("attempt", attempt+1).Str("url", requestUrl).Msg("nakama: request failed, retrying")
			time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond) // Exponential backoff
			continue
		}

		h.App.Logger.Error().Err(err).Str("url", requestUrl).Msg("nakama: failed to proxy request after retries")
		return echo.NewHTTPError(http.StatusBadGateway, "failed to proxy request after retries")
	}
	defer resp.Body.Close()

	// Log authentication failures with more detail
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		h.App.Logger.Warn().Int("status", resp.StatusCode).Str("url", requestUrl).Msg("nakama: authentication failed - verify RemoteServerPassword matches host's HostPassword")
	}

	// Log and handle 406 Not Acceptable errors
	if resp.StatusCode == http.StatusNotAcceptable {
		h.App.Logger.Error().Int("status", resp.StatusCode).Str("url", requestUrl).Str("content-type", resp.Header.Get("Content-Type")).Msg("nakama: 406 Not Acceptable - content negotiation failed")
	}

	// Handle range request errors
	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		h.App.Logger.Warn().Int("status", resp.StatusCode).Str("url", requestUrl).Str("range", c.Request().Header.Get("Range")).Msg("nakama: range request not satisfiable")
	}

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Response().Header().Add(key, value)
		}
	}

	// Set the status code
	c.Response().WriteHeader(resp.StatusCode)

	// Stream the response body with better error handling
	bytesWritten, err := io.Copy(c.Response().Writer, resp.Body)
	if err != nil {
		// Check if it's a network-related error
		if strings.Contains(err.Error(), "connection") || strings.Contains(err.Error(), "broken pipe") ||
			strings.Contains(err.Error(), "wsasend") || strings.Contains(err.Error(), "reset by peer") {
			h.App.Logger.Warn().Err(err).Int64("bytes_written", bytesWritten).Str("url", requestUrl).Msg("nakama: network connection error during streaming")
		} else {
			h.App.Logger.Error().Err(err).Int64("bytes_written", bytesWritten).Str("url", requestUrl).Msg("nakama: error streaming response body")
		}
		// Don't return error here as response has already started
	} else {
		h.App.Logger.Debug().Int64("bytes_written", bytesWritten).Str("url", requestUrl).Msg("nakama: successfully streamed response")
	}
	return nil
}

// HandleNakamaReconnectToHost
//
//	@summary reconnects to the Nakama host.
//	@desc This attempts to reconnect to the configured Nakama host if the connection was lost.
//	@route /api/v1/nakama/reconnect [POST]
//	@returns nakama.MessageResponse
func (h *Handler) HandleNakamaReconnectToHost(c echo.Context) error {
	err := h.App.NakamaManager.ReconnectToHost()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	response := &nakama.MessageResponse{
		Success: true,
		Message: "Reconnection initiated",
	}

	return h.RespondWithData(c, response)
}

// HandleNakamaRemoveStaleConnections
//
//	@summary removes stale peer connections.
//	@desc This removes peer connections that haven't responded to ping messages for a while.
//	@route /api/v1/nakama/cleanup [POST]
//	@returns nakama.MessageResponse
func (h *Handler) HandleNakamaRemoveStaleConnections(c echo.Context) error {
	if !h.App.Settings.GetNakama().IsHost {
		return h.RespondWithError(c, errors.New("not acting as host"))
	}

	h.App.NakamaManager.RemoveStaleConnections()

	response := &nakama.MessageResponse{
		Success: true,
		Message: "Stale connections cleaned up",
	}

	return h.RespondWithData(c, response)
}

// HandleNakamaCreateWatchParty
//
//	@summary creates a new watch party session.
//	@desc This creates a new watch party that peers can join to watch content together in sync.
//	@route /api/v1/nakama/watch-party/create [POST]
//	@returns bool
func (h *Handler) HandleNakamaCreateWatchParty(c echo.Context) error {
	type body struct {
		Settings *nakama.WatchPartySessionSettings `json:"settings"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	if !h.App.Settings.GetNakama().IsHost {
		return h.RespondWithError(c, errors.New("only hosts can create watch parties"))
	}

	// Set default settings if not provided
	if b.Settings == nil {
		b.Settings = &nakama.WatchPartySessionSettings{
			SyncThreshold:     2.0,
			MaxBufferWaitTime: 10,
		}
	}

	_, err := h.App.NakamaManager.GetWatchPartyManager().CreateWatchParty(&nakama.CreateWatchOptions{
		Settings: b.Settings,
	})
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleNakamaJoinWatchParty
//
//	@summary joins an existing watch party.
//	@desc This allows a peer to join an active watch party session.
//	@route /api/v1/nakama/watch-party/join [POST]
//	@returns bool
func (h *Handler) HandleNakamaJoinWatchParty(c echo.Context) error {
	if h.App.Settings.GetNakama().IsHost {
		return h.RespondWithError(c, errors.New("hosts cannot join watch parties"))
	}

	if !h.App.NakamaManager.IsConnectedToHost() {
		return h.RespondWithError(c, errors.New("not connected to host"))
	}

	// Send join request to host
	err := h.App.NakamaManager.GetWatchPartyManager().JoinWatchParty()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleNakamaLeaveWatchParty
//
//	@summary leaves the current watch party.
//	@desc This removes the user from the active watch party session.
//	@route /api/v1/nakama/watch-party/leave [POST]
//	@returns bool
func (h *Handler) HandleNakamaLeaveWatchParty(c echo.Context) error {
	if h.App.Settings.GetNakama().IsHost {
		// Host stopping the watch party
		h.App.NakamaManager.GetWatchPartyManager().StopWatchParty()
	} else {
		// Peer leaving the watch party
		if !h.App.NakamaManager.IsConnectedToHost() {
			return h.RespondWithError(c, errors.New("not connected to host"))
		}

		err := h.App.NakamaManager.GetWatchPartyManager().LeaveWatchParty()
		if err != nil {
			return h.RespondWithError(c, err)
		}
	}

	return h.RespondWithData(c, true)
}
