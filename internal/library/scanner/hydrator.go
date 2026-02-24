package scanner

import (
	"errors"
	"regexp"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/hook"
	"seanime/internal/library/anime"
	"seanime/internal/library/summary"
	"seanime/internal/platforms/platform"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"seanime/internal/util/limiter"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
)

// FileHydrator hydrates the metadata of all (matched) LocalFiles.
// LocalFiles should already have their media ID hydrated.
type FileHydrator struct {
	LocalFiles []*anime.LocalFile       // Local files to hydrate
	AllMedia   []*anime.NormalizedMedia // All media used to hydrate local files
	// Used by media tree analysis
	CompleteAnimeCache  *anilist.CompleteAnimeCache
	PlatformRef         *util.Ref[platform.Platform]
	MetadataProviderRef *util.Ref[metadata_provider.Provider]
	AnilistRateLimiter  *limiter.Limiter
	Logger              *zerolog.Logger
	ScanLogger          *ScanLogger                // optional
	ScanSummaryLogger   *summary.ScanSummaryLogger // optional
	ForceMediaId        int                        // optional - force all local files to have this media ID
	Config              *Config
	hydrationRules      map[string]*compiledHydrationRule
}

type compiledHydrationRule struct {
	regex        *regexp.Regexp
	rule         *HydrationRule
	fileRulesRgx map[string]*compiledHydrationFileRule
}

type compiledHydrationFileRule struct {
	regex *regexp.Regexp
	rule  *HydrationFileRule
}

// HydrateMetadata will hydrate the metadata of each LocalFile with the metadata of the matched anilist.BaseAnime.
// It will divide the LocalFiles into groups based on their media ID and process each group in parallel.
func (fh *FileHydrator) HydrateMetadata() {
	start := time.Now()
	rateLimiter := limiter.NewLimiter(5*time.Second, 20)

	fh.Logger.Debug().Msg("hydrator: Starting metadata hydration")

	fh.precompileRules()

	// Invoke ScanHydrationStarted hook
	event := &ScanHydrationStartedEvent{
		LocalFiles: fh.LocalFiles,
		AllMedia:   fh.AllMedia,
	}
	_ = hook.GlobalHookManager.OnScanHydrationStarted().Trigger(event)
	fh.LocalFiles = event.LocalFiles
	fh.AllMedia = event.AllMedia

	// Default prevented, do not hydrate the metadata
	if event.DefaultPrevented {
		return
	}

	// Group local files by media ID
	groups := lop.GroupBy(fh.LocalFiles, func(localFile *anime.LocalFile) int {
		return localFile.MediaId
	})

	// Remove the group with unmatched media
	delete(groups, 0)

	if fh.ScanLogger != nil {
		fh.ScanLogger.LogFileHydrator(zerolog.InfoLevel).
			Int("entryCount", len(groups)).
			Msg("Starting metadata hydration process")
	}

	const maxConcurrentGroups = 100
	sem := make(chan struct{}, maxConcurrentGroups)
	var wg sync.WaitGroup
	for mId, files := range groups {
		if len(files) == 0 {
			continue
		}
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore
		go func(mId int, files []*anime.LocalFile) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore
			fh.hydrateGroupMetadata(mId, files, rateLimiter)
		}(mId, files)
	}
	wg.Wait()

	if fh.ScanLogger != nil {
		fh.ScanLogger.LogFileHydrator(zerolog.InfoLevel).
			Int64("ms", time.Since(start).Milliseconds()).
			Msg("Finished metadata hydration")
	}
}

func (fh *FileHydrator) hydrateGroupMetadata(
	mId int,
	lfs []*anime.LocalFile, // Grouped local files
	rateLimiter *limiter.Limiter,
) {

	// Get the media
	media, found := lo.Find(fh.AllMedia, func(media *anime.NormalizedMedia) bool {
		return media.ID == mId
	})
	if !found {
		if fh.ScanLogger != nil {
			fh.ScanLogger.LogFileHydrator(zerolog.ErrorLevel).
				Int("mediaId", mId).
				Msg("Could not find media in FileHydrator options")
		}
		return
	}

	// Make sure the media is fetched
	_ = anime.FetchNormalizedMedia(fh.PlatformRef.Get().GetAnilistClient(), fh.AnilistRateLimiter, fh.CompleteAnimeCache, media)

	// Tree contains media relations
	tree := anilist.NewCompleteAnimeRelationTree()
	// Tree analysis used for episode normalization
	var mediaTreeAnalysis *MediaTreeAnalysis
	treeFetched := false
	mediaTreeAnalysisMu := sync.Mutex{}

	// Process each local file in the group sequentially
	lop.ForEach(lfs, func(lf *anime.LocalFile, index int) {

		defer util.HandlePanicInModuleThenS("scanner/hydrator/hydrateGroupMetadata", func(stackTrace string) {
			lf.MediaId = 0
			/*Log*/
			if fh.ScanLogger != nil {
				fh.ScanLogger.LogFileHydrator(zerolog.ErrorLevel).
					Str("filename", lf.Name).
					Msg("Panic occurred, file un-matched")
			}
			fh.ScanSummaryLogger.LogPanic(lf, stackTrace)
		})

		episode := -1

		// Invoke ScanLocalFileHydrationStarted hook
		event := &ScanLocalFileHydrationStartedEvent{
			LocalFile: lf,
			Media:     media,
		}
		_ = hook.GlobalHookManager.OnScanLocalFileHydrationStarted().Trigger(event)
		lf = event.LocalFile
		media = event.Media

		defer func() {
			// Invoke ScanLocalFileHydrated hook
			event := &ScanLocalFileHydratedEvent{
				LocalFile: lf,
				MediaId:   mId,
				Episode:   episode,
			}
			_ = hook.GlobalHookManager.OnScanLocalFileHydrated().Trigger(event)
			lf = event.LocalFile
			mId = event.MediaId
			episode = event.Episode
		}()

		// Handle hook override
		if event.DefaultPrevented {
			if fh.ScanLogger != nil {
				fh.ScanLogger.LogFileHydrator(zerolog.DebugLevel).
					Str("filename", lf.Name).
					Msg("Default hydration skipped by hook")
			}
			fh.ScanSummaryLogger.LogDebug(lf, "Default hydration skipped by hook")
			return
		}

		// Apply hydration rule to the file
		if fh.applyHydrationRule(lf) {
			return
		}

		lf.Metadata.Type = anime.LocalFileTypeMain

		// Get episode number
		if len(lf.ParsedData.Episode) > 0 {
			if ep, ok := util.StringToInt(lf.ParsedData.Episode); ok {
				episode = ep
			}
		}

		// NC metadata
		if lf.IsProbablyNC() {
			lf.Metadata.Episode = 0
			lf.Metadata.AniDBEpisode = ""
			lf.Metadata.Type = anime.LocalFileTypeNC

			if episode > -1 {
				lf.Metadata.Episode = episode
				// Parse OP/ED type from the filename and set AniDBEpisode for better tracking
				// e.g. "OP1", "ED2", etc.
				if ncType, ok := comparison.ExtractNCType(lf.Name); ok {
					lf.Metadata.AniDBEpisode = ncType + strconv.Itoa(episode)
				}
			}

			/*Log */
			if fh.ScanLogger != nil {
				fh.logFileHydration(zerolog.DebugLevel, lf, mId, episode).
					Msg("File has been marked as NC")
			}
			fh.ScanSummaryLogger.LogMetadataNC(lf)
			return
		}

		// Special metadata
		if lf.IsProbablySpecial() {
			lf.Metadata.Type = anime.LocalFileTypeSpecial
			// Sometimes a movie filename could be written as a special episode of the main series
			// anidb rarely if ever adds relevant specials to movie entries
			if *media.Format == anilist.MediaFormatMovie {
				lf.Metadata.Episode = 1
				lf.Metadata.AniDBEpisode = "1"
				lf.Metadata.Type = anime.LocalFileTypeMain
			} else if episode > -1 {
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
			if fh.ScanLogger != nil {
				fh.logFileHydration(zerolog.DebugLevel, lf, mId, episode).
					Msg("File has been marked as special")
			}
			fh.ScanSummaryLogger.LogMetadataSpecial(lf, lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
			return
		}
		// Movie metadata
		if *media.Format == anilist.MediaFormatMovie {
			lf.Metadata.Episode = 1
			lf.Metadata.AniDBEpisode = "1"

			/*Log */
			if fh.ScanLogger != nil {
				fh.logFileHydration(zerolog.DebugLevel, lf, mId, episode).
					Msg("File has been marked as main")
			}
			fh.ScanSummaryLogger.LogMetadataMain(lf, lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
			return
		}

		// No absolute episode count
		// "media.GetTotalEpisodeCount() == -1" is a fix for media with unknown episode count, we will just assume that the episode number is correct
		// devnote(don't know if this is still revelant): We might want to fetch the media when the episode count is unknown in order to get the correct episode count
		if episode > -1 && (episode <= media.GetCurrentEpisodeCount() || media.GetTotalEpisodeCount() == -1) {
			// Episode 0 - Might be a special
			// By default, we will assume that AniDB doesn't include Episode 0 as part of the main episodes (which is often the case)
			if episode == 0 {
				// Leave episode number as 0, assuming that the client will handle tracking correctly
				lf.Metadata.Episode = 0
				lf.Metadata.AniDBEpisode = "S1"

				/*Log */
				if fh.ScanLogger != nil {
					fh.logFileHydration(zerolog.DebugLevel, lf, mId, episode).
						Msg("File has been marked as main")
				}
				fh.ScanSummaryLogger.LogMetadataEpisodeZero(lf, lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
				return
			}

			lf.Metadata.Episode = episode
			lf.Metadata.AniDBEpisode = strconv.Itoa(episode)

			/*Log */
			if fh.ScanLogger != nil {
				fh.logFileHydration(zerolog.DebugLevel, lf, mId, episode).
					Msg("File has been marked as main")
			}
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
			if fh.ScanLogger != nil {
				fh.logFileHydration(zerolog.WarnLevel, lf, mId, episode).
					Str("warning", "File's episode number is higher than the media's episode count, but the media only has 1 episode").
					Msg("File has been marked as main")
			}
			fh.ScanSummaryLogger.LogMetadataMain(lf, lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
			return
		}

		// No episode number, but the media only has 1 episode
		if episode == -1 && media.GetCurrentEpisodeCount() == 1 {
			lf.Metadata.Episode = 1 // Coerce episode number to 1 because it is used for tracking
			lf.Metadata.AniDBEpisode = "1"

			/*Log */
			if fh.ScanLogger != nil {
				fh.logFileHydration(zerolog.WarnLevel, lf, mId, episode).
					Str("warning", "No episode number found, but the media only has 1 episode").
					Msg("File has been marked as main")
			}
			fh.ScanSummaryLogger.LogMetadataMain(lf, lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
			return
		}

		// Still no episode number and the media has more than 1 episode and is not a movie
		// We will mark it as a special episode
		// devnote: Mark as NC?
		if episode == -1 {
			lf.Metadata.Type = anime.LocalFileTypeSpecial
			lf.Metadata.Episode = 1
			lf.Metadata.AniDBEpisode = "S1"

			/*Log */
			if fh.ScanLogger != nil {
				fh.logFileHydration(zerolog.ErrorLevel, lf, mId, episode).
					Msg("No episode number found, file has been marked as special")
			}
			fh.ScanSummaryLogger.LogMetadataEpisodeNormalizationFailed(lf, errors.New("no episode number found"), lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
			return
		}

		// Absolute episode count
		if episode > media.GetCurrentEpisodeCount() && fh.ForceMediaId == 0 {

			// Try part-relative normalization before expensive media tree analysis.
			// This handles cases where filenames use numbering relative to a previous part of the same season
			// (e.g., S04E12 for the first episode of Part 2, where Part 1 had 11 episodes).
			if relativeEp, ok := fh.tryPartRelativeNormalization(lf, episode, mId, rateLimiter); ok {
				lf.Metadata.Episode = relativeEp
				lf.Metadata.AniDBEpisode = strconv.Itoa(relativeEp)

				/*Log */
				if fh.ScanLogger != nil {
					fh.logFileHydration(zerolog.DebugLevel, lf, mId, episode).
						Dict("partRelativeNormalization", zerolog.Dict().
							Bool("normalized", true).
							Int("relativeEpisode", relativeEp),
						).
						Msg("File normalized via part-relative detection")
				}
				fh.ScanSummaryLogger.LogMetadataMain(lf, lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
				return
			}

			mediaTreeAnalysisMu.Lock()
			defer mediaTreeAnalysisMu.Unlock()
			if !treeFetched {

				mediaTreeFetchStart := time.Now()
				// Fetch media tree
				// The media tree will be used to normalize episode numbers
				if err := media.FetchMediaTree(anilist.FetchMediaTreeAll, fh.PlatformRef.Get().GetAnilistClient(), fh.AnilistRateLimiter, tree, fh.CompleteAnimeCache); err == nil {
					// Create a new media tree analysis that will be used for episode normalization
					mta, _ := NewMediaTreeAnalysis(&MediaTreeAnalysisOptions{
						tree:                tree,
						metadataProviderRef: fh.MetadataProviderRef,
						rateLimiter:         rateLimiter,
					})
					// Hoist the media tree analysis, so it will be used by other files
					// We don't care if it's nil because [normalizeEpisodeNumberAndHydrate] will handle it
					mediaTreeAnalysis = mta
					treeFetched = true

					/*Log */
					if mta != nil && mta.branches != nil {
						if fh.ScanLogger != nil {
							fh.ScanLogger.LogFileHydrator(zerolog.DebugLevel).
								Int("mediaId", mId).
								Int64("ms", time.Since(mediaTreeFetchStart).Milliseconds()).
								Int("requests", len(mediaTreeAnalysis.branches)).
								Any("branches", mediaTreeAnalysis.printBranches()).
								Msg("Media tree fetched")
						}
						fh.ScanSummaryLogger.LogMetadataMediaTreeFetched(lf, time.Since(mediaTreeFetchStart).Milliseconds(), len(mediaTreeAnalysis.branches))
					}
				} else {
					if fh.ScanLogger != nil {
						fh.ScanLogger.LogFileHydrator(zerolog.ErrorLevel).
							Int("mediaId", mId).
							Str("error", err.Error()).
							Int64("ms", time.Since(mediaTreeFetchStart).Milliseconds()).
							Msg("Could not fetch media tree")
					}
					fh.ScanSummaryLogger.LogMetadataMediaTreeFetchFailed(lf, err, time.Since(mediaTreeFetchStart).Milliseconds())
				}
			}

			// Normalize episode number
			if err := fh.normalizeEpisodeNumberAndHydrate(mediaTreeAnalysis, lf, episode, media.GetCurrentEpisodeCount()); err != nil {

				/*Log */
				if fh.ScanLogger != nil {
					fh.logFileHydration(zerolog.WarnLevel, lf, mId, episode).
						Dict("mediaTreeAnalysis", zerolog.Dict().
							Bool("normalized", false).
							Str("error", err.Error()).
							Str("reason", "Episode normalization failed"),
						).
						Msg("File has been marked as special")
				}
				fh.ScanSummaryLogger.LogMetadataEpisodeNormalizationFailed(lf, err, lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
			} else {
				/*Log */
				if fh.ScanLogger != nil {
					fh.logFileHydration(zerolog.DebugLevel, lf, mId, episode).
						Dict("mediaTreeAnalysis", zerolog.Dict().
							Bool("normalized", true).
							Bool("hasNewMediaId", lf.MediaId != mId).
							Int("newMediaId", lf.MediaId),
						).
						Msg("File has been marked as main")
				}
				fh.ScanSummaryLogger.LogMetadataEpisodeNormalized(lf, mId, episode, lf.Metadata.Episode, lf.MediaId, lf.Metadata.AniDBEpisode)
			}
			return
		}

		// Absolute episode count with forced media ID
		if fh.ForceMediaId != 0 && episode > media.GetCurrentEpisodeCount() {

			// When we encounter a file with an episode number higher than the media's episode count
			// we have a forced media ID, we will fetch the media from AniList and get the offset
			animeMetadata, err := fh.MetadataProviderRef.Get().GetAnimeMetadata(metadata.AnilistPlatform, fh.ForceMediaId)
			if err != nil {
				/*Log */
				if fh.ScanLogger != nil {
					fh.logFileHydration(zerolog.ErrorLevel, lf, mId, episode).
						Str("error", err.Error()).
						Msg("Could not fetch AniDB metadata")
				}
				lf.Metadata.Episode = episode
				lf.Metadata.AniDBEpisode = strconv.Itoa(episode)
				lf.MediaId = fh.ForceMediaId
				fh.ScanSummaryLogger.LogMetadataEpisodeNormalizationFailed(lf, errors.New("could not fetch AniDB metadata"), lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
				return
			}

			// Get the first episode to calculate the offset
			firstEp, ok := animeMetadata.Episodes["1"]
			if !ok {
				/*Log */
				if fh.ScanLogger != nil {
					fh.logFileHydration(zerolog.ErrorLevel, lf, mId, episode).
						Msg("Could not find absolute episode offset")
				}
				lf.Metadata.Episode = episode
				lf.Metadata.AniDBEpisode = strconv.Itoa(episode)
				lf.MediaId = fh.ForceMediaId
				fh.ScanSummaryLogger.LogMetadataEpisodeNormalizationFailed(lf, errors.New("could not find absolute episode offset"), lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
				return
			}

			// ref: media_tree_analysis.go
			usePartEpisodeNumber := firstEp.EpisodeNumber > 1 && firstEp.AbsoluteEpisodeNumber-firstEp.EpisodeNumber > 1
			minPartAbsoluteEpisodeNumber := 0
			maxPartAbsoluteEpisodeNumber := 0
			if usePartEpisodeNumber {
				minPartAbsoluteEpisodeNumber = firstEp.EpisodeNumber
				maxPartAbsoluteEpisodeNumber = minPartAbsoluteEpisodeNumber + animeMetadata.GetMainEpisodeCount() - 1
			}

			absoluteEpisodeNumber := firstEp.AbsoluteEpisodeNumber

			// Calculate the relative episode number
			relativeEp := episode

			// Let's say the media has 12 episodes and the file is "episode 13"
			// If the [partAbsoluteEpisodeNumber] is 13, then the [relativeEp] will be 1, we can safely ignore the [absoluteEpisodeNumber]
			// e.g. 13 - (13-1) = 1
			if minPartAbsoluteEpisodeNumber <= episode && maxPartAbsoluteEpisodeNumber >= episode {
				relativeEp = episode - (minPartAbsoluteEpisodeNumber - 1)
			} else {
				// Let's say the media has 12 episodes and the file is "episode 38"
				// The [absoluteEpisodeNumber] will be 38 and the [relativeEp] will be 1
				// e.g. 38 - (38-1) = 1
				relativeEp = episode - (absoluteEpisodeNumber - 1)
			}

			if relativeEp < 1 {
				if fh.ScanLogger != nil {
					fh.logFileHydration(zerolog.WarnLevel, lf, mId, episode).
						Dict("normalization", zerolog.Dict().
							Bool("normalized", false).
							Str("reason", "Episode normalization failed, could not find relative episode number"),
						).
						Msg("File has been marked as main")
				}
				lf.Metadata.Episode = episode
				lf.Metadata.AniDBEpisode = strconv.Itoa(episode)
				lf.MediaId = fh.ForceMediaId
				fh.ScanSummaryLogger.LogMetadataEpisodeNormalizationFailed(lf, errors.New("could not find relative episode number"), lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
				return
			}

			if fh.ScanLogger != nil {
				fh.logFileHydration(zerolog.DebugLevel, lf, mId, relativeEp).
					Dict("mediaTreeAnalysis", zerolog.Dict().
						Bool("normalized", true).
						Int("forcedMediaId", fh.ForceMediaId),
					).
					Msg("File has been marked as main")
			}
			lf.Metadata.Episode = relativeEp
			lf.Metadata.AniDBEpisode = strconv.Itoa(relativeEp)
			lf.MediaId = fh.ForceMediaId
			fh.ScanSummaryLogger.LogMetadataMain(lf, lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
			return

		}

	})

}

func (fh *FileHydrator) logFileHydration(level zerolog.Level, lf *anime.LocalFile, mId int, episode int) *zerolog.Event {
	return fh.ScanLogger.LogFileHydrator(level).
		Str("filename", lf.Name).
		Int("mediaId", mId).
		Dict("parsed", zerolog.Dict().
			Str("parsedEpisode", lf.ParsedData.Episode).
			Int("episode", episode),
		).
		Dict("hydrated", zerolog.Dict().
			Int("episode", lf.Metadata.Episode).
			Str("aniDBEpisode", lf.Metadata.AniDBEpisode))
}

// tryPartRelativeNormalization detects and normalizes "part-relative absolute" episode numbers.
//
// Some release groups number episodes from a second cour relative to the start of a season rather than the first season.
// e.g., a file named "S04E12" in a folder for "Danmachi IV Part 2" means episode 12 relative to S04, not S01.
//
// This is detected by fetching the matched media's metadata and checking if its first episode has
// an episodeNumber > 1, which indicates the metadata provider recognizes it as a continuation.
// If the file's episode number falls within the range [firstEp.EpisodeNumber, firstEp.EpisodeNumber + episodeCount - 1],
// we can directly compute the relative episode number without the full media tree analysis.
func (fh *FileHydrator) tryPartRelativeNormalization(
	lf *anime.LocalFile,
	episode int, // parsed episode number (e.g. 12)
	mediaId int, // matched media ID
	rateLimiter *limiter.Limiter,
) (int, bool) {
	if fh.MetadataProviderRef == nil {
		return 0, false
	}

	rateLimiter.Wait()

	animeMetadata, err := fh.MetadataProviderRef.Get().GetAnimeMetadata(metadata.AnilistPlatform, mediaId)
	if err != nil {
		return 0, false
	}

	firstEp, ok := animeMetadata.Episodes["1"]
	if !ok {
		return 0, false
	}

	// If the first episode's episodeNumber is > 1 and differs significantly from its absoluteEpisodeNumber,
	// the media is a "Part 2" (or later) that continues from a previous part of the same season.
	// The episodeNumber represents the part-relative continuation number.
	if firstEp.EpisodeNumber <= 1 || firstEp.AbsoluteEpisodeNumber-firstEp.EpisodeNumber <= 1 {
		return 0, false
	}

	minPartEp := firstEp.EpisodeNumber
	mainEpCount := animeMetadata.GetMainEpisodeCount()
	maxPartEp := minPartEp + mainEpCount - 1

	// Check if the file's episode falls within the part-relative range
	if episode < minPartEp || episode > maxPartEp {
		return 0, false
	}

	relativeEp := episode - (minPartEp - 1)
	if relativeEp < 1 {
		return 0, false
	}

	if fh.ScanLogger != nil {
		fh.ScanLogger.LogFileHydrator(zerolog.DebugLevel).
			Str("filename", lf.Name).
			Int("mediaId", mediaId).
			Int("parsedEpisode", episode).
			Int("partMinEp", minPartEp).
			Int("partMaxEp", maxPartEp).
			Int("relativeEp", relativeEp).
			Msg("Detected part-relative absolute numbering")
	}

	return relativeEp, true
}

// normalizeEpisodeNumberAndHydrate will normalize the episode number and hydrate the metadata of the LocalFile.
// If the MediaTreeAnalysis is nil, the episode number will not be normalized.
func (fh *FileHydrator) normalizeEpisodeNumberAndHydrate(
	mta *MediaTreeAnalysis,
	lf *anime.LocalFile,
	ep int, // The absolute episode number of the media
	maxEp int, // The maximum episode number of the media
) error {
	// No media tree analysis
	if mta == nil {
		diff := ep - maxEp // e.g. 14 - 12 = 2
		// Let's consider this a special episode (it might not exist on AniDB, but it's better than setting everything to "S1")
		lf.Metadata.Episode = diff                          // e.g. 2
		lf.Metadata.AniDBEpisode = "S" + strconv.Itoa(diff) // e.g. S2
		lf.Metadata.Type = anime.LocalFileTypeSpecial
		return errors.New("[hydrator] could not find media tree")
	}

	relativeEp, mediaId, ok := mta.getRelativeEpisodeNumber(ep)
	if !ok {
		diff := ep - maxEp // e.g. 14 - 12 = 2
		// Do the same as above
		lf.Metadata.Episode = diff
		lf.Metadata.AniDBEpisode = "S" + strconv.Itoa(diff) // e.g. S2
		lf.Metadata.Type = anime.LocalFileTypeSpecial
		return errors.New("[hydrator] could not find relative episode number from media tree")
	}

	lf.Metadata.Episode = relativeEp
	lf.Metadata.AniDBEpisode = strconv.Itoa(relativeEp)
	lf.MediaId = mediaId
	return nil
}

func (fh *FileHydrator) precompileRules() {
	defer util.HandlePanicInModuleThenS("scanner/hydration/precompileRules", func(stackTrace string) {
		if fh.ScanLogger != nil {
			fh.ScanLogger.LogMatcher(zerolog.ErrorLevel).
				Msg("Panic occurred, when compiling hydration rules")
		}
	})

	if fh.Config == nil || len(fh.Config.Hydration.Rules) == 0 {
		fh.ScanLogger.LogMatcher(zerolog.DebugLevel).Msg("Zero config")
		return
	}

	for _, rule := range fh.Config.Hydration.Rules {
		// we accept either a patterh or media id
		if rule.Pattern == "" && rule.MediaID == 0 {
			continue
		}

		r := &compiledHydrationRule{
			regex:        nil,
			rule:         rule,
			fileRulesRgx: map[string]*compiledHydrationFileRule{},
		}

		if rule.Pattern != "" {
			rgx, err := regexp.Compile(rule.Pattern)
			if err != nil {
				if fh.ScanLogger != nil {
					fh.ScanLogger.LogMatcher(zerolog.WarnLevel).
						Str("pattern", rule.Pattern).
						Msg("Config: Invalid hydration regex pattern")
				}
				continue
			}
			rgx.Longest()
			r.regex = rgx
		}

		for _, fr := range rule.Files {
			if fr.IsRegex {
				rgx, err := regexp.Compile(fr.Filename)
				if err != nil {
					if fh.ScanLogger != nil {
						fh.ScanLogger.LogMatcher(zerolog.WarnLevel).
							Str("filename", fr.Filename).
							Str("pattern", rule.Pattern).
							Msg("Config: Invalid hydration regex pattern")
					}
					continue
				}
				rgx.Longest()
				r.fileRulesRgx[fr.Filename] = &compiledHydrationFileRule{
					regex: rgx,
					rule:  fr,
				}
			} else {
				r.fileRulesRgx[fr.Filename] = &compiledHydrationFileRule{
					regex: nil,
					rule:  fr,
				}
			}
		}

		fh.hydrationRules[rule.Pattern] = r
	}
}

func (fh *FileHydrator) applyHydrationRule(lf *anime.LocalFile) bool {
	defer util.HandlePanicInModuleThenS("scanner/matcher/applyMatcingRule", func(stackTrace string) {
		if fh.ScanLogger != nil {
			fh.ScanLogger.LogMatcher(zerolog.ErrorLevel).
				Str("filename", lf.Name).
				Msg("Panic occurred when applying matching rule")
		}
		fh.ScanSummaryLogger.LogPanic(lf, stackTrace)
	})

	for _, rule := range fh.hydrationRules {
		if rule.regex == nil && rule.rule.MediaID == 0 {
			continue
		}
		// skip if the regex doesn't match
		if rule.regex != nil && !rule.regex.MatchString(lf.Name) {
			continue
		}
		// skip if the media ids don't match
		if rule.rule.MediaID != 0 && lf.MediaId != rule.rule.MediaID {
			continue
		}

		for _, fileRule := range rule.fileRulesRgx {
			ok := false
			if fileRule.regex != nil && fileRule.regex.MatchString(lf.Name) {
				ok = true
			}
			if !ok && util.NormalizePath(fileRule.rule.Filename) == util.NormalizePath(lf.Name) {
				ok = true
			}
			if !ok {
				continue
			}

			// Apply metadata from the file rule
			episode := fileRule.rule.Episode
			aniDbEpisode := fileRule.rule.AniDbEpisode
			fileType := fileRule.rule.Type

			// Handle regex substitutions for capture groups ($1, $2, etc)
			if fileRule.regex != nil {
				matches := fileRule.regex.FindStringSubmatch(lf.Name)
				if len(matches) > 1 {
					// Replace $1, $2, etc with captured groups
					for i := 1; i < len(matches); i++ {
						placeholder := "$" + strconv.Itoa(i)
						episode = strings.ReplaceAll(episode, placeholder, matches[i])
						aniDbEpisode = strings.ReplaceAll(aniDbEpisode, placeholder, matches[i])
					}
				}
			}

			// Evaluate {calc(...)} expressions in aniDbEpisode
			// e.g. "S{calc($1-11)}"
			aniDbEpisode = fh.evaluateCalcExpressions(aniDbEpisode)

			// Set the metadata
			if episode != "" {
				value, err := util.EvaluateSimpleExpression(episode)
				if err == nil {
					lf.Metadata.Episode = value
				} else {
					if ep, ok := util.StringToInt(episode); ok {
						lf.Metadata.Episode = ep
					}
				}
			}
			if aniDbEpisode != "" {
				lf.Metadata.AniDBEpisode = aniDbEpisode
			}
			if fileType != "" {
				lf.Metadata.Type = fileType
			}

			return true
		}

	}

	return false
}

// evaluateCalcExpressions finds and evaluates {calc(...)} expressions in a string
// e.g. "S{calc(12-11)}" -> "S1"
func (fh *FileHydrator) evaluateCalcExpressions(input string) string {
	// Find all {calc(...)} patterns
	re := regexp.MustCompile(`\{calc\(([^)]+)\)}`)

	result := re.ReplaceAllStringFunc(input, func(match string) string {
		// extract the expression inside calc()
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		expression := submatches[1]

		value, err := util.EvaluateSimpleExpression(expression)
		if err != nil {
			// return the original match
			return match
		}

		return strconv.Itoa(value)
	})

	return result
}
