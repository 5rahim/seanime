package torrentstream

import (
	"cmp"
	"fmt"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/library/anime"
	"slices"
)

type (
	EpisodeCollection struct {
		Episodes []*anime.MediaEntryEpisode `json:"episodes"`
	}
)

// NewEpisodeCollection creates a new episode collection by leveraging anime.MediaEntryDownloadInfo.
// It stores the EpisodeCollection in the repository instance for the lifetime of the repository.
func (r *Repository) NewEpisodeCollection(mId int) (ec *EpisodeCollection, err error) {
	if err = r.FailIfNoSettings(); err != nil {
		return nil, err
	}

	// Get the media info, this is cached
	completeMedia, anizipMedia, err := r.getMediaInfo(mId)
	if err != nil {
		return nil, err
	}

	ec = &EpisodeCollection{
		Episodes: make([]*anime.MediaEntryEpisode, 0),
	}

	// +---------------------+
	// |    Download Info    |
	// +---------------------+

	info, err := anime.NewMediaEntryDownloadInfo(&anime.NewMediaEntryDownloadInfoOptions{
		LocalFiles:       nil,
		AnizipMedia:      anizipMedia,
		Progress:         lo.ToPtr(0), // Progress is 0 because we want the entire list
		Status:           lo.ToPtr(anilist.MediaListStatusCurrent),
		Media:            completeMedia.ToBaseMedia(),
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

	ec.Episodes = lo.Map(info.EpisodesToDownload, func(episode *anime.MediaEntryDownloadEpisode, i int) *anime.MediaEntryEpisode {
		return episode.Episode
	})

	slices.SortStableFunc(ec.Episodes, func(i, j *anime.MediaEntryEpisode) int {
		return cmp.Compare(i.EpisodeNumber, j.EpisodeNumber)
	})

	r.setEpisodeCollection(ec)

	return
}
