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

	"github.com/Eyevinn/hls-m3u8/m3u8"
	"github.com/goccy/go-json"
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

	buffer := bytes.NewBuffer(bodyBytes)
	playlist, listType, decodeErr := m3u8.Decode(*buffer, true)
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
				if !isAlreadyProxied(segment.URI) {
					if segment.URI != "" {
						if !strings.HasPrefix(segment.URI, "http") {
							segment.URI = resolveURL(baseURL, segment.URI)
						}
						segment.URI = rewriteProxyURL(segment.URI, headerMap)
						needsRewrite = true
					}
				}

				// Rewrite encryption key URIs
				for i, key := range segment.Keys {
					if key.URI != "" {
						if !isAlreadyProxied(key.URI) {
							keyURI := key.URI
							if !strings.HasPrefix(key.URI, "http") {
								keyURI = resolveURL(baseURL, key.URI)
							}
							segment.Keys[i].URI = rewriteProxyURL(keyURI, headerMap)
							needsRewrite = true
						}
					}
				}
			}
		}

		// Rewrite playlist-level encryption key URIs
		for i, key := range mediaPl.Keys {
			if key.URI != "" {
				if !isAlreadyProxied(key.URI) {
					keyURI := key.URI
					if !strings.HasPrefix(key.URI, "http") {
						keyURI = resolveURL(baseURL, key.URI)
					}
					mediaPl.Keys[i].URI = rewriteProxyURL(keyURI, headerMap)
					needsRewrite = true
				}
			}
		}

		// Encode the modified media playlist
		buffer := mediaPl.Encode()
		modifiedPlaylistBytes = buffer.Bytes()

	} else if listType == m3u8.MASTER {
		// Rewrite URIs in Master playlists
		masterPl := playlist.(*m3u8.MasterPlaylist)
		baseURL, _ := url2.Parse(url) // Base URL for resolving relative paths

		for _, variant := range masterPl.Variants {
			if variant != nil && variant.URI != "" {
				if !isAlreadyProxied(variant.URI) {
					variantURI := variant.URI
					if !strings.HasPrefix(variant.URI, "http") {
						variantURI = resolveURL(baseURL, variant.URI)
					}
					variant.URI = rewriteProxyURL(variantURI, headerMap)
					needsRewrite = true
				}
			}

			// Handle alternative media groups (audio, subtitles, etc.) for each variant
			if variant != nil {
				for _, alternative := range variant.Alternatives {
					if alternative != nil && alternative.URI != "" {
						if !isAlreadyProxied(alternative.URI) {
							alternativeURI := alternative.URI
							if !strings.HasPrefix(alternative.URI, "http") {
								alternativeURI = resolveURL(baseURL, alternative.URI)
							}
							alternative.URI = rewriteProxyURL(alternativeURI, headerMap)
							needsRewrite = true
						}
					}
				}
			}
		}

		// Rewrite session key URIs
		for i, sessionKey := range masterPl.SessionKeys {
			if sessionKey.URI != "" {
				if !isAlreadyProxied(sessionKey.URI) {
					sessionKeyURI := sessionKey.URI
					if !strings.HasPrefix(sessionKey.URI, "http") {
						sessionKeyURI = resolveURL(baseURL, sessionKey.URI)
					}
					masterPl.SessionKeys[i].URI = rewriteProxyURL(sessionKeyURI, headerMap)
					needsRewrite = true
				}
			}
		}

		// Encode the modified master playlist
		buffer := masterPl.Encode()
		modifiedPlaylistBytes = buffer.Bytes()

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
	proxyURL := "/api/v1/proxy?url=" + url2.QueryEscape(targetMediaURL)
	if len(headerMap) > 0 {
		headersStrB, err := json.Marshal(headerMap)
		// Ignore marshalling errors here? Or log them? For simplicity, ignoring now.
		if err == nil && len(headersStrB) > 2 { // Check > 2 for "{}" empty map
			proxyURL += "&headers=" + url2.QueryEscape(string(headersStrB))
		}
	}
	return proxyURL
}

func isAlreadyProxied(url string) bool {
	// Check if the URL contains the proxy pattern
	return strings.Contains(url, "/api/v1/proxy?url=") || strings.Contains(url, url2.QueryEscape("/api/v1/proxy?url="))
}
