package local_platform

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"seanime/internal/api/anilist"
	"seanime/internal/platforms/platform"
	"seanime/internal/sync"
)

var (
	ErrNoLocalAnimeCollection   = errors.New("no local anime collection")
	ErrorNoLocalMangaCollection = errors.New("no local manga collection")
	// ErrMediaNotFound means the media wasn't found in the local collection
	ErrMediaNotFound = errors.New("media not found")
	// ErrActionNotSupported means the action isn't valid on the local platform
	ErrActionNotSupported = errors.New("action not supported")
)

// LocalPlatform used when offline.
// It provides the same API as the anilist_platform.AnilistPlatform but some methods are no-op.
type LocalPlatform struct {
	logger      *zerolog.Logger
	syncManager sync.Manager
	client      anilist.AnilistClient
}

func NewLocalPlatform(syncManager sync.Manager, client anilist.AnilistClient, logger *zerolog.Logger) (platform.Platform, error) {
	ap := &LocalPlatform{
		logger:      logger,
		syncManager: syncManager,
		client:      client,
	}

	return ap, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (lp *LocalPlatform) SetUsername(username string) {
	// no-op
}

func (lp *LocalPlatform) SetAnilistClient(client anilist.AnilistClient) {
	// no-op
}

// UpdateEntry updates the entry for the given media ID.
// It doesn't add the entry if it doesn't exist.
func (lp *LocalPlatform) UpdateEntry(mediaID int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) error {
	if lp.syncManager.GetLocalAnimeCollection().IsPresent() {
		animeCollection := lp.syncManager.GetLocalAnimeCollection().MustGet()

		// Find the entry
		for _, list := range animeCollection.MediaListCollection.Lists {
			for _, entry := range list.Entries {
				if entry.Media.ID == mediaID {
					// Update the entry
					if status != nil {
						entry.Status = status
					}
					if scoreRaw != nil {
						entry.Score = lo.ToPtr(float64(*scoreRaw))
					}
					if progress != nil {
						entry.Progress = progress
					}
					if startedAt != nil {
						entry.StartedAt = &anilist.AnimeCollection_MediaListCollection_Lists_Entries_StartedAt{
							Year:  startedAt.Year,
							Month: startedAt.Month,
							Day:   startedAt.Day,
						}
					}
					if completedAt != nil {
						entry.CompletedAt = &anilist.AnimeCollection_MediaListCollection_Lists_Entries_CompletedAt{
							Year:  completedAt.Year,
							Month: completedAt.Month,
							Day:   completedAt.Day,
						}
					}

					// Save the collection
					lp.syncManager.SaveLocalAnimeCollection(animeCollection)
					return nil
				}
			}
		}
	}

	if lp.syncManager.GetLocalMangaCollection().IsPresent() {
		mangaCollection := lp.syncManager.GetLocalMangaCollection().MustGet()

		// Find the entry
		for _, list := range mangaCollection.MediaListCollection.Lists {
			for _, entry := range list.Entries {
				if entry.Media.ID == mediaID {
					// Update the entry
					if status != nil {
						entry.Status = status
					}
					if scoreRaw != nil {
						entry.Score = lo.ToPtr(float64(*scoreRaw))
					}
					if progress != nil {
						entry.Progress = progress
					}
					if startedAt != nil {
						entry.StartedAt = &anilist.MangaCollection_MediaListCollection_Lists_Entries_StartedAt{
							Year:  startedAt.Year,
							Month: startedAt.Month,
							Day:   startedAt.Day,
						}
					}
					if completedAt != nil {
						entry.CompletedAt = &anilist.MangaCollection_MediaListCollection_Lists_Entries_CompletedAt{
							Year:  completedAt.Year,
							Month: completedAt.Month,
							Day:   completedAt.Day,
						}
					}

					// Save the collection
					lp.syncManager.SaveLocalMangaCollection(mangaCollection)
					return nil
				}
			}
		}
	}

	return ErrMediaNotFound
}

func (lp *LocalPlatform) UpdateEntryProgress(mediaID int, progress int, totalEpisodes *int) error {
	if lp.syncManager.GetLocalAnimeCollection().IsPresent() {
		animeCollection := lp.syncManager.GetLocalAnimeCollection().MustGet()

		// Find the entry
		for _, list := range animeCollection.MediaListCollection.Lists {
			for _, entry := range list.Entries {
				if entry.Media.ID == mediaID {
					// Update the entry
					entry.Progress = &progress
					if totalEpisodes != nil {
						entry.Media.Episodes = totalEpisodes
					}

					// Save the collection
					lp.syncManager.SaveLocalAnimeCollection(animeCollection)
					return nil
				}
			}
		}
	}

	return ErrMediaNotFound
}

// DeleteEntry isn't supported for the local platform, always returns an error.
func (lp *LocalPlatform) DeleteEntry(mediaID int) error {
	return ErrActionNotSupported
}

func (lp *LocalPlatform) GetAnime(mediaID int) (*anilist.BaseAnime, error) {
	if lp.syncManager.GetLocalAnimeCollection().IsPresent() {
		animeCollection := lp.syncManager.GetLocalAnimeCollection().MustGet()

		// Find the entry
		for _, list := range animeCollection.MediaListCollection.Lists {
			for _, entry := range list.Entries {
				if entry.Media.ID == mediaID {
					return entry.Media, nil
				}
			}
		}
	}

	return nil, ErrMediaNotFound
}

func (lp *LocalPlatform) GetAnimeByMalID(malID int) (*anilist.BaseAnime, error) {
	if lp.syncManager.GetLocalAnimeCollection().IsPresent() {
		animeCollection := lp.syncManager.GetLocalAnimeCollection().MustGet()

		// Find the entry
		for _, list := range animeCollection.MediaListCollection.Lists {
			for _, entry := range list.Entries {
				if entry.Media.IDMal != nil && *entry.Media.IDMal == malID {
					return entry.Media, nil
				}
			}
		}
	}

	return nil, ErrMediaNotFound
}

// GetAnimeDetails isn't supported for the local platform, always returns an empty struct.
func (lp *LocalPlatform) GetAnimeDetails(mediaID int) (*anilist.AnimeDetailsById_Media, error) {
	return &anilist.AnimeDetailsById_Media{}, nil
}

// GetAnimeWithRelations isn't supported for the local platform, always returns an error.
func (lp *LocalPlatform) GetAnimeWithRelations(mediaID int) (*anilist.CompleteAnime, error) {
	return nil, ErrActionNotSupported
}

func (lp *LocalPlatform) GetManga(mediaID int) (*anilist.BaseManga, error) {
	if lp.syncManager.GetLocalMangaCollection().IsPresent() {
		mangaCollection := lp.syncManager.GetLocalMangaCollection().MustGet()

		// Find the entry
		for _, list := range mangaCollection.MediaListCollection.Lists {
			for _, entry := range list.Entries {
				if entry.Media.ID == mediaID {
					return entry.Media, nil
				}
			}
		}
	}

	return nil, ErrMediaNotFound
}

// GetMangaDetails isn't supported for the local platform, always returns an empty struct.
func (lp *LocalPlatform) GetMangaDetails(mediaID int) (*anilist.MangaDetailsById_Media, error) {
	return &anilist.MangaDetailsById_Media{}, nil
}

func (lp *LocalPlatform) GetAnimeCollection(bypassCache bool) (*anilist.AnimeCollection, error) {
	if lp.syncManager.GetLocalAnimeCollection().IsPresent() {
		return lp.syncManager.GetLocalAnimeCollection().MustGet(), nil
	} else {
		return nil, ErrNoLocalAnimeCollection
	}
}

func (lp *LocalPlatform) GetRawAnimeCollection(bypassCache bool) (*anilist.AnimeCollection, error) {
	if lp.syncManager.GetLocalAnimeCollection().IsPresent() {
		return lp.syncManager.GetLocalAnimeCollection().MustGet(), nil
	} else {
		return nil, ErrNoLocalAnimeCollection
	}
}

// RefreshAnimeCollection is a no-op, always returns the local anime collection.
func (lp *LocalPlatform) RefreshAnimeCollection() (*anilist.AnimeCollection, error) {
	animeCollection, ok := lp.syncManager.GetLocalAnimeCollection().Get()
	if !ok {
		return nil, ErrNoLocalAnimeCollection
	}

	return animeCollection, nil
}

func (lp *LocalPlatform) GetAnimeCollectionWithRelations() (*anilist.AnimeCollectionWithRelations, error) {
	return nil, ErrActionNotSupported
}

func (lp *LocalPlatform) GetMangaCollection(bypassCache bool) (*anilist.MangaCollection, error) {
	if lp.syncManager.GetLocalMangaCollection().IsPresent() {
		return lp.syncManager.GetLocalMangaCollection().MustGet(), nil
	} else {
		return nil, ErrorNoLocalMangaCollection
	}
}

func (lp *LocalPlatform) GetRawMangaCollection(bypassCache bool) (*anilist.MangaCollection, error) {
	if lp.syncManager.GetLocalMangaCollection().IsPresent() {
		return lp.syncManager.GetLocalMangaCollection().MustGet(), nil
	} else {
		return nil, ErrorNoLocalMangaCollection
	}
}

func (lp *LocalPlatform) RefreshMangaCollection() (*anilist.MangaCollection, error) {
	mangaCollection, ok := lp.syncManager.GetLocalMangaCollection().Get()
	if !ok {
		return nil, ErrorNoLocalMangaCollection
	}

	return mangaCollection, nil
}

// AddMediaToCollection isn't supported for the local platform, always returns an error.
func (lp *LocalPlatform) AddMediaToCollection(mIds []int) error {
	return ErrActionNotSupported
}

// GetStudioDetails isn't supported for the local platform, always returns an empty struct
func (lp *LocalPlatform) GetStudioDetails(studioID int) (*anilist.StudioDetails, error) {
	return &anilist.StudioDetails{}, nil
}

func (lp *LocalPlatform) GetAnilistClient() anilist.AnilistClient {
	return lp.client
}
