package anilist

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/Yamashou/gqlgenc/graphqljson"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/limiter"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type ClientWrapperInterface interface {
	UpdateEntry(ctx context.Context, mediaID *int, status *MediaListStatus, score *float64, progress *int, repeat *int, private *bool, notes *string, hiddenFromStatusLists *bool, startedAt *FuzzyDateInput, completedAt *FuzzyDateInput, interceptors ...clientv2.RequestInterceptor) (*UpdateEntry, error)
	UpdateMediaListEntry(ctx context.Context, mediaID *int, status *MediaListStatus, scoreRaw *int, progress *int, startedAt *FuzzyDateInput, completedAt *FuzzyDateInput, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntry, error)
	UpdateMediaListEntryProgress(ctx context.Context, mediaID *int, progress *int, totalEpisodes *int) error
	UpdateMediaListEntryStatus(ctx context.Context, mediaID *int, progress *int, status *MediaListStatus, scoreRaw *int, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntryStatus, error)
	DeleteEntry(ctx context.Context, mediaListEntryID *int, interceptors ...clientv2.RequestInterceptor) (*DeleteEntry, error)
	AnimeCollection(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*AnimeCollection, error)
	SearchAnimeShortMedia(ctx context.Context, page *int, perPage *int, sort []*MediaSort, search *string, status []*MediaStatus, interceptors ...clientv2.RequestInterceptor) (*SearchAnimeShortMedia, error)
	BasicMediaByMalID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BasicMediaByMalID, error)
	BasicMediaByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BasicMediaByID, error)
	BaseMediaByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseMediaByID, error)
	MediaDetailsByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*MediaDetailsByID, error)
	CompleteMediaByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*CompleteMediaByID, error)
	ListMedia(ctx context.Context, page *int, search *string, perPage *int, sort []*MediaSort, status []*MediaStatus, genres []*string, averageScoreGreater *int, season *MediaSeason, seasonYear *int, format *MediaFormat, interceptors ...clientv2.RequestInterceptor) (*ListMedia, error)
	ListRecentMedia(ctx context.Context, page *int, perPage *int, airingAtGreater *int, airingAtLesser *int, interceptors ...clientv2.RequestInterceptor) (*ListRecentMedia, error)
	GetViewer(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*GetViewer, error)
	AddMediaToPlanning(mIds []int, rateLimiter *limiter.Limiter, logger *zerolog.Logger) error
	MangaCollection(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*MangaCollection, error)
	SearchBaseManga(ctx context.Context, page *int, perPage *int, sort []*MediaSort, search *string, status []*MediaStatus, interceptors ...clientv2.RequestInterceptor) (*SearchBaseManga, error)
	BaseMangaByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseMangaByID, error)
	MangaDetailsByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*MangaDetailsByID, error)
	ListManga(ctx context.Context, page *int, search *string, perPage *int, sort []*MediaSort, status []*MediaStatus, genres []*string, averageScoreGreater *int, startDateGreater *string, startDateLesser *string, format *MediaFormat, isAdult *bool, interceptors ...clientv2.RequestInterceptor) (*ListManga, error)
	StudioDetails(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*StudioDetails, error)
	ViewerStats(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*ViewerStats, error)
}

type (
	// ClientWrapper is a wrapper around the AniList API client.
	ClientWrapper struct {
		Client *Client
		logger *zerolog.Logger
	}
)

// NewClientWrapper creates a new ClientWrapper with the given token.
// The token is used for authorization when making requests to the AniList API.
func NewClientWrapper(token string) *ClientWrapper {
	cw := &ClientWrapper{
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

	cw.Client.Client.CustomDo = cw.customDoFunc

	return cw
}

func (cw *ClientWrapper) AddMediaToPlanning(mIds []int, rateLimiter *limiter.Limiter, logger *zerolog.Logger) error {
	if len(mIds) == 0 {
		logger.Debug().Msg("anilist: No media added to planning list")
		return nil
	}
	if rateLimiter == nil {
		return errors.New("anilist: no rate limiter provided")
	}

	status := MediaListStatusPlanning

	scoreRaw := 0
	progress := 0

	wg := sync.WaitGroup{}
	for _, _id := range mIds {
		wg.Add(1)
		go func(id int) {
			rateLimiter.Wait()
			defer wg.Done()
			_, err := cw.Client.UpdateMediaListEntry(
				context.Background(),
				&id,
				&status,
				&scoreRaw,
				&progress,
				nil,
				nil,
			)
			if err != nil {
				logger.Error().Msg("anilist: An error occurred while adding media to planning list: " + err.Error())
			}
		}(_id)
	}
	wg.Wait()

	logger.Debug().Any("count", len(mIds)).Msg("anilist: Media added to planning list")

	return nil
}

func (cw *ClientWrapper) UpdateEntry(ctx context.Context, mediaID *int, status *MediaListStatus, score *float64, progress *int, repeat *int, private *bool, notes *string, hiddenFromStatusLists *bool, startedAt *FuzzyDateInput, completedAt *FuzzyDateInput, interceptors ...clientv2.RequestInterceptor) (*UpdateEntry, error) {
	return cw.Client.UpdateEntry(ctx, mediaID, status, score, progress, repeat, private, notes, hiddenFromStatusLists, startedAt, completedAt, interceptors...)
}
func (cw *ClientWrapper) UpdateMediaListEntry(ctx context.Context, mediaID *int, status *MediaListStatus, scoreRaw *int, progress *int, startedAt *FuzzyDateInput, completedAt *FuzzyDateInput, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntry, error) {
	cw.logger.Debug().Int("mediaId", *mediaID).Msg("anilist: Updating media list entry")
	return cw.Client.UpdateMediaListEntry(ctx, mediaID, status, scoreRaw, progress, startedAt, completedAt, interceptors...)
}
func (cw *ClientWrapper) UpdateMediaListEntryStatus(ctx context.Context, mediaID *int, progress *int, status *MediaListStatus, scoreRaw *int, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntryStatus, error) {
	return cw.Client.UpdateMediaListEntryStatus(ctx, mediaID, progress, status, scoreRaw, interceptors...)
}
func (cw *ClientWrapper) DeleteEntry(ctx context.Context, mediaListEntryID *int, interceptors ...clientv2.RequestInterceptor) (*DeleteEntry, error) {
	cw.logger.Debug().Int("entryId", *mediaListEntryID).Msg("anilist: Deleting media list entry")
	return cw.Client.DeleteEntry(ctx, mediaListEntryID, interceptors...)
}
func (cw *ClientWrapper) AnimeCollection(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*AnimeCollection, error) {
	cw.logger.Debug().Str("username", *userName).Msg("anilist: Fetching anime collection")
	return cw.Client.AnimeCollection(ctx, userName, interceptors...)
}
func (cw *ClientWrapper) SearchAnimeShortMedia(ctx context.Context, page *int, perPage *int, sort []*MediaSort, search *string, status []*MediaStatus, interceptors ...clientv2.RequestInterceptor) (*SearchAnimeShortMedia, error) {
	return cw.Client.SearchAnimeShortMedia(ctx, page, perPage, sort, search, status, interceptors...)
}
func (cw *ClientWrapper) BasicMediaByMalID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BasicMediaByMalID, error) {
	return cw.Client.BasicMediaByMalID(ctx, id, interceptors...)
}
func (cw *ClientWrapper) BasicMediaByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BasicMediaByID, error) {
	cw.logger.Debug().Int("mediaId", *id).Msg("anilist: Fetching anime")
	return cw.Client.BasicMediaByID(ctx, id, interceptors...)
}
func (cw *ClientWrapper) BaseMediaByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseMediaByID, error) {
	cw.logger.Debug().Int("mediaId", *id).Msg("anilist: Fetching anime")
	return cw.Client.BaseMediaByID(ctx, id, interceptors...)
}
func (cw *ClientWrapper) MediaDetailsByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*MediaDetailsByID, error) {
	cw.logger.Debug().Int("mediaId", *id).Msg("anilist: Fetching anime details")
	return cw.Client.MediaDetailsByID(ctx, id, interceptors...)
}
func (cw *ClientWrapper) CompleteMediaByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*CompleteMediaByID, error) {
	return cw.Client.CompleteMediaByID(ctx, id, interceptors...)
}
func (cw *ClientWrapper) ListMedia(ctx context.Context, page *int, search *string, perPage *int, sort []*MediaSort, status []*MediaStatus, genres []*string, averageScoreGreater *int, season *MediaSeason, seasonYear *int, format *MediaFormat, interceptors ...clientv2.RequestInterceptor) (*ListMedia, error) {
	cw.logger.Debug().Msg("anilist: Fetching media list")
	return cw.Client.ListMedia(ctx, page, search, perPage, sort, status, genres, averageScoreGreater, season, seasonYear, format, interceptors...)
}
func (cw *ClientWrapper) ListRecentMedia(ctx context.Context, page *int, perPage *int, airingAtGreater *int, airingAtLesser *int, interceptors ...clientv2.RequestInterceptor) (*ListRecentMedia, error) {
	cw.logger.Debug().Msg("anilist: Fetching recent media list")
	return cw.Client.ListRecentMedia(ctx, page, perPage, airingAtGreater, airingAtLesser, interceptors...)
}
func (cw *ClientWrapper) GetViewer(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*GetViewer, error) {
	cw.logger.Debug().Msg("anilist: Fetching viewer")
	return cw.Client.GetViewer(ctx, interceptors...)
}

func (cw *ClientWrapper) MangaCollection(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*MangaCollection, error) {
	cw.logger.Debug().Msg("anilist: Fetching manga collection")
	return cw.Client.MangaCollection(ctx, userName, interceptors...)
}
func (cw *ClientWrapper) SearchBaseManga(ctx context.Context, page *int, perPage *int, sort []*MediaSort, search *string, status []*MediaStatus, interceptors ...clientv2.RequestInterceptor) (*SearchBaseManga, error) {
	return cw.Client.SearchBaseManga(ctx, page, perPage, sort, search, status, interceptors...)
}
func (cw *ClientWrapper) BaseMangaByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseMangaByID, error) {
	cw.logger.Debug().Int("mediaId", *id).Msg("anilist: Fetching manga")
	return cw.Client.BaseMangaByID(ctx, id, interceptors...)
}
func (cw *ClientWrapper) MangaDetailsByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*MangaDetailsByID, error) {
	cw.logger.Debug().Int("mediaId", *id).Msg("anilist: Fetching manga details")
	return cw.Client.MangaDetailsByID(ctx, id, interceptors...)
}
func (cw *ClientWrapper) ListManga(ctx context.Context, page *int, search *string, perPage *int, sort []*MediaSort, status []*MediaStatus, genres []*string, averageScoreGreater *int, startDateGreater *string, startDateLesser *string, format *MediaFormat, isAdult *bool, interceptors ...clientv2.RequestInterceptor) (*ListManga, error) {
	cw.logger.Debug().Msg("anilist: Fetching manga list")
	return cw.Client.ListManga(ctx, page, search, perPage, sort, status, genres, averageScoreGreater, startDateGreater, startDateLesser, format, isAdult, interceptors...)
}

func (cw *ClientWrapper) StudioDetails(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*StudioDetails, error) {
	cw.logger.Debug().Int("studioId", *id).Msg("anilist: Fetching studio details")
	return cw.Client.StudioDetails(ctx, id, interceptors...)
}

func (cw *ClientWrapper) ViewerStats(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*ViewerStats, error) {
	cw.logger.Debug().Msg("anilist: Fetching stats")
	return cw.Client.ViewerStats(ctx, interceptors...)
}

// customDoFunc is a custom request interceptor function that handles rate limiting and retries.
func (cw *ClientWrapper) customDoFunc(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}) (err error) {
	var rlRemainingStr string

	reqTime := time.Now()
	defer func() {
		timeSince := time.Since(reqTime)
		formattedDur := timeSince.Truncate(time.Millisecond).String()
		if err != nil {
			cw.logger.Error().Str("duration", formattedDur).Str("rlr", rlRemainingStr).Err(err).Msg("anilist: Failed Request")
		} else {
			if timeSince > 900*time.Millisecond {
				cw.logger.Warn().Str("rtt", formattedDur).Str("rlr", rlRemainingStr).Msg("anilist: Successful Request (slow)")
			} else {
				cw.logger.Info().Str("rtt", formattedDur).Str("rlr", rlRemainingStr).Msg("anilist: Successful Request")
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
			cw.logger.Warn().Msgf("anilist: Rate limited, retrying in %d seconds", rlRetryAfter+1)
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
