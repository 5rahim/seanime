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
	"io"
	"net/http"
	"strconv"
	"time"
)

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

// customDoFunc is a custom request interceptor function that handles rate limiting and retries.
func (cw *ClientWrapper) customDoFunc(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}) (err error) {

	reqTime := time.Now()
	defer func() {
		timeSince := time.Since(reqTime)
		formattedDur := timeSince.Truncate(time.Millisecond).String()
		if err != nil {
			cw.logger.Error().Str("duration", formattedDur).Err(err).Msg("anilist: Failed Request")
		} else {
			if timeSince > 600*time.Millisecond {
				cw.logger.Warn().Str("rtt", formattedDur).Msg("anilist: Long Request")
			} else {
				cw.logger.Trace().Str("rtt", formattedDur).Msg("anilist: Successful Request")
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

		rlRemainingStr := resp.Header.Get("X-Ratelimit-Remaining")
		rlRetryAfterStr := resp.Header.Get("Retry-After")
		//println("Remaining:", rlRemainingStr, " | RetryAfter:", rlRetryAfterStr)

		// If we have a rate limit, sleep for the time
		rlRetryAfter, err := strconv.Atoi(rlRetryAfterStr)
		if err == nil {
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
