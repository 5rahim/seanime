package manga

import (
	"errors"
	"fmt"
	"seanime/internal/events"
	"seanime/internal/util/filecache"
	"strings"
)

const (
	mangaPreferencesBucket = "manga-preferences"
	mangaPreferencesKey    = "preferences"
)

type MangaProviderFilter struct {
	Scanlators []string `json:"scanlators"`
	Language   string   `json:"language"`
}

type MangaEntryPreference struct {
	Provider string                         `json:"provider"`
	Filters  map[string]MangaProviderFilter `json:"filters"`
}

type MangaPreferences struct {
	Entries map[int]MangaEntryPreference `json:"entries"`
}

type MangaProviderFilterPatch struct {
	Provider   string    `json:"provider"`
	Scanlators *[]string `json:"scanlators,omitempty"`
	Language   *string   `json:"language,omitempty"`
}

type MangaPreferencePatch struct {
	Provider *string                   `json:"provider,omitempty"`
	Filter   *MangaProviderFilterPatch `json:"filter,omitempty"`
}

type MangaPreferencesUpdatedPayload struct {
	MediaIds []int `json:"mediaIds"`
}

func getDefaultPreferences() *MangaPreferences {
	return &MangaPreferences{Entries: make(map[int]MangaEntryPreference)}
}

func normalizeMangaPreferences(preferences *MangaPreferences) *MangaPreferences {
	if preferences == nil {
		return getDefaultPreferences()
	}
	if preferences.Entries == nil {
		preferences.Entries = make(map[int]MangaEntryPreference)
	}
	for mediaId, entry := range preferences.Entries {
		if entry.Filters == nil {
			entry.Filters = make(map[string]MangaProviderFilter)
		}
		preferences.Entries[mediaId] = entry
	}
	return preferences
}

func isValidProvider(provider string) error {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		return errors.New("provider is required")
	}
	if len(provider) > 200 {
		return errors.New("provider is too long")
	}
	return nil
}

func isValidProviderFilter(filter *MangaProviderFilterPatch) error {
	if filter == nil {
		return nil
	}
	if err := isValidProvider(filter.Provider); err != nil {
		return err
	}
	if filter.Scanlators == nil && filter.Language == nil {
		return errors.New("filter update is empty")
	}
	if filter.Language != nil && len(*filter.Language) > 50 {
		return errors.New("language is too long")
	}
	if filter.Scanlators != nil && len(*filter.Scanlators) > 20 {
		return errors.New("too many scanlators")
	}
	if filter.Scanlators != nil {
		for _, scanlator := range *filter.Scanlators {
			if len(scanlator) > 200 {
				return errors.New("scanlator is too long")
			}
		}
	}
	return nil
}

func (r *Repository) GetMangaPreferences() (*MangaPreferences, error) {
	r.preferencesMu.Lock()
	defer r.preferencesMu.Unlock()

	return r.getMangaPreferences()
}

func (r *Repository) getMangaPreferences() (*MangaPreferences, error) {
	bucket := filecache.NewPermanentBucket(mangaPreferencesBucket)
	preferences := getDefaultPreferences()
	found, err := r.fileCacher.GetPerm(bucket, mangaPreferencesKey, preferences)
	if err != nil {
		return nil, err
	}
	if !found {
		if err := r.fileCacher.SetPerm(bucket, mangaPreferencesKey, preferences); err != nil {
			return nil, err
		}
	}
	return normalizeMangaPreferences(preferences), nil
}

func (r *Repository) savePreferences(preferences *MangaPreferences) error {
	bucket := filecache.NewPermanentBucket(mangaPreferencesBucket)
	return r.fileCacher.SetPerm(bucket, mangaPreferencesKey, normalizeMangaPreferences(preferences))
}

func (r *Repository) ImportPreferences(imported *MangaPreferences) (*MangaPreferences, error) {
	imported = normalizeMangaPreferences(imported)
	changedMediaIds := make([]int, 0)

	r.preferencesMu.Lock()
	preferences, err := r.getMangaPreferences()
	if err != nil {
		r.preferencesMu.Unlock()
		return nil, err
	}

	for mediaId, importedEntry := range imported.Entries {
		if mediaId <= 0 {
			continue
		}
		entry, found := preferences.Entries[mediaId]
		entryChanged := false
		if !found {
			entry = MangaEntryPreference{Filters: make(map[string]MangaProviderFilter)}
		}
		if entry.Provider == "" && importedEntry.Provider != "" {
			if err := isValidProvider(importedEntry.Provider); err == nil {
				entry.Provider = strings.TrimSpace(importedEntry.Provider)
				entryChanged = true
			}
		}
		if entry.Filters == nil {
			entry.Filters = make(map[string]MangaProviderFilter)
		}
		for provider, filter := range importedEntry.Filters {
			provider = strings.TrimSpace(provider)
			if _, exists := entry.Filters[provider]; exists {
				continue
			}
			patch := &MangaProviderFilterPatch{
				Provider:   provider,
				Scanlators: &filter.Scanlators,
				Language:   &filter.Language,
			}
			if err := isValidProviderFilter(patch); err != nil {
				continue
			}
			entry.Filters[provider] = MangaProviderFilter{
				Scanlators: append([]string(nil), filter.Scanlators...),
				Language:   filter.Language,
			}
			entryChanged = true
		}
		if entryChanged {
			preferences.Entries[mediaId] = entry
			changedMediaIds = append(changedMediaIds, mediaId)
		}
	}

	if len(changedMediaIds) > 0 {
		err = r.savePreferences(preferences)
	}
	r.preferencesMu.Unlock()
	if err != nil {
		return nil, err
	}

	if len(changedMediaIds) > 0 {
		r.NotifyPreferencesUpdated(changedMediaIds)
	}
	return preferences, nil
}

func (r *Repository) PatchPreference(mediaId int, patch *MangaPreferencePatch, broadcast bool) (*MangaEntryPreference, error) {
	if mediaId <= 0 {
		return nil, errors.New("invalid media id")
	}
	if patch == nil || patch.Provider == nil && patch.Filter == nil {
		return nil, errors.New("preference update is empty")
	}
	if patch.Provider != nil {
		if err := isValidProvider(*patch.Provider); err != nil {
			return nil, err
		}
	}
	if err := isValidProviderFilter(patch.Filter); err != nil {
		return nil, err
	}

	r.preferencesMu.Lock()
	preferences, err := r.getMangaPreferences()
	if err != nil {
		r.preferencesMu.Unlock()
		return nil, err
	}

	entry := preferences.Entries[mediaId]
	if entry.Filters == nil {
		entry.Filters = make(map[string]MangaProviderFilter)
	}
	if patch.Provider != nil {
		entry.Provider = strings.TrimSpace(*patch.Provider)
	}
	if patch.Filter != nil {
		provider := strings.TrimSpace(patch.Filter.Provider)
		filter := entry.Filters[provider]
		if patch.Filter.Scanlators != nil {
			filter.Scanlators = append([]string(nil), (*patch.Filter.Scanlators)...)
		}
		if patch.Filter.Language != nil {
			filter.Language = *patch.Filter.Language
		}
		entry.Filters[provider] = filter
	}
	preferences.Entries[mediaId] = entry
	if err := r.savePreferences(preferences); err != nil {
		r.preferencesMu.Unlock()
		return nil, fmt.Errorf("save manga preferences: %w", err)
	}
	r.preferencesMu.Unlock()

	if broadcast {
		r.NotifyPreferencesUpdated([]int{mediaId})
	}
	return &entry, nil
}

func (r *Repository) NotifyPreferencesUpdated(mediaIds []int) {
	if len(mediaIds) == 0 {
		return
	}
	r.wsEventManager.SendEvent(events.MangaPreferencesUpdated, MangaPreferencesUpdatedPayload{MediaIds: mediaIds})
}
