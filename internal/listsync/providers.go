package listsync

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/mal"
)

type (
	Provider struct {
		Source          Source
		AnimeEntries    []*AnimeEntry
		AnimeEntriesMap map[int]*AnimeEntry
	}
	ProviderRepository struct { // Holds information used for making requests to the providers
		AnilistClient *anilist.Client
		MalToken      string
		Logger        *zerolog.Logger
	}
)

// NewAnilistProvider creates a new provider for Anilist
func NewAnilistProvider(collection *anilist.AnimeCollection) *Provider {
	entries := FromAnilistCollection(collection)
	entriesMap := make(map[int]*AnimeEntry)
	for _, entry := range entries {
		entriesMap[entry.MalID] = entry
	}
	return &Provider{
		Source:          SourceAniList,
		AnimeEntries:    entries,
		AnimeEntriesMap: entriesMap,
	}
}

// NewMALProvider creates a new provider for MyAnimeList
func NewMALProvider(collection []*mal.AnimeListEntry) *Provider {
	entries := FromMALCollection(collection)
	entriesMap := make(map[int]*AnimeEntry)
	for _, entry := range entries {
		entriesMap[entry.MalID] = entry
	}
	return &Provider{
		Source:          SourceMAL,
		AnimeEntries:    entries,
		AnimeEntriesMap: entriesMap,
	}
}

func (pr *ProviderRepository) AddAnime(to Source, entry *AnimeEntry) error {

	// Add the anime to the provider
	switch to {
	case SourceAniList:
		anizipMedia, err := anizip.FetchAniZipMedia("mal", entry.MalID)
		if err != nil {
			return nil
		}
		// Add the anime to the AniList provider
		anilistId := anizipMedia.Mappings.AnilistID
		if anilistId == 0 {
			return nil
		}
		status := ToAnilistListStatus(entry.Status)
		score := entry.Score * 10

		_, err = pr.AnilistClient.UpdateMediaListEntryStatus(
			context.Background(),
			&anilistId,
			&entry.Progress,
			&status,
			&score,
		)
		if err != nil {
			pr.Logger.Error().Err(err).Msgf("listsync: Failed to add anime \"%s\" to AniList", entry.DisplayTitle)
			return err
		}
		pr.Logger.Trace().Msgf("listsync: Added anime \"%s\" to AniList", entry.DisplayTitle)
	case SourceMAL:
		// Add the anime to the MAL provider
		status := ToMALStatusFromAnimeStatus(entry.Status)

		err := mal.UpdateAnimeListStatus(pr.MalToken, &mal.AnimeListStatusParams{
			Status:             &status,
			NumWatchedEpisodes: &entry.Progress,
			Score:              &entry.Score,
		}, entry.MalID)
		if err != nil {
			pr.Logger.Error().Err(err).Msgf("listsync: Failed to add anime \"%s\" to MAL", entry.DisplayTitle)
			return err
		}
		pr.Logger.Trace().Msgf("listsync: Added anime \"%s\" to MAL", entry.DisplayTitle)
	}

	return nil
}

func (pr *ProviderRepository) UpdateAnime(to Source, entry *AnimeEntry) error {

	// Add the anime to the provider
	switch to {
	case SourceAniList:
		anizipMedia, err := anizip.FetchAniZipMedia("mal", entry.MalID)
		if err != nil {
			return nil
		}
		// Add the anime to the AniList provider
		anilistId := anizipMedia.Mappings.AnilistID
		if anilistId == 0 {
			return nil
		}
		status := ToAnilistListStatus(entry.Status)
		score := entry.Score * 10

		_, err = pr.AnilistClient.UpdateMediaListEntryStatus(
			context.Background(),
			&anilistId,
			&entry.Progress,
			&status,
			&score,
		)
		if err != nil {
			pr.Logger.Error().Err(err).Msgf("listsync: Failed to update anime \"%s\" on AniList", entry.DisplayTitle)
			return err
		}
		pr.Logger.Trace().Msgf("listsync: Updated anime \"%s\" on AniList", entry.DisplayTitle)
	case SourceMAL:
		// Add the anime to the MAL provider
		status := ToMALStatusFromAnimeStatus(entry.Status)

		err := mal.UpdateAnimeListStatus(pr.MalToken, &mal.AnimeListStatusParams{
			Status:             &status,
			NumWatchedEpisodes: &entry.Progress,
			Score:              &entry.Score,
		}, entry.MalID)
		if err != nil {
			pr.Logger.Error().Err(err).Msgf("listsync: Failed to update anime \"%s\" on MAL", entry.DisplayTitle)
			return err
		}
		pr.Logger.Trace().Msgf("listsync: Updated anime \"%s\" on MAL", entry.DisplayTitle)
	}

	return nil
}

func (pr *ProviderRepository) DeleteAnime(from Source, entry *AnimeEntry) error {

	// Add the anime to the provider
	switch from {
	case SourceAniList:
		anizipMedia, err := anizip.FetchAniZipMedia("mal", entry.MalID)
		if err != nil {
			return err
		}
		// Delete the anime from the AniList provider
		anilistId := anizipMedia.Mappings.AnilistID
		if anilistId == 0 {
			return errors.New("anilist id not found")
		}

		_, err = pr.AnilistClient.DeleteEntry(
			context.Background(),
			&anilistId,
		)
		if err != nil {
			pr.Logger.Error().Err(err).Msgf("listsync: Failed to delete anime \"%s\" from AniList", entry.DisplayTitle)
			return err
		}
		pr.Logger.Trace().Msgf("listsync: Deleted anime \"%s\" from AniList", entry.DisplayTitle)

	case SourceMAL:
		// Delete the anime from the MAL provider
		err := mal.DeleteAnimeListItem(pr.MalToken, entry.MalID)
		if err != nil {
			pr.Logger.Error().Err(err).Msgf("listsync: Failed to delete anime \"%s\" from MAL", entry.DisplayTitle)
			return err
		}
		pr.Logger.Trace().Msgf("listsync: Deleted anime \"%s\" from MAL", entry.DisplayTitle)
	}

	return nil
}
