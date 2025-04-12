package handlers

import (
	"errors"
	"github.com/labstack/echo/v4"
	"seanime/internal/api/anilist"
	"seanime/internal/onlinestream"
)

// HandleGetOnlineStreamEpisodeList
//
//	@summary returns the episode list for the given media and provider.
//	@desc It returns the episode list for the given media and provider.
//	@desc The episodes are cached using a file cache.
//	@desc The episode list is just a list of episodes with no video sources, it's what the client uses to display the episodes and subsequently fetch the sources.
//	@desc The episode list might be nil or empty if nothing could be found, but the media will always be returned.
//	@route /api/v1/onlinestream/episode-list [POST]
//	@returns onlinestream.EpisodeListResponse
func (h *Handler) HandleGetOnlineStreamEpisodeList(c echo.Context) error {

	type body struct {
		MediaId  int    `json:"mediaId"`
		Dubbed   bool   `json:"dubbed"`
		Provider string `json:"provider,omitempty"` // Can be empty since we still have the media id
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	if h.App.Settings == nil || !h.App.Settings.Library.EnableOnlinestream {
		return h.RespondWithError(c, errors.New("enable online streaming in the settings"))
	}

	// Get media
	// This is cached
	media, err := h.App.OnlinestreamRepository.GetMedia(b.MediaId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	if media.Status == nil || *media.Status == anilist.MediaStatusNotYetReleased {
		return h.RespondWithError(c, errors.New("unavailable"))
	}

	// Get episode list
	// This is cached using file cache
	episodes, err := h.App.OnlinestreamRepository.GetMediaEpisodes(b.Provider, media, b.Dubbed)
	//if err != nil {
	//	return h.RespondWithError(c, err)
	//}

	ret := onlinestream.EpisodeListResponse{
		Episodes: episodes,
		Media:    media,
	}

	h.App.FillerManager.HydrateOnlinestreamFillerData(b.MediaId, ret.Episodes)

	return h.RespondWithData(c, ret)
}

// HandleGetOnlineStreamEpisodeSource
//
//	@summary returns the video sources for the given media, episode number and provider.
//	@route /api/v1/onlinestream/episode-source [POST]
//	@returns onlinestream.EpisodeSource
func (h *Handler) HandleGetOnlineStreamEpisodeSource(c echo.Context) error {

	type body struct {
		EpisodeNumber int    `json:"episodeNumber"`
		MediaId       int    `json:"mediaId"`
		Provider      string `json:"provider"`
		Dubbed        bool   `json:"dubbed"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	// Get media
	// This is cached
	media, err := h.App.OnlinestreamRepository.GetMedia(b.MediaId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	sources, err := h.App.OnlinestreamRepository.GetEpisodeSources(b.Provider, b.MediaId, b.EpisodeNumber, b.Dubbed, media.GetStartYearSafe())
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, sources)
}

// HandleOnlineStreamEmptyCache
//
//	@summary empties the cache for the given media.
//	@route /api/v1/onlinestream/cache [DELETE]
//	@returns bool
func (h *Handler) HandleOnlineStreamEmptyCache(c echo.Context) error {

	type body struct {
		MediaId int `json:"mediaId"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	err := h.App.OnlinestreamRepository.EmptyCache(b.MediaId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandleOnlinestreamManualSearch
//
//	@summary returns search results for a manual search.
//	@desc Returns search results for a manual search.
//	@route /api/v1/onlinestream/search [POST]
//	@returns []hibikeonlinestream.SearchResult
func (h *Handler) HandleOnlinestreamManualSearch(c echo.Context) error {

	type body struct {
		Provider string `json:"provider"`
		Query    string `json:"query"`
		Dubbed   bool   `json:"dubbed"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	ret, err := h.App.OnlinestreamRepository.ManualSearch(b.Provider, b.Query, b.Dubbed)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, ret)
}

// HandleOnlinestreamManualMapping
//
//	@summary manually maps an anime entry to an anime ID from the provider.
//	@desc This is used to manually map an anime entry to an anime ID from the provider.
//	@desc The client should re-fetch the chapter container after this.
//	@route /api/v1/onlinestream/manual-mapping [POST]
//	@returns bool
func (h *Handler) HandleOnlinestreamManualMapping(c echo.Context) error {

	type body struct {
		Provider string `json:"provider"`
		MediaId  int    `json:"mediaId"`
		AnimeId  string `json:"animeId"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	err := h.App.OnlinestreamRepository.ManualMapping(b.Provider, b.MediaId, b.AnimeId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleGetOnlinestreamMapping
//
//	@summary returns the mapping for an anime entry.
//	@desc This is used to get the mapping for an anime entry.
//	@desc An empty string is returned if there's no manual mapping. If there is, the anime ID will be returned.
//	@route /api/v1/onlinestream/get-mapping [POST]
//	@returns onlinestream.MappingResponse
func (h *Handler) HandleGetOnlinestreamMapping(c echo.Context) error {

	type body struct {
		Provider string `json:"provider"`
		MediaId  int    `json:"mediaId"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	mapping := h.App.OnlinestreamRepository.GetMapping(b.Provider, b.MediaId)
	return h.RespondWithData(c, mapping)
}

// HandleRemoveOnlinestreamMapping
//
//	@summary removes the mapping for an anime entry.
//	@desc This is used to remove the mapping for an anime entry.
//	@desc The client should re-fetch the chapter container after this.
//	@route /api/v1/onlinestream/remove-mapping [POST]
//	@returns bool
func (h *Handler) HandleRemoveOnlinestreamMapping(c echo.Context) error {

	type body struct {
		Provider string `json:"provider"`
		MediaId  int    `json:"mediaId"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	err := h.App.OnlinestreamRepository.RemoveMapping(b.Provider, b.MediaId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}
