package scanner

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"runtime"
	"runtime/debug"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/events"
	"seanime/internal/hook"
	"seanime/internal/library/anime"
	"seanime/internal/library/filesystem"
	"seanime/internal/library/summary"
	"seanime/internal/platforms/platform"
	"seanime/internal/util"
	"seanime/internal/util/limiter"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
)

type Scanner struct {
	DirPath                    string
	OtherDirPaths              []string
	Enhanced                   bool
	EnhanceWithOfflineDatabase bool
	PlatformRef                *util.Ref[platform.Platform]
	Logger                     *zerolog.Logger
	WSEventManager             events.WSEventManagerInterface
	ExistingLocalFiles         []*anime.LocalFile
	SkipLockedFiles            bool
	SkipIgnoredFiles           bool
	ScanSummaryLogger          *summary.ScanSummaryLogger
	ScanLogger                 *ScanLogger
	MetadataProviderRef        *util.Ref[metadata_provider.Provider]
	UseLegacyMatching          bool
	MatchingThreshold          float64 // only used by legacy
	MatchingAlgorithm          string  // only used by legacy
	// If true, locked files whose library path doesn't exist will be put aside
	WithShelving         bool
	ExistingShelvedFiles []*anime.LocalFile
	shelvedLocalFiles    []*anime.LocalFile
	Config               *Config
	ConfigAsString       string
	// Optional, used to add custom sources
	AnimeCollection *anilist.AnimeCollection
}

// Scan will scan the directory and return a list of anime.LocalFile.
func (scn *Scanner) Scan(ctx context.Context) (lfs []*anime.LocalFile, err error) {
	defer util.HandlePanicWithError(&err)

	go anime.EpisodeCollectionFromLocalFilesCache.Clear()

	scn.WSEventManager.SendEvent(events.EventScanProgress, 0)
	scn.WSEventManager.SendEvent(events.EventScanStatus, "Retrieving local files...")

	completeAnimeCache := anilist.NewCompleteAnimeCache()

	// Create a new Anilist rate limiter
	anilistRateLimiter := limiter.NewAnilistLimiter()

	if scn.ScanSummaryLogger == nil {
		scn.ScanSummaryLogger = summary.NewScanSummaryLogger()
	}

	if scn.ConfigAsString != "" && scn.Config == nil {
		scn.Config, _ = ToConfig(scn.ConfigAsString)
	}
	if scn.Config == nil {
		scn.Config = &Config{}
	}

	scn.Logger.Debug().Msg("scanner: Starting scan")
	scn.WSEventManager.SendEvent(events.EventScanProgress, 10)
	scn.WSEventManager.SendEvent(events.EventScanStatus, "Retrieving local files...")

	startTime := time.Now()

	if scn.ScanLogger != nil {
		scn.ScanLogger.logger.Info().
			Time("startTime", startTime).
			Msg("Scanning started")

		defer func() {
			now := time.Now()
			scn.ScanLogger.logger.Info().
				Time("endTime", time.Now()).
				Str("duration", now.Sub(startTime).String()).
				Int("localFilesCount", len(lfs)).
				Msg("Ended")
		}()
	}

	// Invoke ScanStarted hook
	event := &ScanStartedEvent{
		LibraryPath:       scn.DirPath,
		OtherLibraryPaths: scn.OtherDirPaths,
		Enhanced:          scn.Enhanced,
		SkipLocked:        scn.SkipLockedFiles,
		SkipIgnored:       scn.SkipIgnoredFiles,
		LocalFiles:        scn.ExistingLocalFiles,
	}
	_ = hook.GlobalHookManager.OnScanStarted().Trigger(event)
	scn.DirPath = event.LibraryPath
	scn.OtherDirPaths = event.OtherLibraryPaths
	scn.Enhanced = event.Enhanced
	scn.SkipLockedFiles = event.SkipLocked
	scn.SkipIgnoredFiles = event.SkipIgnored

	// Default prevented, return the local files
	if event.DefaultPrevented {
		// Invoke ScanCompleted hook
		completedEvent := &ScanCompletedEvent{
			LocalFiles: event.LocalFiles,
			Duration:   int(time.Since(startTime).Milliseconds()),
		}
		_ = hook.GlobalHookManager.OnScanCompleted().Trigger(completedEvent)

		return completedEvent.LocalFiles, nil
	}

	// +---------------------+
	// |     File paths      |
	// +---------------------+

	libraryPaths := append([]string{scn.DirPath}, scn.OtherDirPaths...)
	// Sort library paths by length, so that longer paths are checked first
	sortedLibraryPaths := make([]string, len(libraryPaths))
	copy(sortedLibraryPaths, libraryPaths)
	sort.Slice(sortedLibraryPaths, func(i, j int) bool {
		return len(sortedLibraryPaths[i]) > len(sortedLibraryPaths[j])
	})

	// Create a map of local file paths used to avoid duplicates
	retrievedPathMap := make(map[string]struct{})

	paths := make([]string, 0)
	mu := sync.Mutex{}
	logMu := sync.Mutex{}
	wg := sync.WaitGroup{}

	wg.Add(len(libraryPaths))

	// Get local files from all directories
	for i, dirPath := range libraryPaths {
		go func(dirPath string, i int) {
			defer wg.Done()
			retrievedPaths, err := filesystem.GetMediaFilePathsFromDirS(dirPath)
			if err != nil {
				scn.Logger.Error().Msgf("scanner: An error occurred while retrieving local files from directory: %s", err)
				return
			}

			if scn.ScanLogger != nil {
				logMu.Lock()
				if i == 0 {
					scn.ScanLogger.logger.Info().
						Any("count", len(paths)).
						Msgf("Retrieved file paths from main directory: %s", dirPath)
				} else {
					scn.ScanLogger.logger.Info().
						Any("count", len(retrievedPaths)).
						Msgf("Retrieved file paths from other directory: %s", dirPath)
				}
				logMu.Unlock()
			}

			for _, path := range retrievedPaths {
				if _, ok := retrievedPathMap[util.NormalizePath(path)]; !ok {
					mu.Lock()
					paths = append(paths, path)
					mu.Unlock()
				}
			}
		}(dirPath, i)
	}

	wg.Wait()

	if scn.ScanLogger != nil {
		scn.ScanLogger.logger.Info().
			Any("count", len(paths)).
			Msg("Retrieved file paths from all directories")
	}

	// Invoke ScanFilePathsRetrieved hook
	fpEvent := &ScanFilePathsRetrievedEvent{
		FilePaths: paths,
	}
	_ = hook.GlobalHookManager.OnScanFilePathsRetrieved().Trigger(fpEvent)
	paths = fpEvent.FilePaths

	// +---------------------+
	// |    Local files      |
	// +---------------------+

	localFiles := make([]*anime.LocalFile, 0)

	// Get skipped files depending on options
	skippedLfs := make(map[string]*anime.LocalFile)
	if (scn.SkipLockedFiles || scn.SkipIgnoredFiles) && scn.ExistingLocalFiles != nil {
		// Retrieve skipped files from existing local files
		for _, lf := range scn.ExistingLocalFiles {
			if scn.SkipLockedFiles && lf.IsLocked() {
				skippedLfs[lf.GetNormalizedPath()] = lf
			} else if scn.SkipIgnoredFiles && lf.IsIgnored() {
				skippedLfs[lf.GetNormalizedPath()] = lf
			}
		}
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 20)
	scn.WSEventManager.SendEvent(events.EventScanStatus, "Verifying shelved files...")

	// +---------------------+
	// |    Shelved files    |
	// +---------------------+

	scn.Logger.Debug().Int("count", len(scn.ExistingShelvedFiles)).Msg("scanner: Verifying shelved files")

	// Unshelve shelved files \/
	// Check for shelved files that are now present
	// If a shelved file is found, it is added to the skipped files list (so it's not rescanned)
	for _, shelvedLf := range scn.ExistingShelvedFiles {
		if filesystem.FileExists(shelvedLf.Path) {
			skippedLfs[shelvedLf.GetNormalizedPath()] = shelvedLf
		}
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 30)
	scn.WSEventManager.SendEvent(events.EventScanStatus, "Scanning local files...")

	// Create local files from paths (skipping skipped files)
	localFiles = lop.Map(paths, func(path string, _ int) *anime.LocalFile {
		if _, ok := skippedLfs[util.NormalizePath(path)]; !ok {
			// Create a new local file
			return anime.NewLocalFileS(path, libraryPaths)
		}

		return nil
	})

	// Remove nil values
	localFiles = lo.Filter(localFiles, func(lf *anime.LocalFile, _ int) bool {
		return lf != nil
	})

	// Invoke ScanLocalFilesParsed hook
	parsedEvent := &ScanLocalFilesParsedEvent{
		LocalFiles: localFiles,
	}
	_ = hook.GlobalHookManager.OnScanLocalFilesParsed().Trigger(parsedEvent)
	localFiles = parsedEvent.LocalFiles

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
				Str("filename", lf.Name).
				Interface("parsedData", lf.ParsedData).
				Interface("parsedFolderData", lf.ParsedFolderData).
				Msg("Parsed local file")
		}
	}

	if scn.ScanLogger != nil {
		scn.ScanLogger.logger.Debug().
			Msg("===========================================================================================================")
	}

	// DEVNOTE: Removed library path checking because it causes some issues with symlinks

	// +---------------------+
	// |  No files to scan   |
	// +---------------------+

	// If there are no local files to scan (all files are skipped, or a file was deleted)
	if len(localFiles) == 0 {
		scn.WSEventManager.SendEvent(events.EventScanProgress, 90)
		scn.WSEventManager.SendEvent(events.EventScanStatus, "Verifying file integrity...")

		scn.Logger.Debug().Int("skippedLfs", len(skippedLfs)).Msgf("scanner: Adding skipped local files")
		// Add skipped files
		if len(skippedLfs) > 0 {
			for _, sf := range skippedLfs {
				if filesystem.FileExists(sf.Path) { // Verify that the file still exists
					localFiles = append(localFiles, sf)
				} else if scn.WithShelving && sf.IsLocked() { // If the file is locked and shelving is enabled, shelve it
					scn.shelvedLocalFiles = append(scn.shelvedLocalFiles, sf)
				}
			}
		}

		// Add remaining shelved files
		scn.addRemainingShelvedFiles(skippedLfs, sortedLibraryPaths)

		scn.Logger.Debug().Msg("scanner: Scan completed")
		scn.WSEventManager.SendEvent(events.EventScanProgress, 100)
		scn.WSEventManager.SendEvent(events.EventScanStatus, "Scan completed")

		// Invoke ScanCompleted hook
		completedEvent := &ScanCompletedEvent{
			LocalFiles: localFiles,
			Duration:   int(time.Since(startTime).Milliseconds()),
		}
		hook.GlobalHookManager.OnScanCompleted().Trigger(completedEvent)
		localFiles = completedEvent.LocalFiles

		return localFiles, nil
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 40)
	if scn.Enhanced {
		scn.WSEventManager.SendEvent(events.EventScanStatus, "Fetching additional matching data...")
	} else {
		scn.WSEventManager.SendEvent(events.EventScanStatus, "Fetching media...")
	}

	// +---------------------+
	// |    MediaFetcher     |
	// +---------------------+

	// Fetch media needed for matching
	mf, err := NewMediaFetcher(ctx, &MediaFetcherOptions{
		Enhanced:                   scn.Enhanced,
		EnhanceWithOfflineDatabase: scn.EnhanceWithOfflineDatabase,
		PlatformRef:                scn.PlatformRef,
		MetadataProviderRef:        scn.MetadataProviderRef,
		LocalFiles:                 localFiles,
		CompleteAnimeCache:         completeAnimeCache,
		Logger:                     scn.Logger,
		AnilistRateLimiter:         anilistRateLimiter,
		DisableAnimeCollection:     false,
		ScanLogger:                 scn.ScanLogger,
		OptionalAnimeCollection:    scn.AnimeCollection,
	})
	if err != nil {
		return nil, err
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 50)
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
		LocalFiles:        localFiles,
		MediaContainer:    mc,
		Logger:            scn.Logger,
		ScanLogger:        scn.ScanLogger,
		ScanSummaryLogger: scn.ScanSummaryLogger,
		Algorithm:         scn.MatchingAlgorithm,
		Threshold:         scn.MatchingThreshold,
		UseLegacyMatching: scn.UseLegacyMatching,
		Config:            scn.Config,
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
		AllMedia:            mc.NormalizedMedia,
		LocalFiles:          localFiles,
		MetadataProviderRef: scn.MetadataProviderRef,
		PlatformRef:         scn.PlatformRef,
		CompleteAnimeCache:  completeAnimeCache,
		AnilistRateLimiter:  anilistRateLimiter,
		Logger:              scn.Logger,
		ScanLogger:          scn.ScanLogger,
		ScanSummaryLogger:   scn.ScanSummaryLogger,
		Config:              scn.Config,
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

		if err = scn.PlatformRef.Get().AddMediaToCollection(ctx, mf.UnknownMediaIds); err != nil {
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

	scn.Logger.Debug().Int("skippedLfs", len(skippedLfs)).Msgf("scanner: Adding skipped local files")

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
				} else if scn.WithShelving && skippedLf.IsLocked() { // If the file is locked and shelving is enabled, shelve it
					mu.Lock()
					scn.shelvedLocalFiles = append(scn.shelvedLocalFiles, skippedLf)
					mu.Unlock()
				}
			}(skippedLf)
		}
		wg.Wait()
	}

	// Add remaining shelved files
	scn.addRemainingShelvedFiles(skippedLfs, sortedLibraryPaths)

	scn.Logger.Info().Msg("scanner: Scan completed")
	scn.WSEventManager.SendEvent(events.EventScanProgress, 100)
	scn.WSEventManager.SendEvent(events.EventScanStatus, "Scan completed")

	if scn.ScanLogger != nil {
		scn.ScanLogger.logger.Info().
			Int("count", len(localFiles)).
			Int("unknownMediaCount", len(mf.UnknownMediaIds)).
			Msg("Scan completed")
	}

	// Invoke ScanCompleted hook
	completedEvent := &ScanCompletedEvent{
		LocalFiles: localFiles,
		Duration:   int(time.Since(startTime).Milliseconds()),
	}
	hook.GlobalHookManager.OnScanCompleted().Trigger(completedEvent)
	localFiles = completedEvent.LocalFiles

	runtime.GC()
	debug.FreeOSMemory()

	return localFiles, nil
}

// InLibrariesOnly removes files are not under the library paths
func (scn *Scanner) InLibrariesOnly(lfs []*anime.LocalFile) {

}

func (scn *Scanner) GetShelvedLocalFiles() []*anime.LocalFile {
	return scn.shelvedLocalFiles
}

func (scn *Scanner) addRemainingShelvedFiles(skippedLfs map[string]*anime.LocalFile, sortedLibraryPaths []string) {
	// If a shelved file was not unshelved, it should either:
	// be kept shelved or
	// removed (if its library path exists)

	libraryPathExistsCache := make(map[string]bool)

	for _, shelvedLf := range scn.ExistingShelvedFiles {
		// If not in skippedLfs (meaning it wasn't unshelved), keep it shelved or remove it
		if _, ok := skippedLfs[shelvedLf.GetNormalizedPath()]; !ok {

			// Check if we should really keep it shelved
			keepShelved := false

			// Find which library path this file belongs to
			var matchedLibPath string
			for _, libPath := range sortedLibraryPaths {
				if strings.HasPrefix(shelvedLf.GetNormalizedPath(), util.NormalizePath(libPath)) {
					matchedLibPath = libPath
					break
				}
			}

			if matchedLibPath != "" {
				exists, checked := libraryPathExistsCache[matchedLibPath]
				if !checked {
					_, err := os.Stat(matchedLibPath)
					exists = err == nil || !os.IsNotExist(err)
					libraryPathExistsCache[matchedLibPath] = exists
				}

				if !exists {
					// Library path doesn't exist (e.g. drive disconnected), so keep shelved
					keepShelved = true
				} else {
					// Library path exists, but file was not found, we assume it was deleted
					keepShelved = false
				}
			} else {
				// File doesn't belong to any known library path.
				// Meaning the library path was explicitly removed from the settings
				// default to removing it, doesn't hurt to scan again
				keepShelved = false
			}

			if keepShelved {
				scn.shelvedLocalFiles = append(scn.shelvedLocalFiles, shelvedLf)
			}
		}
	}
}

func ToConfig(c string) (*Config, error) {
	var ret Config
	err := json.Unmarshal([]byte(c), &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
