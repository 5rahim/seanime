package scanner

import (
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/filesystem"
	"github.com/seanime-app/seanime/internal/limiter"
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
	normMediaCache := NewNormalizedMediaCache()
	anizipCache := anizip.NewCache()
	anilistRateLimiter := limiter.NewAnilistLimiter()

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
		Enhanced:             scn.Enhanced,
		Username:             scn.Username,
		AnilistClient:        scn.AnilistClient,
		LocalFiles:           localFiles,
		BaseMediaCache:       baseMediaCache,
		NormalizedMediaCache: normMediaCache,
		AnizipCache:          anizipCache,
		Logger:               scn.Logger,
		AnilistRateLimiter:   anilistRateLimiter,
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
		allMedia: mf.AllMedia,
	})

	scn.Logger.Trace().
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
	}

	scn.WSEventManager.SendEvent(events.EventScanProgress, 60)

	err = matcher.MatchLocalFilesWithMedia()
	if err != nil {
		return nil, err
	}

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
	}
	hydrator.HydrateMetadata()

	scn.WSEventManager.SendEvent(events.EventScanProgress, 90)

	// +---------------------+
	// |  Add missing media  |
	// +---------------------+

	// Add non-added media entries to AniList collection
	if err = scn.AnilistClient.AddMediaToPlanning(mf.UnknownMediaIds, anilistRateLimiter, scn.Logger); err != nil {
		scn.Logger.Warn().Msg("scanner: An error occurred while adding media to planning list: " + err.Error())
	}

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

	return localFiles, nil

}
