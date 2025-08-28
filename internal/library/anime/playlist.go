package anime

type (
	// Playlist holds the data from models.Playlist
	Playlist struct {
		DbId     uint               `json:"dbId"` // DbId is the database ID of the models.Playlist
		Name     string             `json:"name"` // Name is the name of the playlist
		Episodes []*PlaylistEpisode `json:"episodes"`
	}

	WatchType string

	PlaylistEpisode struct {
		Episode     *Episode  `json:"episode"`
		IsCompleted bool      `json:"isCompleted"`
		WatchType   WatchType `json:"watchType"`
		IsNakama    bool      `json:"isNakama"`
	}
)

const (
	WatchTypeLocalFile WatchType = "localfile"
	WatchTypeDebrid    WatchType = "debrid"
	WatchTypeTorrent   WatchType = "torrent"
	WatchTypeNakama    WatchType = "nakama"
)

// NewPlaylist creates a new Playlist instance
func NewPlaylist(name string) *Playlist {
	return &Playlist{
		Name:     name,
		Episodes: make([]*PlaylistEpisode, 0),
	}
}

func (pd *Playlist) SetEpisodes(episodes []*PlaylistEpisode) {
	pd.Episodes = episodes
}

func (pd *Playlist) GetEpisodesToWatch() []*PlaylistEpisode {
	episodes := make([]*PlaylistEpisode, 0)
	for _, episode := range pd.Episodes {
		if !episode.IsCompleted {
			episodes = append(episodes, episode)
		}
	}
	return episodes
}

func (pd *Playlist) SetEpisodeCompleted(mId int, anidbEpisode string) {
	for _, e := range pd.Episodes {
		if e.Episode.BaseAnime.ID == mId && e.Episode.AniDBEpisode == anidbEpisode {
			e.IsCompleted = true
			return
		}
	}
}
