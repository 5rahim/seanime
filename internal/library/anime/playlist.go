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
	WatchTypeOnline    WatchType = "online"
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

func (pd *Playlist) FindEpisode(mId int, episodeNumber int) (*PlaylistEpisode, bool) {
	for _, e := range pd.Episodes {
		if e.Episode.BaseAnime.ID == mId && e.Episode.EpisodeNumber == episodeNumber {
			return e, true
		}
	}
	return nil, false
}

func (pd *Playlist) NextEpisode(episode *PlaylistEpisode) (*PlaylistEpisode, bool) {
	for i, e := range pd.Episodes {
		if isSameEpisode(e, episode) {
			if i+1 < len(pd.Episodes) {
				return pd.Episodes[i+1], true
			}
			return nil, false
		}
	}
	return nil, false
}

func (pd *Playlist) PreviousEpisode(episode *PlaylistEpisode) (*PlaylistEpisode, bool) {
	for i, e := range pd.Episodes {
		if isSameEpisode(e, episode) {
			if i-1 >= 0 {
				return pd.Episodes[i-1], true
			}
			return nil, false
		}
	}
	return nil, false
}

func isSameEpisode(a *PlaylistEpisode, b *PlaylistEpisode) bool {
	if a == nil || b == nil {
		return false
	}

	// If one file is a local file, use progress number for comparison
	if a.Episode.LocalFile != nil || b.Episode.LocalFile != nil {
		return a.Episode.BaseAnime.ID == b.Episode.BaseAnime.ID && a.Episode.ProgressNumber == b.Episode.ProgressNumber
	}

	// Otherwise, use AniDB episode number for comparison
	return a.Episode.BaseAnime.ID == b.Episode.BaseAnime.ID && a.Episode.AniDBEpisode == b.Episode.AniDBEpisode
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
