package handlers

// HandleGetOnlineStreamEpisodes returns the episodes.
// It returns the best available episodes from the online stream providers.
//
//	POST /v1/onlinestream/episodes
func HandleGetOnlineStreamEpisodes(c *RouteCtx) error {

	type body struct {
		MediaId  int    `json:"mediaId"`
		Dubbed   bool   `json:"dubbed"`
		Provider string `json:"provider"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	episodes, err := c.App.Onlinestream.GetMediaEpisodes(b.Provider, b.MediaId, b.Dubbed)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(episodes)
}

// HandleGetOnlineStreamEpisodeSources
//
//	POST /v1/onlinestream/episode-sources
func HandleGetOnlineStreamEpisodeSources(c *RouteCtx) error {

	type body struct {
		EpisodeNumber int    `json:"episodeNumber"`
		MediaId       int    `json:"mediaId"`
		Provider      string `json:"provider"`
		Dubbed        bool   `json:"dubbed"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	sources, err := c.App.Onlinestream.GetEpisodeSources(b.Provider, b.MediaId, b.EpisodeNumber, b.Dubbed)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(sources)
}
