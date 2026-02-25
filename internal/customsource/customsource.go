package customsource

import (
	"context"
	"errors"
	"reflect"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/extension"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"strings"
	"sync"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	AnimeType = "anime"
	MangaType = "manga"
)

const (
	JavascriptMaxSafeInteger uint64 = 1<<53 - 1 // 2^53 - 1
	//ExtensionIdOffset        uint64 = 1 << 31   // 2^31
	//MaxLocalId               uint64 = 0xFFFFFFF // 268,435,455 (28 bits)
	//MaxExtensionIdentifier   uint64 = 0x3FF     // 1,023 (10 bits)
	//LocalIdBitShift          uint64 = 28        // Number of bits allocated for local IDs

	// ExtensionIdOffset uses the sign bit to separate custom source IDs from AniList IDs
	//	AniList IDs: 0 to 2^31-1 (31 bits)
	//	Extension IDs: 2^31 to 2^53-1
	//	Bit allocation: 10 bits for ExtensionIdentifier (1,023 extensions) + 40 bits for LocalId (~1.1 trillion media per extension)
	//	Maximum ID: 2^31 + (2^10 - 1) * 2^40 + (2^40 - 1)
	ExtensionIdOffset      uint64 = 1 << 31       // 2^31
	MaxLocalId             uint64 = (1 << 40) - 1 // ~1.1 trillion (40 bits)
	MaxExtensionIdentifier uint64 = 0x3FF         // 1,023 (10 bits)
	LocalIdBitShift        uint64 = 40            // Number of bits allocated for local IDs
)

type (
	Manager struct {
		extensionBankRef        *util.Ref[*extension.UnifiedBank]
		extensionBankSubscriber *extension.BankSubscriber
		customSources           *result.Map[int, extension.CustomSourceExtension]    // key is extension identifier
		customSourcesById       *result.Map[string, extension.CustomSourceExtension] // key is extension ID
		closedCh                chan struct{}
		once                    sync.Once
		db                      *db.Database
		logger                  *zerolog.Logger
	}
)

// NewManager creates a new custom source manager.
// Should be created each time the extension bank is updated.
func NewManager(extensionBankRef *util.Ref[*extension.UnifiedBank], db *db.Database, logger *zerolog.Logger) *Manager {
	id := uuid.New().String()
	ret := &Manager{
		extensionBankRef:        extensionBankRef,
		extensionBankSubscriber: extensionBankRef.Get().Subscribe(id),
		customSources:           result.NewMap[int, extension.CustomSourceExtension](),
		customSourcesById:       result.NewMap[string, extension.CustomSourceExtension](),
		closedCh:                make(chan struct{}),
		db:                      db,
		logger:                  logger,
	}

	go func() {
		for {
			select {
			case <-ret.extensionBankSubscriber.OnCustomSourcesChanged():
				logger.Debug().Str("id", id).Msg("custom source: Custom sources changed")
				ret.customSources.Clear()
				ret.customSourcesById.Clear()
				extension.RangeExtensions(extensionBankRef.Get(), func(extId string, ext extension.CustomSourceExtension) bool {
					logger.Trace().Str("extension", extId).Str("id", id).Msg("custom source: Updated extension on manager")
					ret.customSources.Set(ext.GetExtensionIdentifier(), ext)
					ret.customSourcesById.Set(ext.GetID(), ext)
					return true
				})
			case <-ret.closedCh:
				logger.Trace().Str("id", id).Msg("custom source: Closed manager")
				ret.customSources.Clear()
				ret.customSourcesById.Clear()
				return
			}
		}
	}()

	logger.Debug().Str("id", id).Msg("custom source: Manager created, loading extensions")
	extension.RangeExtensions(extensionBankRef.Get(), func(id string, ext extension.CustomSourceExtension) bool {
		logger.Trace().Str("extension", ext.GetID()).Str("id", id).Msg("custom source: Extension loaded on manager")
		ret.customSources.Set(ext.GetExtensionIdentifier(), ext)
		ret.customSourcesById.Set(ext.GetID(), ext)
		return true
	})

	return ret
}

func (m *Manager) Close() {
	m.once.Do(func() {
		close(m.closedCh)
		m.extensionBankRef.Get().Unsubscribe(m.extensionBankSubscriber.ID())
	})
}

// GetProviderFromId gets a custom source provider extension from an ID
func (m *Manager) GetProviderFromId(id int) (ext extension.CustomSourceExtension, localId int, isCustom bool, extensionExists bool) {
	return m.getProviderFromId(id)
}

func (m *Manager) GetProviderFromBaseAnime(baseAnime *anilist.BaseAnime) (ext extension.CustomSourceExtension, localId int, isCustom bool, extensionExists bool) {
	if baseAnime == nil {
		return nil, 0, false, false
	}

	id := baseAnime.ID
	if !IsExtensionId(id) {
		return nil, 0, false, false
	}
	return m.getProviderFromId(id)
}

func (m *Manager) GetProviderFromBaseManga(baseManga *anilist.BaseManga) (ext extension.CustomSourceExtension, localId int, isCustom bool, extensionExists bool) {
	if baseManga == nil {
		return nil, 0, false, false
	}

	id := baseManga.ID
	if !IsExtensionId(id) {
		return nil, 0, false, false
	}
	return m.getProviderFromId(id)
}

func (m *Manager) getProviderFromId(id int) (ext extension.CustomSourceExtension, localId int, isCustom bool, extensionExists bool) {
	if !IsExtensionId(id) {
		return nil, 0, false, false
	}

	// Extract the extension identifier and local ID
	extensionIdentifier, localId := ExtractExtensionData(id)

	provider, ok := m.customSources.Get(extensionIdentifier)
	if !ok {
		return nil, 0, true, false
	}

	return provider, localId, true, true
}

// IsExtensionId checks if an ID belongs to an extension using bit-based separation
func IsExtensionId(id int) bool {
	return uint64(id) >= ExtensionIdOffset
}

// GenerateMediaId creates a runtime extension media ID from extension identifier and local ID
func GenerateMediaId(extensionIdentifier, localId int) int {
	ei := uint64(extensionIdentifier) & MaxExtensionIdentifier
	lid := uint64(localId) & MaxLocalId
	id := ExtensionIdOffset + (ei << LocalIdBitShift) + lid
	return int(int64(id))
}

func ExtractExtensionData(mediaId int) (extensionIdentifier int, localId int) {
	u := uint64(mediaId)
	if u < ExtensionIdOffset {
		return 0, 0
	}
	offset := u - ExtensionIdOffset
	extensionIdentifier = int(offset >> LocalIdBitShift)
	localId = int(offset & MaxLocalId)
	return
}

func formatSiteUrl(extId string, siteUrl *string) *string {
	if siteUrl == nil {
		return new("ext_custom_source_" + extId)
	}
	if strings.HasPrefix(*siteUrl, "https://anilist.co") {
		return siteUrl
	}
	return new("ext_custom_source_" + extId + "|END|" + *siteUrl)
}

func GetCustomSourceExtensionIdFromSiteUrl(siteUrl *string) (string, bool) {
	if siteUrl == nil {
		return "", false
	}
	parts := strings.Split(*siteUrl, "|END|")
	if len(parts) != 2 {
		return "", false
	}
	return strings.Replace(parts[0], "ext_custom_source_", "", 1), true
}

func NormalizeMedia(extensionIdentifier int, extId string, obj interface{}) {
	switch v := obj.(type) {
	case *anilist.BaseAnime:
		v.ID = GenerateMediaId(extensionIdentifier, v.ID)
		v.SiteURL = formatSiteUrl(extId, v.SiteURL)
		if v.Title != nil && v.Title.UserPreferred == nil && v.Title.English != nil {
			v.Title.UserPreferred = v.Title.English
		}
		if v.Title == nil {
			v.Title = &anilist.BaseAnime_Title{
				UserPreferred: new("???"),
				English:       new("???"),
				Romaji:        nil,
				Native:        nil,
			}
		}
	case *anilist.CompleteAnime:
		v.ID = GenerateMediaId(extensionIdentifier, v.ID)
		v.SiteURL = formatSiteUrl(extId, v.SiteURL)
		if v.Title != nil && v.Title.UserPreferred == nil && v.Title.English != nil {
			v.Title.UserPreferred = v.Title.English
		}
		if v.Title == nil {
			v.Title = &anilist.CompleteAnime_Title{
				UserPreferred: new("???"),
				English:       new("???"),
				Romaji:        nil,
				Native:        nil,
			}
		}
		if v.Relations != nil {
			for _, edge := range v.Relations.Edges {
				if edge.Node != nil {
					// don't normalize if media comes from anilist
					if edge.Node.SiteURL != nil && strings.HasPrefix(*edge.Node.SiteURL, "https://anilist.co") {
						continue
					}
					NormalizeMedia(extensionIdentifier, extId, edge.Node)
				}
			}
		}
	case *anilist.BaseManga:
		v.ID = GenerateMediaId(extensionIdentifier, v.ID)
		v.SiteURL = formatSiteUrl(extId, v.SiteURL)
		if v.Title != nil && v.Title.UserPreferred == nil {
			v.Title.UserPreferred = v.Title.English
		}
		if v.Title == nil {
			v.Title = &anilist.BaseManga_Title{
				UserPreferred: new("???"),
				English:       new("???"),
				Romaji:        nil,
				Native:        nil,
			}
		}
	case *anilist.AnimeDetailsById_Media:
		v.ID = GenerateMediaId(extensionIdentifier, v.ID)
		v.SiteURL = formatSiteUrl(extId, v.SiteURL)
		if v.Relations != nil {
			for _, edge := range v.Relations.Edges {
				if edge.Node != nil {
					// don't normalize if media comes from anilist
					if edge.Node.SiteURL != nil && strings.HasPrefix(*edge.Node.SiteURL, "https://anilist.co") {
						continue
					}
					NormalizeMedia(extensionIdentifier, extId, edge.Node)
				}
			}
		}
		if v.Recommendations != nil {
			for _, edge := range v.Recommendations.Edges {
				if edge.Node != nil && edge.Node.MediaRecommendation != nil {
					// don't normalize if media comes from anilist
					if edge.Node.MediaRecommendation.SiteURL != nil && strings.HasPrefix(*edge.Node.MediaRecommendation.SiteURL, "https://anilist.co") {
						continue
					}
					NormalizeMedia(extensionIdentifier, extId, edge.Node.MediaRecommendation)
				}
			}
		}
	case *anilist.MangaDetailsById_Media:
		v.ID = GenerateMediaId(extensionIdentifier, v.ID)
		v.SiteURL = formatSiteUrl(extId, v.SiteURL)
		if v.Relations != nil {
			for _, edge := range v.Relations.Edges {
				if edge.Node != nil {
					// don't normalize if media comes from anilist
					if edge.Node.SiteURL != nil && strings.HasPrefix(*edge.Node.SiteURL, "https://anilist.co") {
						continue
					}
					NormalizeMedia(extensionIdentifier, extId, edge.Node)
				}
			}
		}
		if v.Recommendations != nil {
			for _, edge := range v.Recommendations.Edges {
				if edge.Node != nil && edge.Node.MediaRecommendation != nil {
					// don't normalize if media comes from anilist
					if edge.Node.MediaRecommendation.SiteURL != nil && strings.HasPrefix(*edge.Node.MediaRecommendation.SiteURL, "https://anilist.co") {
						continue
					}
					NormalizeMedia(extensionIdentifier, extId, edge.Node.MediaRecommendation)
				}
			}
		}
	default:
		// fallback using reflection
		rv := reflect.ValueOf(obj)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}

		if rv.Kind() == reflect.Struct {
			// Handle ID
			if field := rv.FieldByName("ID"); field.IsValid() && field.CanSet() && field.Kind() == reflect.Int {
				oldID := int(field.Int())
				newID := GenerateMediaId(extensionIdentifier, oldID)
				field.SetInt(int64(newID))
			}

			// Handle SiteURL
			if field := rv.FieldByName("SiteURL"); field.IsValid() && field.CanSet() {
				if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.String {
					val, _ := field.Interface().(*string)
					newVal := *formatSiteUrl(extId, val)
					field.Set(reflect.ValueOf(&newVal))
				}
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Custom source collections
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *Manager) _getCustomSourceEntries(extId string, collectionType string) (*models.CustomSourceCollection, bool) {
	var lc models.CustomSourceCollection
	err := m.db.Gorm().Where("type = ? AND extension_id = ?", collectionType, extId).First(&lc).Error
	return &lc, err == nil
}
func (m *Manager) _getAllCustomSourceEntries(collectionType string) ([]*models.CustomSourceCollection, bool) {
	var lc []*models.CustomSourceCollection
	err := m.db.Gorm().Where("type = ?", collectionType).Find(&lc).Error
	return lc, err == nil
}

func (m *Manager) _saveCustomSourceEntries(extId string, collectionType string, value interface{}) error {

	marshalledValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Check if collection already exists
	lc, ok := m._getCustomSourceEntries(extId, collectionType)
	if ok {
		lc.Value = marshalledValue
		return m.db.Gorm().Save(&lc).Error
	}

	lcN := models.CustomSourceCollection{
		ExtensionId: extId,
		Type:        collectionType,
		Value:       marshalledValue,
	}

	return m.db.Gorm().Save(&lcN).Error
}

func (m *Manager) SaveCustomSourceAnimeEntries(extId string, input map[int]*anilist.AnimeListEntry) error {
	return m._saveCustomSourceEntries(extId, AnimeType, input)
}

func (m *Manager) SaveCustomSourceMangaEntries(extId string, input map[int]*anilist.MangaListEntry) error {
	return m._saveCustomSourceEntries(extId, MangaType, input)
}

func (m *Manager) GetCustomSourceAnimeEntries() (map[string]map[int]*anilist.AnimeListEntry, bool) {
	lc, ok := m._getAllCustomSourceEntries(AnimeType)
	if !ok {
		return nil, false
	}

	extEntries := make(map[string]map[int]*anilist.AnimeListEntry)

	for _, source := range lc {
		var entries map[int]*anilist.AnimeListEntry
		err := json.Unmarshal(source.Value, &entries)
		if err != nil {
			continue
		}

		// Check if the extension still exists
		customSource, found := m.customSourcesById.Get(source.ExtensionId)
		if !found {
			// Extension no longer exists, skip these entries
			continue
		}

		// Refresh media data from the extension
		refreshedEntries := make(map[int]*anilist.AnimeListEntry)
		localIds := make([]int, 0, len(entries))
		for localId := range entries {
			localIds = append(localIds, localId)
		}

		if len(localIds) > 0 {
			// Fetch fresh media data from the extension
			media, err := customSource.GetProvider().GetAnime(context.Background(), localIds)
			if err == nil {
				mediaMap := make(map[int]*anilist.BaseAnime)
				for _, m := range media {
					mediaMap[m.ID] = m
				}

				// Update entries with fresh media data
				for localId, entry := range entries {
					if freshMedia, exists := mediaMap[localId]; exists {
						updatedEntry := *entry
						updatedEntry.Media = freshMedia
						refreshedEntries[localId] = &updatedEntry
					}
					// If media not found, skip this entry (it might have been removed from the extension)
				}
			}
		}

		if len(refreshedEntries) > 0 {
			extEntries[source.ExtensionId] = refreshedEntries
		}
	}

	return extEntries, true
}

func (m *Manager) GetCustomSourceMangaCollection() (map[string]map[int]*anilist.MangaListEntry, bool) {
	lc, ok := m._getAllCustomSourceEntries(MangaType)
	if !ok {
		return nil, false
	}

	extEntries := make(map[string]map[int]*anilist.MangaListEntry)

	for _, source := range lc {
		var entries map[int]*anilist.MangaListEntry
		err := json.Unmarshal(source.Value, &entries)
		if err != nil {
			continue
		}

		// Check if the extension still exists
		customSource, found := m.customSourcesById.Get(source.ExtensionId)
		if !found {
			// Extension no longer exists, skip these entries
			continue
		}

		// Refresh media data from the extension
		refreshedEntries := make(map[int]*anilist.MangaListEntry)
		localIds := make([]int, 0, len(entries))
		for localId := range entries {
			localIds = append(localIds, localId)
		}

		if len(localIds) > 0 {
			// Fetch fresh media data from the extension
			media, err := customSource.GetProvider().GetManga(context.Background(), localIds)
			if err == nil {
				mediaMap := make(map[int]*anilist.BaseManga)
				for _, m := range media {
					mediaMap[m.ID] = m
				}

				// Update entries with fresh media data
				for localId, entry := range entries {
					if freshMedia, exists := mediaMap[localId]; exists {
						updatedEntry := *entry
						updatedEntry.Media = freshMedia
						refreshedEntries[localId] = &updatedEntry
					}
					// If media not found, skip this entry (it might have been removed from the extension)
				}
			}
		}

		if len(refreshedEntries) > 0 {
			extEntries[source.ExtensionId] = refreshedEntries
		}
	}

	return extEntries, true
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Custom source entry management
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// UpdateEntry handles updating a custom source entry
func (m *Manager) UpdateEntry(ctx context.Context, mediaID int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) error {
	customSource, localId, isCustom, extensionExists := m.GetProviderFromId(mediaID)
	if !extensionExists || !isCustom {
		return errors.New("custom source extension not found for media ID")
	}

	// Get the current custom source entries
	extId := customSource.GetID()

	// Try anime first
	animeEntries, hasAnime := m.GetCustomSourceAnimeEntries()
	if hasAnime {
		if entries, exists := animeEntries[extId]; exists {
			if entry, found := entries[localId]; found {
				// Update the entry
				if status != nil {
					entry.Status = status
				}
				if scoreRaw != nil {
					entry.Score = new(float64(*scoreRaw))
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

				// Save the updated entries
				return m.SaveCustomSourceAnimeEntries(extId, entries)
			}
		}
	}

	// Try manga
	mangaEntries, hasManga := m.GetCustomSourceMangaCollection()
	if hasManga {
		if entries, exists := mangaEntries[extId]; exists {
			if entry, found := entries[localId]; found {
				// Update the entry
				if status != nil {
					entry.Status = status
				}
				if scoreRaw != nil {
					entry.Score = new(float64(*scoreRaw))
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

				// Save the updated entries
				return m.SaveCustomSourceMangaEntries(extId, entries)
			}
		}
	}

	// Entry doesn't exist, create it
	// Determine if it's anime or manga by trying to get the media
	media, err := customSource.GetProvider().GetAnime(ctx, []int{localId})
	if err == nil && len(media) > 0 {
		// It's an anime, create entry
		entries := make(map[int]*anilist.AnimeListEntry)
		if hasAnime {
			if existingEntries, exists := animeEntries[extId]; exists {
				entries = existingEntries
			}
		}

		newEntry := &anilist.AnimeListEntry{
			ID:       localId,
			Status:   status,
			Progress: progress,
			Media:    media[0],
		}
		if scoreRaw != nil {
			newEntry.Score = new(float64(*scoreRaw))
		}
		if startedAt != nil {
			newEntry.StartedAt = &anilist.AnimeCollection_MediaListCollection_Lists_Entries_StartedAt{
				Year:  startedAt.Year,
				Month: startedAt.Month,
				Day:   startedAt.Day,
			}
		}
		if completedAt != nil {
			newEntry.CompletedAt = &anilist.AnimeCollection_MediaListCollection_Lists_Entries_CompletedAt{
				Year:  completedAt.Year,
				Month: completedAt.Month,
				Day:   completedAt.Day,
			}
		}
		entries[localId] = newEntry

		return m.SaveCustomSourceAnimeEntries(extId, entries)
	}

	// Try manga
	mangaMedia, err := customSource.GetProvider().GetManga(ctx, []int{localId})
	if err == nil && len(mangaMedia) > 0 {
		// It's a manga, create entry
		entries := make(map[int]*anilist.MangaListEntry)
		if hasManga {
			if existingEntries, exists := mangaEntries[extId]; exists {
				entries = existingEntries
			}
		}

		newEntry := &anilist.MangaListEntry{
			Status:   status,
			Progress: progress,
			Media:    mangaMedia[0],
		}
		if scoreRaw != nil {
			newEntry.Score = new(float64(*scoreRaw))
		}
		if startedAt != nil {
			newEntry.StartedAt = &anilist.MangaCollection_MediaListCollection_Lists_Entries_StartedAt{
				Year:  startedAt.Year,
				Month: startedAt.Month,
				Day:   startedAt.Day,
			}
		}
		if completedAt != nil {
			newEntry.CompletedAt = &anilist.MangaCollection_MediaListCollection_Lists_Entries_CompletedAt{
				Year:  completedAt.Year,
				Month: completedAt.Month,
				Day:   completedAt.Day,
			}
		}
		entries[localId] = newEntry

		return m.SaveCustomSourceMangaEntries(extId, entries)
	}

	return errors.New("unable to determine media type for custom source entry")
}

// UpdateEntryProgress handles updating progress for a custom source entry
func (m *Manager) UpdateEntryProgress(ctx context.Context, mediaID int, progress int, totalCount *int) error {
	status := anilist.MediaListStatusCurrent
	if totalCount != nil && *totalCount > 0 && progress >= *totalCount {
		status = anilist.MediaListStatusCompleted
	}

	return m.UpdateEntry(ctx, mediaID, &status, nil, &progress, nil, nil)
}

// UpdateEntryRepeat handles updating repeat count for a custom source entry
func (m *Manager) UpdateEntryRepeat(_ context.Context, mediaID int, repeat int) error {
	customSource, localId, isCustom, extensionExists := m.GetProviderFromId(mediaID)
	if !isCustom || !extensionExists {
		return errors.New("custom source extension not found for media ID")
	}

	extId := customSource.GetID()

	// Try anime first
	animeEntries, hasAnime := m.GetCustomSourceAnimeEntries()
	if hasAnime {
		if entries, exists := animeEntries[extId]; exists {
			if entry, found := entries[localId]; found {
				entry.Repeat = &repeat
				return m.SaveCustomSourceAnimeEntries(extId, entries)
			}
		}
	}

	// Try manga
	mangaEntries, hasManga := m.GetCustomSourceMangaCollection()
	if hasManga {
		if entries, exists := mangaEntries[extId]; exists {
			if entry, found := entries[localId]; found {
				entry.Repeat = &repeat
				return m.SaveCustomSourceMangaEntries(extId, entries)
			}
		}
	}

	return errors.New("custom source entry not found")
}

// DeleteEntry handles deleting a custom source entry
func (m *Manager) DeleteEntry(_ context.Context, mediaID int, _ int) error {
	customSource, localId, isCustom, extensionExists := m.GetProviderFromId(mediaID)
	if !isCustom || !extensionExists {
		return errors.New("custom source extension not found for media ID")
	}

	extId := customSource.GetID()

	// Try anime first
	animeEntries, hasAnime := m.GetCustomSourceAnimeEntries()
	if hasAnime {
		if entries, exists := animeEntries[extId]; exists {
			if _, found := entries[localId]; found {
				delete(entries, localId)
				return m.SaveCustomSourceAnimeEntries(extId, entries)
			}
		}
	}

	// Try manga
	mangaEntries, hasManga := m.GetCustomSourceMangaCollection()
	if hasManga {
		if entries, exists := mangaEntries[extId]; exists {
			if _, found := entries[localId]; found {
				delete(entries, localId)
				return m.SaveCustomSourceMangaEntries(extId, entries)
			}
		}
	}

	return errors.New("custom source entry not found")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Custom source collection merging
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// MergeAnimeEntries merges custom source anime entries into the anime collection
func (m *Manager) MergeAnimeEntries(collection *anilist.AnimeCollection) {
	customEntries, ok := m.GetCustomSourceAnimeEntries()
	if !ok {
		return
	}

	for extId, entries := range customEntries {
		// Check if the extension still exists
		extIdentifier := 0
		found := false
		m.customSources.Range(func(key int, ext extension.CustomSourceExtension) bool {
			if ext.GetID() == extId {
				extIdentifier = key
				found = true
				return false
			}
			return true
		})

		if !found {
			// Extension no longer exists, skip these entries
			continue
		}

		// Group entries by status for collection lists
		entriesByStatus := make(map[anilist.MediaListStatus][]*anilist.AnimeCollection_MediaListCollection_Lists_Entries)

		for localId, entry := range entries {
			if entry == nil || entry.Media == nil {
				continue
			}

			mediaId := GenerateMediaId(extIdentifier, localId)

			// clone
			mediaCopy := *entry.Media
			NormalizeMedia(extIdentifier, extId, &mediaCopy)

			// Create collection entry
			collectionEntry := &anilist.AnimeCollection_MediaListCollection_Lists_Entries{
				ID:          mediaId,
				Status:      entry.Status,
				Score:       entry.Score,
				Progress:    entry.Progress,
				Repeat:      entry.Repeat,
				StartedAt:   entry.StartedAt,
				CompletedAt: entry.CompletedAt,
				Media:       &mediaCopy,
			}

			// Default to planning if no status
			status := anilist.MediaListStatusPlanning
			if entry.Status != nil {
				status = *entry.Status
			}

			entriesByStatus[status] = append(entriesByStatus[status], collectionEntry)
		}

		// Add entries to appropriate lists in the collection
		for status, statusEntries := range entriesByStatus {
			// Find or create the list for this status
			var targetList *anilist.AnimeCollection_MediaListCollection_Lists
			for _, list := range collection.MediaListCollection.Lists {
				if list.Status != nil && *list.Status == status {
					targetList = list
					break
				}
			}

			if targetList == nil {
				// Create new list for this status
				targetList = &anilist.AnimeCollection_MediaListCollection_Lists{
					Status:  &status,
					Entries: []*anilist.AnimeCollection_MediaListCollection_Lists_Entries{},
				}
				collection.MediaListCollection.Lists = append(collection.MediaListCollection.Lists, targetList)
			}

			// Add entries to the list
			targetList.Entries = append(targetList.Entries, statusEntries...)
		}
	}
}

// MergeMangaEntries merges custom source manga entries into the manga collection
func (m *Manager) MergeMangaEntries(collection *anilist.MangaCollection) {
	customEntries, ok := m.GetCustomSourceMangaCollection()
	if !ok {
		return
	}

	for extId, entries := range customEntries {
		// Check if the extension still exists
		extIdentifier := 0
		found := false
		m.customSources.Range(func(key int, ext extension.CustomSourceExtension) bool {
			if ext.GetID() == extId {
				extIdentifier = key
				found = true
				return false
			}
			return true
		})

		if !found {
			// Extension no longer exists, skip these entries
			continue
		}

		// Group entries by status for collection lists
		entriesByStatus := make(map[anilist.MediaListStatus][]*anilist.MangaCollection_MediaListCollection_Lists_Entries)

		for localId, entry := range entries {
			if entry == nil || entry.Media == nil {
				continue
			}

			mediaId := GenerateMediaId(extIdentifier, localId)

			// clone
			mediaCopy := *entry.Media
			NormalizeMedia(extIdentifier, extId, &mediaCopy)

			// Create collection entry
			collectionEntry := &anilist.MangaCollection_MediaListCollection_Lists_Entries{
				ID:          mediaId,
				Status:      entry.Status,
				Score:       entry.Score,
				Progress:    entry.Progress,
				Repeat:      entry.Repeat,
				StartedAt:   entry.StartedAt,
				CompletedAt: entry.CompletedAt,
				Media:       &mediaCopy,
			}

			// Default to planning if no status
			status := anilist.MediaListStatusPlanning
			if entry.Status != nil {
				status = *entry.Status
			}

			entriesByStatus[status] = append(entriesByStatus[status], collectionEntry)
		}

		// Add entries to appropriate lists in the collection
		for status, statusEntries := range entriesByStatus {
			// Find or create the list for this status
			var targetList *anilist.MangaCollection_MediaListCollection_Lists
			for _, list := range collection.MediaListCollection.Lists {
				if list.Status != nil && *list.Status == status {
					targetList = list
					break
				}
			}

			if targetList == nil {
				// Create new list for this status
				targetList = &anilist.MangaCollection_MediaListCollection_Lists{
					Status:  &status,
					Entries: []*anilist.MangaCollection_MediaListCollection_Lists_Entries{},
				}
				collection.MediaListCollection.Lists = append(collection.MediaListCollection.Lists, targetList)
			}

			// Add entries to the list
			targetList.Entries = append(targetList.Entries, statusEntries...)
		}
	}
}
