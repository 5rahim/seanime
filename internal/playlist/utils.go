package playlist

import "seanime/internal/library/anime"

func isLocalFile(e *anime.PlaylistEpisode) bool {
	return e.Episode.LocalFile != nil && !e.IsNakama
}
