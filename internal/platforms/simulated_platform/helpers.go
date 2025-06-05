package simulated_platform

import (
	"context"
	"errors"
	"seanime/internal/api/anilist"
	"time"

	"github.com/samber/lo"
)

// CollectionWrapper provides an ambivalent interface for anime and manga collections
type CollectionWrapper struct {
	platform *SimulatedPlatform
	isAnime  bool
}

func (sp *SimulatedPlatform) GetAnimeCollectionWrapper() *CollectionWrapper {
	return &CollectionWrapper{platform: sp, isAnime: true}
}

func (sp *SimulatedPlatform) GetMangaCollectionWrapper() *CollectionWrapper {
	return &CollectionWrapper{platform: sp, isAnime: false}
}

// AddEntry adds a new entry to the collection
func (cw *CollectionWrapper) AddEntry(mediaId int, status anilist.MediaListStatus) error {
	if cw.isAnime {
		return cw.addAnimeEntry(mediaId, status)
	}
	return cw.addMangaEntry(mediaId, status)
}

// UpdateEntry updates an existing entry in the collection
func (cw *CollectionWrapper) UpdateEntry(mediaId int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) error {
	if cw.isAnime {
		return cw.updateAnimeEntry(mediaId, status, scoreRaw, progress, startedAt, completedAt)
	}
	return cw.updateMangaEntry(mediaId, status, scoreRaw, progress, startedAt, completedAt)
}

// UpdateEntryProgress updates the progress of an entry
func (cw *CollectionWrapper) UpdateEntryProgress(mediaId int, progress int, totalCount *int) error {
	status := anilist.MediaListStatusCurrent
	if totalCount != nil && progress >= *totalCount {
		status = anilist.MediaListStatusCompleted
	}

	return cw.UpdateEntry(mediaId, &status, nil, &progress, nil, nil)
}

// DeleteEntry removes an entry from the collection
func (cw *CollectionWrapper) DeleteEntry(mediaId int, isEntryId ...bool) error {
	if cw.isAnime {
		return cw.deleteAnimeEntry(mediaId, isEntryId...)
	}
	return cw.deleteMangaEntry(mediaId, isEntryId...)
}

// FindEntry finds an entry by media ID
func (cw *CollectionWrapper) FindEntry(mediaId int, isEntryId ...bool) (interface{}, error) {
	if cw.isAnime {
		return cw.findAnimeEntry(mediaId, isEntryId...)
	}
	return cw.findMangaEntry(mediaId, isEntryId...)
}

// UpdateMediaData updates the media data for an entry
func (cw *CollectionWrapper) UpdateMediaData(mediaId int, mediaData interface{}) error {
	if cw.isAnime {
		if baseAnime, ok := mediaData.(*anilist.BaseAnime); ok {
			return cw.updateAnimeMediaData(mediaId, baseAnime)
		}
		return errors.New("invalid anime data type")
	}

	if baseManga, ok := mediaData.(*anilist.BaseManga); ok {
		return cw.updateMangaMediaData(mediaId, baseManga)
	}
	return errors.New("invalid manga data type")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Anime Collection Helper Methods
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (cw *CollectionWrapper) addAnimeEntry(mediaId int, status anilist.MediaListStatus) error {
	collection, err := cw.platform.getOrCreateAnimeCollection()
	if err != nil {
		return err
	}

	// Check if entry already exists
	if _, err := cw.findAnimeEntry(mediaId); err == nil {
		return errors.New("entry already exists")
	}

	// Fetch media data
	mediaResp, err := cw.platform.client.BaseAnimeByID(context.Background(), &mediaId)
	if err != nil {
		return err
	}

	// Find or create the appropriate list
	var targetList *anilist.AnimeCollection_MediaListCollection_Lists
	for _, list := range collection.GetMediaListCollection().GetLists() {
		if list.GetStatus() != nil && *list.GetStatus() == status {
			targetList = list
			break
		}
	}

	if targetList == nil {
		// Create new list
		targetList = &anilist.AnimeCollection_MediaListCollection_Lists{
			Status:       &status,
			Name:         lo.ToPtr(string(status)),
			IsCustomList: lo.ToPtr(false),
			Entries:      []*anilist.AnimeCollection_MediaListCollection_Lists_Entries{},
		}
		collection.GetMediaListCollection().Lists = append(collection.GetMediaListCollection().Lists, targetList)
	}

	// Create new entry
	newEntry := &anilist.AnimeCollection_MediaListCollection_Lists_Entries{
		ID:          int(time.Now().UnixNano()), // Generate unique ID
		Status:      &status,
		Progress:    lo.ToPtr(0),
		Media:       mediaResp.GetMedia(),
		Score:       lo.ToPtr(0.0),
		Notes:       nil,
		Repeat:      lo.ToPtr(0),
		Private:     lo.ToPtr(false),
		StartedAt:   &anilist.AnimeCollection_MediaListCollection_Lists_Entries_StartedAt{},
		CompletedAt: &anilist.AnimeCollection_MediaListCollection_Lists_Entries_CompletedAt{},
	}

	targetList.Entries = append(targetList.Entries, newEntry)

	// Save collection
	cw.platform.localManager.SaveSimulatedAnimeCollection(collection)
	return nil
}

func (cw *CollectionWrapper) updateAnimeEntry(mediaId int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) error {
	collection, err := cw.platform.getOrCreateAnimeCollection()
	if err != nil {
		return err
	}

	var foundEntry *anilist.AnimeCollection_MediaListCollection_Lists_Entries
	var sourceList *anilist.AnimeCollection_MediaListCollection_Lists
	var entryIndex int

	// Find the entry
	for _, list := range collection.GetMediaListCollection().GetLists() {
		for i, entry := range list.GetEntries() {
			if entry.GetMedia().GetID() == mediaId {
				foundEntry = entry
				sourceList = list
				entryIndex = i
				break
			}
		}
		if foundEntry != nil {
			break
		}
	}

	if foundEntry == nil || sourceList == nil {
		return ErrMediaNotFound
	}

	// Update entry fields
	if progress != nil {
		foundEntry.Progress = progress
	}
	if scoreRaw != nil {
		foundEntry.Score = lo.ToPtr(float64(*scoreRaw))
	}
	if startedAt != nil {
		foundEntry.StartedAt = &anilist.AnimeCollection_MediaListCollection_Lists_Entries_StartedAt{
			Year:  startedAt.Year,
			Month: startedAt.Month,
			Day:   startedAt.Day,
		}
	}
	if completedAt != nil {
		foundEntry.CompletedAt = &anilist.AnimeCollection_MediaListCollection_Lists_Entries_CompletedAt{
			Year:  completedAt.Year,
			Month: completedAt.Month,
			Day:   completedAt.Day,
		}
	}

	// If status changed, move entry to different list
	if status != nil && foundEntry.GetStatus() != nil && *status != *foundEntry.GetStatus() {
		foundEntry.Status = status

		// Remove from current list
		sourceList.Entries = append(sourceList.Entries[:entryIndex], sourceList.Entries[entryIndex+1:]...)

		// Find or create target list
		var targetList *anilist.AnimeCollection_MediaListCollection_Lists
		for _, list := range collection.GetMediaListCollection().GetLists() {
			if list.GetStatus() != nil && *list.GetStatus() == *status {
				targetList = list
				break
			}
		}

		if targetList == nil {
			targetList = &anilist.AnimeCollection_MediaListCollection_Lists{
				Status:       status,
				Name:         lo.ToPtr(string(*status)),
				IsCustomList: lo.ToPtr(false),
				Entries:      []*anilist.AnimeCollection_MediaListCollection_Lists_Entries{},
			}
			collection.GetMediaListCollection().Lists = append(collection.GetMediaListCollection().Lists, targetList)
		}

		targetList.Entries = append(targetList.Entries, foundEntry)
	}

	cw.platform.localManager.SaveSimulatedAnimeCollection(collection)
	return nil
}

func (cw *CollectionWrapper) deleteAnimeEntry(mediaId int, isEntryId ...bool) error {
	collection, err := cw.platform.getOrCreateAnimeCollection()
	if err != nil {
		return err
	}

	// Find and remove entry
	for _, list := range collection.GetMediaListCollection().GetLists() {
		for i, entry := range list.GetEntries() {
			if len(isEntryId) > 0 && isEntryId[0] {
				// If isEntryId is true, we assume mediaId is actually the entry ID
				if entry.GetID() == mediaId {
					list.Entries = append(list.Entries[:i], list.Entries[i+1:]...)
					cw.platform.localManager.SaveSimulatedAnimeCollection(collection)
					return nil
				}
			} else {
				if entry.GetMedia().GetID() == mediaId {
					list.Entries = append(list.Entries[:i], list.Entries[i+1:]...)
					cw.platform.localManager.SaveSimulatedAnimeCollection(collection)
					return nil
				}
			}

		}
	}

	return ErrMediaNotFound
}

func (cw *CollectionWrapper) findAnimeEntry(mediaId int, isEntryId ...bool) (*anilist.AnimeCollection_MediaListCollection_Lists_Entries, error) {
	collection, err := cw.platform.getOrCreateAnimeCollection()
	if err != nil {
		return nil, err
	}

	for _, list := range collection.GetMediaListCollection().GetLists() {
		for _, entry := range list.GetEntries() {
			if len(isEntryId) > 0 && isEntryId[0] {
				if entry.GetID() == mediaId {
					return entry, nil
				}
			} else {
				if entry.GetMedia().GetID() == mediaId {
					return entry, nil
				}
			}
		}
	}

	return nil, ErrMediaNotFound
}

func (cw *CollectionWrapper) updateAnimeMediaData(mediaId int, mediaData *anilist.BaseAnime) error {
	collection, err := cw.platform.getOrCreateAnimeCollection()
	if err != nil {
		return err
	}

	for _, list := range collection.GetMediaListCollection().GetLists() {
		for _, entry := range list.GetEntries() {
			if entry.GetMedia().GetID() == mediaId {
				entry.Media = mediaData
				cw.platform.localManager.SaveSimulatedAnimeCollection(collection)
				return nil
			}
		}
	}

	return ErrMediaNotFound
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Manga Collection Helper Methods
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (cw *CollectionWrapper) addMangaEntry(mediaId int, status anilist.MediaListStatus) error {
	collection, err := cw.platform.getOrCreateMangaCollection()
	if err != nil {
		return err
	}

	// Check if entry already exists
	if _, err := cw.findMangaEntry(mediaId); err == nil {
		return errors.New("entry already exists")
	}

	// Fetch media data
	mediaResp, err := cw.platform.client.BaseMangaByID(context.Background(), &mediaId)
	if err != nil {
		return err
	}

	// Find or create the appropriate list
	var targetList *anilist.MangaCollection_MediaListCollection_Lists
	for _, list := range collection.GetMediaListCollection().GetLists() {
		if list.GetStatus() != nil && *list.GetStatus() == status {
			targetList = list
			break
		}
	}

	if targetList == nil {
		// Create new list
		targetList = &anilist.MangaCollection_MediaListCollection_Lists{
			Status:       &status,
			Name:         lo.ToPtr(string(status)),
			IsCustomList: lo.ToPtr(false),
			Entries:      []*anilist.MangaCollection_MediaListCollection_Lists_Entries{},
		}
		collection.GetMediaListCollection().Lists = append(collection.GetMediaListCollection().Lists, targetList)
	}

	// Create new entry
	newEntry := &anilist.MangaCollection_MediaListCollection_Lists_Entries{
		ID:          int(time.Now().UnixNano()),
		Status:      &status,
		Progress:    lo.ToPtr(0),
		Media:       mediaResp.GetMedia(),
		Score:       lo.ToPtr(0.0),
		Notes:       nil,
		Repeat:      lo.ToPtr(0),
		Private:     lo.ToPtr(false),
		StartedAt:   &anilist.MangaCollection_MediaListCollection_Lists_Entries_StartedAt{},
		CompletedAt: &anilist.MangaCollection_MediaListCollection_Lists_Entries_CompletedAt{},
	}

	targetList.Entries = append(targetList.Entries, newEntry)

	// Save collection
	cw.platform.localManager.SaveSimulatedMangaCollection(collection)
	return nil
}

func (cw *CollectionWrapper) updateMangaEntry(mediaId int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) error {
	collection, err := cw.platform.getOrCreateMangaCollection()
	if err != nil {
		return err
	}

	var foundEntry *anilist.MangaCollection_MediaListCollection_Lists_Entries
	var sourceList *anilist.MangaCollection_MediaListCollection_Lists
	var entryIndex int

	// Find the entry
	for _, list := range collection.GetMediaListCollection().GetLists() {
		for i, entry := range list.GetEntries() {
			if entry.GetMedia().GetID() == mediaId {
				foundEntry = entry
				sourceList = list
				entryIndex = i
				break
			}
		}
		if foundEntry != nil {
			break
		}
	}

	if foundEntry == nil || sourceList == nil {
		return ErrMediaNotFound
	}

	// Update entry fields
	if progress != nil {
		foundEntry.Progress = progress
	}
	if scoreRaw != nil {
		foundEntry.Score = lo.ToPtr(float64(*scoreRaw))
	}
	if startedAt != nil {
		foundEntry.StartedAt = &anilist.MangaCollection_MediaListCollection_Lists_Entries_StartedAt{
			Year:  startedAt.Year,
			Month: startedAt.Month,
			Day:   startedAt.Day,
		}
	}
	if completedAt != nil {
		foundEntry.CompletedAt = &anilist.MangaCollection_MediaListCollection_Lists_Entries_CompletedAt{
			Year:  completedAt.Year,
			Month: completedAt.Month,
			Day:   completedAt.Day,
		}
	}

	// If status changed, move entry to different list
	if status != nil && foundEntry.GetStatus() != nil && *status != *foundEntry.GetStatus() {
		foundEntry.Status = status

		// Remove from current list
		sourceList.Entries = append(sourceList.Entries[:entryIndex], sourceList.Entries[entryIndex+1:]...)

		// Find or create target list
		var targetList *anilist.MangaCollection_MediaListCollection_Lists
		for _, list := range collection.GetMediaListCollection().GetLists() {
			if list.GetStatus() != nil && *list.GetStatus() == *status {
				targetList = list
				break
			}
		}

		if targetList == nil {
			targetList = &anilist.MangaCollection_MediaListCollection_Lists{
				Status:       status,
				Name:         lo.ToPtr(string(*status)),
				IsCustomList: lo.ToPtr(false),
				Entries:      []*anilist.MangaCollection_MediaListCollection_Lists_Entries{},
			}
			collection.GetMediaListCollection().Lists = append(collection.GetMediaListCollection().Lists, targetList)
		}

		targetList.Entries = append(targetList.Entries, foundEntry)
	}

	cw.platform.localManager.SaveSimulatedMangaCollection(collection)
	return nil
}

func (cw *CollectionWrapper) deleteMangaEntry(mediaId int, isEntryId ...bool) error {
	collection, err := cw.platform.getOrCreateMangaCollection()
	if err != nil {
		return err
	}

	// Find and remove entry
	for _, list := range collection.GetMediaListCollection().GetLists() {
		for i, entry := range list.GetEntries() {
			if len(isEntryId) > 0 && isEntryId[0] {
				if entry.GetID() == mediaId {
					list.Entries = append(list.Entries[:i], list.Entries[i+1:]...)
					cw.platform.localManager.SaveSimulatedMangaCollection(collection)
					return nil
				}
			} else {
				if entry.GetMedia().GetID() == mediaId {
					list.Entries = append(list.Entries[:i], list.Entries[i+1:]...)
					cw.platform.localManager.SaveSimulatedMangaCollection(collection)
					return nil
				}
			}
		}
	}

	return ErrMediaNotFound
}

func (cw *CollectionWrapper) findMangaEntry(mediaId int, isEntryId ...bool) (*anilist.MangaCollection_MediaListCollection_Lists_Entries, error) {
	collection, err := cw.platform.getOrCreateMangaCollection()
	if err != nil {
		return nil, err
	}

	for _, list := range collection.GetMediaListCollection().GetLists() {
		for _, entry := range list.GetEntries() {
			if len(isEntryId) > 0 && isEntryId[0] {
				if entry.GetID() == mediaId {
					return entry, nil
				}
			} else {
				if entry.GetMedia().GetID() == mediaId {
					return entry, nil
				}
			}
		}
	}

	return nil, ErrMediaNotFound
}

func (cw *CollectionWrapper) updateMangaMediaData(mediaId int, mediaData *anilist.BaseManga) error {
	collection, err := cw.platform.getOrCreateMangaCollection()
	if err != nil {
		return err
	}

	for _, list := range collection.GetMediaListCollection().GetLists() {
		for _, entry := range list.GetEntries() {
			if entry.GetMedia().GetID() == mediaId {
				entry.Media = mediaData
				cw.platform.localManager.SaveSimulatedMangaCollection(collection)
				return nil
			}
		}
	}

	return ErrMediaNotFound
}
