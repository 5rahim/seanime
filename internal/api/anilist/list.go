package anilist

import (
	"fmt"
	"seanime/internal/hook"

	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

func ListMissedSequels(
	animeCollectionWithRelations *AnimeCollectionWithRelations,
	logger *zerolog.Logger,
	token string,
) (ret []*BaseAnime, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	variables := map[string]interface{}{}
	variables["page"] = 1
	variables["perPage"] = 50

	ids := make(map[int]struct{})
	for _, list := range animeCollectionWithRelations.GetMediaListCollection().GetLists() {
		if list.Status == nil || !(*list.Status == MediaListStatusCompleted || *list.Status == MediaListStatusRepeating || *list.Status == MediaListStatusPaused) || list.Entries == nil {
			continue
		}
		for _, entry := range list.Entries {
			if _, ok := ids[entry.GetMedia().GetID()]; !ok {
				edges := entry.GetMedia().GetRelations().GetEdges()
				var sequel *BaseAnime
				for _, edge := range edges {
					if edge.GetRelationType() != nil && *edge.GetRelationType() == MediaRelationSequel {
						sequel = edge.GetNode()
						break
					}
				}

				if sequel == nil {
					continue
				}

				// Check if sequel is already in the list
				_, found := animeCollectionWithRelations.FindAnime(sequel.GetID())
				if found {
					continue
				}

				if *sequel.GetStatus() == MediaStatusFinished || *sequel.GetStatus() == MediaStatusReleasing {
					ids[sequel.GetID()] = struct{}{}
				}
			}

		}
	}

	idsSlice := make([]int, 0, len(ids))
	for id := range ids {
		idsSlice = append(idsSlice, id)
	}

	if len(idsSlice) == 0 {
		return []*BaseAnime{}, nil
	}

	if len(idsSlice) > 10 {
		idsSlice = idsSlice[:10]
	}

	variables["ids"] = idsSlice
	variables["inCollection"] = false
	variables["sort"] = MediaSortStartDateDesc

	// Event
	reqEvent := &ListMissedSequelsRequestedEvent{
		AnimeCollectionWithRelations: animeCollectionWithRelations,
		Variables:                    variables,
		List:                         make([]*BaseAnime, 0),
		Query:                        SearchBaseAnimeByIdsDocument,
	}
	err = hook.GlobalHookManager.OnListMissedSequelsRequested().Trigger(reqEvent)
	if err != nil {
		return nil, err
	}

	// If the hook prevented the default behavior, return the data
	if reqEvent.DefaultPrevented {
		return reqEvent.List, nil
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"query":     reqEvent.Query,
		"variables": reqEvent.Variables,
	})
	if err != nil {
		return nil, err
	}

	data, err := customQuery(requestBody, logger, token)
	if err != nil {
		return nil, err
	}

	m, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var searchRes *SearchBaseAnimeByIds
	if err := json.Unmarshal(m, &searchRes); err != nil {
		return nil, err
	}

	if searchRes == nil || searchRes.Page == nil || searchRes.Page.Media == nil {
		return nil, fmt.Errorf("no data found")
	}

	// Event
	event := &ListMissedSequelsEvent{
		List: searchRes.Page.Media,
	}
	err = hook.GlobalHookManager.OnListMissedSequels().Trigger(event)
	if err != nil {
		return nil, err
	}

	return event.List, nil
}

func ListAnimeM(
	Page *int,
	Search *string,
	PerPage *int,
	Sort []*MediaSort,
	Status []*MediaStatus,
	Genres []*string,
	AverageScoreGreater *int,
	Season *MediaSeason,
	SeasonYear *int,
	Format *MediaFormat,
	IsAdult *bool,
	logger *zerolog.Logger,
	token string,
) (*ListAnime, error) {

	variables := map[string]interface{}{}
	if Page != nil {
		variables["page"] = *Page
	}
	if Search != nil {
		variables["search"] = *Search
	}
	if PerPage != nil {
		variables["perPage"] = *PerPage
	}
	if Sort != nil {
		variables["sort"] = Sort
	}
	if Status != nil {
		variables["status"] = Status
	}
	if Genres != nil {
		variables["genres"] = Genres
	}
	if AverageScoreGreater != nil {
		variables["averageScore_greater"] = *AverageScoreGreater
	}
	if Season != nil {
		variables["season"] = *Season
	}
	if SeasonYear != nil {
		variables["seasonYear"] = *SeasonYear
	}
	if Format != nil {
		variables["format"] = *Format
	}
	if IsAdult != nil {
		variables["isAdult"] = *IsAdult
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"query":     ListAnimeDocument,
		"variables": variables,
	})
	if err != nil {
		return nil, err
	}

	data, err := customQuery(requestBody, logger, token)
	if err != nil {
		return nil, err
	}

	var listMediaF ListAnime
	m, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(m, &listMediaF); err != nil {
		return nil, err
	}

	return &listMediaF, nil
}

func ListMangaM(
	Page *int,
	Search *string,
	PerPage *int,
	Sort []*MediaSort,
	Status []*MediaStatus,
	Genres []*string,
	AverageScoreGreater *int,
	Year *int,
	Format *MediaFormat,
	CountryOfOrigin *string,
	IsAdult *bool,
	logger *zerolog.Logger,
	token string,
) (*ListManga, error) {

	variables := map[string]interface{}{}
	if Page != nil {
		variables["page"] = *Page
	}
	if Search != nil {
		variables["search"] = *Search
	}
	if PerPage != nil {
		variables["perPage"] = *PerPage
	}
	if Sort != nil {
		variables["sort"] = Sort
	}
	if Status != nil {
		variables["status"] = Status
	}
	if Genres != nil {
		variables["genres"] = Genres
	}
	if AverageScoreGreater != nil {
		variables["averageScore_greater"] = *AverageScoreGreater * 10
	}
	if Year != nil {
		variables["startDate_greater"] = lo.ToPtr(fmt.Sprintf("%d0000", *Year))
		variables["startDate_lesser"] = lo.ToPtr(fmt.Sprintf("%d0000", *Year+1))
	}
	if Format != nil {
		variables["format"] = *Format
	}
	if CountryOfOrigin != nil {
		variables["countryOfOrigin"] = *CountryOfOrigin
	}
	if IsAdult != nil {
		variables["isAdult"] = *IsAdult
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"query":     ListMangaDocument,
		"variables": variables,
	})
	if err != nil {
		return nil, err
	}

	data, err := customQuery(requestBody, logger, token)
	if err != nil {
		return nil, err
	}

	var listMediaF ListManga
	m, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(m, &listMediaF); err != nil {
		return nil, err
	}

	return &listMediaF, nil
}

func ListRecentAiringAnimeM(
	Page *int,
	Search *string,
	PerPage *int,
	AiringAtGreater *int,
	AiringAtLesser *int,
	NotYetAired *bool,
	Sort []*AiringSort,
	logger *zerolog.Logger,
	token string,
) (*ListRecentAnime, error) {

	variables := map[string]interface{}{}
	if Page != nil {
		variables["page"] = *Page
	}
	if Search != nil {
		variables["search"] = *Search
	}
	if PerPage != nil {
		variables["perPage"] = *PerPage
	}
	if AiringAtGreater != nil {
		variables["airingAt_greater"] = *AiringAtGreater
	}
	if AiringAtLesser != nil {
		variables["airingAt_lesser"] = *AiringAtLesser
	}
	if NotYetAired != nil {
		variables["notYetAired"] = *NotYetAired
	}
	if Sort != nil {
		variables["sort"] = Sort
	} else {
		variables["sort"] = []*AiringSort{lo.ToPtr(AiringSortTimeDesc)}
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"query":     ListRecentAiringAnimeQuery,
		"variables": variables,
	})
	if err != nil {
		return nil, err
	}

	data, err := customQuery(requestBody, logger, token)
	if err != nil {
		return nil, err
	}

	var listMediaF ListRecentAnime
	m, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(m, &listMediaF); err != nil {
		return nil, err
	}

	return &listMediaF, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func ListAnimeCacheKey(
	Page *int,
	Search *string,
	PerPage *int,
	Sort []*MediaSort,
	Status []*MediaStatus,
	Genres []*string,
	AverageScoreGreater *int,
	Season *MediaSeason,
	SeasonYear *int,
	Format *MediaFormat,
	IsAdult *bool,
) string {

	key := "ListAnime"
	if Page != nil {
		key += fmt.Sprintf("_%d", *Page)
	}
	if Search != nil {
		key += fmt.Sprintf("_%s", *Search)
	}
	if PerPage != nil {
		key += fmt.Sprintf("_%d", *PerPage)
	}
	if Sort != nil {
		key += fmt.Sprintf("_%v", Sort)
	}
	if Status != nil {
		key += fmt.Sprintf("_%v", Status)
	}
	if Genres != nil {
		key += fmt.Sprintf("_%v", Genres)
	}
	if AverageScoreGreater != nil {
		key += fmt.Sprintf("_%d", *AverageScoreGreater)
	}
	if Season != nil {
		key += fmt.Sprintf("_%s", *Season)
	}
	if SeasonYear != nil {
		key += fmt.Sprintf("_%d", *SeasonYear)
	}
	if Format != nil {
		key += fmt.Sprintf("_%s", *Format)
	}
	if IsAdult != nil {
		key += fmt.Sprintf("_%t", *IsAdult)
	}

	return key

}
func ListMangaCacheKey(
	Page *int,
	Search *string,
	PerPage *int,
	Sort []*MediaSort,
	Status []*MediaStatus,
	Genres []*string,
	AverageScoreGreater *int,
	Season *MediaSeason,
	SeasonYear *int,
	Format *MediaFormat,
	CountryOfOrigin *string,
	IsAdult *bool,
) string {

	key := "ListAnime"
	if Page != nil {
		key += fmt.Sprintf("_%d", *Page)
	}
	if Search != nil {
		key += fmt.Sprintf("_%s", *Search)
	}
	if PerPage != nil {
		key += fmt.Sprintf("_%d", *PerPage)
	}
	if Sort != nil {
		key += fmt.Sprintf("_%v", Sort)
	}
	if Status != nil {
		key += fmt.Sprintf("_%v", Status)
	}
	if Genres != nil {
		key += fmt.Sprintf("_%v", Genres)
	}
	if AverageScoreGreater != nil {
		key += fmt.Sprintf("_%d", *AverageScoreGreater)
	}
	if Season != nil {
		key += fmt.Sprintf("_%s", *Season)
	}
	if SeasonYear != nil {
		key += fmt.Sprintf("_%d", *SeasonYear)
	}
	if Format != nil {
		key += fmt.Sprintf("_%s", *Format)
	}
	if CountryOfOrigin != nil {
		key += fmt.Sprintf("_%s", *CountryOfOrigin)
	}
	if IsAdult != nil {
		key += fmt.Sprintf("_%t", *IsAdult)
	}

	return key

}

const ListRecentAiringAnimeQuery = `query ListRecentAnime ($page: Int, $perPage: Int, $airingAt_greater: Int, $airingAt_lesser: Int, $sort: [AiringSort], $notYetAired: Boolean = false) {
	Page(page: $page, perPage: $perPage) {
		pageInfo {
			hasNextPage
			total
			perPage
			currentPage
			lastPage
		}
		airingSchedules(notYetAired: $notYetAired, sort: $sort, airingAt_greater: $airingAt_greater, airingAt_lesser: $airingAt_lesser) {
			id
			airingAt
			episode
			timeUntilAiring
			media {
				... baseAnime
			}
		}
	}
}
fragment baseAnime on Media {
	id
	idMal
	siteUrl
	status(version: 2)
	season
	type
	format
	bannerImage
	episodes
	synonyms
	isAdult
	countryOfOrigin
	meanScore
	description
	genres
	duration
	trailer {
		id
		site
		thumbnail
	}
	title {
		userPreferred
		romaji
		english
		native
	}
	coverImage {
		extraLarge
		large
		medium
		color
	}
	startDate {
		year
		month
		day
	}
	endDate {
		year
		month
		day
	}
	nextAiringEpisode {
		airingAt
		timeUntilAiring
		episode
	}
}
  `
