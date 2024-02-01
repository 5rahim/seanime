package scanner

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/filesystem"
	"github.com/seanime-app/seanime/internal/limiter"
	"time"
)

type Scanner struct {
	DirPath            string
	Username           string
	Enhanced           bool
	AnilistClient      *anilist.Client
	Logger             *zerolog.Logger
	WSEventManager     events.IWSEventManager
	ExistingLocalFiles []*entities.LocalFile
	SkipLockedFiles    bool
	SkipIgnoredFiles   bool
}

// Scan will scan the directory and return a list of entities.LocalFile.
func (scn *Scanner) Scan() ([]*entities.LocalFile, error) {

	baseMediaCache := anilist.NewBaseMediaCache()
	anizipCache := anizip.NewCache()
	anilistRateLimiter := limiter.NewAnilistLimiter()
	scanLogger, err := NewScanLogger()
	if err != nil {
		return nil, err
	}

	scn.Logger.Debug().Msg("scanner: Starting scan")
	scn.WSEventManager.SendEvent(events.EventScanProgress, 10)

	// +---------------------+
	// |     Local Files     |
	// +---------------------+

	// Get local files
	localFiles, err := GetLocalFilesFromDir(scn.DirPath, scn.Logger)
	if err != nil {
		return nil, err
	}

	scanLogger.logger.Info().
		Any("count", len(localFiles)).
		Msg("Retrieved and parsed local files")

	for _, lf := range localFiles {
		scanLogger.logger.Trace().
			Str("path", lf.Path).
			Any("parsedData", spew.Sdump(lf.ParsedData)).
			Any("parsedFolderData", spew.Sdump(lf.ParsedFolderData)).
			Msg("Parsed local file")
	}

	scanLogger.logger.Debug().
		Msg("===========================================================================================================")

	// +---------------------+
	// | Filter local files  |
	// +---------------------+

	// Get skipped files depending on options
	skippedLfs := make([]*entities.LocalFile, 0)
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
		localFiles = lo.Filter(localFiles, func(lf *entities.LocalFile, _ int) bool {
			if lf.IsIncluded(skippedLfs) {
				return false
			}
			return true
		})
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 20)

	// +---------------------+
	// |    MediaFetcher     |
	// +---------------------+

	// Fetch media needed for matching
	mf, err := NewMediaFetcher(&MediaFetcherOptions{
		Enhanced:           scn.Enhanced,
		Username:           scn.Username,
		AnilistClient:      scn.AnilistClient,
		LocalFiles:         localFiles,
		BaseMediaCache:     baseMediaCache,
		AnizipCache:        anizipCache,
		Logger:             scn.Logger,
		AnilistRateLimiter: anilistRateLimiter,
		ScanLogger:         scanLogger,
	})
	if err != nil {
		return nil, err
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 40)

	// +---------------------+
	// |   MediaContainer    |
	// +---------------------+

	// Create a new container for media
	mc := NewMediaContainer(&MediaContainerOptions{
		allMedia:   mf.AllMedia,
		ScanLogger: scanLogger,
	})

	scn.Logger.Debug().
		Any("count", len(mc.NormalizedMedia)).
		Msg("media container: Media container created")

	// +---------------------+
	// |      Matcher        |
	// +---------------------+

	// Create a new matcher
	matcher := &Matcher{
		localFiles:     localFiles,
		mediaContainer: mc,
		baseMediaCache: baseMediaCache,
		logger:         scn.Logger,
		ScanLogger:     scanLogger,
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 60)

	err = matcher.MatchLocalFilesWithMedia()
	if err != nil {
		return nil, err
	}

	// When enhancing is on, wait a minute before hydrating files
	// This is due to issues with rate limiting
	if scn.Enhanced {
		select {
		case <-time.After(time.Minute):
			break
		}
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 70)

	// +---------------------+
	// |    FileHydrator     |
	// +---------------------+

	// Create a new hydrator
	hydrator := &FileHydrator{
		AllMedia:           mc.NormalizedMedia,
		LocalFiles:         localFiles,
		AnizipCache:        anizipCache,
		AnilistClient:      scn.AnilistClient,
		BaseMediaCache:     baseMediaCache,
		AnilistRateLimiter: anilistRateLimiter,
		Logger:             scn.Logger,
		ScanLogger:         scanLogger,
	}
	hydrator.HydrateMetadata()

	scn.WSEventManager.SendEvent(events.EventScanProgress, 80)
	if scn.Enhanced {
		select {
		case <-time.After(time.Minute):
			break
		}
	}

	// +---------------------+
	// |  Add missing media  |
	// +---------------------+

	// Add non-added media entries to AniList collection
	// Max of 4 to avoid rate limit issues
	if len(mf.UnknownMediaIds) < 5 {
		if err = scn.AnilistClient.AddMediaToPlanning(mf.UnknownMediaIds, anilistRateLimiter, scn.Logger); err != nil {
			scn.Logger.Warn().Msg("scanner: An error occurred while adding media to planning list: " + err.Error())
		}
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 90)

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

	scn.Logger.Debug().Msg("scanner: Scan completed")
	scn.WSEventManager.SendEvent(events.EventScanProgress, 100)

	scanLogger.logger.Info().
		Int("scannedFileCount", len(localFiles)).
		Int("skippedFileCount", len(skippedLfs)).
		Int("unknownMediaCount", len(mf.UnknownMediaIds)).
		Msg("Scan completed")

	return localFiles, nil

}
