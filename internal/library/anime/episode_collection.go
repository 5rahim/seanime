package anime

type (
	// AnimeEntryEpisodeCollection represents a collection of episodes.
	AnimeEntryEpisodeCollection struct {
		Episodes []*AnimeEntryEpisode `json:"episodes"`
	}
)
