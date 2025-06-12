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

// HandleGetNakamaStatus
//
//	@summary gets the current Nakama connection status.
//	@desc This returns the current status of Nakama connections including host mode and peer connections.
//	@route /api/v1/nakama/status [GET]
//	@returns nakama.NakamaStatus
func (h *Handler) HandleGetNakamaStatus(c echo.Context) error {
	status := &nakama.NakamaStatus{
		IsHost:               h.App.Settings.GetNakama().IsHost,
		ConnectedPeers:       h.App.NakamaManager.GetConnectedPeers(),
		IsConnectedToHost:    h.App.NakamaManager.IsConnectedToHost(),
		HostConnectionStatus: h.App.NakamaManager.GetHostConnectionStatus(),
	}

	return h.RespondWithData(c, status)
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
//	@desc This includes episodes and metadata (if any), AniList list data, download info...
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
	w := c.Response().Writer
	r := c.Request()

	if c.Request().Method == http.MethodHead {
		h.App.TorrentstreamRepository.HTTPStreamHandler().ServeHTTP(w, r)
		return nil
	}

	// Check password
	password := r.Header.Get("X-Seanime-Nakama-Password")
	if password != h.App.Settings.GetNakama().HostPassword {
		return h.RespondWithError(c, errors.New("invalid password"))
	}

	h.App.TorrentstreamRepository.HTTPStreamHandler().ServeHTTP(w, r)
	return nil
}

// route /api/v1/nakama/host/debridstream/stream
// Allows peers to stream the currently playing torrent.
func (h *Handler) HandleNakamaHostDebridstreamServeStream(c echo.Context) error {
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

	return c.File(string(decodedPath))
}

//-------------------------------------------
// Perr
//-------------------------------------------

var videoProxyClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		ForceAttemptHTTP2:   false, // Fixes issues on Linux
	},
	Timeout: 60 * time.Second,
}

// route /api/v1/nakama/stream
// Proxies stream requests to the host. It inserts the Nakama password in the headers.
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
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create request")
		}

		// Add Nakama password for authentication
		req.Header.Set("X-Seanime-Nakama-Password", h.App.Settings.GetNakama().RemoteServerPassword)

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

		return c.NoContent(resp.StatusCode)
	}

	// Proxy the request
	req, err := http.NewRequest(c.Request().Method, requestUrl, c.Request().Body)
	if err != nil {
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
