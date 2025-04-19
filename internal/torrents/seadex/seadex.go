package seadex

import (
	"fmt"
	"net/http"
	"seanime/internal/extension"
	"seanime/internal/util"
	"strings"

	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
)

type (
	SeaDex struct {
		logger *zerolog.Logger
		uri    string
	}

	Torrent struct {
		Name         string `json:"name"`
		Date         string `json:"date"`
		Size         int64  `json:"size"`
		Link         string `json:"link"`
		InfoHash     string `json:"infoHash"`
		ReleaseGroup string `json:"releaseGroup,omitempty"`
	}
)

func New(logger *zerolog.Logger) *SeaDex {
	return &SeaDex{
		logger: logger,
		uri:    util.Decode("aHR0cHM6Ly9yZWxlYXNlcy5tb2UvYXBpL2NvbGxlY3Rpb25zL2VudHJpZXMvcmVjb3Jkcw=="),
	}
}

func (s *SeaDex) SetSavedUserConfig(savedConfig *extension.SavedUserConfig) {
	url, _ := savedConfig.Values["apiUrl"]
	if url != "" {
		s.uri = url
	}
}

func (s *SeaDex) FetchTorrents(mediaId int, title string) (ret []*Torrent, err error) {

	ret = make([]*Torrent, 0)

	records, err := s.fetchRecords(mediaId)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return ret, nil
	}

	if len(records[0].Expand.Trs) == 0 {
		return ret, nil
	}
	for _, tr := range records[0].Expand.Trs {
		if tr.InfoHash == "" || tr.InfoHash == "<redacted>" || tr.Tracker != "Nyaa" || !strings.Contains(tr.URL, "nyaa.si") {
			continue
		}
		ret = append(ret, &Torrent{
			Name:         fmt.Sprintf("[%s] %s%s", tr.ReleaseGroup, title, map[bool]string{true: " [Dual-Audio]", false: ""}[tr.DualAudio]),
			Date:         tr.Created,
			Size:         int64(s.getTorrentSize(tr.Files)),
			Link:         tr.URL,
			InfoHash:     tr.InfoHash,
			ReleaseGroup: tr.ReleaseGroup,
		})
	}

	return ret, nil

}

func (s *SeaDex) fetchRecords(mediaId int) (ret []*RecordItem, err error) {

	uri := fmt.Sprintf("%s?page=1&perPage=1&filter=alID%%3D%%22%d%%22&skipTotal=1&expand=trs", s.uri, mediaId)

	resp, err := http.Get(uri)
	if err != nil {
		s.logger.Error().Err(err).Msgf("seadex: error getting media records: %v", mediaId)
		return nil, err
	}
	defer resp.Body.Close()

	var res RecordsResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		s.logger.Error().Err(err).Msgf("seadex: error decoding response: %v", mediaId)
		return nil, err
	}

	return res.Items, nil
}

func (s *SeaDex) getTorrentSize(fls []*TrFile) int {
	if fls == nil || len(fls) == 0 {
		return 0
	}

	var size int
	for _, f := range fls {
		size += f.Length
	}

	return size
}
