package core

import (
	"errors"
	"seanime/internal/api/anilist"
	"seanime/internal/database/models"
)

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

func (a *App) GetAccountToken() string {
	if a.account == nil {
		return ""
	}

	return a.account.Token
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// UpdateAnilistClientToken will update the Anilist Client Wrapper token.
// This function should be called when a user logs in
func (a *App) UpdateAnilistClientToken(token string) {
	a.AnilistClient = anilist.NewAnilistClient(token)
	a.AnilistPlatform.SetAnilistClient(a.AnilistClient) // Update Anilist Client Wrapper in Platform
}

// GetAnimeCollection returns the user's Anilist collection if it in the cache, otherwise it queries Anilist for the user's collection.
// When bypassCache is true, it will always query Anilist for the user's collection
func (a *App) GetAnimeCollection(bypassCache bool) (*anilist.AnimeCollection, error) {
	return a.AnilistPlatform.GetAnimeCollection(bypassCache)
}

// GetRawAnimeCollection is the same as GetAnimeCollection but returns the raw collection that includes custom lists
func (a *App) GetRawAnimeCollection(bypassCache bool) (*anilist.AnimeCollection, error) {
	return a.AnilistPlatform.GetRawAnimeCollection(bypassCache)
}

// RefreshAnimeCollection queries Anilist for the user's collection
func (a *App) RefreshAnimeCollection() (*anilist.AnimeCollection, error) {
	ret, err := a.AnilistPlatform.RefreshAnimeCollection()

	if err != nil {
		return nil, err
	}

	// Save the collection to PlaybackManager
	a.PlaybackManager.SetAnimeCollection(ret)

	// Save the collection to AutoDownloader
	a.AutoDownloader.SetAnimeCollection(ret)

	a.SyncManager.SetAnimeCollection(ret)

	return ret, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetMangaCollection is the same as GetAnimeCollection but for manga
func (a *App) GetMangaCollection(bypassCache bool) (*anilist.MangaCollection, error) {
	return a.AnilistPlatform.GetMangaCollection(bypassCache)
}

// GetRawMangaCollection does not exclude custom lists
func (a *App) GetRawMangaCollection(bypassCache bool) (*anilist.MangaCollection, error) {
	return a.AnilistPlatform.GetRawMangaCollection(bypassCache)
}

// RefreshMangaCollection queries Anilist for the user's manga collection
func (a *App) RefreshMangaCollection() (*anilist.MangaCollection, error) {
	mc, err := a.AnilistPlatform.RefreshMangaCollection()

	if err != nil {
		return nil, err
	}

	a.SyncManager.SetMangaCollection(mc)

	return mc, nil
}
