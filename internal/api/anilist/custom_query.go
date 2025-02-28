package anilist

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"net/http"
	"seanime/internal/util"
	"strconv"
	"time"

	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
)

func CustomQuery(body map[string]interface{}, logger *zerolog.Logger, token string) (data interface{}, err error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return customQuery(bodyBytes, logger, token)
}

func customQuery(body []byte, logger *zerolog.Logger, token ...string) (data interface{}, err error) {

	var rlRemainingStr string

	reqTime := time.Now()
	defer func() {
		timeSince := time.Since(reqTime)
		formattedDur := timeSince.Truncate(time.Millisecond).String()
		if err != nil {
			logger.Error().Str("duration", formattedDur).Str("rlr", rlRemainingStr).Err(err).Msg("anilist: Failed Request")
		} else {
			if timeSince > 600*time.Millisecond {
				logger.Warn().Str("rtt", formattedDur).Str("rlr", rlRemainingStr).Msg("anilist: Long Request")
			} else {
				logger.Trace().Str("rtt", formattedDur).Str("rlr", rlRemainingStr).Msg("anilist: Successful Request")
			}
		}
	}()

	defer util.HandlePanicInModuleThen("api/anilist/custom_query", func() {
		err = errors.New("panic in customQuery")
	})

	client := http.DefaultClient

	var req *http.Request
	req, err = http.NewRequest("POST", "https://graphql.anilist.co", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if len(token) > 0 {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token[0]))
	}

	// Send request
	retryCount := 2

	var resp *http.Response
	for i := 0; i < retryCount; i++ {

		// Reset response body for retry
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

		// Recreate the request body if it was read in a previous attempt
		if req.GetBody != nil {
			newBody, err := req.GetBody()
			if err != nil {
				return nil, fmt.Errorf("failed to get request body: %w", err)
			}
			req.Body = newBody
		}

		resp, err = client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}

		rlRemainingStr = resp.Header.Get("X-Ratelimit-Remaining")
		rlRetryAfterStr := resp.Header.Get("Retry-After")
		rlRetryAfter, err := strconv.Atoi(rlRetryAfterStr)
		if err == nil {
			logger.Warn().Msgf("anilist: Rate limited, retrying in %d seconds", rlRetryAfter+1)
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
			return nil, fmt.Errorf("gzip decode failed: %w", err)
		}
	}

	var res interface{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var ok bool

	reqErrors, ok := res.(map[string]interface{})["errors"].([]interface{})

	if ok && len(reqErrors) > 0 {
		firstError, foundErr := reqErrors[0].(map[string]interface{})
		if foundErr {
			return nil, errors.New(firstError["message"].(string))
		}
	}

	data, ok = res.(map[string]interface{})["data"]
	if !ok {
		return nil, errors.New("failed to parse data")
	}

	return data, nil
}
