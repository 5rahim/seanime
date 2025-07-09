package handlers

import (
	"encoding/base64"
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
		return echo.NewHTTPError(http.StatusForbidden, "file not in library")
	}

	return c.File(string(decodedPath))
}

// route /api/v1/nakama/stream
// Proxies stream requests to the host. It inserts the Nakama password in the headers.
// It checks if the password is valid.
func (h *Handler) HandleNakamaProxyStream(c echo.Context) error {

	streamType := c.QueryParam("type") // "file", "torrent", "debrid"
	if streamType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "type is required")
	}

	hostServerUrl := h.App.Settings.GetNakama().RemoteServerURL
	if strings.HasSuffix(hostServerUrl, "/") {
		hostServerUrl = hostServerUrl[:len(hostServerUrl)-1]
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
	case "debrid":
		requestUrl = hostServerUrl + "/api/v1/nakama/host/debridstream/stream"
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "invalid type")
	}

	if c.Request().Method == http.MethodHead {
		req, err := http.NewRequest(http.MethodHead, requestUrl, nil)
		if err != nil {
			h.App.Logger.Error().Msgf("nakama: failed to create request: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create request")
		}

		// Add Nakama password for authentication
		req.Header.Set("X-Seanime-Nakama-Password", h.App.Settings.GetNakama().RemoteServerPassword)

		resp, err := videoProxyClient.Do(req)
		if err != nil {
			h.App.Logger.Error().Msgf("nakama: failed to proxy request: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to proxy request")
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Response().Header().Add(key, value)
			}
		}

		return c.NoContent(resp.StatusCode)
	}

	// Proxy the request
	req, err := http.NewRequest(c.Request().Method, requestUrl, c.Request().Body)
	if err != nil {
		h.App.Logger.Error().Msgf("nakama: failed to create request: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create request")
	}

	// Copy request headers but add authentication
	for key, values := range c.Request().Header {
		// Skip certain headers
		if key == "Host" || key == "Content-Length" {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Add Nakama password for authentication
	req.Header.Set("X-Seanime-Nakama-Password", h.App.Settings.GetNakama().RemoteServerPassword)

	resp, err := videoProxyClient.Do(req)
	if err != nil {
		h.App.Logger.Error().Msgf("nakama: failed to proxy request: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to proxy request")
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Response().Header().Add(key, value)
		}
	}

	// Stream the response body
	c.Response().WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Response().Writer, resp.Body)
	return err
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
			MaxBufferWaitTime: 5,
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
