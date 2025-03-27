package util

import (
	"bytes"
	"io"
	"net/http"
	url2 "net/url"
	"seanime/internal/util"
	"strings"

	"github.com/goccy/go-json"
	"github.com/rs/zerolog/log"

	"github.com/grafov/m3u8"
	"github.com/labstack/echo/v4"
)

var proxyUA = util.GetRandomUserAgent()

func M3U8Proxy(c echo.Context) (err error) {
	defer util.HandlePanicInModuleWithError("util/EchoM3U8Proxy", &err)

	url := c.QueryParam("url")
	headers := c.QueryParam("headers")

	client := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2: false,
		},
	}

	// Always use GET request internally, even for HEAD requests
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Error().Err(err).Msg("proxy: Error creating request")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	var headerMap map[string]string
	if headers != "" {
		if err := json.Unmarshal([]byte(headers), &headerMap); err != nil {
			log.Error().Err(err).Msg("proxy: Error unmarshalling headers")
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		for key, value := range headerMap {
			req.Header.Set(key, value)
		}
	}

	req.Header.Set("User-Agent", proxyUA)
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("proxy: Error sending request")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	// Copy response headers
	for k, vs := range resp.Header {
		for _, v := range vs {
			if !strings.EqualFold(k, "Content-Length") { // Skip Content-Length header, fixes net::ERR_CONTENT_LENGTH_MISMATCH
				c.Response().Header().Set(k, v)
			}
		}
	}

	// Set CORS headers
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Response().Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

	// For HEAD requests, return only headers
	if c.Request().Method == http.MethodHead {
		return c.NoContent(http.StatusOK)
	}

	// If the URL is not an HLS stream, stream directly without loading into memory
	if !strings.HasSuffix(url, ".m3u8") {
		return c.Stream(http.StatusOK, resp.Header.Get("Content-Type"), resp.Body)
	}

	var ret []byte

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("proxy: Error reading response body")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if strings.HasSuffix(url, ".m3u8") {
		playlist, listType, err := m3u8.DecodeFrom(bytes.NewReader(b), true)
		if err != nil {
			log.Error().Err(err).Msg("proxy: Error decoding m3u8 playlist")
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		if listType == m3u8.MASTER {
			ret = b
		} else if listType == m3u8.MEDIA {
			media := playlist.(*m3u8.MediaPlaylist)
			for _, segment := range media.Segments {
				if segment != nil {
					// get base url
					if !strings.HasPrefix(segment.URI, "http") {
						baseUrl := url[:strings.LastIndex(url, "/")+1]
						segment.URI = baseUrl + segment.URI
					}
					segment.URI = "/api/v1/proxy?url=" + url2.QueryEscape(segment.URI)
					headersStrB, _ := json.Marshal(headerMap)
					if len(headersStrB) > 0 {
						segment.URI += "&headers=" + url2.QueryEscape(string(headersStrB))
					}
					if segment.Key != nil {
						segment.Key.URI = "/api/v1/proxy?url=" + url2.QueryEscape(segment.Key.URI)
						headersStrB, _ := json.Marshal(headerMap)
						if len(headersStrB) > 0 {
							segment.Key.URI += "&headers=" + url2.QueryEscape(string(headersStrB))
						}
					}
				}
			}
			if media.Key != nil {
				media.Key.URI = "/api/v1/proxy?url=" + url2.QueryEscape(media.Key.URI)
				headersStrB, _ := json.Marshal(headerMap)
				if len(headersStrB) > 0 {
					media.Key.URI += "&headers=" + url2.QueryEscape(string(headersStrB))
				}
			}
			ret = []byte(media.String())
		}
	} else {
		ret = b
	}

	return c.Blob(http.StatusOK, c.Response().Header().Get("Content-Type"), ret)
}
