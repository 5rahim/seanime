package onlinestream_sources

import (
	"errors"
	"io"
	"net/http"
	"regexp"
	"seanime/internal/util"
	"strings"

	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
)

type (
	Streamtape struct {
		Client *http.Client
	}
)

func NewStreamtape() *Streamtape {
	return &Streamtape{
		Client: &http.Client{},
	}
}

func (s *Streamtape) Extract(uri string) (vs []*hibikeonlinestream.VideoSource, err error) {
	defer util.HandlePanicInModuleThen("onlinestream/sources/streamtape/Extract", func() {
		err = ErrVideoSourceExtraction
	})

	var ret []*hibikeonlinestream.VideoSource

	resp, err := s.Client.Get(uri)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`robotlink'\).innerHTML = (.*)'`)
	match := re.FindStringSubmatch(string(body))
	if len(match) == 0 {
		return nil, errors.New("could not find robotlink")
	}

	fhsh := strings.Split(match[1], "+ ('")
	fh := fhsh[0]
	sh := fhsh[1][3:]

	fh = strings.ReplaceAll(fh, "'", "")

	url := "https:" + fh + sh

	ret = append(ret, &hibikeonlinestream.VideoSource{
		URL:     url,
		Type:    map[bool]hibikeonlinestream.VideoSourceType{true: hibikeonlinestream.VideoSourceM3U8, false: hibikeonlinestream.VideoSourceMP4}[strings.Contains(url, ".m3u8")],
		Quality: QualityAuto,
	})

	return ret, nil
}
