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
	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
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

func (m *MegaCloud) Extract(uri string) (vs []*hibikeonlinestream.VideoSource, err error) {
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

	subtitles := make([]*hibikeonlinestream.VideoSubtitle, 0)
	for idx, s := range srcData["tracks"].([]interface{}) {
		sub := s.(map[string]interface{})
		label, ok := sub["label"].(string)
		if ok {
			subtitle := &hibikeonlinestream.VideoSubtitle{
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
			videoSources := make([]*hibikeonlinestream.VideoSource, 0)
			if e, ok := encryptedString.([]interface{})[0].(map[string]interface{}); ok {
				file, ok := e["file"].(string)
				if ok {
					videoSources = append(videoSources, &hibikeonlinestream.VideoSource{
						URL:       file,
						Type:      map[bool]hibikeonlinestream.VideoSourceType{true: hibikeonlinestream.VideoSourceM3U8, false: hibikeonlinestream.VideoSourceMP4}[strings.Contains(file, ".m3u8")],
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
				videoSources := make([]*hibikeonlinestream.VideoSource, 0)
				for _, e := range encryptedString.([]map[string]interface{}) {
					videoSources = append(videoSources, &hibikeonlinestream.VideoSource{
						URL:       e["file"].(string),
						Type:      map[bool]hibikeonlinestream.VideoSourceType{true: hibikeonlinestream.VideoSourceM3U8, false: hibikeonlinestream.VideoSourceMP4}[strings.Contains(e["file"].(string), ".m3u8")],
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

			secret, encryptedSource := m.getSecret(encryptedString.(string), values)
			//if err != nil {
			//	return nil, err
			//}

			decrypted, err := m.decrypt(encryptedSource, secret)
			if err != nil {
				return nil, err
			}

			var decryptedData []map[string]interface{}
			err = json.Unmarshal([]byte(decrypted), &decryptedData)
			if err != nil {
				return nil, err
			}

			sources := make([]*hibikeonlinestream.VideoSource, 0)
			for _, e := range decryptedData {
				sources = append(sources, &hibikeonlinestream.VideoSource{
					URL:       e["file"].(string),
					Type:      map[bool]hibikeonlinestream.VideoSourceType{true: hibikeonlinestream.VideoSourceM3U8, false: hibikeonlinestream.VideoSourceMP4}[strings.Contains(e["file"].(string), ".m3u8")],
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

func (m *MegaCloud) extractVariables(text string) ([][]int, error) {
	re := regexp.MustCompile(`case\s*0x[0-9a-f]+:\s*\w+\s*=\s*(\w+)\s*,\s*\w+\s*=\s*(\w+);`)
	matches := re.FindAllStringSubmatch(text, -1)

	var vars [][]int

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		caseLine := match[0]
		if strings.Contains(caseLine, "partKey") {
			continue
		}

		matchKey1, err1 := m.matchingKey(match[1], text)
		matchKey2, err2 := m.matchingKey(match[2], text)

		if err1 != nil || err2 != nil {
			continue
		}

		key1, err1 := strconv.ParseInt(matchKey1, 16, 64)
		key2, err2 := strconv.ParseInt(matchKey2, 16, 64)

		if err1 != nil || err2 != nil {
			continue
		}

		vars = append(vars, []int{int(key1), int(key2)})
	}

	return vars, nil
}

func (m *MegaCloud) matchingKey(value, script string) (string, error) {
	regexPattern := `,` + regexp.QuoteMeta(value) + `=((?:0x)?([0-9a-fA-F]+))`
	re := regexp.MustCompile(regexPattern)

	match := re.FindStringSubmatch(script)
	if len(match) > 1 {
		return strings.TrimPrefix(match[1], "0x"), nil
	}

	return "", errors.New("failed to match the key")
}

func (m *MegaCloud) getSecret(encryptedString string, values [][]int) (string, string) {
	secret := ""
	encryptedSourceArray := strings.Split(encryptedString, "")
	currentIndex := 0

	for _, index := range values {
		start := index[0] + currentIndex
		end := start + index[1]

		for i := start; i < end; i++ {
			secret += string(encryptedString[i])
			encryptedSourceArray[i] = ""
		}

		currentIndex += index[1]
	}

	encryptedSource := strings.Join(encryptedSourceArray, "")

	return secret, encryptedSource
}

//func (m *MegaCloud) getSecret(encryptedString string, values []int) (string, string, error) {
//	var secret string
//	var encryptedSource = encryptedString
//	var totalInc int
//
//	for i := 0; i < values[0]; i++ {
//		var start, inc int
//
//		switch i {
//		case 0:
//			start = values[2]
//			inc = values[1]
//		case 1:
//			start = values[4]
//			inc = values[3]
//		case 2:
//			start = values[6]
//			inc = values[5]
//		case 3:
//			start = values[8]
//			inc = values[7]
//		case 4:
//			start = values[10]
//			inc = values[9]
//		case 5:
//			start = values[12]
//			inc = values[11]
//		case 6:
//			start = values[14]
//			inc = values[13]
//		case 7:
//			start = values[16]
//			inc = values[15]
//		case 8:
//			start = values[18]
//			inc = values[17]
//		default:
//			return "", "", errors.New("invalid index")
//		}
//
//		from := start + totalInc
//		to := from + inc
//
//		secret += encryptedString[from:to]
//		encryptedSource = strings.Replace(encryptedSource, encryptedString[from:to], "", 1)
//		totalInc += inc
//	}
//
//	return secret, encryptedSource, nil
//}

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
