package scanner

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	"github.com/seanime-app/seanime-server/internal/entities"
	"github.com/seanime-app/seanime-server/internal/limiter"
)

type Scanner struct {
	Token         string
	DirPath       string
	Username      string
	Enhanced      bool
	AnilistClient *anilist.Client
	Logger        *zerolog.Logger
}

func (scn *Scanner) Scan() ([]*entities.LocalFile, error) {

	baseMediaCache := anilist.NewBaseMediaCache()
	anizipCache := anizip.NewCache()
	anilistRateLimiter := limiter.NewAnilistLimiter()

	scn.Logger.Debug().Msg("scanner: Starting scan")

	// Get local files
	localFiles, err := GetLocalFilesFromDir(scn.DirPath, scn.Logger)
	if err != nil {
		return nil, err
	}

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

	// Add non-added media entries to AniList collection
	if err = scn.AnilistClient.AddMediaToPlanning(mf.UnknownMediaIds, anilistRateLimiter, scn.Logger); err != nil {
		scn.Logger.Warn().Msg("scanner: An error occurred while adding media to planning list: " + err.Error())
	}

	scn.Logger.Debug().Msg("scanner: Scan completed")

	return localFiles, nil

}
