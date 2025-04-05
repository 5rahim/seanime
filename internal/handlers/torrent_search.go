package handlers

import (
	"seanime/internal/api/anilist"
	"seanime/internal/debrid/debrid"
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
		Type           string            `json:"type,omitempty"`
		Provider       string            `json:"provider,omitempty"`
		Query          string            `json:"query,omitempty"`
		EpisodeNumber  int               `json:"episodeNumber,omitempty"`
		Batch          bool              `json:"batch,omitempty"`
		Media          anilist.BaseAnime `json:"media,omitempty"`
		AbsoluteOffset int               `json:"absoluteOffset,omitempty"`
		Resolution     string            `json:"resolution,omitempty"`
		BestRelease    bool              `json:"bestRelease,omitempty"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	data, err := h.App.TorrentRepository.SearchAnime(c.Request().Context(), torrent.AnimeSearchOptions{
		Provider:      b.Provider,
		Type:          torrent.AnimeSearchType(b.Type),
		Media:         &b.Media,
		Query:         b.Query,
		Batch:         b.Batch,
		EpisodeNumber: b.EpisodeNumber,
		BestReleases:  b.BestRelease,
		Resolution:    b.Resolution,
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
