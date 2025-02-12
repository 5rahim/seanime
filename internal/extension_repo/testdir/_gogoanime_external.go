package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/gocolly/colly"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
)

const (
	DefaultServer      = "default"
	GogoanimeProvider  = "gogoanime-external"
	GogocdnServer      = "gogocdn"
	VidstreamingServer = "vidstreaming"
	StreamSBServer     = "streamsb"
)

type Gogoanime struct {
	BaseURL   string
	AjaxURL   string
	Client    http.Client
	UserAgent string
	logger    *zerolog.Logger
}

func NewProvider(logger *zerolog.Logger) hibikeonlinestream.Provider {
	return &Gogoanime{
		BaseURL:   "https://anitaku.to",
		AjaxURL:   "https://ajax.gogocdn.net",
		Client:    http.Client{},
		UserAgent: util.GetRandomUserAgent(),
		logger:    logger,
	}
}

func (g *Gogoanime) GetEpisodeServers() []string {
	return []string{GogocdnServer, VidstreamingServer}
}

func (g *Gogoanime) Search(query string, dubbed bool) ([]*hibikeonlinestream.SearchResult, error) {
	var results []*hibikeonlinestream.SearchResult

	g.logger.Debug().Str("query", query).Bool("dubbed", dubbed).Msg("gogoanime: Searching anime")

	c := colly.NewCollector(
		colly.UserAgent(g.UserAgent),
	)

	c.OnHTML(".last_episodes > ul > li", func(e *colly.HTMLElement) {
		id := ""
		idParts := strings.Split(e.ChildAttr("p.name > a", "href"), "/")
		if len(idParts) > 2 {
			id = idParts[2]
		}
		title := e.ChildText("p.name > a")
		url := g.BaseURL + e.ChildAttr("p.name > a", "href")
		subOrDub := hibikeonlinestream.Sub
		if strings.Contains(strings.ToLower(e.ChildText("p.name > a")), "dub") {
			subOrDub = hibikeonlinestream.Dub
		}
		results = append(results, &hibikeonlinestream.SearchResult{
			ID:       id,
			Title:    title,
			URL:      url,
			SubOrDub: subOrDub,
		})
	})

	searchURL := g.BaseURL + "/search.html?keyword=" + url.QueryEscape(query)
	if dubbed {
		searchURL += "%20(Dub)"
	}

	err := c.Visit(searchURL)
	if err != nil {
		return nil, err
	}

	g.logger.Debug().Int("count", len(results)).Msg("gogoanime: Fetched anime")

	return results, nil
}

func (g *Gogoanime) FindEpisode(id string) ([]*hibikeonlinestream.EpisodeDetails, error) {
	var episodes []*hibikeonlinestream.EpisodeDetails

	g.logger.Debug().Str("id", id).Msg("gogoanime: Fetching episodes")

	if !strings.Contains(id, "gogoanime") {
		id = fmt.Sprintf("%s/category/%s", g.BaseURL, id)
	}

	c := colly.NewCollector(
		colly.UserAgent(g.UserAgent),
	)

	var epStart, epEnd, movieID, alias string

	c.OnHTML("#episode_page > li > a", func(e *colly.HTMLElement) {
		if epStart == "" {
			epStart = e.Attr("ep_start")
		}
		epEnd = e.Attr("ep_end")
	})

	c.OnHTML("#movie_id", func(e *colly.HTMLElement) {
		movieID = e.Attr("value")
	})

	c.OnHTML("#alias", func(e *colly.HTMLElement) {
		alias = e.Attr("value")
	})

	err := c.Visit(id)
	if err != nil {
		g.logger.Error().Err(err).Msg("gogoanime: Failed to fetch episodes")
		return nil, err
	}

	c2 := colly.NewCollector(
		colly.UserAgent(g.UserAgent),
	)

	c2.OnHTML("#episode_related > li", func(e *colly.HTMLElement) {
		episodeIDParts := strings.Split(e.ChildAttr("a", "href"), "/")
		if len(episodeIDParts) < 2 {
			return
		}
		episodeID := strings.TrimSpace(episodeIDParts[1])
		episodeNumberStr := strings.TrimPrefix(e.ChildText("div.name"), "EP ")
		episodeNumber, err := strconv.Atoi(episodeNumberStr)
		if err != nil {
			g.logger.Error().Err(err).Str("episodeID", episodeID).Msg("failed to parse episode number")
			return
		}
		episodes = append(episodes, &hibikeonlinestream.EpisodeDetails{
			Provider: GogoanimeProvider,
			ID:       episodeID,
			Number:   episodeNumber,
			URL:      g.BaseURL + "/" + episodeID,
		})
	})

	ajaxURL := fmt.Sprintf("%s/ajax/load-list-episode", g.AjaxURL)
	ajaxParams := url.Values{
		"ep_start":   {epStart},
		"ep_end":     {epEnd},
		"id":         {movieID},
		"alias":      {alias},
		"default_ep": {"0"},
	}
	ajaxURLWithParams := fmt.Sprintf("%s?%s", ajaxURL, ajaxParams.Encode())

	err = c2.Visit(ajaxURLWithParams)
	if err != nil {
		g.logger.Error().Err(err).Msg("gogoanime: Failed to fetch episodes")
		return nil, err
	}

	g.logger.Debug().Int("count", len(episodes)).Msg("gogoanime: Fetched episodes")

	return episodes, nil
}

func (g *Gogoanime) FindEpisodeServer(episodeInfo *hibikeonlinestream.EpisodeDetails, server string) (*hibikeonlinestream.EpisodeServer, error) {
	var source *hibikeonlinestream.EpisodeServer

	if server == DefaultServer {
		server = GogocdnServer
	}
	g.logger.Debug().Str("server", string(server)).Str("episodeID", episodeInfo.ID).Msg("gogoanime: Fetching server sources")

	c := colly.NewCollector()

	switch server {
	case VidstreamingServer:
		c.OnHTML(".anime_muti_link > ul > li.vidcdn > a", func(e *colly.HTMLElement) {
			src := e.Attr("data-video")
			gogocdn := NewGogoCDN()
			videoSources, err := gogocdn.Extract(src)
			if err == nil {
				source = &hibikeonlinestream.EpisodeServer{
					Provider: GogoanimeProvider,
					Server:   server,
					Headers: map[string]string{
						"Referer": g.BaseURL + "/" + episodeInfo.ID,
					},
					VideoSources: videoSources,
				}
			}
		})
	case GogocdnServer, "":
		c.OnHTML("#load_anime > div > div > iframe", func(e *colly.HTMLElement) {
			src := e.Attr("src")
			gogocdn := NewGogoCDN()
			videoSources, err := gogocdn.Extract(src)
			if err == nil {
				source = &hibikeonlinestream.EpisodeServer{
					Provider: GogoanimeProvider,
					Server:   server,
					Headers: map[string]string{
						"Referer": g.BaseURL + "/" + episodeInfo.ID,
					},
					VideoSources: videoSources,
				}
			}
		})
	case StreamSBServer:
		c.OnHTML(".anime_muti_link > ul > li.streamsb > a", func(e *colly.HTMLElement) {
			src := e.Attr("data-video")
			streamsb := NewStreamSB()
			videoSources, err := streamsb.Extract(src)
			if err == nil {
				source = &hibikeonlinestream.EpisodeServer{
					Provider: GogoanimeProvider,
					Server:   server,
					Headers: map[string]string{
						"Referer":    g.BaseURL + "/" + episodeInfo.ID,
						"watchsb":    "streamsb",
						"User-Agent": g.UserAgent,
					},
					VideoSources: videoSources,
				}
			}
		})
	}

	err := c.Visit(g.BaseURL + "/" + episodeInfo.ID)
	if err != nil {
		return nil, err
	}

	if source == nil {
		g.logger.Warn().Str("server", string(server)).Msg("gogoanime: No sources found")
		return nil, fmt.Errorf("no sources found")
	}

	g.logger.Debug().Str("server", string(server)).Int("videoSources", len(source.VideoSources)).Msg("gogoanime: Fetched server sources")

	return source, nil

}

type cdnKeys struct {
	key       []byte
	secondKey []byte
	iv        []byte
}

type GogoCDN struct {
	client     *http.Client
	serverName string
	keys       cdnKeys
	referrer   string
}

func NewGogoCDN() *GogoCDN {
	return &GogoCDN{
		client:     &http.Client{},
		serverName: "goload",
		keys: cdnKeys{
			key:       []byte("37911490979715163134003223491201"),
			secondKey: []byte("54674138327930866480207815084989"),
			iv:        []byte("3134003223491201"),
		},
	}
}

// Extract fetches and extracts video sources from the provided URI.
func (g *GogoCDN) Extract(uri string) (vs []*hibikeonlinestream.VideoSource, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to extract video sources")
		}
	}()

	// Instantiate a new collector
	c := colly.NewCollector(
		// Allow visiting the same page multiple times
		colly.AllowURLRevisit(),
	)
	ur, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	// Variables to hold extracted values
	var scriptValue, id string

	id = ur.Query().Get("id")

	// Find and extract the script value and id
	c.OnHTML("script[data-name='episode']", func(e *colly.HTMLElement) {
		scriptValue = e.Attr("data-value")

	})

	// Start scraping
	err = c.Visit(uri)
	if err != nil {
		return nil, err
	}

	// Check if scriptValue and id are found
	if scriptValue == "" || id == "" {
		return nil, errors.New("script value or id not found")
	}

	// Extract video sources
	ajaxUrl := fmt.Sprintf("%s://%s/encrypt-ajax.php?%s", ur.Scheme, ur.Host, g.generateEncryptedAjaxParams(id, scriptValue))

	req, err := http.NewRequest("GET", ajaxUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")

	encryptedData, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer encryptedData.Body.Close()

	encryptedDataBytesRes, err := io.ReadAll(encryptedData.Body)
	if err != nil {
		return nil, err
	}

	var encryptedDataBytes map[string]string
	err = json.Unmarshal(encryptedDataBytesRes, &encryptedDataBytes)
	if err != nil {
		return nil, err
	}

	data, err := g.decryptAjaxData(encryptedDataBytes["data"])

	source, ok := data["source"].([]interface{})

	// Check if source is found
	if !ok {
		return nil, errors.New("source not found")
	}

	var results []*hibikeonlinestream.VideoSource

	urls := make([]string, 0)
	for _, src := range source {
		s := src.(map[string]interface{})
		urls = append(urls, s["file"].(string))
	}

	sourceBK, ok := data["source_bk"].([]interface{})
	if ok {
		for _, src := range sourceBK {
			s := src.(map[string]interface{})
			urls = append(urls, s["file"].(string))
		}
	}

	for _, url := range urls {

		vs, ok := g.urlToVideoSource(url, source, sourceBK)
		if ok {
			results = append(results, vs...)
		}

	}

	return results, nil
}

func (g *GogoCDN) urlToVideoSource(url string, source []interface{}, sourceBK []interface{}) (vs []*hibikeonlinestream.VideoSource, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()
	ret := make([]*hibikeonlinestream.VideoSource, 0)
	if strings.Contains(url, ".m3u8") {
		resResult, err := http.Get(url)
		if err != nil {
			return nil, false
		}
		defer resResult.Body.Close()

		bodyBytes, err := io.ReadAll(resResult.Body)
		if err != nil {
			return nil, false
		}
		bodyString := string(bodyBytes)

		resolutions := regexp.MustCompile(`(RESOLUTION=)(.*)(\s*?)(\s.*)`).FindAllStringSubmatch(bodyString, -1)
		baseURL := url[:strings.LastIndex(url, "/")]

		for _, res := range resolutions {
			quality := strings.Split(strings.Split(res[2], "x")[1], ",")[0]
			url := fmt.Sprintf("%s/%s", baseURL, strings.TrimSpace(res[4]))
			ret = append(ret, &hibikeonlinestream.VideoSource{URL: url, Type: hibikeonlinestream.VideoSourceM3U8, Quality: quality + "p"})
		}

		ret = append(ret, &hibikeonlinestream.VideoSource{URL: url, Type: hibikeonlinestream.VideoSourceM3U8, Quality: "default"})
	} else {
		for _, src := range source {
			s := src.(map[string]interface{})
			if s["file"].(string) == url {
				quality := strings.Split(s["label"].(string), " ")[0] + "p"
				ret = append(ret, &hibikeonlinestream.VideoSource{URL: url, Type: hibikeonlinestream.VideoSourceMP4, Quality: quality})
			}
		}
		if sourceBK != nil {
			for _, src := range sourceBK {
				s := src.(map[string]interface{})
				if s["file"].(string) == url {
					ret = append(ret, &hibikeonlinestream.VideoSource{URL: url, Type: hibikeonlinestream.VideoSourceMP4, Quality: "backup"})
				}
			}
		}
	}

	return ret, true
}

// generateEncryptedAjaxParams generates encrypted AJAX parameters.
func (g *GogoCDN) generateEncryptedAjaxParams(id, scriptValue string) string {
	encryptedKey := g.encrypt(id, g.keys.iv, g.keys.key)
	decryptedToken := g.decrypt(scriptValue, g.keys.iv, g.keys.key)
	return fmt.Sprintf("id=%s&alias=%s", encryptedKey, decryptedToken)
}

// encrypt encrypts the given text using AES CBC mode.
func (g *GogoCDN) encrypt(text string, iv []byte, key []byte) string {
	block, _ := aes.NewCipher(key)
	textBytes := []byte(text)
	textBytes = pkcs7Padding(textBytes, aes.BlockSize)
	cipherText := make([]byte, len(textBytes))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, textBytes)

	return base64.StdEncoding.EncodeToString(cipherText)
}

// decrypt decrypts the given text using AES CBC mode.
func (g *GogoCDN) decrypt(text string, iv []byte, key []byte) string {
	block, _ := aes.NewCipher(key)
	cipherText, _ := base64.StdEncoding.DecodeString(text)
	plainText := make([]byte, len(cipherText))

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plainText, cipherText)
	plainText = pkcs7Trimming(plainText)

	return string(plainText)
}

func (g *GogoCDN) decryptAjaxData(encryptedData string) (map[string]interface{}, error) {
	decodedData, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(g.keys.secondKey)
	if err != nil {
		return nil, err
	}

	if len(decodedData) < aes.BlockSize {
		return nil, fmt.Errorf("cipher text too short")
	}

	iv := g.keys.iv
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decodedData, decodedData)

	// Remove padding
	decodedData = pkcs7Trimming(decodedData)

	var data map[string]interface{}
	err = json.Unmarshal(decodedData, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// pkcs7Padding pads the text to be a multiple of blockSize using Pkcs7 padding.
func pkcs7Padding(text []byte, blockSize int) []byte {
	padding := blockSize - len(text)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(text, padText...)
}

// pkcs7Trimming removes Pkcs7 padding from the text.
func pkcs7Trimming(text []byte) []byte {
	length := len(text)
	unpadding := int(text[length-1])
	return text[:(length - unpadding)]
}

type StreamSB struct {
	Host      string
	Host2     string
	UserAgent string
}

func NewStreamSB() *StreamSB {
	return &StreamSB{
		Host:      "https://streamsss.net/sources50",
		Host2:     "https://watchsb.com/sources50",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.116 Safari/537.36",
	}
}

func (s *StreamSB) Payload(hex string) string {
	return "566d337678566f743674494a7c7c" + hex + "7c7c346b6767586d6934774855537c7c73747265616d7362/6565417268755339773461447c7c346133383438333436313335376136323337373433383634376337633465366534393338373136643732373736343735373237613763376334363733353737303533366236333463353333363534366137633763373337343732363536313664373336327c7c6b586c3163614468645a47617c7c73747265616d7362"
}

func (s *StreamSB) Extract(uri string) (vs []*hibikeonlinestream.VideoSource, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("failed to extract video sources")
		}
	}()

	var ret []*hibikeonlinestream.VideoSource

	id := strings.Split(uri, "/e/")[1]
	if strings.Contains(id, "html") {
		id = strings.Split(id, ".html")[0]
	}

	if id == "" {
		return nil, errors.New("cannot find ID")
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/%s", s.Host, s.Payload(hex.EncodeToString([]byte(id)))), nil)
	req.Header.Add("watchsb", "sbstream")
	req.Header.Add("User-Agent", s.UserAgent)
	req.Header.Add("Referer", uri)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, err
	}

	streamData, ok := jsonResponse["stream_data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("stream data not found")
	}

	m3u8Urls, err := client.Get(streamData["file"].(string))
	if err != nil {
		return nil, err
	}
	defer m3u8Urls.Body.Close()

	m3u8Body, err := io.ReadAll(m3u8Urls.Body)
	if err != nil {
		return nil, err
	}
	videoList := strings.Split(string(m3u8Body), "#EXT-X-STREAM-INF:")

	for _, video := range videoList {
		if !strings.Contains(video, "m3u8") {
			continue
		}

		url := strings.Split(video, "\n")[1]
		quality := strings.Split(strings.Split(video, "RESOLUTION=")[1], ",")[0]
		quality = strings.Split(quality, "x")[1]

		ret = append(ret, &hibikeonlinestream.VideoSource{
			URL:     url,
			Quality: quality + "p",
			Type:    hibikeonlinestream.VideoSourceM3U8,
		})
	}

	ret = append(ret, &hibikeonlinestream.VideoSource{
		URL:     streamData["file"].(string),
		Quality: "auto",
		Type:    map[bool]hibikeonlinestream.VideoSourceType{true: hibikeonlinestream.VideoSourceM3U8, false: hibikeonlinestream.VideoSourceMP4}[strings.Contains(streamData["file"].(string), ".m3u8")],
	})

	return ret, nil
}
