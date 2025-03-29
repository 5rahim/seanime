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
		Episodes        []*anime.Episode `json:"episodes"`
		HasMappingError bool             `json:"hasMappingError"`
	}
)

// NewEpisodeCollection creates a new episode collection by leveraging anime.EntryDownloadInfo.
// It stores the EpisodeCollection in the repository instance for the lifetime of the repository.
//
// Note: This is also used by the Debrid streaming view.
func (r *Repository) NewEpisodeCollection(mId int) (ec *EpisodeCollection, err error) {

	// Get the media info, this is cached
	// Note: animeMetadata is always defined, even if it's not found on AniDB
	completeAnime, animeMetadata, err := r.getMediaInfo(mId)
	if err != nil {
		return nil, err
	}

	ec = &EpisodeCollection{
		HasMappingError: false,
		Episodes:        make([]*anime.Episode, 0),
	}

	// +---------------------+
	// |    Download Info    |
	// +---------------------+

	info, err := anime.NewEntryDownloadInfo(&anime.NewEntryDownloadInfoOptions{
		LocalFiles:       nil,
		AnimeMetadata:    animeMetadata,
		Progress:         lo.ToPtr(0), // Progress is 0 because we want the entire list
		Status:           lo.ToPtr(anilist.MediaListStatusCurrent),
		Media:            completeAnime.ToBaseAnime(),
		MetadataProvider: r.metadataProvider,
	})
	if err != nil {
		r.logger.Error().Err(err).Msg("torrentstream: could not get media entry info")
		return nil, err
	}

	// As of v2.8.0, this should never happen, getMediaInfo always returns an anime metadata struct, even if it's not found
	// causing NewEntryDownloadInfo to return a valid list of episodes to download
	if info == nil || info.EpisodesToDownload == nil {
		r.logger.Debug().Msg("torrentstream: no episodes found from AniDB, using AniList")
		baseAnime := completeAnime.ToBaseAnime()
		for epIdx := range baseAnime.GetCurrentEpisodeCount() {
			episodeNumber := epIdx + 1

			mediaWrapper := r.metadataProvider.GetAnimeMetadataWrapper(baseAnime, nil)
			episodeMetadata := mediaWrapper.GetEpisodeMetadata(episodeNumber)

			episode := &anime.Episode{
				Type:                  anime.LocalFileTypeMain,
				DisplayTitle:          fmt.Sprintf("Episode %d", episodeNumber),
				EpisodeTitle:          baseAnime.GetPreferredTitle(),
				EpisodeNumber:         episodeNumber,
				AniDBEpisode:          fmt.Sprintf("%d", episodeNumber),
				AbsoluteEpisodeNumber: episodeNumber,
				ProgressNumber:        episodeNumber,
				LocalFile:             nil,
				IsDownloaded:          false,
				EpisodeMetadata: &anime.EpisodeMetadata{
					AnidbId:  0,
					Image:    episodeMetadata.Image,
					AirDate:  "",
					Length:   0,
					Summary:  "",
					Overview: "",
					IsFiller: false,
				},
				FileMetadata:  nil,
				IsInvalid:     false,
				MetadataIssue: "",
				BaseAnime:     baseAnime,
			}
			ec.Episodes = append(ec.Episodes, episode)
		}
		ec.HasMappingError = true
		r.setEpisodeCollection(ec)
		return
	}

	if len(info.EpisodesToDownload) == 0 {
		r.logger.Error().Msg("torrentstream: no episodes found")
		return nil, fmt.Errorf("no episodes found")
	}

	ec.Episodes = lo.Map(info.EpisodesToDownload, func(episode *anime.EntryDownloadEpisode, i int) *anime.Episode {
		return episode.Episode
	})

	slices.SortStableFunc(ec.Episodes, func(i, j *anime.Episode) int {
		return cmp.Compare(i.EpisodeNumber, j.EpisodeNumber)
	})

	r.setEpisodeCollection(ec)

	return
}
