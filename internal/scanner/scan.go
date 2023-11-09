package scanner

import (
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	"github.com/seanime-app/seanime-server/internal/entities"
	"github.com/seanime-app/seanime-server/internal/events"
	"github.com/seanime-app/seanime-server/internal/filesystem"
	"github.com/seanime-app/seanime-server/internal/limiter"
)

type Scanner struct {
	DirPath            string
	Username           string
	Enhanced           bool
	AnilistClient      *anilist.Client
	Logger             *zerolog.Logger
	WSEventManager     *events.WSEventManager
	ExistingLocalFiles []*entities.LocalFile
	SkipLockedFiles    bool
	SkipIgnoredFiles   bool
}

func (scn *Scanner) Scan() ([]*entities.LocalFile, error) {

	baseMediaCache := anilist.NewBaseMediaCache()
	anizipCache := anizip.NewCache()
	anilistRateLimiter := limiter.NewAnilistLimiter()

	scn.Logger.Debug().Msg("scanner: Starting scan")
	scn.WSEventManager.SendEvent(events.EventScanProgress, 10)

	// Get local files
	localFiles, err := GetLocalFilesFromDir(scn.DirPath, scn.Logger)
	if err != nil {
		return nil, err
	}

	// TODO: If SkipLockedFiles or SkipIgnoredFiles is true, filter out skipped files (1)

	// Get skipped files depending on options
	skippedLfs := make([]*entities.LocalFile, 0)
	if (scn.SkipLockedFiles || scn.SkipIgnoredFiles) && scn.ExistingLocalFiles != nil {
		// Retrive skipped files from existing local files
		for _, lf := range scn.ExistingLocalFiles {
			if scn.SkipLockedFiles && lf.IsLocked() {
				skippedLfs = append(skippedLfs, lf)
			} else if scn.SkipIgnoredFiles && lf.IsIgnored() {
				skippedLfs = append(skippedLfs, lf)
			}
		}

		// Remove skipped files for local files that will be hydrated
		localFiles = lo.Filter(localFiles, func(lf *entities.LocalFile, _ int) bool {
			for _, sf := range skippedLfs {
				if lf.Path == sf.Path {
					return false
				}
			}
			return true
		})
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 20)

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
	})
	if err != nil {
		return nil, err
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 40)

	// Create a new container for media
	mc := NewMediaContainer(&MediaContainerOptions{
		allMedia: mf.AllMedia,
	})

	scn.Logger.Trace().
		Any("count", len(mc.allMedia)).
		Msg("media container: Media container created")

	// Create a new matcher
	matcher := &Matcher{
		localFiles:     localFiles,
		mediaContainer: mc,
		baseMediaCache: baseMediaCache,
		logger:         scn.Logger,
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 60)

	err = matcher.MatchLocalFilesWithMedia()
	if err != nil {
		return nil, err
	}

	// Create a new hydrator
	hydrator := &FileHydrator{
		media:              mc.allMedia,
		localFiles:         localFiles,
		anizipCache:        anizipCache,
		anilistClient:      scn.AnilistClient,
		baseMediaCache:     baseMediaCache,
		anilistRateLimiter: anilistRateLimiter,
		logger:             scn.Logger,
	}
	hydrator.HydrateMetadata()

	scn.WSEventManager.SendEvent(events.EventScanProgress, 90)

	// Add non-added media entries to AniList collection
	if err = scn.AnilistClient.AddMediaToPlanning(mf.UnknownMediaIds, anilistRateLimiter, scn.Logger); err != nil {
		scn.Logger.Warn().Msg("scanner: An error occurred while adding media to planning list: " + err.Error())
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 100)

	scn.Logger.Debug().Msg("scanner: Scan completed")

	// Merge skipped files with scanned files
	if len(skippedLfs) > 0 {
		for _, sf := range skippedLfs {
			if filesystem.FileExists(sf.Path) {
				localFiles = append(localFiles, sf)
			}
		}
	}

	return localFiles, nil

}
