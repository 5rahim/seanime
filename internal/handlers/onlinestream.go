package handlers

// HandleGetOnlineStreamEpisodes returns the episodes.
// It returns the best available episodes from the online stream providers.
//
//	POST /v1/onlinestream/episodes
func HandleGetOnlineStreamEpisodes(c *RouteCtx) error {

	type body struct {
		MediaId string `json:"mediaId"`
	}

	panic("not implemented")
}

// HandleGetOnlineStreamEpisode returns the online stream episode data.
//
//	POST /v1/onlinestream/episode
func HandleGetOnlineStreamEpisode(c *RouteCtx) error {

	type body struct {
		EpisodeNumber string `json:"episode"`
	}

	panic("not implemented")
}
