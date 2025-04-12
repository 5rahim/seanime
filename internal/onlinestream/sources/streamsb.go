package onlinestream_sources

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"seanime/internal/util"
	"strings"

	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
)

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

	defer util.HandlePanicInModuleThen("onlinestream/sources/streamsb/Extract", func() {
		err = ErrVideoSourceExtraction
	})

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
		return nil, ErrNoVideoSourceFound
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
