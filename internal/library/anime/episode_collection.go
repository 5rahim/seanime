package anime

type (
	// EpisodeCollection represents a collection of episodes.
	EpisodeCollection struct {
		Episodes []*Episode `json:"episodes"`
	}
)
