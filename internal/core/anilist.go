package core

import (
	"context"
	"errors"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/database/models"
)

// UpdateAnilistClientToken will update the Anilist Client Wrapper token.
// This function should be called when a user logs in
func (a *App) UpdateAnilistClientToken(token string) {
	a.AnilistClientWrapper = anilist.NewClientWrapper(token)
	a.PlaybackManager.SetAnilistClientWrapper(a.AnilistClientWrapper) // Update Anilist Client Wrapper in Playback Manager
}

// GetAnimeCollection returns the user's Anilist collection if it in the cache, otherwise it queries Anilist for the user's collection.
// When bypassCache is true, it will always query Anilist for the user's collection
func (a *App) GetAnimeCollection(bypassCache bool) (*anilist.AnimeCollection, error) {

	// Get Anilist Collection from App if it exists
	if !bypassCache && a.animeCollection != nil {
		return a.animeCollection, nil
	}

	return a.RefreshAnimeCollection()
}

// GetRawAnimeCollection is the same as GetAnimeCollection but returns the raw collection that includes custom lists
func (a *App) GetRawAnimeCollection(bypassCache bool) (*anilist.AnimeCollection, error) {

	// Get Anilist Collection from App if it exists
	if !bypassCache && a.rawAnimeCollection != nil {
		return a.rawAnimeCollection, nil
	}

	_, err := a.RefreshAnimeCollection()
	if err != nil {
		return nil, err
	}

	return a.rawAnimeCollection, nil
}

// RefreshAnimeCollection queries Anilist for the user's collection
func (a *App) RefreshAnimeCollection() (*anilist.AnimeCollection, error) {

	// If the account is nil, return false
	if a.account == nil {
		return nil, nil
	}

	// Else, get the collection from Anilist
	collection, err := a.AnilistClientWrapper.AnimeCollection(context.Background(), &a.account.Username)
	if err != nil {
		return nil, err
	}

	// Save the raw collection to App (retains the lists with no status)
	collectionCopy := *collection
	a.rawAnimeCollection = &collectionCopy
	listCollectionCopy := *collection.MediaListCollection
	a.rawAnimeCollection.MediaListCollection = &listCollectionCopy
	listsCopy := make([]*anilist.AnimeCollection_MediaListCollection_Lists, len(collection.MediaListCollection.Lists))
	copy(listsCopy, collection.MediaListCollection.Lists)
	a.rawAnimeCollection.MediaListCollection.Lists = listsCopy

	// Remove lists with no status (custom lists)
	collection.MediaListCollection.Lists = lo.Filter(collection.MediaListCollection.Lists, func(list *anilist.AnimeCollection_MediaListCollection_Lists, _ int) bool {
		return list.Status != nil
	})

	// Save the collection to App
	a.animeCollection = collection

	// Save the collection to AutoDownloader
	a.AutoDownloader.SetAnimeCollection(collection)

	// Save the collection to PlaybackManager
	a.PlaybackManager.SetAnimeCollection(collection)

	// Save the collection to TorrentstreamRepository
	a.TorrentstreamRepository.SetAnimeCollection(collection)

	return collection, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (a *App) GetAccount() (*models.Account, error) {

	if a.account == nil {
		return nil, nil
	}

	if a.account.Username == "" {
		return nil, errors.New("no username was found")
	}

	if a.account.Token == "" {
		return nil, errors.New("no token was found")
	}

	return a.account, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetMangaCollection is the same as GetAnimeCollection but for manga
func (a *App) GetMangaCollection(bypassCache bool) (*anilist.MangaCollection, error) {

	// Get Anilist Collection from App if it exists
	if !bypassCache && a.mangaCollection != nil {
		return a.mangaCollection, nil
	}

	return a.RefreshMangaCollection()

}

// GetRawMangaCollection does not exclude custom lists
func (a *App) GetRawMangaCollection(bypassCache bool) (*anilist.MangaCollection, error) {

	// Get Anilist Collection from App if it exists
	if !bypassCache && a.mangaCollection != nil {
		return a.mangaCollection, nil
	}

	_, err := a.RefreshMangaCollection()
	if err != nil {
		return nil, err
	}

	return a.rawMangaCollection, nil
}

// RefreshMangaCollection queries Anilist for the user's manga collection
func (a *App) RefreshMangaCollection() (*anilist.MangaCollection, error) {

	// If the account is nil, return false
	if a.account == nil {
		return nil, nil
	}

	// Else, get the collection from Anilist
	collection, err := a.AnilistClientWrapper.MangaCollection(context.Background(), &a.account.Username)
	if err != nil {
		return nil, err
	}

	collectionCopy := *collection
	a.rawMangaCollection = &collectionCopy
	listCollectionCopy := *collection.MediaListCollection
	a.rawMangaCollection.MediaListCollection = &listCollectionCopy
	listsCopy := make([]*anilist.MangaCollection_MediaListCollection_Lists, len(collection.MediaListCollection.Lists))
	copy(listsCopy, collection.MediaListCollection.Lists)
	a.rawMangaCollection.MediaListCollection.Lists = listsCopy

	// Remove lists with no status
	collection.MediaListCollection.Lists = lo.Filter(collection.MediaListCollection.Lists, func(list *anilist.MangaList, _ int) bool {
		return list.Status != nil
	})

	// Save the collection to App
	a.mangaCollection = collection

	return collection, nil
}
