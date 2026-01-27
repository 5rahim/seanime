package handlers

import (
	"seanime/internal/api/anilist"
	"seanime/internal/database/db_bridge"
	"seanime/internal/debrid/debrid"
	"seanime/internal/library/anime"
	"seanime/internal/torrents/torrent"
	"seanime/internal/util/result"
	"strings"

	"github.com/labstack/echo/v4"
)

var debridInstantAvailabilityCache = result.NewCache[string, map[string]debrid.TorrentItemInstantAvailability]()

// HandleSearchTorrent
//
//	@summary searches torrents and returns a list of torrents and their previews.
//	@desc This will search for torrents and return a list of torrents with previews.
//	@desc If smart search is enabled, it will filter the torrents based on search parameters.
//	@route /api/v1/torrent/search [POST]
//	@returns torrent.SearchData
func (h *Handler) HandleSearchTorrent(c echo.Context) error {

	type body struct {
		// "smart" or "simple"
		Type                    string            `json:"type,omitempty"`
		Provider                string            `json:"provider,omitempty"`
		Query                   string            `json:"query,omitempty"`
		EpisodeNumber           int               `json:"episodeNumber,omitempty"`
		Batch                   bool              `json:"batch,omitempty"`
		Media                   anilist.BaseAnime `json:"media,omitempty"`
		AbsoluteOffset          int               `json:"absoluteOffset,omitempty"`
		Resolution              string            `json:"resolution,omitempty"`
		BestRelease             bool              `json:"bestRelease,omitempty"`
		IncludeSpecialProviders bool              `json:"includeSpecialProviders,omitempty"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	data, err := h.App.TorrentRepository.SearchAnime(c.Request().Context(), torrent.AnimeSearchOptions{
		Provider:                b.Provider,
		Type:                    torrent.AnimeSearchType(b.Type),
		Media:                   &b.Media,
		Query:                   b.Query,
		Batch:                   b.Batch,
		EpisodeNumber:           b.EpisodeNumber,
		BestReleases:            b.BestRelease,
		Resolution:              b.Resolution,
		IncludeSpecialProviders: false,
	})
	if err != nil {
		return h.RespondWithError(c, err)
	}

	//
	// Debrid torrent instant availability
	//
	if h.App.SecondarySettings.Debrid.Enabled {
		hashes := make([]string, 0)
		for _, t := range data.Torrents {
			if t.InfoHash == "" {
				continue
			}
			hashes = append(hashes, t.InfoHash)
		}
		hashesKey := strings.Join(hashes, ",")
		var found bool
		data.DebridInstantAvailability, found = debridInstantAvailabilityCache.Get(hashesKey)
		if !found {
			provider, err := h.App.DebridClientRepository.GetProvider()
			if err == nil {
				instantAvail := provider.GetInstantAvailability(hashes)
				data.DebridInstantAvailability = instantAvail
				debridInstantAvailabilityCache.Set(hashesKey, instantAvail)
			}
		}
	}

	return h.RespondWithData(c, data)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandleGetAutoSelectProfile
//
//	@summary returns the autoselect profile.
//	@desc This returns the single autoselect profile if it exists.
//	@route /api/v1/auto-select/profile [GET]
//	@returns anime.AutoSelectProfile
func (h *Handler) HandleGetAutoSelectProfile(c echo.Context) error {
	profile, err := db_bridge.GetAutoSelectProfile(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, profile)
}

// HandleSaveAutoSelectProfile
//
//	@summary creates or updates the autoselect profile.
//	@desc Since there's only one profile at all time, this will create or update it.
//	@route /api/v1/auto-select/profile [POST]
//	@returns anime.AutoSelectProfile
func (h *Handler) HandleSaveAutoSelectProfile(c echo.Context) error {
	type body struct {
		Profile *anime.AutoSelectProfile `json:"profile"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	if err := db_bridge.SaveAutoSelectProfile(h.App.Database, b.Profile); err != nil {
		return h.RespondWithError(c, err)
	}

	// Get the saved profile to return it with the DB ID
	profile, err := db_bridge.GetAutoSelectProfile(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, profile)
}

// HandleDeleteAutoSelectProfile
//
//	@summary deletes the autoselect profile.
//	@route /api/v1/auto-select/profile [DELETE]
//	@returns bool
func (h *Handler) HandleDeleteAutoSelectProfile(c echo.Context) error {
	if err := db_bridge.DeleteAutoSelectProfile(h.App.Database); err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}
