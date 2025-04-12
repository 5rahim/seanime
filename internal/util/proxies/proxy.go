package util

import (
	"bytes"
	"io"
	"net/http"
	url2 "net/url"
	"seanime/internal/util"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/grafov/m3u8"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

var proxyUA = util.GetRandomUserAgent()

var videoProxyClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		ForceAttemptHTTP2:   false, // Fixes issues on Linux
	},
	Timeout: 60 * time.Second,
}

func VideoProxy(c echo.Context) (err error) {
	defer util.HandlePanicInModuleWithError("util/VideoProxy", &err)

	url := c.QueryParam("url")
	headers := c.QueryParam("headers")

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
	if rangeHeader := c.Request().Header.Get("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	resp, err := videoProxyClient.Do(req)

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

	isHlsPlaylist := strings.HasSuffix(url, ".m3u8") || strings.Contains(resp.Header.Get("Content-Type"), "mpegurl")

	if !isHlsPlaylist {
		return c.Stream(resp.StatusCode, c.Response().Header().Get("Content-Type"), resp.Body)
	}

	// HLS Playlist
	//log.Debug().Str("url", url).Msg("proxy: Processing HLS playlist")

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Error().Err(readErr).Str("url", url).Msg("proxy: Error reading HLS response body")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read HLS playlist")
	}

	playlist, listType, decodeErr := m3u8.DecodeFrom(bytes.NewReader(bodyBytes), true)
	if decodeErr != nil {
		// Playlist might be valid but not decodable by the library, or simply corrupted.
		// Option 1: Proxy as-is (might be preferred if decoding fails unexpectedly)
		log.Warn().Err(decodeErr).Str("url", url).Msg("proxy: Failed to decode M3U8 playlist, proxying raw content")
		c.Response().Header().Set(echo.HeaderContentType, resp.Header.Get("Content-Type")) // Use original Content-Type
		c.Response().Header().Set(echo.HeaderContentLength, strconv.Itoa(len(bodyBytes)))
		c.Response().WriteHeader(resp.StatusCode)
		_, writeErr := c.Response().Writer.Write(bodyBytes)
		return writeErr
	}

	var modifiedPlaylistBytes []byte
	needsRewrite := false // Flag to check if we actually need to rewrite

	if listType == m3u8.MEDIA {
		mediaPl := playlist.(*m3u8.MediaPlaylist)
		baseURL, _ := url2.Parse(url) // Base URL for resolving relative paths

		for _, segment := range mediaPl.Segments {
			if segment != nil {
				// Rewrite Segment URI
				if segment.URI != "" && !strings.HasPrefix(segment.URI, "http") {
					absURI := resolveURL(baseURL, segment.URI)
					segment.URI = rewriteProxyURL(absURI, headerMap)
					needsRewrite = true
				} else if segment.URI != "" {
					// Rewrite absolute URL only if it doesn't already point to the proxy
					if !strings.Contains(segment.URI, "/api/v1/proxy?url=") {
						segment.URI = rewriteProxyURL(segment.URI, headerMap)
						needsRewrite = true
					}
				}

				// Rewrite Key URI
				if segment.Key != nil && segment.Key.URI != "" {
					if !strings.HasPrefix(segment.Key.URI, "http") {
						absKeyURI := resolveURL(baseURL, segment.Key.URI)
						segment.Key.URI = rewriteProxyURL(absKeyURI, headerMap)
						needsRewrite = true
					} else {
						// Rewrite absolute URL only if it doesn't already point to the proxy
						if !strings.Contains(segment.Key.URI, "/api/v1/proxy?url=") {
							segment.Key.URI = rewriteProxyURL(segment.Key.URI, headerMap)
							needsRewrite = true
						}
					}
				}
			}
		}

		// Rewrite Playlist Key URI (if present)
		if mediaPl.Key != nil && mediaPl.Key.URI != "" {
			if !strings.HasPrefix(mediaPl.Key.URI, "http") {
				absKeyURI := resolveURL(baseURL, mediaPl.Key.URI)
				mediaPl.Key.URI = rewriteProxyURL(absKeyURI, headerMap)
				needsRewrite = true
			} else {
				if !strings.Contains(mediaPl.Key.URI, "/api/v1/proxy?url=") {
					mediaPl.Key.URI = rewriteProxyURL(mediaPl.Key.URI, headerMap)
					needsRewrite = true
				}
			}
		}

		modifiedPlaylistBytes = []byte(mediaPl.String())

	} else if listType == m3u8.MASTER {
		// Optionally rewrite URIs in Master playlists as well if needed
		// Currently, just passes the master playlist through potentially unmodified
		masterPl := playlist.(*m3u8.MasterPlaylist)
		baseURL, _ := url2.Parse(url) // Base URL for resolving relative paths

		for _, variant := range masterPl.Variants {
			if variant != nil && variant.URI != "" {
				if !strings.HasPrefix(variant.URI, "http") {
					absURI := resolveURL(baseURL, variant.URI)
					variant.URI = rewriteProxyURL(absURI, headerMap)
					needsRewrite = true
				} else {
					if !strings.Contains(variant.URI, "/api/v1/proxy?url=") {
						variant.URI = rewriteProxyURL(variant.URI, headerMap)
						needsRewrite = true
					}
				}
			}
		}
		modifiedPlaylistBytes = []byte(masterPl.String())
	} else {
		// Unknown type, pass through
		modifiedPlaylistBytes = bodyBytes
	}

	// Set headers *after* potential modification
	contentType := "application/vnd.apple.mpegurl"
	c.Response().Header().Set(echo.HeaderContentType, contentType)
	// Set Content-Length based on the *modified* playlist
	c.Response().Header().Set(echo.HeaderContentLength, strconv.Itoa(len(modifiedPlaylistBytes)))

	// Set Cache-Control headers appropriate for playlists (often no-cache for live)
	if resp.Header.Get("Cache-Control") == "" {
		c.Response().Header().Set("Cache-Control", "no-cache")
	}

	log.Debug().Bool("rewritten", needsRewrite).Str("url", url).Msg("proxy: Sending modified HLS playlist")
	c.Response().WriteHeader(resp.StatusCode)

	return c.Blob(http.StatusOK, c.Response().Header().Get("Content-Type"), modifiedPlaylistBytes)
}

func resolveURL(base *url2.URL, relativeURI string) string {
	if base == nil {
		return relativeURI // Cannot resolve without a base
	}
	relativeURL, err := url2.Parse(relativeURI)
	if err != nil {
		return relativeURI // Invalid relative URI
	}
	return base.ResolveReference(relativeURL).String()
}

func rewriteProxyURL(targetMediaURL string, headerMap map[string]string) string {
	proxyURL := "/api/v1/proxy?url=" + url2.QueryEscape(targetMediaURL) // Use your proxy path
	if len(headerMap) > 0 {
		headersStrB, err := json.Marshal(headerMap)
		// Ignore marshalling errors here? Or log them? For simplicity, ignoring now.
		if err == nil && len(headersStrB) > 2 { // Check > 2 for "{}" empty map
			proxyURL += "&headers=" + url2.QueryEscape(string(headersStrB))
		}
	}
	return proxyURL
}
