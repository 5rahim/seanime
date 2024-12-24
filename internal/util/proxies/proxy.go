package util

import (
	"bytes"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	url2 "net/url"
	"seanime/internal/util"
	"strings"

	"github.com/grafov/m3u8"
)

func M3U8Proxy(c *fiber.Ctx) (err error) {
	defer util.HandlePanicInModuleWithError("util/M3U8Proxy", &err)

	url := c.Query("url")
	headers := c.Query("headers")

	client := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2: false,
		},
	}

	req, err := http.NewRequest(c.Method(), url, nil)
	if err != nil {
		log.Error().Err(err).Msg("proxy: Error creating request")
		return fiber.ErrInternalServerError
	}

	var headerMap map[string]string
	if headers != "" {
		if err := json.Unmarshal([]byte(headers), &headerMap); err != nil {
			log.Error().Err(err).Msg("proxy: Error unmarshalling headers")
			return fiber.ErrInternalServerError
		}
		for key, value := range headerMap {
			req.Header.Set(key, value)
		}
	}

	req.Header.Set("User-Agent", "AppleCoreMedia/1.0.0.16F203 (iPod touch; U; CPU OS 12_3_1 like Mac OS X; zh_cn)")
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("proxy: Error sending request")
		return fiber.ErrInternalServerError
	}
	defer resp.Body.Close()

	var ret []byte

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("proxy: Error reading response body")
		return fiber.ErrInternalServerError
	}

	if strings.HasSuffix(url, ".m3u8") {
		playlist, listType, err := m3u8.DecodeFrom(bytes.NewReader(b), true)
		if err != nil {
			log.Error().Err(err).Msg("proxy: Error decoding m3u8 playlist")
			return fiber.ErrInternalServerError
		}

		if listType == m3u8.MASTER {
			ret = b
			//master := playlist.(*m3u8.MasterPlaylist)
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

	for k, vs := range resp.Header {
		for _, v := range vs {
			c.Set(k, v)
		}
	}

	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

	return c.Send(ret)
}
