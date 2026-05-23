package anilist

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"net/http"
	"seanime/internal/util"
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
			logger.Error().Str("duration", formattedDur).Str("rlr", rlRemainingStr).Err(err).Msg("anilist: Failed Request (custom query)")
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

	var req *http.Request
	req, err = http.NewRequest("POST", alApiUrl(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	authToken := ""
	if len(token) > 0 {
		authToken = token[0]
	}

	if err = initAnilistReq(req.Context(), req, authToken); err != nil {
		return nil, err
	}

	var resp *http.Response
	resp, rlRemainingStr, err = doAniListRequestWithRetries(
		alHttpClient(),
		req,
		sharedAniListRateBlocker,
		sleepWithContext,
		func(waitSeconds int) {
			notifyAniListRateLimit(logger, waitSeconds)
		},
	)
	if err != nil {
		return nil, err
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
