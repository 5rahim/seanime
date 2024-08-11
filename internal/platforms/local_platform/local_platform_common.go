package local_platform

import (
	"github.com/samber/lo"
	"github.com/samber/mo"
	"seanime/internal/api/anilist"
)

func (pm *LocalPlatform) getAnimeCollection(bypassCache bool) (ret *anilist.AnimeCollection, ok bool) {
	if !bypassCache {
		if ret, ok = pm.animeCollection.Get(); ok {
			return ret, true
		}
	}
	pm.loadAnimeCollection()
	return pm.animeCollection.Get()
}

func (pm *LocalPlatform) getRawAnimeCollection(bypassCache bool) (ret *anilist.AnimeCollection, ok bool) {
	if !bypassCache {
		if ret, ok = pm.rawAnimeCollection.Get(); ok {
			return ret, true
		}
	}
	pm.loadAnimeCollection()
	return pm.rawAnimeCollection.Get()
}

func (pm *LocalPlatform) getMangaCollection(bypassCache bool) (ret *anilist.MangaCollection, ok bool) {
	if !bypassCache {
		if ret, ok = pm.mangaCollection.Get(); ok {
			return ret, true
		}
	}
	pm.loadMangaCollection()
	return pm.mangaCollection.Get()
}

func (pm *LocalPlatform) getRawMangaCollection(bypassCache bool) (ret *anilist.MangaCollection, ok bool) {
	if !bypassCache {
		if ret, ok = pm.rawMangaCollection.Get(); ok {
			return ret, true
		}
	}
	pm.loadMangaCollection()
	return pm.rawMangaCollection.Get()
}

func (pm *LocalPlatform) loadAnimeCollection() {
	// Load the anime collection from the local database
	collection, ok := pm.localDb.getLocalAnimeCollection()
	if !ok {
		return
	}

	pm.animeMu.Lock()
	// Save the raw collection to App (retains the lists with no status)
	collectionCopy := *collection
	pm.rawAnimeCollection = mo.Some(&collectionCopy)
	listCollectionCopy := *collection.MediaListCollection
	pm.rawAnimeCollection.MustGet().MediaListCollection = &listCollectionCopy
	listsCopy := make([]*anilist.AnimeCollection_MediaListCollection_Lists, len(collection.MediaListCollection.Lists))
	copy(listsCopy, collection.MediaListCollection.Lists)
	pm.rawAnimeCollection.MustGet().MediaListCollection.Lists = listsCopy
	// Remove lists with no status (custom lists)
	collection.MediaListCollection.Lists = lo.Filter(collection.MediaListCollection.Lists, func(list *anilist.AnimeCollection_MediaListCollection_Lists, _ int) bool {
		return list.Status != nil
	})
	// Save the collection to App
	pm.animeCollection = mo.Some(collection)
	pm.animeMu.Unlock()
	return
}

func (pm *LocalPlatform) loadMangaCollection() {
	// Load the manga collection from the local database
	collection, ok := pm.localDb.getLocalMangaCollection()
	if !ok {
		return
	}
	pm.mangaMu.Lock()
	// Save the raw collection to App (retains the lists with no status)
	collectionCopy := *collection
	pm.rawMangaCollection = mo.Some(&collectionCopy)
	listCollectionCopy := *collection.MediaListCollection
	pm.rawMangaCollection.MustGet().MediaListCollection = &listCollectionCopy
	listsCopy := make([]*anilist.MangaCollection_MediaListCollection_Lists, len(collection.MediaListCollection.Lists))
	copy(listsCopy, collection.MediaListCollection.Lists)
	pm.rawMangaCollection.MustGet().MediaListCollection.Lists = listsCopy
	// Remove lists with no status (custom lists)
	collection.MediaListCollection.Lists = lo.Filter(collection.MediaListCollection.Lists, func(list *anilist.MangaCollection_MediaListCollection_Lists, _ int) bool {
		return list.Status != nil
	})
	// Save the collection to App
	pm.mangaCollection = mo.Some(collection)
	pm.mangaMu.Unlock()
	return
}

// UpdateLocalAnimeCollection updates the local anime collection with the current collection from Anilist.
func (pm *LocalPlatform) UpdateLocalAnimeCollection(current *anilist.AnimeCollection) error {
	err := pm.localDb.saveLocalCollection("anime", current)
	if err != nil {
		return err
	}
	go pm.loadAnimeCollection()
	return nil
}

// UpdateLocalMangaCollection updates the local manga collection with the current collection from Anilist.
func (pm *LocalPlatform) UpdateLocalMangaCollection(current *anilist.MangaCollection) error {
	err := pm.localDb.saveLocalCollection("manga", current)
	if err != nil {
		return err
	}
	go pm.loadMangaCollection()
	return nil
}
