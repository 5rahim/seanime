package anilist

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"seanime/internal/events"
	"seanime/internal/util"
	"strconv"
	"time"

	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/Yamashou/gqlgenc/graphqljson"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
)

var (
	// ErrNotAuthenticated is returned when trying to access an Anilist API endpoint that requires authentication,
	// but the client is not authenticated.
	ErrNotAuthenticated = errors.New("not authenticated")
)

type AnilistClient interface {
	IsAuthenticated() bool
	AnimeCollection(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*AnimeCollection, error)
	AnimeCollectionWithRelations(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*AnimeCollectionWithRelations, error)
	BaseAnimeByMalID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseAnimeByMalID, error)
	BaseAnimeByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseAnimeByID, error)
	SearchBaseAnimeByIds(ctx context.Context, ids []*int, page *int, perPage *int, status []*MediaStatus, inCollection *bool, sort []*MediaSort, season *MediaSeason, year *int, genre *string, format *MediaFormat, interceptors ...clientv2.RequestInterceptor) (*SearchBaseAnimeByIds, error)
	CompleteAnimeByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*CompleteAnimeByID, error)
	AnimeDetailsByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*AnimeDetailsByID, error)
	ListAnime(ctx context.Context, page *int, search *string, perPage *int, sort []*MediaSort, status []*MediaStatus, genres []*string, averageScoreGreater *int, season *MediaSeason, seasonYear *int, format *MediaFormat, isAdult *bool, interceptors ...clientv2.RequestInterceptor) (*ListAnime, error)
	ListRecentAnime(ctx context.Context, page *int, perPage *int, airingAtGreater *int, airingAtLesser *int, notYetAired *bool, interceptors ...clientv2.RequestInterceptor) (*ListRecentAnime, error)
	UpdateMediaListEntry(ctx context.Context, mediaID *int, status *MediaListStatus, scoreRaw *int, progress *int, startedAt *FuzzyDateInput, completedAt *FuzzyDateInput, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntry, error)
	UpdateMediaListEntryProgress(ctx context.Context, mediaID *int, progress *int, status *MediaListStatus, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntryProgress, error)
	UpdateMediaListEntryRepeat(ctx context.Context, mediaID *int, repeat *int, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntryRepeat, error)
	DeleteEntry(ctx context.Context, mediaListEntryID *int, interceptors ...clientv2.RequestInterceptor) (*DeleteEntry, error)
	MangaCollection(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*MangaCollection, error)
	SearchBaseManga(ctx context.Context, page *int, perPage *int, sort []*MediaSort, search *string, status []*MediaStatus, interceptors ...clientv2.RequestInterceptor) (*SearchBaseManga, error)
	BaseMangaByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseMangaByID, error)
	MangaDetailsByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*MangaDetailsByID, error)
	ListManga(ctx context.Context, page *int, search *string, perPage *int, sort []*MediaSort, status []*MediaStatus, genres []*string, averageScoreGreater *int, startDateGreater *string, startDateLesser *string, format *MediaFormat, countryOfOrigin *string, isAdult *bool, interceptors ...clientv2.RequestInterceptor) (*ListManga, error)
	ViewerStats(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*ViewerStats, error)
	StudioDetails(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*StudioDetails, error)
	GetViewer(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*GetViewer, error)
	AnimeAiringSchedule(ctx context.Context, ids []*int, season *MediaSeason, seasonYear *int, previousSeason *MediaSeason, previousSeasonYear *int, nextSeason *MediaSeason, nextSeasonYear *int, interceptors ...clientv2.RequestInterceptor) (*AnimeAiringSchedule, error)
	AnimeAiringScheduleRaw(ctx context.Context, ids []*int, interceptors ...clientv2.RequestInterceptor) (*AnimeAiringScheduleRaw, error)
}

type (
	// AnilistClientImpl is a wrapper around the AniList API client.
	AnilistClientImpl struct {
		Client *Client
		logger *zerolog.Logger
		token  string // The token used for authentication with the AniList API
	}
)

// NewAnilistClient creates a new AnilistClientImpl with the given token.
// The token is used for authorization when making requests to the AniList API.
func NewAnilistClient(token string) *AnilistClientImpl {
	ac := &AnilistClientImpl{
		token: token,
		Client: &Client{
			Client: clientv2.NewClient(http.DefaultClient, "https://graphql.anilist.co", nil,
				func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}, next clientv2.RequestInterceptorFunc) error {
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Accept", "application/json")
					if len(token) > 0 {
						req.Header.Set("Authorization", "Bearer "+token)
					}
					return next(ctx, req, gqlInfo, res)
				}),
		},
		logger: util.NewLogger(),
	}

	ac.Client.Client.CustomDo = ac.customDoFunc

	return ac
}

func (ac *AnilistClientImpl) IsAuthenticated() bool {
	if ac.Client == nil || ac.Client.Client == nil {
		return false
	}
	if len(ac.token) == 0 {
		return false
	}
	// If the token is not empty, we are authenticated
	return true
}

////////////////////////////////
// Authenticated
////////////////////////////////

func (ac *AnilistClientImpl) UpdateMediaListEntry(ctx context.Context, mediaID *int, status *MediaListStatus, scoreRaw *int, progress *int, startedAt *FuzzyDateInput, completedAt *FuzzyDateInput, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntry, error) {
	if !ac.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}
	ac.logger.Debug().Int("mediaId", *mediaID).Msg("anilist: Updating media list entry")
	return ac.Client.UpdateMediaListEntry(ctx, mediaID, status, scoreRaw, progress, startedAt, completedAt, interceptors...)
}

func (ac *AnilistClientImpl) UpdateMediaListEntryProgress(ctx context.Context, mediaID *int, progress *int, status *MediaListStatus, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntryProgress, error) {
	if !ac.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}
	ac.logger.Debug().Int("mediaId", *mediaID).Msg("anilist: Updating media list entry progress")
	return ac.Client.UpdateMediaListEntryProgress(ctx, mediaID, progress, status, interceptors...)
}

func (ac *AnilistClientImpl) UpdateMediaListEntryRepeat(ctx context.Context, mediaID *int, repeat *int, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntryRepeat, error) {
	if !ac.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}
	ac.logger.Debug().Int("mediaId", *mediaID).Msg("anilist: Updating media list entry repeat")
	return ac.Client.UpdateMediaListEntryRepeat(ctx, mediaID, repeat, interceptors...)
}

func (ac *AnilistClientImpl) DeleteEntry(ctx context.Context, mediaListEntryID *int, interceptors ...clientv2.RequestInterceptor) (*DeleteEntry, error) {
	if !ac.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}
	ac.logger.Debug().Int("entryId", *mediaListEntryID).Msg("anilist: Deleting media list entry")
	return ac.Client.DeleteEntry(ctx, mediaListEntryID, interceptors...)
}

func (ac *AnilistClientImpl) AnimeCollection(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*AnimeCollection, error) {
	if !ac.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}
	ac.logger.Debug().Msg("anilist: Fetching anime collection")
	return ac.Client.AnimeCollection(ctx, userName, interceptors...)
}

func (ac *AnilistClientImpl) AnimeCollectionWithRelations(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*AnimeCollectionWithRelations, error) {
	if !ac.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}
	ac.logger.Debug().Msg("anilist: Fetching anime collection with relations")
	return ac.Client.AnimeCollectionWithRelations(ctx, userName, interceptors...)
}

func (ac *AnilistClientImpl) GetViewer(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*GetViewer, error) {
	if !ac.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}
	ac.logger.Debug().Msg("anilist: Fetching viewer")
	return ac.Client.GetViewer(ctx, interceptors...)
}

func (ac *AnilistClientImpl) MangaCollection(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*MangaCollection, error) {
	if !ac.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}
	ac.logger.Debug().Msg("anilist: Fetching manga collection")
	return ac.Client.MangaCollection(ctx, userName, interceptors...)
}

func (ac *AnilistClientImpl) ViewerStats(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*ViewerStats, error) {
	if !ac.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}
	ac.logger.Debug().Msg("anilist: Fetching stats")
	return ac.Client.ViewerStats(ctx, interceptors...)
}

////////////////////////////////
// Not authenticated
////////////////////////////////

func (ac *AnilistClientImpl) BaseAnimeByMalID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseAnimeByMalID, error) {
	return ac.Client.BaseAnimeByMalID(ctx, id, interceptors...)
}

func (ac *AnilistClientImpl) BaseAnimeByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseAnimeByID, error) {
	ac.logger.Debug().Int("mediaId", *id).Msg("anilist: Fetching anime")
	return ac.Client.BaseAnimeByID(ctx, id, interceptors...)
}

func (ac *AnilistClientImpl) AnimeDetailsByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*AnimeDetailsByID, error) {
	ac.logger.Debug().Int("mediaId", *id).Msg("anilist: Fetching anime details")
	return ac.Client.AnimeDetailsByID(ctx, id, interceptors...)
}

func (ac *AnilistClientImpl) CompleteAnimeByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*CompleteAnimeByID, error) {
	ac.logger.Debug().Int("mediaId", *id).Msg("anilist: Fetching complete media")
	return ac.Client.CompleteAnimeByID(ctx, id, interceptors...)
}

func (ac *AnilistClientImpl) ListAnime(ctx context.Context, page *int, search *string, perPage *int, sort []*MediaSort, status []*MediaStatus, genres []*string, averageScoreGreater *int, season *MediaSeason, seasonYear *int, format *MediaFormat, isAdult *bool, interceptors ...clientv2.RequestInterceptor) (*ListAnime, error) {
	ac.logger.Debug().Msg("anilist: Fetching media list")
	return ac.Client.ListAnime(ctx, page, search, perPage, sort, status, genres, averageScoreGreater, season, seasonYear, format, isAdult, interceptors...)
}

func (ac *AnilistClientImpl) ListRecentAnime(ctx context.Context, page *int, perPage *int, airingAtGreater *int, airingAtLesser *int, notYetAired *bool, interceptors ...clientv2.RequestInterceptor) (*ListRecentAnime, error) {
	ac.logger.Debug().Msg("anilist: Fetching recent media list")
	return ac.Client.ListRecentAnime(ctx, page, perPage, airingAtGreater, airingAtLesser, notYetAired, interceptors...)
}

func (ac *AnilistClientImpl) SearchBaseManga(ctx context.Context, page *int, perPage *int, sort []*MediaSort, search *string, status []*MediaStatus, interceptors ...clientv2.RequestInterceptor) (*SearchBaseManga, error) {
	ac.logger.Debug().Msg("anilist: Searching manga")
	return ac.Client.SearchBaseManga(ctx, page, perPage, sort, search, status, interceptors...)
}

func (ac *AnilistClientImpl) BaseMangaByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseMangaByID, error) {
	ac.logger.Debug().Int("mediaId", *id).Msg("anilist: Fetching manga")
	return ac.Client.BaseMangaByID(ctx, id, interceptors...)
}

func (ac *AnilistClientImpl) MangaDetailsByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*MangaDetailsByID, error) {
	ac.logger.Debug().Int("mediaId", *id).Msg("anilist: Fetching manga details")
	return ac.Client.MangaDetailsByID(ctx, id, interceptors...)
}

func (ac *AnilistClientImpl) ListManga(ctx context.Context, page *int, search *string, perPage *int, sort []*MediaSort, status []*MediaStatus, genres []*string, averageScoreGreater *int, startDateGreater *string, startDateLesser *string, format *MediaFormat, countryOfOrigin *string, isAdult *bool, interceptors ...clientv2.RequestInterceptor) (*ListManga, error) {
	ac.logger.Debug().Msg("anilist: Fetching manga list")
	return ac.Client.ListManga(ctx, page, search, perPage, sort, status, genres, averageScoreGreater, startDateGreater, startDateLesser, format, countryOfOrigin, isAdult, interceptors...)
}

func (ac *AnilistClientImpl) StudioDetails(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*StudioDetails, error) {
	ac.logger.Debug().Int("studioId", *id).Msg("anilist: Fetching studio details")
	return ac.Client.StudioDetails(ctx, id, interceptors...)
}

func (ac *AnilistClientImpl) SearchBaseAnimeByIds(ctx context.Context, ids []*int, page *int, perPage *int, status []*MediaStatus, inCollection *bool, sort []*MediaSort, season *MediaSeason, year *int, genre *string, format *MediaFormat, interceptors ...clientv2.RequestInterceptor) (*SearchBaseAnimeByIds, error) {
	ac.logger.Debug().Msg("anilist: Searching anime by ids")
	return ac.Client.SearchBaseAnimeByIds(ctx, ids, page, perPage, status, inCollection, sort, season, year, genre, format, interceptors...)
}

func (ac *AnilistClientImpl) AnimeAiringSchedule(ctx context.Context, ids []*int, season *MediaSeason, seasonYear *int, previousSeason *MediaSeason, previousSeasonYear *int, nextSeason *MediaSeason, nextSeasonYear *int, interceptors ...clientv2.RequestInterceptor) (*AnimeAiringSchedule, error) {
	ac.logger.Debug().Msg("anilist: Fetching schedule")
	return ac.Client.AnimeAiringSchedule(ctx, ids, season, seasonYear, previousSeason, previousSeasonYear, nextSeason, nextSeasonYear, interceptors...)
}

func (ac *AnilistClientImpl) AnimeAiringScheduleRaw(ctx context.Context, ids []*int, interceptors ...clientv2.RequestInterceptor) (*AnimeAiringScheduleRaw, error) {
	ac.logger.Debug().Msg("anilist: Fetching schedule")
	return ac.Client.AnimeAiringScheduleRaw(ctx, ids, interceptors...)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var sentRateLimitWarningTime = time.Now().Add(-10 * time.Second)

// customDoFunc is a custom request interceptor function that handles rate limiting and retries.
func (ac *AnilistClientImpl) customDoFunc(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}) (err error) {
	var rlRemainingStr string

	reqTime := time.Now()
	defer func() {
		timeSince := time.Since(reqTime)
		formattedDur := timeSince.Truncate(time.Millisecond).String()
		if err != nil {
			ac.logger.Error().Str("duration", formattedDur).Str("rlr", rlRemainingStr).Err(err).Msg("anilist: Failed Request")
		} else {
			if timeSince > 900*time.Millisecond {
				ac.logger.Warn().Str("rtt", formattedDur).Str("rlr", rlRemainingStr).Msg("anilist: Successful Request (slow)")
			} else {
				ac.logger.Info().Str("rtt", formattedDur).Str("rlr", rlRemainingStr).Msg("anilist: Successful Request")
			}
		}
	}()

	client := http.DefaultClient
	var resp *http.Response

	retryCount := 2

	for i := 0; i < retryCount; i++ {

		// Reset response body for retry
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

		// Recreate the request body if it was read in a previous attempt
		if req.GetBody != nil {
			newBody, err := req.GetBody()
			if err != nil {
				return fmt.Errorf("failed to get request body: %w", err)
			}
			req.Body = newBody
		}

		resp, err = client.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}

		rlRemainingStr = resp.Header.Get("X-Ratelimit-Remaining")
		rlRetryAfterStr := resp.Header.Get("Retry-After")
		//println("Remaining:", rlRemainingStr, " | RetryAfter:", rlRetryAfterStr)

		// If we have a rate limit, sleep for the time
		rlRetryAfter, err := strconv.Atoi(rlRetryAfterStr)
		if err == nil {
			ac.logger.Warn().Msgf("anilist: Rate limited, retrying in %d seconds", rlRetryAfter+1)
			if time.Since(sentRateLimitWarningTime) > 10*time.Second {
				events.GlobalWSEventManager.SendEvent(events.WarningToast, "anilist: Rate limited, retrying in "+strconv.Itoa(rlRetryAfter+1)+" seconds")
				sentRateLimitWarningTime = time.Now()
			}
			select {
			case <-time.After(time.Duration(rlRetryAfter+1) * time.Second):
				continue
			}
		}

		if rlRemainingStr == "" {
			select {
			case <-time.After(5 * time.Second):
				continue
			}
		}

		break
	}

	defer resp.Body.Close()

	if resp.Header.Get("Content-Encoding") == "gzip" {
		resp.Body, err = gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("gzip decode failed: %w", err)
		}
	}

	var body []byte
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	err = parseResponse(body, resp.StatusCode, res)
	return
}

func parseResponse(body []byte, httpCode int, result interface{}) error {
	errResponse := &clientv2.ErrorResponse{}
	isKOCode := httpCode < 200 || 299 < httpCode
	if isKOCode {
		errResponse.NetworkError = &clientv2.HTTPError{
			Code:    httpCode,
			Message: fmt.Sprintf("Response body %s", string(body)),
		}
	}

	// some servers return a graphql error with a non OK http code, try anyway to parse the body
	if err := unmarshal(body, result); err != nil {
		var gqlErr *clientv2.GqlErrorList
		if errors.As(err, &gqlErr) {
			errResponse.GqlErrors = &gqlErr.Errors
		} else if !isKOCode {
			return err
		}
	}

	if errResponse.HasErrors() {
		return errResponse
	}

	return nil
}

// response is a GraphQL layer response from a handler.
type response struct {
	Data   json.RawMessage `json:"data"`
	Errors json.RawMessage `json:"errors"`
}

func unmarshal(data []byte, res interface{}) error {
	ParseDataWhenErrors := false
	resp := response{}
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("failed to decode data %s: %w", string(data), err)
	}

	var err error
	if resp.Errors != nil && len(resp.Errors) > 0 {
		// try to parse standard graphql error
		err = &clientv2.GqlErrorList{}
		if e := json.Unmarshal(data, err); e != nil {
			return fmt.Errorf("faild to parse graphql errors. Response content %s - %w", string(data), e)
		}

		// if ParseDataWhenErrors is true, try to parse data as well
		if !ParseDataWhenErrors {
			return err
		}
	}

	if errData := graphqljson.UnmarshalData(resp.Data, res); errData != nil {
		// if ParseDataWhenErrors is true, and we failed to unmarshal data, return the actual error
		if ParseDataWhenErrors {
			return err
		}

		return fmt.Errorf("failed to decode data into response %s: %w", string(data), errData)
	}

	return err
}
