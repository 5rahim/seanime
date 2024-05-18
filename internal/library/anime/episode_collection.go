package anime

type (
	// MediaEntryEpisodeCollection represents a collection of episodes.
	MediaEntryEpisodeCollection struct {
		Episodes []*MediaEntryEpisode `json:"episodes"`
	}
)
