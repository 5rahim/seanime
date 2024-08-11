package scanner

import (
	"errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/library/filesystem"
	"seanime/internal/library/summary"
	"seanime/internal/platforms/platform"
	"seanime/internal/util"
	"seanime/internal/util/limiter"
)

type Scanner struct {
	DirPath            string
	Enhanced           bool
	Platform           platform.Platform
	Logger             *zerolog.Logger
	WSEventManager     events.WSEventManagerInterface
	ExistingLocalFiles []*anime.LocalFile
	SkipLockedFiles    bool
	SkipIgnoredFiles   bool
	ScanSummaryLogger  *summary.ScanSummaryLogger
	ScanLogger         *ScanLogger
}

// Scan will scan the directory and return a list of anime.LocalFile.
func (scn *Scanner) Scan() (lfs []*anime.LocalFile, err error) {

	defer util.HandlePanicWithError(&err)

	completeAnimeCache := anilist.NewCompleteAnimeCache()
	anizipCache := anizip.NewCache()

	// Create a new Anilist rate limiter
	anilistRateLimiter := limiter.NewAnilistLimiter()

	if scn.ScanSummaryLogger == nil {
		scn.ScanSummaryLogger = summary.NewScanSummaryLogger()
	}

	scn.Logger.Debug().Msg("scanner: Starting scan")
	scn.WSEventManager.SendEvent(events.EventScanProgress, 10)
	scn.WSEventManager.SendEvent(events.EventScanStatus, "Retrieving local files...")

	// +---------------------+
	// |     Local Files     |
	// +---------------------+

	// Get local files
	localFiles, err := GetLocalFilesFromDir(scn.DirPath, scn.Logger)
	if err != nil {
		return nil, err
	}

	if scn.ScanLogger != nil {
		scn.ScanLogger.logger.Info().
			Any("count", len(localFiles)).
			Msg("Retrieved and parsed local files")
	}

	for _, lf := range localFiles {
		if scn.ScanLogger != nil {
			scn.ScanLogger.logger.Trace().
				Str("path", lf.Path).
				Any("parsedData", spew.Sdump(lf.ParsedData)).
				Any("parsedFolderData", spew.Sdump(lf.ParsedFolderData)).
				Msg("Parsed local file")
		}
	}

	if scn.ScanLogger != nil {
		scn.ScanLogger.logger.Debug().
			Msg("===========================================================================================================")
	}

	// +---------------------+
	// | Filter local files  |
	// +---------------------+

	// Get skipped files depending on options
	skippedLfs := make([]*anime.LocalFile, 0)
	if (scn.SkipLockedFiles || scn.SkipIgnoredFiles) && scn.ExistingLocalFiles != nil {
		// Retrieve skipped files from existing local files
		for _, lf := range scn.ExistingLocalFiles {
			if scn.SkipLockedFiles && lf.IsLocked() {
				skippedLfs = append(skippedLfs, lf)
			} else if scn.SkipIgnoredFiles && lf.IsIgnored() {
				skippedLfs = append(skippedLfs, lf)
			}
		}

		// Remove skipped files from local files that will be hydrated
		localFiles = lo.Filter(localFiles, func(lf *anime.LocalFile, _ int) bool {
			if lf.IsIncluded(skippedLfs) {
				return false
			}
			return true
		})
	}

	// +---------------------+
	// |  No files to scan   |
	// +---------------------+

	// If there are no local files to scan (all files are skipped, or a file was deleted)
	if len(localFiles) == 0 {
		scn.WSEventManager.SendEvent(events.EventScanProgress, 90)
		scn.WSEventManager.SendEvent(events.EventScanStatus, "Verifying file integrity...")
		// Add skipped files
		if len(skippedLfs) > 0 {
			for _, sf := range skippedLfs {
				if filesystem.FileExists(sf.Path) { // Verify that the file still exists
					localFiles = append(localFiles, sf)
				}
			}
		}
		scn.Logger.Debug().Msg("scanner: Scan completed")
		scn.WSEventManager.SendEvent(events.EventScanProgress, 100)
		scn.WSEventManager.SendEvent(events.EventScanStatus, "Scan completed")

		return localFiles, nil
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 20)
	if scn.Enhanced {
		scn.WSEventManager.SendEvent(events.EventScanStatus, "Fetching media detected from file titles...")
	} else {
		scn.WSEventManager.SendEvent(events.EventScanStatus, "Fetching media...")
	}

	// +---------------------+
	// |    MediaFetcher     |
	// +---------------------+

	// Fetch media needed for matching
	mf, err := NewMediaFetcher(&MediaFetcherOptions{
		Enhanced:           scn.Enhanced,
		Platform:           scn.Platform,
		LocalFiles:         localFiles,
		CompleteAnimeCache: completeAnimeCache,
		AnizipCache:        anizipCache,
		Logger:             scn.Logger,
		AnilistRateLimiter: anilistRateLimiter,
		ScanLogger:         scn.ScanLogger,
	})
	if err != nil {
		return nil, err
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 40)
	scn.WSEventManager.SendEvent(events.EventScanStatus, "Matching local files...")

	// +---------------------+
	// |   MediaContainer    |
	// +---------------------+

	// Create a new container for media
	mc := NewMediaContainer(&MediaContainerOptions{
		AllMedia:   mf.AllMedia,
		ScanLogger: scn.ScanLogger,
	})

	scn.Logger.Debug().
		Any("count", len(mc.NormalizedMedia)).
		Msg("media container: Media container created")

	// +---------------------+
	// |      Matcher        |
	// +---------------------+

	// Create a new matcher
	matcher := &Matcher{
		LocalFiles:         localFiles,
		MediaContainer:     mc,
		CompleteAnimeCache: completeAnimeCache,
		Logger:             scn.Logger,
		ScanLogger:         scn.ScanLogger,
		ScanSummaryLogger:  scn.ScanSummaryLogger,
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 60)

	err = matcher.MatchLocalFilesWithMedia()
	if err != nil {
		// If the matcher received no local files, return an error
		if errors.Is(err, ErrNoLocalFiles) {
			scn.Logger.Debug().Msg("scanner: Scan completed")
			scn.WSEventManager.SendEvent(events.EventScanProgress, 100)
			scn.WSEventManager.SendEvent(events.EventScanStatus, "Scan completed")
		}
		return nil, err
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 70)
	scn.WSEventManager.SendEvent(events.EventScanStatus, "Hydrating metadata...")

	// +---------------------+
	// |    FileHydrator     |
	// +---------------------+

	// Create a new hydrator
	hydrator := &FileHydrator{
		AllMedia:           mc.NormalizedMedia,
		LocalFiles:         localFiles,
		AnizipCache:        anizipCache,
		Platform:           scn.Platform,
		CompleteAnimeCache: completeAnimeCache,
		AnilistRateLimiter: anilistRateLimiter,
		Logger:             scn.Logger,
		ScanLogger:         scn.ScanLogger,
		ScanSummaryLogger:  scn.ScanSummaryLogger,
	}
	hydrator.HydrateMetadata()

	scn.WSEventManager.SendEvent(events.EventScanProgress, 80)

	// +---------------------+
	// |  Add missing media  |
	// +---------------------+

	// Add non-added media entries to AniList collection
	// Max of 4 to avoid rate limit issues
	if len(mf.UnknownMediaIds) < 5 {
		scn.WSEventManager.SendEvent(events.EventScanStatus, "Adding missing media to AniList...")

		if err = scn.Platform.AddMediaToCollection(mf.UnknownMediaIds); err != nil {
			scn.Logger.Warn().Msg("scanner: An error occurred while adding media to planning list: " + err.Error())
		}
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 90)
	scn.WSEventManager.SendEvent(events.EventScanStatus, "Verifying file integrity...")

	// Hydrate the summary logger before merging files
	scn.ScanSummaryLogger.HydrateData(localFiles, mc.NormalizedMedia, mf.AnimeCollectionWithRelations)

	// +---------------------+
	// |    Merge files      |
	// +---------------------+

	// Merge skipped files with scanned files
	// Only files that exist (this removes deleted/moved files)
	if len(skippedLfs) > 0 {
		for _, sf := range skippedLfs {
			if filesystem.FileExists(sf.Path) {
				localFiles = append(localFiles, sf)
			}
		}
	}

	scn.Logger.Info().Msg("scanner: Scan completed")
	scn.WSEventManager.SendEvent(events.EventScanProgress, 100)
	scn.WSEventManager.SendEvent(events.EventScanStatus, "Scan completed")

	if scn.ScanLogger != nil {
		scn.ScanLogger.logger.Info().
			Int("scannedFileCount", len(localFiles)).
			Int("skippedFileCount", len(skippedLfs)).
			Int("unknownMediaCount", len(mf.UnknownMediaIds)).
			Msg("Scan completed")
	}

	return localFiles, nil
}
