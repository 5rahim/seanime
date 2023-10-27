package scanner

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	"github.com/seanime-app/seanime-server/internal/db"
)

type Scanner struct {
	Token         string
	DirPath       string
	Username      string
	Enhanced      bool
	AnilistClient *anilist.Client
	Logger        *zerolog.Logger
	DB            *db.Database
}

func (scn Scanner) Scan() (any, error) {

	baseMediaCache := anilist.NewBaseMediaCache()
	anizipCache := anizip.NewCache()

	// Get local files
	localFiles, err := GetLocalFilesFromDir(scn.DirPath, scn.Logger)
	if err != nil {
		return nil, err
	}

	// Fetch media needed for matching
	mf, err := NewMediaFetcher(&MediaFetcherOptions{
		Enhanced:       scn.Enhanced,
		Username:       scn.Username,
		AnilistClient:  scn.AnilistClient,
		LocalFiles:     localFiles,
		BaseMediaCache: baseMediaCache,
		AnizipCache:    anizipCache,
		Logger:         scn.Logger,
	})
	if err != nil {
		return nil, err
	}

	// Create a new container for media
	mc := NewMediaContainer(&MediaContainerOptions{
		allMedia: mf.AllMedia,
	})

	// Create a new matcher
	matcher := NewMatcher(&MatcherOptions{
		localFiles:     localFiles,
		mediaContainer: mc,
		baseMediaCache: baseMediaCache,
	})

	err = matcher.MatchLocalFilesWithMedia()
	if err != nil {
		return nil, err
	}

	// Create a new hydrator
	hydrator := &FileHydrator{
		localFiles:     localFiles,
		media:          mc.allMedia,
		baseMediaCache: baseMediaCache,
		anizipCache:    anizipCache,
	}
	hydrator.HydrateMetadata()

	return localFiles, nil

}
