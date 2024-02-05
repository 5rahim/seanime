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
	"github.com/seanime-app/seanime/internal/summary"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/sourcegraph/conc/pool"
	"strconv"
	"time"
)

// FileHydrator hydrates the metadata of all (matched) LocalFiles.
// LocalFiles should already have their media ID hydrated.
type FileHydrator struct {
	LocalFiles           []*entities.LocalFile       // Local files to hydrate
	AllMedia             []*entities.NormalizedMedia // All media used to hydrate local files
	BaseMediaCache       *anilist.BaseMediaCache
	AnizipCache          *anizip.Cache
	AnilistClientWrapper *anilist.ClientWrapper
	AnilistRateLimiter   *limiter.Limiter
	Logger               *zerolog.Logger
	ScanLogger           *ScanLogger
	ScanSummaryLogger    *summary.ScanSummaryLogger // optional
}

// HydrateMetadata will hydrate the metadata of each LocalFile with the metadata of the matched anilist.BaseMedia.
// It will divide the LocalFiles into groups based on their media ID and process each group in parallel.
func (fh *FileHydrator) HydrateMetadata() {
	start := time.Now()
	rateLimiter := limiter.NewLimiter(5*time.Second, 20)

	fh.Logger.Debug().Msg("hydrator: Starting metadata hydration")

	// Group local files by media ID
	groups := lop.GroupBy(fh.LocalFiles, func(localFile *entities.LocalFile) int {
		return localFile.MediaId
	})

	// Remove the group with unmatched media
	delete(groups, 0)

	fh.ScanLogger.LogFileHydrator(zerolog.InfoLevel).
		Int("entryCount", len(groups)).
		Msg("Starting metadata hydration process")

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

	fh.ScanLogger.LogFileHydrator(zerolog.InfoLevel).
		Any("ms", time.Since(start).Milliseconds()).
		Msg("Finished metadata hydration")
}

func (fh *FileHydrator) hydrateGroupMetadata(
	mId int,
	lfs []*entities.LocalFile, // Grouped local files
	rateLimiter *limiter.Limiter,
) {

	// Get the media
	media, found := lo.Find(fh.AllMedia, func(media *entities.NormalizedMedia) bool {
		return media.ID == mId
	})
	if !found {
		fh.ScanLogger.LogFileHydrator(zerolog.ErrorLevel).
			Int("mediaId", mId).
			Msg("Could not find media in FileHydrator options")
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

		// NC metadata
		if comparison.ValueContainsNC(lf.Name) {
			lf.Metadata.Episode = 0
			lf.Metadata.AniDBEpisode = ""
			lf.Metadata.Type = entities.LocalFileTypeNC

			/*Log */
			fh.logFileHydration(zerolog.DebugLevel, lf, mId, episode).
				Msg("File has been marked as NC")
			fh.ScanSummaryLogger.LogMetadataNC(lf)
			return
		}

		// Special metadata
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

			/*Log */
			fh.logFileHydration(zerolog.DebugLevel, lf, mId, episode).
				Msg("File has been marked as special")
			fh.ScanSummaryLogger.LogMetadataSpecial(lf, lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
			return
		}
		// Movie metadata
		if *media.Format == anilist.MediaFormatMovie {
			lf.Metadata.Episode = 1
			lf.Metadata.AniDBEpisode = "1"

			/*Log */
			fh.logFileHydration(zerolog.DebugLevel, lf, mId, episode).
				Msg("File has been marked as main")
			fh.ScanSummaryLogger.LogMetadataMain(lf, lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
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

			/*Log */
			fh.logFileHydration(zerolog.DebugLevel, lf, mId, episode).
				Msg("File has been marked as main")
			fh.ScanSummaryLogger.LogMetadataMain(lf, lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
			return
		}

		// Episode number is higher but media only has 1 episode
		// - Might be a movie that was not correctly identified as such
		// - Or, the torrent files were divided into multiple episodes from a media that is listed as a movie on AniList
		if episode > media.GetCurrentEpisodeCount() && media.GetTotalEpisodeCount() == 1 {
			lf.Metadata.Episode = 1 // Coerce episode number to 1 because it is used for tracking
			lf.Metadata.AniDBEpisode = "1"

			/*Log */
			fh.logFileHydration(zerolog.WarnLevel, lf, mId, episode).
				Str("warning", "File's episode number is higher than the media's episode count, but the media only has 1 episode").
				Msg("File has been marked as main")
			fh.ScanSummaryLogger.LogMetadataMain(lf, lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
			return
		}

		// Absolute episode count
		if episode > media.GetCurrentEpisodeCount() {
			if !treeFetched {

				mediaTreeFetchStart := time.Now()
				// Fetch media tree
				// The media tree will be used to normalize episode numbers
				if err := media.FetchMediaTree(anilist.FetchMediaTreeAll, fh.AnilistClientWrapper, fh.AnilistRateLimiter, tree, fh.BaseMediaCache); err == nil {
					// Create a new media tree analysis that will be used for episode normalization
					mta, _ := NewMediaTreeAnalysis(&MediaTreeAnalysisOptions{
						tree:        tree,
						anizipCache: fh.AnizipCache,
						rateLimiter: rateLimiter,
					})
					// Hoist the media tree analysis, so it will be used by other files
					mediaTreeAnalysis = mta
					treeFetched = true

					/*Log */
					fh.ScanLogger.LogFileHydrator(zerolog.DebugLevel).
						Int("mediaId", mId).
						Any("ms", time.Since(mediaTreeFetchStart).Milliseconds()).
						Int("requests", len(mediaTreeAnalysis.branches)).
						Any("branches", mediaTreeAnalysis.printBranches()).
						Msg("Media tree fetched")
					fh.ScanSummaryLogger.LogMetadataMediaTreeFetched(lf, time.Since(mediaTreeFetchStart).Milliseconds(), len(mediaTreeAnalysis.branches))
				} else {
					fh.ScanLogger.LogFileHydrator(zerolog.ErrorLevel).
						Int("mediaId", mId).
						Str("error", err.Error()).
						Any("ms", time.Since(mediaTreeFetchStart).Milliseconds()).
						Msg("Could not fetch media tree")
					fh.ScanSummaryLogger.LogMetadataMediaTreeFetchFailed(lf, err, time.Since(mediaTreeFetchStart).Milliseconds())
				}
			}

			// Normalize episode number
			if err := fh.normalizeEpisodeNumberAndHydrate(mediaTreeAnalysis, lf, episode); err != nil {

				/*Log */
				fh.logFileHydration(zerolog.WarnLevel, lf, mId, episode).
					Dict("mediaTreeAnalysis", zerolog.Dict().
						Bool("normalized", false).
						Str("error", err.Error()).
						Str("reason", "Episode normalization failed"),
					).
					Msg("File has been marked as main")
				fh.ScanSummaryLogger.LogMetadataEpisodeNormalizationFailed(lf, err, lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
			} else {
				/*Log */
				fh.logFileHydration(zerolog.DebugLevel, lf, mId, episode).
					Dict("mediaTreeAnalysis", zerolog.Dict().
						Bool("normalized", true).
						Bool("hasNewMediaId", lf.MediaId != mId).
						Int("newMediaId", lf.MediaId),
					).
					Msg("File has been marked as main")
				fh.ScanSummaryLogger.LogMetadataEpisodeNormalized(lf, mId, episode, lf.Metadata.Episode, lf.MediaId, lf.Metadata.AniDBEpisode)
			}
			return
		}

	})

}

func (fh *FileHydrator) logFileHydration(level zerolog.Level, lf *entities.LocalFile, mId int, episode int) *zerolog.Event {
	return fh.ScanLogger.LogFileHydrator(level).
		Str("filename", lf.Name).
		Int("mediaId", mId).
		Dict("vars", zerolog.Dict().
			Str("parsedEpisode", lf.ParsedData.Episode).
			Int("episode", episode),
		).
		Dict("metadata", zerolog.Dict().
			Int("episode", lf.Metadata.Episode).
			Str("aniDBEpisode", lf.Metadata.AniDBEpisode))
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
		return errors.New("[hydrator] could not find relative episode number from branches")
	}

	lf.Metadata.Episode = relativeEp
	lf.Metadata.AniDBEpisode = strconv.Itoa(relativeEp)
	lf.MediaId = mediaId
	return nil
}
