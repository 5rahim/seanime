package scanner

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/comparison"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/limiter"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/sourcegraph/conc/pool"
	"strconv"
	"time"
)

// FileHydrator hydrates the metadata of all (matched) LocalFiles.
// LocalFiles should already have their media ID hydrated.
type FileHydrator struct {
	LocalFiles         []*entities.LocalFile // Local files to hydrate
	AllMedia           []*NormalizedMedia    // All media used to hydrate local files
	BaseMediaCache     *anilist.BaseMediaCache
	AnizipCache        *anizip.Cache
	AnilistClient      *anilist.Client
	AnilistRateLimiter *limiter.Limiter
	Logger             *zerolog.Logger
}

// HydrateMetadata will hydrate the metadata of each LocalFile with the metadata of the matched anilist.BaseMedia.
// It will divide the LocalFiles into groups based on their media ID and process each group in parallel.
func (fh *FileHydrator) HydrateMetadata() {
	rateLimiter := limiter.NewLimiter(5*time.Second, 20)

	fh.Logger.Debug().Msg("hydrator: Starting metadata hydration process")

	// Group local files by media ID
	groups := lop.GroupBy(fh.LocalFiles, func(localFile *entities.LocalFile) int {
		return localFile.MediaId
	})

	// Remove the group with unmatched media
	delete(groups, 0)

	// Process each group in parallel
	p := pool.New()
	for mId, files := range groups {
		mId := mId
		files := files
		p.Go(func() {
			if len(files) > 0 {
				fh.hydrateGroupMetadata(mId, files, rateLimiter)
			}
		})
	}
	p.Wait()
}

func (fh *FileHydrator) hydrateGroupMetadata(
	mId int,
	lfs []*entities.LocalFile, // Grouped local files
	rateLimiter *limiter.Limiter,
) {

	// Get the media
	media, found := lo.Find(fh.AllMedia, func(media *NormalizedMedia) bool {
		return media.ID == mId
	})
	if !found {
		return
	}

	// Tree contains media relations
	tree := anilist.NewBaseMediaRelationTree()
	// Tree analysis used for episode normalization
	var mediaTreeAnalysis *MediaTreeAnalysis
	treeFetched := false

	// Process each local file in the group sequentially
	lo.ForEach(lfs, func(lf *entities.LocalFile, index int) {

		lf.Metadata.Type = entities.LocalFileTypeMain

		// Get episode number
		episode := -1
		if len(lf.ParsedData.Episode) > 0 {
			if ep, ok := util.StringToInt(lf.ParsedData.Episode); ok {
				episode = ep
			}
		}

		if comparison.ValueContainsNC(lf.Name) {
			lf.Metadata.Episode = 0
			lf.Metadata.AniDBEpisode = ""
			lf.Metadata.Type = entities.LocalFileTypeNC
			return
		}
		if comparison.ValueContainsSpecial(lf.Name) {
			lf.Metadata.Type = entities.LocalFileTypeSpecial
			if episode > -1 {
				// ep14 (13 original) -> ep1 s1
				if episode > media.GetCurrentEpisodeCount() {
					lf.Metadata.Episode = episode - media.GetCurrentEpisodeCount()
					lf.Metadata.AniDBEpisode = "S" + strconv.Itoa(episode-media.GetCurrentEpisodeCount())
				} else {
					lf.Metadata.Episode = episode
					lf.Metadata.AniDBEpisode = "S" + strconv.Itoa(episode)
				}
			} else {
				lf.Metadata.Episode = 1
				lf.Metadata.AniDBEpisode = "S1"
			}
			return
		}
		// Movie metadata
		if *media.Format == anilist.MediaFormatMovie {
			lf.Metadata.Episode = 1
			lf.Metadata.AniDBEpisode = "1"
			return
		}

		// No absolute episode count
		if episode <= media.GetCurrentEpisodeCount() {
			// Episode 0 - Might be a special
			// By default, we will assume that AniDB doesn't include Episode 0 as part of the main episodes (which is often the case)
			// If this proves to be wrong, media_entry.go will offset the AniDBEpisode by 1 and treat "S1" as "1" when it is a main episode
			if episode == 0 {
				// Leave episode number as 0, assuming that the client will handle tracking correctly
				lf.Metadata.Episode = 0
				lf.Metadata.AniDBEpisode = "S1"
				return
			}

			lf.Metadata.Episode = episode
			lf.Metadata.AniDBEpisode = strconv.Itoa(episode)
			return
		}

		// Episode number is higher but media only has 1 episode
		// - Might be a movie that was not correctly identified as such
		// - Or, the torrent files were divided into multiple episodes from a media that is listed as a movie on AniList
		if episode > media.GetCurrentEpisodeCount() && media.GetTotalEpisodeCount() == 1 {
			lf.Metadata.Episode = 1 // Coerce episode number to 1 because it is used for tracking
			lf.Metadata.AniDBEpisode = "1"
			return
		}

		// Absolute episode count
		if episode > media.GetCurrentEpisodeCount() {
			if !treeFetched {
				// Fetch media tree
				// The media tree will be used to normalize episode numbers
				if err := media.FetchMediaTree(anilist.FetchMediaTreeAll, fh.AnilistClient, fh.AnilistRateLimiter, tree, fh.BaseMediaCache); err == nil {
					// Create a new media tree analysis that will be used for episode normalization
					mta, _ := NewMediaTreeAnalysis(&MediaTreeAnalysisOptions{
						tree:        tree,
						anizipCache: fh.AnizipCache,
						rateLimiter: rateLimiter,
					})
					// Hoist the media tree analysis, so it will be used by other files
					mediaTreeAnalysis = mta
					treeFetched = true
				}
			}
			if err := fh.normalizeEpisodeNumberAndHydrate(mediaTreeAnalysis, lf, episode); err != nil {
				fh.Logger.Warn().Str("filename", lf.Name).Msg("hydrator: Could not normalize episode number")
			}
			return
		}

	})

}

// normalizeEpisodeNumberAndHydrate will normalize the episode number and hydrate the metadata of the LocalFile.
// If the MediaTreeAnalysis is nil, the episode number will not be normalized.
func (fh *FileHydrator) normalizeEpisodeNumberAndHydrate(
	mta *MediaTreeAnalysis,
	lf *entities.LocalFile,
	ep int,
) error {
	if mta == nil {
		lf.Metadata.Episode = ep
		lf.Metadata.AniDBEpisode = strconv.Itoa(ep)
		return errors.New("[hydrator] could not find media tree analysis")
	}

	relativeEp, mediaId, ok := mta.getRelativeEpisodeNumber(ep)
	if !ok {
		lf.Metadata.Episode = ep
		lf.Metadata.AniDBEpisode = strconv.Itoa(ep)
		return errors.New("[hydrator] could not normalize episode number")
	}

	lf.Metadata.Episode = relativeEp
	lf.Metadata.AniDBEpisode = strconv.Itoa(relativeEp)
	lf.MediaId = mediaId
	return nil
}
