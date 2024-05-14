package torrentstream

import (
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/torrents/torrent"
)

func (r *Repository) findBestTorrent(media *anilist.BaseMedia, anizipMedia *anizip.Media, anizipEpisode *anizip.Episode, episodeNumber int) (string, error) {
	searchBatch := false
	// Search batch if not a movie and finished
	if media.Format != nil && *media.Format != anilist.MediaFormatMovie && // Not a movie
		media.Status != nil && *media.Status == anilist.MediaStatusFinished { // Finished
		searchBatch = true
	}

	_, err := torrent.NewSmartSearch(&torrent.SmartSearchOptions{
		SmartSearchQueryOptions: torrent.SmartSearchQueryOptions{
			SmartSearch:    lo.ToPtr(true),
			Query:          lo.ToPtr(""),
			EpisodeNumber:  &episodeNumber,
			Batch:          &searchBatch,
			Media:          media,
			AbsoluteOffset: lo.ToPtr(anizipMedia.GetOffset()),
			Resolution:     lo.ToPtr(r.settings.MustGet().PreferredResolution),
			Provider:       "animetosho",
			Best:           lo.ToPtr(true),
		},
		NyaaSearchCache:       r.nyaaSearchCache,
		AnimeToshoSearchCache: r.animeToshoSearchCache,
		AnizipCache:           r.anizipCache,
		Logger:                r.logger,
		MetadataProvider:      r.metadataProvider,
	})
	if err != nil {
		return "", err
	}

	// Go through the top 5 torrents
	// - For each torrent, add it, get the files, and check if it has the episode
	// - If it does, return the magnet link

	return "", nil
}
