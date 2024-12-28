package scanner

import (
	"errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/library/filesystem"
	"seanime/internal/library/summary"
	"seanime/internal/platforms/platform"
	"seanime/internal/util"
	"seanime/internal/util/limiter"
	"strings"
	"sync"
)

type Scanner struct {
	DirPath            string
	OtherDirPaths      []string
	Enhanced           bool
	Platform           platform.Platform
	Logger             *zerolog.Logger
	WSEventManager     events.WSEventManagerInterface
	ExistingLocalFiles []*anime.LocalFile
	SkipLockedFiles    bool
	SkipIgnoredFiles   bool
	ScanSummaryLogger  *summary.ScanSummaryLogger
	ScanLogger         *ScanLogger
	MetadataProvider   metadata.Provider
}

// Scan will scan the directory and return a list of anime.LocalFile.
func (scn *Scanner) Scan() (lfs []*anime.LocalFile, err error) {
	defer util.HandlePanicWithError(&err)

	scn.WSEventManager.SendEvent(events.EventScanProgress, 0)
	scn.WSEventManager.SendEvent(events.EventScanStatus, "Retrieving local files...")

	completeAnimeCache := anilist.NewCompleteAnimeCache()

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

	paths, err := filesystem.GetMediaFilePathsFromDirS(scn.DirPath)
	if err != nil {
		return nil, err
	}

	if scn.ScanLogger != nil {
		scn.ScanLogger.logger.Info().
			Any("count", len(paths)).
			Msg("Retrieved file paths from main directory")
	}

	// Get local files
	//localFiles, err := GetLocalFilesFromDir(scn.DirPath, scn.Logger)
	//if err != nil {
	//	return nil, err
	//}

	localFilePathsMap := make(map[string]struct{})
	for _, path := range paths {
		localFilePathsMap[util.NormalizePath(path)] = struct{}{}
	}

	// Get local files from other directories
	for _, dirPath := range scn.OtherDirPaths {
		//otherLocalFiles, err := GetLocalFilesFromDir(dirPath, scn.Logger)
		otherPaths, err := filesystem.GetMediaFilePathsFromDirS(dirPath)
		if scn.ScanLogger != nil {
			scn.ScanLogger.logger.Info().
				Any("count", len(otherPaths)).
				Msgf("Retrieved file paths from other directory: %s", dirPath)
		}
		if err != nil {
			return nil, err
		}
		for _, path := range otherPaths {
			if _, ok := localFilePathsMap[util.NormalizePath(path)]; !ok {
				paths = append(paths, path)
			}
		}
	}

	if scn.ScanLogger != nil {
		scn.ScanLogger.logger.Info().
			Any("count", len(paths)).
			Msg("Retrieved file paths from all directories")
	}

	// +---------------------+
	// | Filter local files  |
	// +---------------------+

	localFiles := make([]*anime.LocalFile, 0)

	// Get skipped files depending on options
	skippedLfs := make(map[string]*anime.LocalFile, 0)
	if (scn.SkipLockedFiles || scn.SkipIgnoredFiles) && scn.ExistingLocalFiles != nil {
		// Retrieve skipped files from existing local files
		for _, lf := range scn.ExistingLocalFiles {
			if scn.SkipLockedFiles && lf.IsLocked() {
				skippedLfs[lf.GetNormalizedPath()] = lf
			} else if scn.SkipIgnoredFiles && lf.IsIgnored() {
				skippedLfs[lf.GetNormalizedPath()] = lf
			}
		}

		// Remove skipped files from local files that will be hydrated
		localFiles = lop.Map(paths, func(path string, _ int) *anime.LocalFile {
			if _, ok := skippedLfs[strings.ToLower(path)]; !ok {
				// Create a new local file
				return anime.NewLocalFile(path, scn.DirPath)
			} else {
				return nil
			}
		})

		// Remove nil values
		localFiles = lo.Filter(localFiles, func(lf *anime.LocalFile, _ int) bool {
			return lf != nil
		})

	}
	if scn.ScanLogger != nil {
		scn.ScanLogger.logger.Debug().
			Any("count", len(localFiles)).
			Msg("Local files to be scanned")
		scn.ScanLogger.logger.Debug().
			Any("count", len(skippedLfs)).
			Msg("Skipped files")

		scn.ScanLogger.logger.Debug().
			Msg("===========================================================================================================")
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

	// Remove local files from both skipped and un-skipped files if they are not under any of the directories
	allLibraries := []string{scn.DirPath}
	allLibraries = append(allLibraries, scn.OtherDirPaths...)
	localFiles = lo.Filter(localFiles, func(lf *anime.LocalFile, _ int) bool {
		if !util.IsSubdirectoryOfAny(allLibraries, lf.Path) {
			return false
		}
		return true
	})
	//skippedLfs = lo.Filter(skippedLfs, func(lf *anime.LocalFile, _ int) bool {
	//	if !util.IsSubdirectoryOfAny(allLibraries, lf.Path) {
	//		return false
	//	}
	//	return true
	//})
	for _, lf := range skippedLfs {
		if !util.IsSubdirectoryOfAny(allLibraries, lf.Path) {
			delete(skippedLfs, lf.GetNormalizedPath())
		}
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
		Enhanced:               scn.Enhanced,
		Platform:               scn.Platform,
		MetadataProvider:       scn.MetadataProvider,
		LocalFiles:             localFiles,
		CompleteAnimeCache:     completeAnimeCache,
		Logger:                 scn.Logger,
		AnilistRateLimiter:     anilistRateLimiter,
		DisableAnimeCollection: false,
		ScanLogger:             scn.ScanLogger,
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
		MetadataProvider:   scn.MetadataProvider,
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
		wg := sync.WaitGroup{}
		mu := sync.Mutex{}
		wg.Add(len(skippedLfs))
		for _, skippedLf := range skippedLfs {
			go func(skippedLf *anime.LocalFile) {
				defer wg.Done()
				if filesystem.FileExists(skippedLf.Path) {
					mu.Lock()
					localFiles = append(localFiles, skippedLf)
					mu.Unlock()
				}
			}(skippedLf)
		}
		wg.Wait()
	}

	scn.Logger.Info().Msg("scanner: Scan completed")
	scn.WSEventManager.SendEvent(events.EventScanProgress, 100)
	scn.WSEventManager.SendEvent(events.EventScanStatus, "Scan completed")

	if scn.ScanLogger != nil {
		scn.ScanLogger.logger.Info().
			Int("count", len(localFiles)).
			Int("unknownMediaCount", len(mf.UnknownMediaIds)).
			Msg("Scan completed")
	}

	return localFiles, nil
}
