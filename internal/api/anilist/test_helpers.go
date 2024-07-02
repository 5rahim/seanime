package anilist

import (
	"context"
	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/limiter"
	"log"
	"os"
)

// This file contains helper functions for testing the anilist package

func TestGetMockAnilistClientWrapper() ClientWrapperInterface {
	return NewMockClientWrapper()
}

// MockClientWrapper is a mock implementation of the ClientWrapperInterface, used for tests.
// It uses the real implementation of the ClientWrapperInterface to make requests then populates a cache with the results.
// This is to avoid making repeated requests to the AniList API during tests but still have realistic data.
type MockClientWrapper struct {
	realClientWrapper ClientWrapperInterface
	logger            *zerolog.Logger
}

func NewMockClientWrapper() *MockClientWrapper {
	return &MockClientWrapper{
		realClientWrapper: NewClientWrapper(test_utils.ConfigData.Provider.AnilistJwt),
		logger:            util.NewLogger(),
	}
}

func (c *MockClientWrapper) BasicMediaByMalID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BasicMediaByMalID, error) {
	file, err := os.Open(test_utils.GetTestDataPath("BasicMediaByMalID"))
	defer file.Close()
	if err != nil {
		if os.IsNotExist(err) {
			c.logger.Warn().Msgf("MockClientWrapper: CACHE MISS [BasicMediaByMalID]: %d", *id)
			ret, err := c.realClientWrapper.BasicMediaByMalID(context.Background(), id)
			if err != nil {
				return nil, err
			}
			data, err := json.Marshal([]*BasicMediaByMalID{ret})
			if err != nil {
				log.Fatal(err)
			}
			err = os.WriteFile(test_utils.GetTestDataPath("BasicMediaByMalID"), data, 0644)
			if err != nil {
				log.Fatal(err)
			}
			return ret, nil
		}
	}

	var media []*BasicMediaByMalID
	err = json.NewDecoder(file).Decode(&media)
	if err != nil {
		log.Fatal(err)
	}
	var ret *BasicMediaByMalID
	for _, m := range media {
		if m.GetMedia().ID == *id {
			ret = m
			break
		}
	}

	if ret == nil {
		c.logger.Warn().Msgf("MockClientWrapper: CACHE MISS [BasicMediaByMalID]: %d", *id)
		ret, err := c.realClientWrapper.BasicMediaByMalID(context.Background(), id)
		if err != nil {
			return nil, err
		}
		media = append(media, ret)
		data, err := json.Marshal(media)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(test_utils.GetTestDataPath("BasicMediaByMalID"), data, 0644)
		if err != nil {
			log.Fatal(err)
		}
		return ret, nil
	}

	c.logger.Trace().Msgf("MockClientWrapper: CACHE HIT [BasicMediaByMalID]: %d", *id)
	return ret, nil
}
func (c *MockClientWrapper) BasicMediaByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BasicMediaByID, error) {
	file, err := os.Open(test_utils.GetTestDataPath("BasicMediaByID"))
	defer file.Close()
	if err != nil {
		if os.IsNotExist(err) {
			c.logger.Warn().Msgf("MockClientWrapper: CACHE MISS [BasicMediaByID]: %d", *id)
			ret, err := c.realClientWrapper.BasicMediaByID(context.Background(), id)
			if err != nil {
				return nil, err
			}
			data, err := json.Marshal([]*BasicMediaByID{ret})
			if err != nil {
				log.Fatal(err)
			}
			err = os.WriteFile(test_utils.GetTestDataPath("BasicMediaByID"), data, 0644)
			if err != nil {
				log.Fatal(err)
			}
			return ret, nil
		}
	}

	var media []*BasicMediaByID
	err = json.NewDecoder(file).Decode(&media)
	if err != nil {
		log.Fatal(err)
	}
	var ret *BasicMediaByID
	for _, m := range media {
		if m.GetMedia().ID == *id {
			ret = m
			break
		}
	}

	if ret == nil {
		c.logger.Warn().Msgf("MockClientWrapper: CACHE MISS [BasicMediaByID]: %d", *id)
		ret, err := c.realClientWrapper.BasicMediaByID(context.Background(), id)
		if err != nil {
			return nil, err
		}
		media = append(media, ret)
		data, err := json.Marshal(media)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(test_utils.GetTestDataPath("BasicMediaByID"), data, 0644)
		if err != nil {
			log.Fatal(err)
		}
		return ret, nil
	}

	c.logger.Trace().Msgf("MockClientWrapper: CACHE HIT [BasicMediaByID]: %d", *id)
	return ret, nil
}

func (c *MockClientWrapper) BaseMediaByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseMediaByID, error) {
	file, err := os.Open(test_utils.GetTestDataPath("BaseMediaByID"))
	defer file.Close()
	if err != nil {
		if os.IsNotExist(err) {
			c.logger.Warn().Msgf("MockClientWrapper: CACHE MISS [BaseMediaByID]: %d", *id)
			baseMedia, err := c.realClientWrapper.BaseMediaByID(context.Background(), id)
			if err != nil {
				return nil, err
			}
			data, err := json.Marshal([]*BaseMediaByID{baseMedia})
			if err != nil {
				log.Fatal(err)
			}
			err = os.WriteFile(test_utils.GetTestDataPath("BaseMediaByID"), data, 0644)
			if err != nil {
				log.Fatal(err)
			}
			return baseMedia, nil
		}
	}

	var media []*BaseMediaByID
	err = json.NewDecoder(file).Decode(&media)
	if err != nil {
		log.Fatal(err)
	}
	var baseMedia *BaseMediaByID
	for _, m := range media {
		if m.GetMedia().ID == *id {
			baseMedia = m
			break
		}
	}

	if baseMedia == nil {
		c.logger.Warn().Msgf("MockClientWrapper: CACHE MISS [BaseMediaByID]: %d", *id)
		baseMedia, err := c.realClientWrapper.BaseMediaByID(context.Background(), id)
		if err != nil {
			return nil, err
		}
		media = append(media, baseMedia)
		data, err := json.Marshal(media)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(test_utils.GetTestDataPath("BaseMediaByID"), data, 0644)
		if err != nil {
			log.Fatal(err)
		}
		return baseMedia, nil
	}

	c.logger.Trace().Msgf("MockClientWrapper: CACHE HIT [BaseMediaByID]: %d", *id)
	return baseMedia, nil
}

// AnimeCollection
//   - Set userName to nil to use the boilerplate AnimeCollection
//   - Set userName to a specific username to fetch and cache
func (c *MockClientWrapper) AnimeCollection(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*AnimeCollection, error) {

	if userName == nil {
		file, err := os.Open(test_utils.GetDataPath("BoilerplateAnimeCollection"))
		defer file.Close()

		var ret *AnimeCollection
		err = json.NewDecoder(file).Decode(&ret)
		if err != nil {
			log.Fatal(err)
		}

		c.logger.Trace().Msgf("MockClientWrapper: Using [BoilerplateAnimeCollection]")
		return ret, nil
	}

	file, err := os.Open(test_utils.GetTestDataPath("AnimeCollection"))
	defer file.Close()
	if err != nil {
		if os.IsNotExist(err) {
			c.logger.Warn().Msgf("MockClientWrapper: CACHE MISS [AnimeCollection]: %s", *userName)
			ret, err := c.realClientWrapper.AnimeCollection(context.Background(), userName)
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
		c.logger.Warn().Msgf("MockClientWrapper: CACHE MISS [AnimeCollection]: %s", *userName)
		ret, err := c.realClientWrapper.AnimeCollection(context.Background(), userName)
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

	c.logger.Trace().Msgf("MockClientWrapper: CACHE HIT [AnimeCollection]: %s", *userName)
	return ret, nil

}

type TestModifyAnimeCollectionEntryInput struct {
	Status            *MediaListStatus
	Progress          *int
	Score             *float64
	AiredEpisodes     *int
	NextAiringEpisode *BaseMedia_NextAiringEpisode
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
			if list.Status == nil {
				continue
			}
			if list.Entries == nil {
				continue
			}
			entries := list.GetEntries()
			for idx, entry := range entries {
				if entry.GetMedia().ID == mId {
					if *list.Status != *input.Status {
						removedFromList = true
						entries = append(entries[:idx], entries[idx+1:]...)
						rEntry = entry
						break
					}
				}
			}
		}
		if removedFromList {
			for _, list := range lists {
				if list.Status == nil {
					continue
				}
				if *list.Status == *input.Status {
					if list.Entries == nil {
						list.Entries = make([]*AnimeCollection_MediaListCollection_Lists_Entries, 0)
					}
					list.Entries = append(list.Entries, rEntry)
					break
				}
			}
		}
	}

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

//
// WILL NOT IMPLEMENT
//

func (c *MockClientWrapper) AddMediaToPlanning(mIds []int, rateLimiter *limiter.Limiter, logger *zerolog.Logger) error {
	panic("not implemented")
}
func (c *MockClientWrapper) UpdateMediaListEntryProgress(ctx context.Context, mediaID *int, progress *int, totalEpisodes *int) error {
	panic("not implemented")
}
func (c *MockClientWrapper) UpdateEntry(ctx context.Context, mediaID *int, status *MediaListStatus, score *float64, progress *int, repeat *int, private *bool, notes *string, hiddenFromStatusLists *bool, startedAt *FuzzyDateInput, completedAt *FuzzyDateInput, interceptors ...clientv2.RequestInterceptor) (*UpdateEntry, error) {
	panic("not implemented")
}
func (c *MockClientWrapper) UpdateMediaListEntry(ctx context.Context, mediaID *int, status *MediaListStatus, scoreRaw *int, progress *int, startedAt *FuzzyDateInput, completedAt *FuzzyDateInput, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntry, error) {
	panic("not implemented")
}
func (c *MockClientWrapper) UpdateMediaListEntryStatus(ctx context.Context, mediaID *int, progress *int, status *MediaListStatus, scoreRaw *int, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntryStatus, error) {
	panic("not implemented")
}
func (c *MockClientWrapper) DeleteEntry(ctx context.Context, mediaListEntryID *int, interceptors ...clientv2.RequestInterceptor) (*DeleteEntry, error) {
	panic("not implemented")
}
func (c *MockClientWrapper) SearchAnimeShortMedia(ctx context.Context, page *int, perPage *int, sort []*MediaSort, search *string, status []*MediaStatus, interceptors ...clientv2.RequestInterceptor) (*SearchAnimeShortMedia, error) {
	panic("not implemented")
}
func (c *MockClientWrapper) MediaDetailsByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*MediaDetailsByID, error) {
	panic("not implemented")
}
func (c *MockClientWrapper) CompleteMediaByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*CompleteMediaByID, error) {
	panic("not implemented")
}
func (c *MockClientWrapper) ListMedia(ctx context.Context, page *int, search *string, perPage *int, sort []*MediaSort, status []*MediaStatus, genres []*string, averageScoreGreater *int, season *MediaSeason, seasonYear *int, format *MediaFormat, interceptors ...clientv2.RequestInterceptor) (*ListMedia, error) {
	return c.realClientWrapper.ListMedia(ctx, page, search, perPage, sort, status, genres, averageScoreGreater, season, seasonYear, format, interceptors...)
}
func (c *MockClientWrapper) ListRecentMedia(ctx context.Context, page *int, perPage *int, airingAtGreater *int, airingAtLesser *int, interceptors ...clientv2.RequestInterceptor) (*ListRecentMedia, error) {
	panic("not implemented")
}
func (c *MockClientWrapper) GetViewer(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*GetViewer, error) {
	panic("not implemented")
}
func (c *MockClientWrapper) MangaCollection(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*MangaCollection, error) {
	return c.realClientWrapper.MangaCollection(ctx, userName, interceptors...)
}
func (c *MockClientWrapper) SearchBaseManga(ctx context.Context, page *int, perPage *int, sort []*MediaSort, search *string, status []*MediaStatus, interceptors ...clientv2.RequestInterceptor) (*SearchBaseManga, error) {
	return c.realClientWrapper.SearchBaseManga(ctx, page, perPage, sort, search, status, interceptors...)
}
func (c *MockClientWrapper) BaseMangaByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseMangaByID, error) {
	return c.realClientWrapper.BaseMangaByID(ctx, id, interceptors...)
}
func (c *MockClientWrapper) MangaDetailsByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*MangaDetailsByID, error) {
	return c.realClientWrapper.MangaDetailsByID(ctx, id, interceptors...)
}
func (c *MockClientWrapper) ListManga(ctx context.Context, page *int, search *string, perPage *int, sort []*MediaSort, status []*MediaStatus, genres []*string, averageScoreGreater *int, startDateGreater *string, startDateLesser *string, format *MediaFormat, isAdult *bool, interceptors ...clientv2.RequestInterceptor) (*ListManga, error) {
	return c.realClientWrapper.ListManga(ctx, page, search, perPage, sort, status, genres, averageScoreGreater, startDateGreater, startDateLesser, format, isAdult, interceptors...)
}
func (c *MockClientWrapper) StudioDetails(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*StudioDetails, error) {
	c.logger.Debug().Int("studioId", *id).Msg("anilist: Fetching studio details")
	return c.realClientWrapper.StudioDetails(ctx, id, interceptors...)
}

func (c *MockClientWrapper) ViewerStats(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*ViewerStats, error) {
	c.logger.Debug().Msg("anilist: Fetching stats")
	return c.realClientWrapper.ViewerStats(ctx, interceptors...)
}
