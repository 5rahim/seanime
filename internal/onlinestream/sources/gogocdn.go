package onlinestream_sources

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gocolly/colly"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"seanime/internal/util"
	"strings"

	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
)

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

	defer util.HandlePanicInModuleThen("onlinestream/sources/gogocdn/Extract", func() {
		err = ErrVideoSourceExtraction
	})

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
		return nil, ErrNoVideoSourceFound
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
	defer util.HandlePanicInModuleThen("onlinestream/sources/gogocdn/urlToVideoSource", func() {
		ok = false
	})
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
