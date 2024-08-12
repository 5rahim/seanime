package anilist_platform

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"seanime/internal/api/anilist"
	"seanime/internal/platforms/platform"
	"seanime/internal/util/limiter"
	"sync"
	"time"
)

type (
	AnilistPlatform struct {
		logger               *zerolog.Logger
		username             mo.Option[string]
		anilistClient        anilist.AnilistClient
		animeCollection      mo.Option[*anilist.AnimeCollection]
		rawAnimeCollection   mo.Option[*anilist.AnimeCollection]
		mangaCollection      mo.Option[*anilist.MangaCollection]
		rawMangaCollection   mo.Option[*anilist.MangaCollection]
		isOffline            bool
		localPlatformEnabled bool
	}
)

func NewAnilistPlatform(anilistClient anilist.AnilistClient, logger *zerolog.Logger) platform.Platform {
	ap := &AnilistPlatform{
		anilistClient:      anilistClient,
		logger:             logger,
		username:           mo.None[string](),
		animeCollection:    mo.None[*anilist.AnimeCollection](),
		rawAnimeCollection: mo.None[*anilist.AnimeCollection](),
		mangaCollection:    mo.None[*anilist.MangaCollection](),
		rawMangaCollection: mo.None[*anilist.MangaCollection](),
	}

	return ap
}

func (ap *AnilistPlatform) SetUsername(username string) {
	// Set the username for the AnilistPlatform
	if username == "" {
		ap.username = mo.Some[string]("")
		return
	}

	ap.username = mo.Some(username)
	return
}

func (ap *AnilistPlatform) SetAnilistClient(client anilist.AnilistClient) {
	// Set the AnilistClient for the AnilistPlatform
	ap.anilistClient = client
}

func (ap *AnilistPlatform) UpdateEntry(mediaID int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) error {
	ap.logger.Trace().Msg("anilist platform: Updating entry")
	_, err := ap.anilistClient.UpdateMediaListEntry(context.Background(), &mediaID, status, scoreRaw, progress, startedAt, completedAt)
	if err != nil {
		return err
	}
	return nil
}

func (ap *AnilistPlatform) UpdateEntryProgress(mediaID int, progress int, totalEpisodes *int) error {
	ap.logger.Trace().Msg("anilist platform: Updating entry progress")

	totalEp := 0
	if totalEpisodes != nil && *totalEpisodes > 0 {
		totalEp = *totalEpisodes
	}

	status := anilist.MediaListStatusCurrent
	// Check if the anime is in the repeating list
	// If it is, set the status to repeating
	if ap.rawAnimeCollection.IsPresent() {
		for _, list := range ap.rawAnimeCollection.MustGet().MediaListCollection.Lists {
			if list.Status != nil && *list.Status == anilist.MediaListStatusRepeating {
				if list.Entries != nil {
					for _, entry := range list.Entries {
						if entry.GetMedia().GetID() == mediaID {
							status = anilist.MediaListStatusRepeating
							break
						}
					}
				}
			}
		}
	}
	if totalEp > 0 && progress >= totalEp {
		status = anilist.MediaListStatusCompleted
	}

	if totalEp > 0 && progress > totalEp {
		progress = totalEp
	}

	_, err := ap.anilistClient.UpdateMediaListEntryProgress(
		context.Background(),
		&mediaID,
		&progress,
		&status,
	)
	if err != nil {
		return err
	}

	return nil
}

func (ap *AnilistPlatform) DeleteEntry(mediaID int) error {
	ap.logger.Trace().Msg("anilist platform: Deleting entry")
	_, err := ap.anilistClient.DeleteEntry(context.Background(), &mediaID)
	if err != nil {
		return err
	}
	return nil
}

func (ap *AnilistPlatform) GetAnime(mediaID int) (*anilist.BaseAnime, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching anime")
	ret, err := ap.anilistClient.BaseAnimeByID(context.Background(), &mediaID)
	if err != nil {
		return nil, err
	}
	return ret.GetMedia(), nil
}

func (ap *AnilistPlatform) GetAnimeByMalID(malID int) (*anilist.BaseAnime, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching anime by MAL ID")
	ret, err := ap.anilistClient.BaseAnimeByMalID(context.Background(), &malID)
	if err != nil {
		return nil, err
	}
	return ret.GetMedia(), nil
}

func (ap *AnilistPlatform) GetAnimeDetails(mediaID int) (*anilist.AnimeDetailsById_Media, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching anime details")
	ret, err := ap.anilistClient.AnimeDetailsByID(context.Background(), &mediaID)
	if err != nil {
		return nil, err
	}
	return ret.GetMedia(), nil
}

func (ap *AnilistPlatform) GetAnimeWithRelations(mediaID int) (*anilist.CompleteAnime, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching anime with relations")
	ret, err := ap.anilistClient.CompleteAnimeByID(context.Background(), &mediaID)
	if err != nil {
		return nil, err
	}
	return ret.GetMedia(), nil
}

func (ap *AnilistPlatform) GetManga(mediaID int) (*anilist.BaseManga, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching manga")
	ret, err := ap.anilistClient.BaseMangaByID(context.Background(), &mediaID)
	if err != nil {
		return nil, err
	}
	return ret.GetMedia(), nil
}

func (ap *AnilistPlatform) GetMangaDetails(mediaID int) (*anilist.MangaDetailsById_Media, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching manga details")
	ret, err := ap.anilistClient.MangaDetailsByID(context.Background(), &mediaID)
	if err != nil {
		return nil, err
	}
	return ret.GetMedia(), nil
}

func (ap *AnilistPlatform) GetAnimeCollection(bypassCache bool) (*anilist.AnimeCollection, error) {
	if !bypassCache && ap.animeCollection.IsPresent() {
		return ap.animeCollection.MustGet(), nil
	}

	if ap.username.IsAbsent() {
		return nil, nil
	}

	err := ap.refreshAnimeCollection()
	if err != nil {
		return nil, err
	}

	return ap.animeCollection.MustGet(), nil
}

func (ap *AnilistPlatform) GetRawAnimeCollection(bypassCache bool) (*anilist.AnimeCollection, error) {
	if !bypassCache && ap.rawAnimeCollection.IsPresent() {
		return ap.rawAnimeCollection.MustGet(), nil
	}

	if ap.username.IsAbsent() {
		return nil, nil
	}

	err := ap.refreshAnimeCollection()
	if err != nil {
		return nil, err
	}

	return ap.rawAnimeCollection.MustGet(), nil
}

func (ap *AnilistPlatform) RefreshAnimeCollection() (*anilist.AnimeCollection, error) {
	if ap.username.IsAbsent() {
		return nil, nil
	}

	err := ap.refreshAnimeCollection()
	if err != nil {
		return nil, err
	}

	return ap.animeCollection.MustGet(), nil
}

func (ap *AnilistPlatform) refreshAnimeCollection() error {
	if ap.username.IsAbsent() {
		return errors.New("anilist: Username is not set")
	}

	// Else, get the collection from Anilist
	collection, err := ap.anilistClient.AnimeCollection(context.Background(), ap.username.ToPointer())
	if err != nil {
		return err
	}

	// Save the raw collection to App (retains the lists with no status)
	collectionCopy := *collection
	ap.rawAnimeCollection = mo.Some(&collectionCopy)
	listCollectionCopy := *collection.MediaListCollection
	ap.rawAnimeCollection.MustGet().MediaListCollection = &listCollectionCopy
	listsCopy := make([]*anilist.AnimeCollection_MediaListCollection_Lists, len(collection.MediaListCollection.Lists))
	copy(listsCopy, collection.MediaListCollection.Lists)
	ap.rawAnimeCollection.MustGet().MediaListCollection.Lists = listsCopy

	// Remove lists with no status (custom lists)
	collection.MediaListCollection.Lists = lo.Filter(collection.MediaListCollection.Lists, func(list *anilist.AnimeCollection_MediaListCollection_Lists, _ int) bool {
		return list.Status != nil
	})

	// Save the collection to App
	ap.animeCollection = mo.Some(collection)

	return nil
}

func (ap *AnilistPlatform) GetAnimeCollectionWithRelations() (*anilist.AnimeCollectionWithRelations, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching anime collection with relations")

	if ap.username.IsAbsent() {
		return nil, nil
	}

	ret, err := ap.anilistClient.AnimeCollectionWithRelations(context.Background(), ap.username.ToPointer())
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (ap *AnilistPlatform) GetMangaCollection(bypassCache bool) (*anilist.MangaCollection, error) {

	if !bypassCache && ap.mangaCollection.IsPresent() {
		return ap.mangaCollection.MustGet(), nil
	}

	if ap.username.IsAbsent() {
		return nil, nil
	}

	err := ap.refreshMangaCollection()
	if err != nil {
		return nil, err
	}

	return ap.mangaCollection.MustGet(), nil
}

func (ap *AnilistPlatform) GetRawMangaCollection(bypassCache bool) (*anilist.MangaCollection, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching raw manga collection")

	if !bypassCache && ap.rawMangaCollection.IsPresent() {
		ap.logger.Trace().Msg("anilist platform: Returning raw manga collection from cache")
		return ap.rawMangaCollection.MustGet(), nil
	}

	if ap.username.IsAbsent() {
		return nil, nil
	}

	err := ap.refreshMangaCollection()
	if err != nil {
		return nil, err
	}

	return ap.rawMangaCollection.MustGet(), nil
}

func (ap *AnilistPlatform) RefreshMangaCollection() (*anilist.MangaCollection, error) {
	if ap.username.IsAbsent() {
		return nil, nil
	}

	err := ap.refreshMangaCollection()
	if err != nil {
		return nil, err
	}

	return ap.mangaCollection.MustGet(), nil
}

func (ap *AnilistPlatform) refreshMangaCollection() error {
	if ap.username.IsAbsent() {
		return errors.New("anilist: Username is not set")
	}

	collection, err := ap.anilistClient.MangaCollection(context.Background(), ap.username.ToPointer())
	if err != nil {
		return err
	}

	// Save the raw collection to App (retains the lists with no status)
	collectionCopy := *collection
	ap.rawMangaCollection = mo.Some(&collectionCopy)
	listCollectionCopy := *collection.MediaListCollection
	ap.rawMangaCollection.MustGet().MediaListCollection = &listCollectionCopy
	listsCopy := make([]*anilist.MangaCollection_MediaListCollection_Lists, len(collection.MediaListCollection.Lists))
	copy(listsCopy, collection.MediaListCollection.Lists)
	ap.rawMangaCollection.MustGet().MediaListCollection.Lists = listsCopy

	// Remove lists with no status (custom lists)
	collection.MediaListCollection.Lists = lo.Filter(collection.MediaListCollection.Lists, func(list *anilist.MangaCollection_MediaListCollection_Lists, _ int) bool {
		return list.Status != nil
	})

	// Save the collection to App
	ap.mangaCollection = mo.Some(collection)

	return nil
}

func (ap *AnilistPlatform) AddMediaToCollection(mIds []int) error {
	ap.logger.Trace().Msg("anilist platform: Adding media to collection")
	if len(mIds) == 0 {
		ap.logger.Debug().Msg("anilist: No media added to planning list")
		return nil
	}

	rateLimiter := limiter.NewLimiter(1*time.Second, 1) // 1 request per second

	wg := sync.WaitGroup{}
	for _, _id := range mIds {
		wg.Add(1)
		go func(id int) {
			rateLimiter.Wait()
			defer wg.Done()
			_, err := ap.anilistClient.UpdateMediaListEntry(
				context.Background(),
				&id,
				lo.ToPtr(anilist.MediaListStatusPlanning),
				lo.ToPtr(0),
				lo.ToPtr(0),
				nil,
				nil,
			)
			if err != nil {
				ap.logger.Error().Msg("anilist: An error occurred while adding media to planning list: " + err.Error())
			}
		}(_id)
	}
	wg.Wait()

	ap.logger.Debug().Any("count", len(mIds)).Msg("anilist: Media added to planning list")
	return nil
}

func (ap *AnilistPlatform) GetStudioDetails(studioID int) (*anilist.StudioDetails, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching studio details")
	ret, err := ap.anilistClient.StudioDetails(context.Background(), &studioID)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (ap *AnilistPlatform) GetAnilistClient() anilist.AnilistClient {
	return ap.anilistClient
}
