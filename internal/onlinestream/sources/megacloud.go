package onlinestream_sources

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"seanime/internal/util"
	"strconv"
	"strings"
)

type MegaCloud struct {
	Script    string
	Sources   string
	UserAgent string
}

func NewMegaCloud() *MegaCloud {
	return &MegaCloud{
		Script:    "https://megacloud.tv/js/player/a/prod/e1-player.min.js",
		Sources:   "https://megacloud.tv/embed-2/ajax/e-1/getSources?id=",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
	}
}

func (m *MegaCloud) Extract(uri string) (vs []*VideoSource, err error) {
	defer util.HandlePanicInModuleThen("onlinestream/sources/megacloud/Extract", func() {
		err = ErrVideoSourceExtraction
	})

	videoIdParts := strings.Split(uri, "/")
	videoId := videoIdParts[len(videoIdParts)-1]
	videoId = strings.Split(videoId, "?")[0]

	client := &http.Client{}
	req, err := http.NewRequest("GET", m.Sources+videoId, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", m.UserAgent)
	req.Header.Set("Referer", uri)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var srcData map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&srcData)
	if err != nil {
		return nil, err
	}

	subtitles := make([]*VideoSubtitle, 0)
	for idx, s := range srcData["tracks"].([]interface{}) {
		sub := s.(map[string]interface{})
		label, ok := sub["label"].(string)
		if ok {
			subtitle := &VideoSubtitle{
				URL:       sub["file"].(string),
				ID:        label,
				Language:  label,
				IsDefault: idx == 0,
			}
			subtitles = append(subtitles, subtitle)
		}
	}
	if encryptedString, ok := srcData["sources"]; ok {

		switch encryptedString.(type) {
		case []interface{}:
			if len(encryptedString.([]interface{})) == 0 {
				return nil, ErrNoVideoSourceFound
			}
			videoSources := make([]*VideoSource, 0)
			if e, ok := encryptedString.([]interface{})[0].(map[string]interface{}); ok {
				file, ok := e["file"].(string)
				if ok {
					videoSources = append(videoSources, &VideoSource{
						URL:       file,
						Type:      map[bool]VideoSourceType{true: VideoSourceM3U8, false: VideoSourceMP4}[strings.Contains(file, ".m3u8")],
						Subtitles: subtitles,
						Quality:   QualityAuto,
					})
				}
			}

			if len(videoSources) == 0 {
				return nil, ErrNoVideoSourceFound
			}

			return videoSources, nil

		case []map[string]interface{}:
			if srcData["encrypted"].(bool) && ok {
				videoSources := make([]*VideoSource, 0)
				for _, e := range encryptedString.([]map[string]interface{}) {
					videoSources = append(videoSources, &VideoSource{
						URL:       e["file"].(string),
						Type:      map[bool]VideoSourceType{true: VideoSourceM3U8, false: VideoSourceMP4}[strings.Contains(e["file"].(string), ".m3u8")],
						Subtitles: subtitles,
						Quality:   QualityAuto,
					})
				}
				if len(videoSources) == 0 {
					return nil, ErrNoVideoSourceFound
				}
				return videoSources, nil
			}
		case string:
			res, err = client.Get(m.Script)
			if err != nil {
				return nil, err
			}
			defer res.Body.Close()

			text, err := io.ReadAll(res.Body)
			if err != nil {
				return nil, errors.New("couldn't fetch script to decrypt resource")
			}

			values, err := m.extractVariables(string(text))
			if err != nil {
				return nil, err
			}

			secret, encryptedSource, err := m.getSecret(encryptedString.(string), values)
			if err != nil {
				return nil, err
			}

			decrypted, err := m.decrypt(encryptedSource, secret)
			if err != nil {
				return nil, err
			}

			var decryptedData []map[string]interface{}
			err = json.Unmarshal([]byte(decrypted), &decryptedData)
			if err != nil {
				return nil, err
			}

			sources := make([]*VideoSource, 0)
			for _, e := range decryptedData {
				sources = append(sources, &VideoSource{
					URL:       e["file"].(string),
					Type:      map[bool]VideoSourceType{true: VideoSourceM3U8, false: VideoSourceMP4}[strings.Contains(e["file"].(string), ".m3u8")],
					Subtitles: subtitles,
					Quality:   QualityAuto,
				})
			}

			if len(sources) == 0 {
				return nil, ErrNoVideoSourceFound
			}

			return sources, nil
		}

	}

	return nil, ErrNoVideoSourceFound
}

func (m *MegaCloud) extractVariables(text string) ([]int, error) {
	var allVars string

	re := regexp.MustCompile(`const \w{1,2}=new URLSearchParams.+?;function`)
	matches := re.FindAllString(text, -1)
	if len(matches) > 0 {
		allVars = matches[len(matches)-1]
		if strings.HasSuffix(allVars, "function") {
			allVars = strings.TrimSuffix(allVars, "function")
		}
	}

	pairs := strings.Split(allVars[:len(allVars)-1], "=")[1:]
	var values []int
	for _, pair := range pairs {
		value, err := strconv.ParseInt(strings.Split(pair, ",")[0][2:], 16, 64)
		if err != nil || value == 0 {
			continue
		}
		values = append(values, int(value))
	}

	return values, nil
}

func (m *MegaCloud) getSecret(encryptedString string, values []int) (string, string, error) {
	var secret string
	var encryptedSource = encryptedString
	var totalInc int

	for i := 0; i < values[0]; i++ {
		var start, inc int

		switch i {
		case 0:
			start = values[2]
			inc = values[1]
		case 1:
			start = values[4]
			inc = values[3]
		case 2:
			start = values[6]
			inc = values[5]
		case 3:
			start = values[8]
			inc = values[7]
		case 4:
			start = values[10]
			inc = values[9]
		case 5:
			start = values[12]
			inc = values[11]
		case 6:
			start = values[14]
			inc = values[13]
		case 7:
			start = values[16]
			inc = values[15]
		case 8:
			start = values[18]
			inc = values[17]
		default:
			return "", "", errors.New("invalid index")
		}

		from := start + totalInc
		to := from + inc

		secret += encryptedString[from:to]
		encryptedSource = strings.Replace(encryptedSource, encryptedString[from:to], "", 1)
		totalInc += inc
	}

	return secret, encryptedSource, nil
}

func (m *MegaCloud) decrypt(encrypted, keyOrSecret string) (string, error) {
	cypher, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	salt := cypher[8:16]
	password := append([]byte(keyOrSecret), salt...)

	md5Hashes := make([][]byte, 3)
	digest := password
	for i := 0; i < 3; i++ {
		hash := md5.Sum(digest)
		md5Hashes[i] = hash[:]
		digest = append(hash[:], password...)
	}

	key := append(md5Hashes[0], md5Hashes[1]...)
	iv := md5Hashes[2]
	contents := cypher[16:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(contents, contents)

	contents, err = pkcs7Unpad(contents, block.BlockSize())
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if blockSize <= 0 {
		return nil, errors.New("invalid blocksize")
	}
	if len(data)%blockSize != 0 || len(data) == 0 {
		return nil, errors.New("invalid PKCS7 data (block size must be a multiple of input length)")
	}
	padLen := int(data[len(data)-1])
	if padLen > blockSize || padLen == 0 {
		return nil, errors.New("invalid PKCS7 padding")
	}
	for i := 0; i < padLen; i++ {
		if data[len(data)-1-i] != byte(padLen) {
			return nil, errors.New("invalid PKCS7 padding")
		}
	}
	return data[:len(data)-padLen], nil
}
