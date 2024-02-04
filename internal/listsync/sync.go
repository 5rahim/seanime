package listsync

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/db"
	"github.com/seanime-app/seanime/internal/mal"
	"github.com/seanime-app/seanime/internal/result"
)

const (
	ErrSettingsNotSet            = "list sync settings not set"
	ErrOriginNotSet              = "list sync origin not set"
	ErrNotAuthenticatedToAnilist = "not authenticated to AniList"
	ErrMalAccountNotConnected    = "MAL account not connected"
)

type (
	Cache struct {
		*result.Cache[int, *ListSync]
	}
)

func NewCache() *Cache {
	return &Cache{result.NewCache[int, *ListSync]()}
}

func BuildListSync(db *db.Database, logger *zerolog.Logger) (*ListSync, error) {
	settings, err := db.GetSettings()
	if err != nil {
		return nil, err
	}

	if settings.ListSync == nil {
		return nil, errors.New(ErrSettingsNotSet)
	}

	origin := settings.ListSync.Origin
	if origin == "" {
		return nil, errors.New(ErrOriginNotSet)
	}

	// Anilist provider
	anilistProvider := &Provider{}
	account, err := db.GetAccount()
	if err != nil {
		return nil, err // AniList provider is required
	}
	if account.Token == "" {
		return nil, errors.New(ErrNotAuthenticatedToAnilist)
	}

	anilistClient := anilist.NewAuthedClient(account.Token)
	collection, err := anilistClient.AnimeCollection(context.Background(), &account.Username)
	if err != nil {
		return nil, err
	}
	anilistProvider = NewAnilistProvider(collection)

	// MAL provider
	malProvider := &Provider{}
	malProvider = nil
	malInfo, err := db.GetMalInfo()
	if err == nil && malInfo != nil {
		collection, err := mal.GetAnimeCollection(malInfo.AccessToken)
		if err == nil {
			malProvider = NewMALProvider(collection)
		}
	}

	ls := &ListSync{}

	targets := make([]*Provider, 0)

	providerRepo := &ProviderRepository{
		AnilistClient: anilistClient,
		MalToken:      malInfo.AccessToken,
		Logger:        logger,
	}

	switch origin {
	case "anilist":
		if malProvider != nil {
			targets = append(targets, malProvider)
		}
		// ... Add more providers here
		ls = NewListSync(anilistProvider, targets, providerRepo)
	case "mal":
		if malProvider == nil {
			return nil, errors.New(ErrMalAccountNotConnected)
		}
		targets = append(targets, anilistProvider)
		// ... Add more providers here
		ls = NewListSync(malProvider, targets, providerRepo)
	}

	return ls, nil
}

// NewListSync creates a new list sync
func NewListSync(origin *Provider, targets []*Provider, providerRepo *ProviderRepository) *ListSync {
	ls := &ListSync{
		Origin:             origin,
		Targets:            targets,
		ProviderRepository: providerRepo,
	}

	ls.AnimeDiffs = ls.getAnimeDiffs()

	return ls
}

func (ls *ListSync) getAnimeDiffs() []*AnimeDiff {
	diff := make([]*AnimeDiff, 0)

	for _, target := range ls.Targets {
		// First, check for missing anime in the target
		missing, ok := checkMissingFrom(ls.Origin, target)
		if ok {
			for _, entry := range missing.OriginEntries {
				diff = append(diff, &AnimeDiff{
					ID:                uuid.NewString(),
					TargetSource:      target.Source,
					OriginEntry:       entry,
					TargetEntry:       nil,
					Kind:              AnimeDiffKindMissingTarget,
					MetadataDiffKinds: make([]AnimeMetadataDiffKind, 0),
				})
			}
		}

		// Then, check for missing anime in the origin
		missing, ok = checkMissingFrom(target, ls.Origin)
		if ok {
			for _, entry := range missing.OriginEntries {
				diff = append(diff, &AnimeDiff{
					ID:                uuid.NewString(),
					TargetSource:      target.Source,
					OriginEntry:       nil,
					TargetEntry:       entry,
					Kind:              AnimeDiffKindMissingOrigin,
					MetadataDiffKinds: make([]AnimeMetadataDiffKind, 0),
				})
			}
		}

		// Finally, check for different metadata
		for _, entry := range ls.Origin.AnimeEntries {
			if targetEntry, ok := target.AnimeEntriesMap[entry.MalID]; ok {
				diffs, found := entry.FindMetadataDiffs(targetEntry)
				if found {
					diff = append(diff, &AnimeDiff{
						ID:                uuid.NewString(),
						TargetSource:      target.Source,
						OriginEntry:       entry,
						TargetEntry:       targetEntry,
						Kind:              AnimeDiffKindMetadata,
						MetadataDiffKinds: diffs,
					})
				}
			}
		}
	}

	return diff
}

// CheckMissingFrom checks for anime that are present in the origin but not in the target.
func checkMissingFrom(origin *Provider, target *Provider) (*MissingAnime, bool) {
	missing := make([]*AnimeEntry, 0)

	for _, entry := range origin.AnimeEntries {
		if _, ok := target.AnimeEntriesMap[entry.MalID]; !ok {
			missing = append(missing, entry)
		}
	}

	if len(missing) == 0 {
		return nil, false
	}

	return &MissingAnime{
		Provider:      target,
		OriginEntries: missing,
	}, true
}

// SyncAnime syncs anime between the origin and targets
func (ls *ListSync) SyncAnime(diff *AnimeDiff) error {
	var err error
	switch diff.Kind {
	case AnimeDiffKindMissingTarget:
		err = ls.ProviderRepository.AddAnime(diff.TargetSource, diff.OriginEntry)
	case AnimeDiffKindMissingOrigin:
		err = ls.ProviderRepository.DeleteAnime(diff.TargetSource, diff.TargetEntry)
	case AnimeDiffKindMetadata:
		err = ls.ProviderRepository.UpdateAnime(diff.TargetSource, diff.OriginEntry)
	}

	if err != nil {
		return err
	}

	// Remove the diff
	ls.AnimeDiffs = lo.Filter(ls.AnimeDiffs, func(ad *AnimeDiff, _ int) bool {
		return ad.ID != diff.ID
	})

	return nil
}
