package handlers

import (
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/downloader"
	"github.com/seanime-app/seanime-server/internal/nyaa"
	"github.com/sourcegraph/conc/pool"
)

func HandleDownloadNyaaTorrents(c *RouteCtx) error {

	type body struct {
		Urls        []string `json:"urls"`
		Destination string   `json:"destination"`
		SmartSelect struct {
			Enabled               bool  `json:"enabled"`
			MissingEpisodeNumbers []int `json:"missingEpisodeNumbers"`
			AbsoluteOffset        int   `json:"absoluteOffset"`
		} `json:"smartSelect"`
		Media *anilist.BaseMedia `json:"media"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	// try to start qbittorrent if it's not running
	err := c.App.QBittorrent.Start()
	if err != nil {
		return c.RespondWithError(err)
	}

	// get magnets
	p := pool.NewWithResults[string]().WithErrors()
	for _, url := range b.Urls {
		url := url
		p.Go(func() (string, error) {
			return nyaa.TorrentMagnet(url)
		})
	}
	// if we couldn't get a magnet, return error
	magnets, err := p.Wait()
	if err != nil {
		return c.RespondWithError(err)
	}

	// create repository
	repo := &downloader.QbittorrentRepository{
		Logger:         c.App.Logger,
		Client:         c.App.QBittorrent,
		WSEventManager: c.App.WSEventManager,
		Destination:    b.Destination,
	}

	// try to add torrents to qbittorrent, on error return error
	err = repo.AddMagnets(magnets)
	if err != nil {
		return c.RespondWithError(err)
	}

	err = repo.SmartSelect(&downloader.SmartSelect{
		Magnets:               magnets,
		Enabled:               b.SmartSelect.Enabled,
		MissingEpisodeNumbers: b.SmartSelect.MissingEpisodeNumbers,
		AbsoluteOffset:        b.SmartSelect.AbsoluteOffset,
		Media:                 b.Media,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)

}
