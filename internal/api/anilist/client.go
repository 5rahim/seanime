package anilist

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"seanime/internal/events"
	"seanime/internal/util"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goccy/go-json"
	"github.com/gqlgo/gqlgenc/clientv2"
	"github.com/gqlgo/gqlgenc/graphqljson"
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
	AnimeCollectionTags(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*AnimeCollectionTags, error)
	AnimeCollectionWithRelations(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*AnimeCollectionWithRelations, error)
	BaseAnimeByMalID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseAnimeByMalID, error)
	BaseAnimeByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseAnimeByID, error)
	SearchBaseAnimeByIds(ctx context.Context, ids []*int, page *int, perPage *int, status []*MediaStatus, inCollection *bool, sort []*MediaSort, season *MediaSeason, year *int, genre *string, format *MediaFormat, interceptors ...clientv2.RequestInterceptor) (*SearchBaseAnimeByIds, error)
	CompleteAnimeByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*CompleteAnimeByID, error)
	AnimeDetailsByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*AnimeDetailsByID, error)
	ListAnime(ctx context.Context, page *int, search *string, perPage *int, sort []*MediaSort, status []*MediaStatus, genres []*string, tags []*string, averageScoreGreater *int, season *MediaSeason, seasonYear *int, format *MediaFormat, isAdult *bool, interceptors ...clientv2.RequestInterceptor) (*ListAnime, error)
	ListRecentAnime(ctx context.Context, page *int, perPage *int, airingAtGreater *int, airingAtLesser *int, notYetAired *bool, interceptors ...clientv2.RequestInterceptor) (*ListRecentAnime, error)
	UpdateMediaListEntry(ctx context.Context, mediaID *int, status *MediaListStatus, scoreRaw *int, progress *int, startedAt *FuzzyDateInput, completedAt *FuzzyDateInput, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntry, error)
	UpdateMediaListEntryProgress(ctx context.Context, mediaID *int, progress *int, status *MediaListStatus, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntryProgress, error)
	UpdateMediaListEntryRepeat(ctx context.Context, mediaID *int, repeat *int, interceptors ...clientv2.RequestInterceptor) (*UpdateMediaListEntryRepeat, error)
	DeleteEntry(ctx context.Context, mediaListEntryID *int, interceptors ...clientv2.RequestInterceptor) (*DeleteEntry, error)
	MangaCollection(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*MangaCollection, error)
	MangaCollectionTags(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*MangaCollectionTags, error)
	SearchBaseManga(ctx context.Context, page *int, perPage *int, sort []*MediaSort, search *string, status []*MediaStatus, interceptors ...clientv2.RequestInterceptor) (*SearchBaseManga, error)
	BaseMangaByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseMangaByID, error)
	MangaDetailsByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*MangaDetailsByID, error)
	ListManga(ctx context.Context, page *int, search *string, perPage *int, sort []*MediaSort, status []*MediaStatus, genres []*string, tags []*string, averageScoreGreater *int, startDateGreater *string, startDateLesser *string, format *MediaFormat, countryOfOrigin *string, isAdult *bool, interceptors ...clientv2.RequestInterceptor) (*ListManga, error)
	ViewerStats(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*ViewerStats, error)
	StudioDetails(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*StudioDetails, error)
	GetViewer(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*GetViewer, error)
	AnimeAiringSchedule(ctx context.Context, ids []*int, season *MediaSeason, seasonYear *int, previousSeason *MediaSeason, previousSeasonYear *int, nextSeason *MediaSeason, nextSeasonYear *int, interceptors ...clientv2.RequestInterceptor) (*AnimeAiringSchedule, error)
	AnimeAiringScheduleRaw(ctx context.Context, ids []*int, interceptors ...clientv2.RequestInterceptor) (*AnimeAiringScheduleRaw, error)
	GetCacheDir() string
	CustomQuery(body []byte, logger *zerolog.Logger, token ...string) (interface{}, error)
}

type (
	// AnilistClientImpl is a wrapper around the AniList API client.
	AnilistClientImpl struct {
		Client   *Client
		logger   *zerolog.Logger
		token    string // The token used for authentication with the AniList API
		cacheDir string
	}
)

// NewAnilistClient creates a new AnilistClientImpl with the given token.
// The token is used for authorization when making requests to the AniList API.
func NewAnilistClient(token string, cacheDir string) *AnilistClientImpl {
	ac := &AnilistClientImpl{
		token:    token,
		cacheDir: cacheDir,
		Client: &Client{
			Client: clientv2.NewClient(alHttpClient(), alApiUrl(), nil,
				func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}, next clientv2.RequestInterceptorFunc) error {
					if err := initAnilistReq(ctx, req, token); err != nil {
						return err
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
	provider := currentRequestProvider()
	if provider == nil {
		return false
	}
	return provider.IsAuthenticated(ac.token)
}

func (ac *AnilistClientImpl) GetCacheDir() string {
	return ac.cacheDir
}

func (ac *AnilistClientImpl) CustomQuery(body []byte, logger *zerolog.Logger, token ...string) (data interface{}, err error) {
	return customQuery(body, logger, token...)
}

func alApiUrl() string {
	return currentRequestProvider().ApiUrl()
}

func alHttpClient() *http.Client {
	return requestProviderHTTPClient(currentRequestProvider())
}

func initAnilistReq(ctx context.Context, req *http.Request, token string) error {
	provider := currentRequestProvider()
	if err := setAnilistReqUrl(req, provider.ApiUrl()); err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return provider.PrepareRequest(ctx, req, token)
}

func setAnilistReqUrl(req *http.Request, rawURL string) error {
	apiURL, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	req.URL = apiURL
	req.Host = apiURL.Host
	return nil
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

func (ac *AnilistClientImpl) AnimeCollectionTags(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*AnimeCollectionTags, error) {
	if !ac.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}
	ac.logger.Debug().Msg("anilist: Fetching anime collection tags")
	return ac.Client.AnimeCollectionTags(ctx, userName, interceptors...)
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

func (ac *AnilistClientImpl) MangaCollectionTags(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*MangaCollectionTags, error) {
	if !ac.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}
	ac.logger.Debug().Msg("anilist: Fetching manga collection tags")
	return ac.Client.MangaCollectionTags(ctx, userName, interceptors...)
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

var noErrLogs = atomic.Bool{}

func (ac *AnilistClientImpl) BaseAnimeByMalID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseAnimeByMalID, error) {
	return ac.Client.BaseAnimeByMalID(ctx, id, interceptors...)
}

func (ac *AnilistClientImpl) BaseAnimeByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*BaseAnimeByID, error) {
	if id != nil && *id == 1 {
		noErrLogs.Store(true)
		defer noErrLogs.Store(false)
	}
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

func (ac *AnilistClientImpl) ListAnime(ctx context.Context, page *int, search *string, perPage *int, sort []*MediaSort, status []*MediaStatus, genres []*string, tags []*string, averageScoreGreater *int, season *MediaSeason, seasonYear *int, format *MediaFormat, isAdult *bool, interceptors ...clientv2.RequestInterceptor) (*ListAnime, error) {
	ac.logger.Debug().Msg("anilist: Fetching media list")
	if isAdult == nil {
		ret, err := ac.Client.ListAnimeAll(ctx, page, search, perPage, sort, status, genres, tags, averageScoreGreater, season, seasonYear, format, interceptors...)
		if err != nil {
			return nil, err
		}
		return FromListAnimeAll(ret), nil
	}
	return ac.Client.ListAnime(ctx, page, search, perPage, sort, status, genres, tags, averageScoreGreater, season, seasonYear, format, isAdult, interceptors...)
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

func (ac *AnilistClientImpl) ListManga(ctx context.Context, page *int, search *string, perPage *int, sort []*MediaSort, status []*MediaStatus, genres []*string, tags []*string, averageScoreGreater *int, startDateGreater *string, startDateLesser *string, format *MediaFormat, countryOfOrigin *string, isAdult *bool, interceptors ...clientv2.RequestInterceptor) (*ListManga, error) {
	ac.logger.Debug().Msg("anilist: Fetching manga list")
	if isAdult == nil {
		ret, err := ac.Client.ListMangaAll(ctx, page, search, perPage, sort, status, genres, tags, averageScoreGreater, startDateGreater, startDateLesser, format, countryOfOrigin, interceptors...)
		if err != nil {
			return nil, err
		}
		return FromListMangaAll(ret), nil
	}
	return ac.Client.ListManga(ctx, page, search, perPage, sort, status, genres, tags, averageScoreGreater, startDateGreater, startDateLesser, format, countryOfOrigin, isAdult, interceptors...)
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

type requestRateBlocker interface {
	Wait(ctx context.Context, sleep requestSleepFunc) error
	BlockUntil(until time.Time) bool
}

type requestSleepFunc func(ctx context.Context, delay time.Duration) error

type aniListRateBlocker struct {
	mu           sync.Mutex
	blockedUntil time.Time
	now          func() time.Time
}

func newAniListRateBlocker() *aniListRateBlocker {
	return &aniListRateBlocker{now: time.Now}
}

func (b *aniListRateBlocker) Wait(ctx context.Context, sleep requestSleepFunc) error {
	if sleep == nil {
		sleep = sleepWithContext
	}

	for {
		b.mu.Lock()
		blockedUntil := b.blockedUntil
		now := b.currentTime()
		b.mu.Unlock()

		if blockedUntil.IsZero() || !now.Before(blockedUntil) {
			return nil
		}

		if err := sleep(ctx, blockedUntil.Sub(now)); err != nil {
			return err
		}
	}
}

func (b *aniListRateBlocker) BlockUntil(until time.Time) bool {
	if until.IsZero() {
		return false
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.currentTime()
	if !until.After(now) || !until.After(b.blockedUntil) {
		return false
	}

	b.blockedUntil = until
	return true
}

func (b *aniListRateBlocker) currentTime() time.Time {
	if b.now != nil {
		return b.now()
	}
	return time.Now()
}

func parseResponseDate(headers http.Header) (time.Time, bool) {
	raw := headers.Get("Date")
	if raw == "" {
		return time.Time{}, false
	}

	parsed, err := http.ParseTime(raw)
	if err != nil {
		return time.Time{}, false
	}

	return parsed, true
}

func parseAniListRateLimitResetTime(headers http.Header, now time.Time) (time.Time, bool) {
	if resetAt, ok := parseRetryAfterTime(headers, now); ok {
		return resetAt, true
	}

	raw := headers.Get("X-RateLimit-Reset")
	if raw == "" {
		return time.Time{}, false
	}

	if unixSeconds, err := strconv.ParseInt(raw, 10, 64); err == nil && unixSeconds > 0 {
		return time.Unix(unixSeconds, 0), true
	}

	parsed, err := http.ParseTime(raw)
	if err != nil {
		return time.Time{}, false
	}

	return parsed, true
}

func parseRetryAfterTime(headers http.Header, now time.Time) (time.Time, bool) {
	raw := headers.Get("Retry-After")
	if raw == "" {
		return time.Time{}, false
	}

	if retryAfterSeconds, err := strconv.Atoi(raw); err == nil {
		return now.Truncate(time.Second).Add(time.Duration(retryAfterSeconds+1) * time.Second), true
	}

	parsed, err := http.ParseTime(raw)
	if err != nil {
		return time.Time{}, false
	}

	return parsed, true
}

var (
	sentRateLimitWarningTime                    = time.Now().Add(-10 * time.Second)
	sharedAniListRateBlocker requestRateBlocker = newAniListRateBlocker()
)

func doAniListRequestWithRetries(
	client *http.Client,
	req *http.Request,
	rateBlocker requestRateBlocker,
	sleep requestSleepFunc,
	onRateLimited func(waitSeconds int),
) (resp *http.Response, rlRemainingStr string, err error) {
	if client == nil {
		client = http.DefaultClient
	}
	if sleep == nil {
		sleep = sleepWithContext
	}

	const retryCount = 2

	for i := 0; i < retryCount; i++ {
		if err := req.Context().Err(); err != nil {
			return nil, rlRemainingStr, err
		}

		if rateBlocker != nil {
			if err := rateBlocker.Wait(req.Context(), sleep); err != nil {
				return nil, rlRemainingStr, err
			}
		}

		if i > 0 && req.Body != nil {
			if req.GetBody == nil {
				return nil, rlRemainingStr, errors.New("failed to retry request: request body is not replayable")
			}

			newBody, err := req.GetBody()
			if err != nil {
				return nil, rlRemainingStr, fmt.Errorf("failed to get request body: %w", err)
			}
			req.Body = newBody
		}

		resp, err = client.Do(req)
		if err != nil {
			return nil, rlRemainingStr, fmt.Errorf("request failed: %w", err)
		}

		rlRemainingStr = resp.Header.Get("X-Ratelimit-Remaining")
		responseTime := time.Now()
		if responseDate, ok := parseResponseDate(resp.Header); ok {
			responseTime = responseDate
		}
		if resetAt, ok := parseAniListRateLimitResetTime(resp.Header, responseTime); ok {
			if rateBlocker == nil || rateBlocker.BlockUntil(resetAt) {
				if onRateLimited != nil {
					waitSeconds := int(resetAt.Sub(responseTime).Round(time.Second) / time.Second)
					if waitSeconds < 1 {
						waitSeconds = 1
					}
					onRateLimited(waitSeconds)
				}
			}
			closeAniListResponseBody(resp)
			continue
		}

		return resp, rlRemainingStr, nil
	}

	return nil, rlRemainingStr, errors.New("anilist: rate limit exceeded, retries exhausted")
}

func closeAniListResponseBody(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}

	_ = resp.Body.Close()
	resp.Body = nil
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func notifyAniListRateLimit(logger *zerolog.Logger, waitSeconds int) {
	if logger != nil {
		logger.Warn().Msgf("anilist: Rate limited, retrying in %d seconds", waitSeconds)
	}

	if events.GlobalWSEventManager != nil {
		events.GlobalWSEventManager.SendEvent(events.AnilistRateLimit, waitSeconds)
	}

	if time.Since(sentRateLimitWarningTime) <= 10*time.Second {
		return
	}

	if events.GlobalWSEventManager != nil {
		events.GlobalWSEventManager.SendEvent(events.WarningToast, "anilist: Rate limited, retrying in "+strconv.Itoa(waitSeconds)+" seconds")
	}
	sentRateLimitWarningTime = time.Now()
}

// customDoFunc is a custom request interceptor function that handles rate limiting and retries.
func (ac *AnilistClientImpl) customDoFunc(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}) (err error) {
	var rlRemainingStr string

	reqTime := time.Now()
	defer func() {
		timeSince := time.Since(reqTime)
		formattedDur := timeSince.Truncate(time.Millisecond).String()
		if err != nil {
			if !noErrLogs.Load() {
				ac.logger.Error().Str("duration", formattedDur).Str("rlr", rlRemainingStr).Err(err).Str("document", gqlInfo.Request.OperationName).Msg("anilist: Failed Request")
			}
		} else {
			if timeSince > 900*time.Millisecond {
				ac.logger.Warn().Str("rtt", formattedDur).Str("rlr", rlRemainingStr).Msg("anilist: Successful Request (slow)")
			} else {
				ac.logger.Info().Str("rtt", formattedDur).Str("rlr", rlRemainingStr).Msg("anilist: Successful Request")
			}
		}
	}()

	var resp *http.Response
	resp, rlRemainingStr, err = doAniListRequestWithRetries(
		alHttpClient(),
		req,
		sharedAniListRateBlocker,
		sleepWithContext,
		func(waitSeconds int) {
			notifyAniListRateLimit(ac.logger, waitSeconds)
		},
	)
	if err != nil {
		return err
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
		if gqlErr, ok := errors.AsType[*clientv2.GqlErrorList](err); ok {
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
	if len(resp.Errors) > 0 {
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
