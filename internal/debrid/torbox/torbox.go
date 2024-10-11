package torbox

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"io"
	"mime/multipart"
	"net/http"
	"seanime/internal/debrid/debrid"
	"seanime/internal/util"
	"slices"
	"strconv"
	"strings"
	"time"
)

type (
	TorBox struct {
		baseUrl string
		apiKey  mo.Option[string]
		client  *http.Client
		logger  *zerolog.Logger
	}

	Response struct {
		Success bool        `json:"success"`
		Detail  string      `json:"detail"`
		Data    interface{} `json:"data"`
	}

	File struct {
		ID        int    `json:"id"`
		MD5       string `json:"md5"`
		S3Path    string `json:"s3_path"`
		Name      string `json:"name"`
		Size      int    `json:"size"`
		MimeType  string `json:"mimetype"`
		ShortName string `json:"short_name"`
	}

	Torrent struct {
		ID               int     `json:"id"`
		Hash             string  `json:"hash"`
		CreatedAt        string  `json:"created_at"`
		UpdatedAt        string  `json:"updated_at"`
		Magnet           string  `json:"magnet"`
		Size             int64   `json:"size"`
		Active           bool    `json:"active"`
		AuthID           string  `json:"auth_id"`
		DownloadState    string  `json:"download_state"`
		Seeds            int     `json:"seeds"`
		Peers            int     `json:"peers"`
		Ratio            float64 `json:"ratio"`
		Progress         float64 `json:"progress"`
		DownloadSpeed    float64 `json:"download_speed"`
		UploadSpeed      float64 `json:"upload_speed"`
		Name             string  `json:"name"`
		ETA              int64   `json:"eta"`
		Server           float64 `json:"server"`
		TorrentFile      bool    `json:"torrent_file"`
		ExpiresAt        string  `json:"expires_at"`
		DownloadPresent  bool    `json:"download_present"`
		DownloadFinished bool    `json:"download_finished"`
		Files            []*File `json:"files"`
		InactiveCheck    int     `json:"inactive_check"`
		Availability     float64 `json:"availability"`
	}

	InstantAvailabilityItem struct {
		Name string `json:"name"`
		Hash string `json:"hash"`
		Size int64  `json:"size"`
	}
)

func NewTorBox(logger *zerolog.Logger) debrid.Provider {
	return &TorBox{
		baseUrl: "https://api.torbox.app/v1/api",
		apiKey:  mo.None[string](),
		client:  &http.Client{},
		logger:  logger,
	}
}

func (t *TorBox) GetSettings() debrid.Settings {
	return debrid.Settings{
		ID:                  "torbox",
		Name:                "TorBox",
		CanStream:           false,
		CanSelectStreamFile: false,
	}
}

func (t *TorBox) doQuery(method, uri string, body io.Reader, contentType string) (*Response, error) {
	apiKey, found := t.apiKey.Get()
	if !found {
		return nil, debrid.ErrNotAuthenticated
	}

	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Authorization", "Bearer "+apiKey)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ret Response

	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		t.logger.Error().Err(err).Msg("debrid: Failed to decode response")
		return nil, err
	}

	if !ret.Success {
		return nil, fmt.Errorf("request failed: %s", ret.Detail)
	}

	return &ret, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (t *TorBox) Authenticate(apiKey string) error {
	t.apiKey = mo.Some(apiKey)
	return nil
}

func (t *TorBox) GetInstantAvailability(hashes []string) map[string]bool {

	t.logger.Trace().Strs("hashes", hashes).Msg("torbox: Checking instant availability")

	availability := make(map[string]bool)

	var hashBatches [][]string

	for i := 0; i < len(hashes); i += 100 {
		end := i + 100
		if end > len(hashes) {
			end = len(hashes)
		}
		hashBatches = append(hashBatches, hashes[i:end])
	}

	for _, batch := range hashBatches {
		resp, err := t.doQuery("GET", t.baseUrl+fmt.Sprintf("/torrents/checkcached?hash=%s&format=list&list_files=false", strings.Join(batch, ",")), nil, "application/json")
		if err != nil {
			return availability
		}

		marshaledData, _ := json.Marshal(resp.Data)

		var items []*InstantAvailabilityItem
		err = json.Unmarshal(marshaledData, &items)
		if err != nil {
			return availability
		}

		for _, item := range items {
			availability[item.Hash] = true
		}

	}

	return availability
}

func (t *TorBox) AddTorrent(opts debrid.AddTorrentOptions) (string, error) {

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	t.logger.Trace().Str("magnetLink", opts.MagnetLink).Msg("torbox: Adding torrent")

	err := writer.WriteField("magnet", opts.MagnetLink)
	if err != nil {
		return "", fmt.Errorf("torbox: Failed to add torrent: %w", err)
	}

	err = writer.WriteField("seed", "1")
	if err != nil {
		return "", fmt.Errorf("torbox: Failed to add torrent: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("torbox: Failed to add torrent: %w", err)
	}

	resp, err := t.doQuery("POST", t.baseUrl+"/torrents/createtorrent", &body, writer.FormDataContentType())
	if err != nil {
		return "", fmt.Errorf("torbox: Failed to add torrent: %w", err)
	}

	type data struct {
		ID   int    `json:"torrent_id"`
		Name string `json:"name"`
		Hash string `json:"hash"`
	}

	marshaledData, _ := json.Marshal(resp.Data)

	var d data
	err = json.Unmarshal(marshaledData, &d)
	if err != nil {
		return "", fmt.Errorf("torbox: Failed to add torrent: %w", err)
	}

	t.logger.Debug().Str("torrentId", strconv.Itoa(d.ID)).Str("torrentName", d.Name).Str("torrentHash", d.Hash).Msg("torbox: Torrent added")

	return strconv.Itoa(d.ID), nil
}

func (t *TorBox) GetTorrentStreamUrl(_ debrid.StreamTorrentOptions) (streamUrl string, err error) {
	return "", fmt.Errorf("torbox: Streaming is not supported")
}

func (t *TorBox) GetTorrentDownloadUrl(opts debrid.DownloadTorrentOptions) (downloadUrl string, err error) {

	t.logger.Trace().Str("torrentId", opts.ID).Msg("torbox: Retrieving download link")

	apiKey, found := t.apiKey.Get()
	if !found {
		return "", fmt.Errorf("torbox: Failed to downloaded torrent: %w", debrid.ErrNotAuthenticated)
	}

	resp, err := t.doQuery("GET", t.baseUrl+fmt.Sprintf("/torrents/requestdl?token=%s&torrent_id=%s&zip_link=true", apiKey, opts.ID), nil, "application/json")
	if err != nil {
		return "", fmt.Errorf("torbox: Failed to download torrent: %w", err)
	}

	marshaledData, _ := json.Marshal(resp.Data)

	var d string
	err = json.Unmarshal(marshaledData, &d)
	if err != nil {
		return "", fmt.Errorf("torbox: Failed to download torrent: %w", err)
	}

	t.logger.Debug().Str("downloadUrl", d).Msg("torbox: Download link retrieved")

	return d, nil
}

func (t *TorBox) GetTorrent(id string) (ret *debrid.TorrentItem, err error) {

	resp, err := t.doQuery("GET", t.baseUrl+fmt.Sprintf("/torrents/mylist?bypass_cache=true&id=%s", id), nil, "application/json")
	if err != nil {
		return nil, fmt.Errorf("torbox: Failed to get torrent: %w", err)
	}

	marshaledData, _ := json.Marshal(resp.Data)

	var torrent Torrent
	err = json.Unmarshal(marshaledData, &torrent)
	if err != nil {
		return nil, fmt.Errorf("torbox: Failed to parse torrent: %w", err)
	}

	ret = toDebridTorrent(&torrent)

	return ret, nil
}

func (t *TorBox) GetTorrents() (ret []*debrid.TorrentItem, err error) {

	resp, err := t.doQuery("GET", t.baseUrl+"/torrents/mylist?bypass_cache=true", nil, "application/json")
	if err != nil {
		return nil, fmt.Errorf("torbox: Failed to get torrents: %w", err)
	}

	marshaledData, _ := json.Marshal(resp.Data)

	var torrents []*Torrent
	err = json.Unmarshal(marshaledData, &torrents)
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to parse torrents")
		return nil, fmt.Errorf("torbox: Failed to parse torrents: %w", err)
	}

	for _, t := range torrents {
		ret = append(ret, toDebridTorrent(t))
	}

	// Limit the number of torrents to 500
	if len(ret) > 500 {
		ret = ret[:500]
	}

	slices.SortFunc(ret, func(i, j *debrid.TorrentItem) int {
		return cmp.Compare(j.AddedAt, i.AddedAt)
	})

	return ret, nil
}

func toDebridTorrent(t *Torrent) (ret *debrid.TorrentItem) {

	addedAt, _ := time.Parse(time.RFC3339Nano, t.CreatedAt)

	completionPercentage := int(t.Progress * 100)

	ret = &debrid.TorrentItem{
		ID:                   strconv.Itoa(t.ID),
		Name:                 t.Name,
		Hash:                 t.Hash,
		Size:                 t.Size,
		FormattedSize:        humanize.Bytes(uint64(t.Size)),
		CompletionPercentage: completionPercentage,
		ETA:                  util.FormatETA(int(t.ETA)),
		Status:               toDebridTorrentStatus(t),
		AddedAt:              addedAt.Format(time.RFC3339),
		Speed:                util.ToHumanReadableSpeed(int(t.DownloadSpeed)),
		Seeders:              t.Seeds,
		IsReady:              t.DownloadPresent,
	}

	return
}

func toDebridTorrentStatus(t *Torrent) debrid.TorrentItemStatus {
	if t.DownloadFinished && t.DownloadPresent {
		switch t.DownloadState {
		case "uploading":
			return debrid.TorrentItemStatusSeeding
		default:
			return debrid.TorrentItemStatusCompleted
		}
	}

	switch t.DownloadState {
	case "downloading", "metaDL":
		return debrid.TorrentItemStatusDownloading
	case "stalled", "stalled (no seeds)":
		return debrid.TorrentItemStatusStalled
	case "completed", "cached":
		return debrid.TorrentItemStatusCompleted
	case "uploading":
		return debrid.TorrentItemStatusSeeding
	case "paused":
		return debrid.TorrentItemStatusPaused
	default:
		return debrid.TorrentItemStatusOther
	}
}

func (t *TorBox) DeleteTorrent(id string) error {

	type body = struct {
		ID        int    `json:"torrent_id"`
		Operation string `json:"operation"`
	}

	b := body{
		ID:        util.StringToIntMust(id),
		Operation: "delete",
	}

	marshaledData, _ := json.Marshal(b)

	_, err := t.doQuery("POST", t.baseUrl+fmt.Sprintf("/torrents/controltorrent"), bytes.NewReader(marshaledData), "application/json")
	if err != nil {
		return fmt.Errorf("torbox: Failed to delete torrent: %w", err)
	}

	return nil
}
