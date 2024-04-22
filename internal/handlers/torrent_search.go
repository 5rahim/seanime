package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/torrents/torrent"
)

// HandleSearchTorrent
//
//	@summary searches torrents and returns a list of torrents and their previews.
//	@desc This will search for torrents and return a list of torrents with previews.
//	@desc If smart search is enabled, it will filter the torrents based on search parameters.
//	@route /api/v1/torrent/search [POST]
//	@returns torrent.SearchData
func HandleSearchTorrent(c *RouteCtx) error {

	type body struct {
		SmartSearch    *bool              `json:"smartSearch"`
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

	if b.SmartSearch == nil ||
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
			SmartSearch:    b.SmartSearch,
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

// HandleSearchNsfwTorrent
//
//	@summary searches NSFW torrents and returns a list of torrents without previews.
//	@desc This will search for NSFW torrents and return a list of torrents without previews.
//	@route /api/v1/torrent/nsfw-search [POST]
//	@returns torrent.SearchData
func HandleSearchNsfwTorrent(c *RouteCtx) error {

	type body struct {
		Query string `json:"query"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	data, err := torrent.NewNsfwSearch(b.Query, c.App.NyaaSearchCache)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(data)

}
