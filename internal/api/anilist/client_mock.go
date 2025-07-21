package anilist

import (
	"context"
	"log"
	"os"
	"seanime/internal/test_utils"
	"seanime/internal/util"

	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
)

// This file contains helper functions for testing the anilist package

func TestGetMockAnilistClient() AnilistClient {
	return NewMockAnilistClient()
}

// MockAnilistClientImpl is a mock implementation of the AnilistClient, used for tests.
// It uses the real implementation of the AnilistClient to make requests then populates a cache with the results.
// This is to avoid making repeated requests to the AniList API during tests but still have realistic data.
type MockAnilistClientImpl struct {
	realAnilistClient AnilistClient
	logger            *zerolog.Logger
}

func NewMockAnilistClient() *MockAnilistClientImpl {
	return &MockAnilistClientImpl{
		realAnilistClient: NewAnilistClient(test_utils.ConfigData.Provider.AnilistJwt),
		logger:            util.NewLogger(),
	}
}

func (ac *MockAnilistClientImpl) IsAuthenticated() bool {
	return ac.realAnilistClient.IsAuthenticated()
}

func (ac *MockAnilistClientImpl) BaseAnimeByMalID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseAnimeByMalID, error) {
	file, err := os.Open(test_utils.GetTestDataPath("BaseAnimeByMalID"))
	defer file.Close()
	if err != nil {
		if os.IsNotExist(err) {
			ac.logger.Warn().Msgf("MockAnilistClientImpl: CACHE MISS [BaseAnimeByMalID]: %d", *id)
			ret, err := ac.realAnilistClient.BaseAnimeByMalID(context.Background(), id)
			if err != nil {
				return nil, err
			}
			data, err := json.Marshal([]*BaseAnimeByMalID{ret})
			if err != nil {
				log.Fatal(err)
			}
			err = os.WriteFile(test_utils.GetTestDataPath("BaseAnimeByMalID"), data, 0644)
			if err != nil {
				log.Fatal(err)
			}
			return ret, nil
		}
	}

	var media []*BaseAnimeByMalID
	err = json.NewDecoder(file).Decode(&media)
	if err != nil {
		log.Fatal(err)
	}
	var ret *BaseAnimeByMalID
	for _, m := range media {
		if m.GetMedia().ID == *id {
			ret = m
			break
		}
	}

	if ret == nil {
		ac.logger.Warn().Msgf("MockAnilistClientImpl: CACHE MISS [BaseAnimeByMalID]: %d", *id)
		ret, err := ac.realAnilistClient.BaseAnimeByMalID(context.Background(), id)
		if err != nil {
			return nil, err
		}
		media = append(media, ret)
		data, err := json.Marshal(media)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(test_utils.GetTestDataPath("BaseAnimeByMalID"), data, 0644)
		if err != nil {
			log.Fatal(err)
		}
		return ret, nil
	}

	ac.logger.Trace().Msgf("MockAnilistClientImpl: CACHE HIT [BaseAnimeByMalID]: %d", *id)
	return ret, nil
}

func (ac *MockAnilistClientImpl) BaseAnimeByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseAnimeByID, error) {
	file, err := os.Open(test_utils.GetTestDataPath("BaseAnimeByID"))
	defer file.Close()
	if err != nil {
		if os.IsNotExist(err) {
			ac.logger.Warn().Msgf("MockAnilistClientImpl: CACHE MISS [BaseAnimeByID]: %d", *id)
			baseAnime, err := ac.realAnilistClient.BaseAnimeByID(context.Background(), id)
			if err != nil {
				return nil, err
			}
			data, err := json.Marshal([]*BaseAnimeByID{baseAnime})
			if err != nil {
				log.Fatal(err)
			}
			err = os.WriteFile(test_utils.GetTestDataPath("BaseAnimeByID"), data, 0644)
			if err != nil {
				log.Fatal(err)
			}
			return baseAnime, nil
		}
	}

	var media []*BaseAnimeByID
	err = json.NewDecoder(file).Decode(&media)
	if err != nil {
		log.Fatal(err)
	}
	var baseAnime *BaseAnimeByID
	for _, m := range media {
		if m.GetMedia().ID == *id {
			baseAnime = m
			break
		}
	}

	if baseAnime == nil {
		ac.logger.Warn().Msgf("MockAnilistClientImpl: CACHE MISS [BaseAnimeByID]: %d", *id)
		baseAnime, err := ac.realAnilistClient.BaseAnimeByID(context.Background(), id)
		if err != nil {
			return nil, err
		}
		media = append(media, baseAnime)
		data, err := json.Marshal(media)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(test_utils.GetTestDataPath("BaseAnimeByID"), data, 0644)
		if err != nil {
			log.Fatal(err)
		}
		return baseAnime, nil
	}

	ac.logger.Trace().Msgf("MockAnilistClientImpl: CACHE HIT [BaseAnimeByID]: %d", *id)
	return baseAnime, nil
}

// AnimeCollection
//   - Set userName to nil to use the boilerplate AnimeCollection
//   - Set userName to a specific username to fetch and cache
func (ac *MockAnilistClientImpl) AnimeCollection(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*AnimeCollection, error) {

	if userName == nil {
		file, err := os.Open(test_utils.GetDataPath("BoilerplateAnimeCollection"))
		defer file.Close()

		var ret *AnimeCollection
		err = json.NewDecoder(file).Decode(&ret)
		if err != nil {
			log.Fatal(err)
		}

		ac.logger.Trace().Msgf("MockAnilistClientImpl: Using [BoilerplateAnimeCollection]")
		return ret, nil
	}

	file, err := os.Open(test_utils.GetTestDataPath("AnimeCollection"))
	defer file.Close()
	if err != nil {
		if os.IsNotExist(err) {
			ac.logger.Warn().Msgf("MockAnilistClientImpl: CACHE MISS [AnimeCollection]: %s", *userName)
			ret, err := ac.realAnilistClient.AnimeCollection(context.Background(), userName)
			if err != nil {
				return nil, err
			}
			data, err := json.Marshal(ret)
			if err != nil {
				log.Fatal(err)
			}
			err = os.WriteFile(test_utils.GetTestDataPath("AnimeCollection"), data, 0644)
			if err != nil {
				log.Fatal(err)
			}
			return ret, nil
		}
	}

	var ret *AnimeCollection
	err = json.NewDecoder(file).Decode(&ret)
	if err != nil {
		log.Fatal(err)
	}

	if ret == nil {
		ac.logger.Warn().Msgf("MockAnilistClientImpl: CACHE MISS [AnimeCollection]: %s", *userName)
		ret, err := ac.realAnilistClient.AnimeCollection(context.Background(), userName)
		if err != nil {
			return nil, err
		}
		data, err := json.Marshal(ret)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(test_utils.GetTestDataPath("AnimeCollection"), data, 0644)
		if err != nil {
			log.Fatal(err)
		}
		return ret, nil
	}

	ac.logger.Trace().Msgf("MockAnilistClientImpl: CACHE HIT [AnimeCollection]: %s", *userName)
	return ret, nil

}

func (ac *MockAnilistClientImpl) AnimeCollectionWithRelations(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*AnimeCollectionWithRelations, error) {

	if userName == nil {
		file, err := os.Open(test_utils.GetDataPath("BoilerplateAnimeCollectionWithRelations"))
		defer file.Close()

		var ret *AnimeCollectionWithRelations
		err = json.NewDecoder(file).Decode(&ret)
		if err != nil {
			log.Fatal(err)
		}

		ac.logger.Trace().Msgf("MockAnilistClientImpl: Using [BoilerplateAnimeCollectionWithRelations]")
		return ret, nil
	}

	file, err := os.Open(test_utils.GetTestDataPath("AnimeCollectionWithRelations"))
	defer file.Close()
	if err != nil {
		if os.IsNotExist(err) {
			ac.logger.Warn().Msgf("MockAnilistClientImpl: CACHE MISS [AnimeCollectionWithRelations]: %s", *userName)
			ret, err := ac.realAnilistClient.AnimeCollectionWithRelations(context.Background(), userName)
			if err != nil {
				return nil, err
			}
			data, err := json.Marshal(ret)
			if err != nil {
				log.Fatal(err)
			}
			err = os.WriteFile(test_utils.GetTestDataPath("AnimeCollectionWithRelations"), data, 0644)
			if err != nil {
				log.Fatal(err)
			}
			return ret, nil
		}
	}

	var ret *AnimeCollectionWithRelations
	err = json.NewDecoder(file).Decode(&ret)
	if err != nil {
		log.Fatal(err)
	}

	if ret == nil {
		ac.logger.Warn().Msgf("MockAnilistClientImpl: CACHE MISS [AnimeCollectionWithRelations]: %s", *userName)
		ret, err := ac.realAnilistClient.AnimeCollectionWithRelations(context.Background(), userName)
		if err != nil {
			return nil, err
		}
		data, err := json.Marshal(ret)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(test_utils.GetTestDataPath("AnimeCollectionWithRelations"), data, 0644)
		if err != nil {
			log.Fatal(err)
		}
		return ret, nil
	}

	ac.logger.Trace().Msgf("MockAnilistClientImpl: CACHE HIT [AnimeCollectionWithRelations]: %s", *userName)
	return ret, nil

}

type TestModifyAnimeCollectionEntryInput struct {
	Status            *MediaListStatus
	Progress          *int
	Score             *float64
	AiredEpisodes     *int
	NextAiringEpisode *BaseAnime_NextAiringEpisode
}

// TestModifyAnimeCollectionEntry will modify an entry in the fetched anime collection.
// This is used to fine-tune the anime collection for testing purposes.
//
// Example: Setting a specific progress in case the origin anime collection has no progress
func TestModifyAnimeCollectionEntry(ac *AnimeCollection, mId int, input TestModifyAnimeCollectionEntryInput) *AnimeCollection {
	if ac == nil {
		panic("AnimeCollection is nil")
	}

	lists := ac.GetMediaListCollection().GetLists()

	removedFromList := false
	var rEntry *AnimeCollection_MediaListCollection_Lists_Entries

	// Move the entry to the correct list
	if input.Status != nil {
		for _, list := range lists {
			if list.Status == nil || list.Entries == nil {
				continue
			}
			entries := list.GetEntries()
			for idx, entry := range entries {
				if entry.GetMedia().ID == mId {
					// Remove from current list if status differs
					if *list.Status != *input.Status {
						removedFromList = true
						rEntry = entry
						// Ensure we're not going out of bounds
						if idx >= 0 && idx < len(entries) {
							// Safely remove the entry by re-slicing
							list.Entries = append(entries[:idx], entries[idx+1:]...)
						}
						break
					}
				}
			}
		}

		// Add the entry to the correct list if it was removed
		if removedFromList && rEntry != nil {
			for _, list := range lists {
				if list.Status == nil {
					continue
				}
				if *list.Status == *input.Status {
					if list.Entries == nil {
						list.Entries = make([]*AnimeCollection_MediaListCollection_Lists_Entries, 0)
					}
					// Add the removed entry to the new list
					list.Entries = append(list.Entries, rEntry)
					break
				}
			}
		}
	}

	// Update the entry details
out:
	for _, list := range lists {
		entries := list.GetEntries()
		for _, entry := range entries {
			if entry.GetMedia().ID == mId {
				if input.Status != nil {
					entry.Status = input.Status
				}
				if input.Progress != nil {
					entry.Progress = input.Progress
				}
				if input.Score != nil {
					entry.Score = input.Score
				}
				if input.AiredEpisodes != nil {
					entry.Media.Episodes = input.AiredEpisodes
				}
				if input.NextAiringEpisode != nil {
					entry.Media.NextAiringEpisode = input.NextAiringEpisode
				}
				break out
			}
		}
	}

	return ac
}

func TestAddAnimeCollectionEntry(ac *AnimeCollection, mId int, input TestModifyAnimeCollectionEntryInput, realClient AnilistClient) *AnimeCollection {
	if ac == nil {
		panic("AnimeCollection is nil")
	}

	// Fetch the anime details
	baseAnime, err := realClient.BaseAnimeByID(context.Background(), &mId)
	if err != nil {
		log.Fatal(err)
	}
	anime := baseAnime.GetMedia()

	if input.NextAiringEpisode != nil {
		anime.NextAiringEpisode = input.NextAiringEpisode
	}

	if input.AiredEpisodes != nil {
		anime.Episodes = input.AiredEpisodes
	}

	lists := ac.GetMediaListCollection().GetLists()

	// Add the entry to the correct list
	if input.Status != nil {
		for _, list := range lists {
			if list.Status == nil {
				continue
			}
			if *list.Status == *input.Status {
				if list.Entries == nil {
					list.Entries = make([]*AnimeCollection_MediaListCollection_Lists_Entries, 0)
				}
				list.Entries = append(list.Entries, &AnimeCollection_MediaListCollection_Lists_Entries{
					Media:    baseAnime.GetMedia(),
					Status:   input.Status,
					Progress: input.Progress,
					Score:    input.Score,
				})
				break
			}
		}
	}

	return ac
}

func TestAddAnimeCollectionWithRelationsEntry(ac *AnimeCollectionWithRelations, mId int, input TestModifyAnimeCollectionEntryInput, realClient AnilistClient) *AnimeCollectionWithRelations {
	if ac == nil {
		panic("AnimeCollection is nil")
	}

	// Fetch the anime details
	baseAnime, err := realClient.CompleteAnimeByID(context.Background(), &mId)
	if err != nil {
		log.Fatal(err)
	}
	anime := baseAnime.GetMedia()

	//if input.NextAiringEpisode != nil {
	//	anime.NextAiringEpisode = input.NextAiringEpisode
	//}

	if input.AiredEpisodes != nil {
		anime.Episodes = input.AiredEpisodes
	}

	lists := ac.GetMediaListCollection().GetLists()

	// Add the entry to the correct list
	if input.Status != nil {
		for _, list := range lists {
			if list.Status == nil {
				continue
			}
			if *list.Status == *input.Status {
				if list.Entries == nil {
					list.Entries = make([]*AnimeCollectionWithRelations_MediaListCollection_Lists_Entries, 0)
				}
				list.Entries = append(list.Entries, &AnimeCollectionWithRelations_MediaListCollection_Lists_Entries{
					Media:    baseAnime.GetMedia(),
					Status:   input.Status,
					Progress: input.Progress,
					Score:    input.Score,
				})
				break
			}
		}
	}

	return ac
}

//
// WILL NOT IMPLEMENT
//

func (ac *MockAnilistClientImpl) UpdateMediaListEntry(ctx context.Context, mediaID *int, status *MediaListStatus, scoreRaw *int, progress *int, startedAt *FuzzyDateInput, completedAt *FuzzyDateInput, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntry, error) {
	ac.logger.Debug().Int("mediaId", *mediaID).Msg("anilist: Updating media list entry")
	return &UpdateMediaListEntry{}, nil
}

func (ac *MockAnilistClientImpl) UpdateMediaListEntryProgress(ctx context.Context, mediaID *int, progress *int, status *MediaListStatus, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntryProgress, error) {
	ac.logger.Debug().Int("mediaId", *mediaID).Msg("anilist: Updating media list entry progress")
	return &UpdateMediaListEntryProgress{}, nil
}

func (ac *MockAnilistClientImpl) UpdateMediaListEntryRepeat(ctx context.Context, mediaID *int, repeat *int, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntryRepeat, error) {
	ac.logger.Debug().Int("mediaId", *mediaID).Msg("anilist: Updating media list entry repeat")
	return &UpdateMediaListEntryRepeat{}, nil
}

func (ac *MockAnilistClientImpl) DeleteEntry(ctx context.Context, mediaListEntryID *int, interceptors ...clientv2.RequestInterceptor) (*DeleteEntry, error) {
	ac.logger.Debug().Int("entryId", *mediaListEntryID).Msg("anilist: Deleting media list entry")
	return &DeleteEntry{}, nil
}

func (ac *MockAnilistClientImpl) AnimeDetailsByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*AnimeDetailsByID, error) {
	ac.logger.Debug().Int("mediaId", *id).Msg("anilist: Fetching anime details")
	return ac.realAnilistClient.AnimeDetailsByID(ctx, id, interceptors...)
}

func (ac *MockAnilistClientImpl) CompleteAnimeByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*CompleteAnimeByID, error) {
	ac.logger.Debug().Int("mediaId", *id).Msg("anilist: Fetching complete media")
	return ac.realAnilistClient.CompleteAnimeByID(ctx, id, interceptors...)
}

func (ac *MockAnilistClientImpl) ListAnime(ctx context.Context, page *int, search *string, perPage *int, sort []*MediaSort, status []*MediaStatus, genres []*string, averageScoreGreater *int, season *MediaSeason, seasonYear *int, format *MediaFormat, isAdult *bool, interceptors ...clientv2.RequestInterceptor) (*ListAnime, error) {
	ac.logger.Debug().Msg("anilist: Fetching media list")
	return ac.realAnilistClient.ListAnime(ctx, page, search, perPage, sort, status, genres, averageScoreGreater, season, seasonYear, format, isAdult, interceptors...)
}

func (ac *MockAnilistClientImpl) ListRecentAnime(ctx context.Context, page *int, perPage *int, airingAtGreater *int, airingAtLesser *int, notYetAired *bool, interceptors ...clientv2.RequestInterceptor) (*ListRecentAnime, error) {
	ac.logger.Debug().Msg("anilist: Fetching recent media list")
	return ac.realAnilistClient.ListRecentAnime(ctx, page, perPage, airingAtGreater, airingAtLesser, notYetAired, interceptors...)
}

func (ac *MockAnilistClientImpl) GetViewer(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*GetViewer, error) {
	ac.logger.Debug().Msg("anilist: Fetching viewer")
	return ac.realAnilistClient.GetViewer(ctx, interceptors...)
}

func (ac *MockAnilistClientImpl) MangaCollection(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*MangaCollection, error) {
	ac.logger.Debug().Msg("anilist: Fetching manga collection")
	return ac.realAnilistClient.MangaCollection(ctx, userName, interceptors...)
}

func (ac *MockAnilistClientImpl) SearchBaseManga(ctx context.Context, page *int, perPage *int, sort []*MediaSort, search *string, status []*MediaStatus, interceptors ...clientv2.RequestInterceptor) (*SearchBaseManga, error) {
	ac.logger.Debug().Msg("anilist: Searching manga")
	return ac.realAnilistClient.SearchBaseManga(ctx, page, perPage, sort, search, status, interceptors...)
}

func (ac *MockAnilistClientImpl) BaseMangaByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseMangaByID, error) {
	ac.logger.Debug().Int("mediaId", *id).Msg("anilist: Fetching manga")
	return ac.realAnilistClient.BaseMangaByID(ctx, id, interceptors...)
}

func (ac *MockAnilistClientImpl) MangaDetailsByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*MangaDetailsByID, error) {
	ac.logger.Debug().Int("mediaId", *id).Msg("anilist: Fetching manga details")
	return ac.realAnilistClient.MangaDetailsByID(ctx, id, interceptors...)
}

func (ac *MockAnilistClientImpl) ListManga(ctx context.Context, page *int, search *string, perPage *int, sort []*MediaSort, status []*MediaStatus, genres []*string, averageScoreGreater *int, startDateGreater *string, startDateLesser *string, format *MediaFormat, countryOfOrigin *string, isAdult *bool, interceptors ...clientv2.RequestInterceptor) (*ListManga, error) {
	ac.logger.Debug().Msg("anilist: Fetching manga list")
	return ac.realAnilistClient.ListManga(ctx, page, search, perPage, sort, status, genres, averageScoreGreater, startDateGreater, startDateLesser, format, countryOfOrigin, isAdult, interceptors...)
}

func (ac *MockAnilistClientImpl) StudioDetails(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*StudioDetails, error) {
	ac.logger.Debug().Int("studioId", *id).Msg("anilist: Fetching studio details")
	return ac.realAnilistClient.StudioDetails(ctx, id, interceptors...)
}

func (ac *MockAnilistClientImpl) ViewerStats(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*ViewerStats, error) {
	ac.logger.Debug().Msg("anilist: Fetching stats")
	return ac.realAnilistClient.ViewerStats(ctx, interceptors...)
}

func (ac *MockAnilistClientImpl) SearchBaseAnimeByIds(ctx context.Context, ids []*int, page *int, perPage *int, status []*MediaStatus, inCollection *bool, sort []*MediaSort, season *MediaSeason, year *int, genre *string, format *MediaFormat, interceptors ...clientv2.RequestInterceptor) (*SearchBaseAnimeByIds, error) {
	ac.logger.Debug().Msg("anilist: Searching anime by ids")
	return ac.realAnilistClient.SearchBaseAnimeByIds(ctx, ids, page, perPage, status, inCollection, sort, season, year, genre, format, interceptors...)
}

func (ac *MockAnilistClientImpl) AnimeAiringSchedule(ctx context.Context, ids []*int, season *MediaSeason, seasonYear *int, previousSeason *MediaSeason, previousSeasonYear *int, nextSeason *MediaSeason, nextSeasonYear *int, interceptors ...clientv2.RequestInterceptor) (*AnimeAiringSchedule, error) {
	ac.logger.Debug().Msg("anilist: Fetching schedule")
	return ac.realAnilistClient.AnimeAiringSchedule(ctx, ids, season, seasonYear, previousSeason, previousSeasonYear, nextSeason, nextSeasonYear, interceptors...)
}

func (ac *MockAnilistClientImpl) AnimeAiringScheduleRaw(ctx context.Context, ids []*int, interceptors ...clientv2.RequestInterceptor) (*AnimeAiringScheduleRaw, error) {
	ac.logger.Debug().Msg("anilist: Fetching schedule")
	return ac.realAnilistClient.AnimeAiringScheduleRaw(ctx, ids, interceptors...)
}
