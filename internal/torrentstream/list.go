package torrentstream

import (
	"cmp"
	"fmt"
	"github.com/samber/lo"
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
	"slices"
)

type (
	EpisodeCollection struct {
		Episodes []*anime.AnimeEntryEpisode `json:"episodes"`
	}
)

// NewEpisodeCollection creates a new episode collection by leveraging anime.AnimeEntryDownloadInfo.
// It stores the EpisodeCollection in the repository instance for the lifetime of the repository.
func (r *Repository) NewEpisodeCollection(mId int) (ec *EpisodeCollection, err error) {
	if err = r.FailIfNoSettings(); err != nil {
		return nil, err
	}

	// Get the media info, this is cached
	completeAnime, anizipMedia, err := r.getMediaInfo(mId)
	if err != nil {
		return nil, err
	}

	ec = &EpisodeCollection{
		Episodes: make([]*anime.AnimeEntryEpisode, 0),
	}

	// +---------------------+
	// |    Download Info    |
	// +---------------------+

	info, err := anime.NewAnimeEntryDownloadInfo(&anime.NewAnimeEntryDownloadInfoOptions{
		LocalFiles:       nil,
		AnizipMedia:      anizipMedia,
		Progress:         lo.ToPtr(0), // Progress is 0 because we want the entire list
		Status:           lo.ToPtr(anilist.MediaListStatusCurrent),
		Media:            completeAnime.ToBaseAnime(),
		MetadataProvider: r.metadataProvider,
	})
	if err != nil {
		r.logger.Error().Err(err).Msg("torrentstream: could not get media entry info")
		return nil, err
	}

	if info == nil || info.EpisodesToDownload == nil {
		r.logger.Error().Msg("torrentstream: could not get media entry info, episodes to download is nil")
		return nil, fmt.Errorf("could not get media entry info")
	}

	if len(info.EpisodesToDownload) == 0 {
		r.logger.Error().Msg("torrentstream: no episodes found")
		return nil, fmt.Errorf("no episodes found")
	}

	ec.Episodes = lo.Map(info.EpisodesToDownload, func(episode *anime.AnimeEntryDownloadEpisode, i int) *anime.AnimeEntryEpisode {
		return episode.Episode
	})

	slices.SortStableFunc(ec.Episodes, func(i, j *anime.AnimeEntryEpisode) int {
		return cmp.Compare(i.EpisodeNumber, j.EpisodeNumber)
	})

	r.setEpisodeCollection(ec)

	return
}
