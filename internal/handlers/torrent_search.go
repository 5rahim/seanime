package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/torrents/torrent"
)

// HandleTorrentSearch will search torrents.
// It will return a list of torrents and their previews (TorrentSearchData).
//
//	POST /v1/torrent/search
func HandleTorrentSearch(c *RouteCtx) error {

	type body struct {
		QuickSearch    *bool              `json:"quickSearch"`
		Query          *string            `json:"query"`
		EpisodeNumber  *int               `json:"episodeNumber"`
		Batch          *bool              `json:"batch"`
		Media          *anilist.BaseMedia `json:"media"`
		AbsoluteOffset *int               `json:"absoluteOffset"`
		Resolution     *string            `json:"resolution"`
		Best           *bool              `json:"best"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	if b.QuickSearch == nil ||
		b.Media == nil ||
		b.Batch == nil ||
		b.EpisodeNumber == nil ||
		b.AbsoluteOffset == nil ||
		b.Resolution == nil ||
		b.Query == nil {
		return c.RespondWithError(errors.New("missing arguments"))
	}

	data, err := torrent.NewSmartSearch(&torrent.SmartSearchOptions{
		SmartSearchQueryOptions: torrent.SmartSearchQueryOptions{
			QuickSearch:    b.QuickSearch,
			Query:          b.Query,
			EpisodeNumber:  b.EpisodeNumber,
			Batch:          b.Batch,
			Media:          b.Media,
			AbsoluteOffset: b.AbsoluteOffset,
			Resolution:     b.Resolution,
			Provider:       c.App.Settings.Library.TorrentProvider,
			Best:           b.Best,
		},
		NyaaSearchCache:       c.App.NyaaSearchCache,
		AnimeToshoSearchCache: c.App.AnimeToshoSearchCache,
		AnizipCache:           c.App.AnizipCache,
		Logger:                c.App.Logger,
		MetadataProvider:      c.App.MetadataProvider,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(data)

}
