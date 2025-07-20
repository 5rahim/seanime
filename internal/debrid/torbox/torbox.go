package torbox

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"seanime/internal/constants"
	"seanime/internal/debrid/debrid"
	"seanime/internal/util"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
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

	TorrentInfo struct {
		Name  string             `json:"name"`
		Hash  string             `json:"hash"`
		Size  int64              `json:"size"`
		Files []*TorrentInfoFile `json:"files"`
	}

	TorrentInfoFile struct {
		Name string `json:"name"` // e.g. "Big Buck Bunny/Big Buck Bunny.mp4"
		Size int64  `json:"size"`
	}

	InstantAvailabilityItem struct {
		Name  string `json:"name"`
		Hash  string `json:"hash"`
		Size  int64  `json:"size"`
		Files []struct {
			Name string `json:"name"`
			Size int64  `json:"size"`
		} `json:"files"`
	}
)

func NewTorBox(logger *zerolog.Logger) debrid.Provider {
	return &TorBox{
		baseUrl: "https://api.torbox.app/v1/api",
		apiKey:  mo.None[string](),
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		logger: logger,
	}
}

func (t *TorBox) GetSettings() debrid.Settings {
	return debrid.Settings{
		ID:   "torbox",
		Name: "TorBox",
	}
}

func (t *TorBox) doQuery(method, uri string, body io.Reader, contentType string) (*Response, error) {
	return t.doQueryCtx(context.Background(), method, uri, body, contentType)
}

func (t *TorBox) doQueryCtx(ctx context.Context, method, uri string, body io.Reader, contentType string) (*Response, error) {
	apiKey, found := t.apiKey.Get()
	if !found {
		return nil, debrid.ErrNotAuthenticated
	}

	req, err := http.NewRequestWithContext(ctx, method, uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Add("User-Agent", "Seanime/"+constants.Version)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyB, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed: code %d, body: %s", resp.StatusCode, string(bodyB))
	}

	bodyB, err := io.ReadAll(resp.Body)
	if err != nil {
		t.logger.Error().Err(err).Msg("torbox: Failed to read response body")
		return nil, err
	}

	var ret Response
	if err := json.Unmarshal(bodyB, &ret); err != nil {
		trimmedBody := string(bodyB)
		if len(trimmedBody) > 2000 {
			trimmedBody = trimmedBody[:2000] + "..."
		}
		t.logger.Error().Err(err).Msg("torbox: Failed to decode response, response body: " + trimmedBody)
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

func (t *TorBox) GetInstantAvailability(hashes []string) map[string]debrid.TorrentItemInstantAvailability {

	t.logger.Trace().Strs("hashes", hashes).Msg("torbox: Checking instant availability")

	availability := make(map[string]debrid.TorrentItemInstantAvailability)

	if len(hashes) == 0 {
		return availability
	}

	var hashBatches [][]string

	for i := 0; i < len(hashes); i += 100 {
		end := i + 100
		if end > len(hashes) {
			end = len(hashes)
		}
		hashBatches = append(hashBatches, hashes[i:end])
	}

	for _, batch := range hashBatches {
		resp, err := t.doQuery("GET", t.baseUrl+fmt.Sprintf("/torrents/checkcached?hash=%s&format=list&list_files=true", strings.Join(batch, ",")), nil, "application/json")
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
			availability[item.Hash] = debrid.TorrentItemInstantAvailability{
				CachedFiles: make(map[string]*debrid.CachedFile),
			}

			for idx, file := range item.Files {
				availability[item.Hash].CachedFiles[strconv.Itoa(idx)] = &debrid.CachedFile{
					Name: file.Name,
					Size: file.Size,
				}
			}
		}

	}

	return availability
}

func (t *TorBox) AddTorrent(opts debrid.AddTorrentOptions) (string, error) {

	// Check if the torrent is already added by checking existing torrents
	if opts.InfoHash != "" {
		// First check if it's already in our account using a more efficient approach
		torrents, err := t.getTorrents()
		if err == nil {
			for _, torrent := range torrents {
				if torrent.Hash == opts.InfoHash {
					return strconv.Itoa(torrent.ID), nil
				}
			}
		}
		// Small delay to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}

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

// GetTorrentStreamUrl blocks until the torrent is downloaded and returns the stream URL for the torrent file by calling GetTorrentDownloadUrl.
func (t *TorBox) GetTorrentStreamUrl(ctx context.Context, opts debrid.StreamTorrentOptions, itemCh chan debrid.TorrentItem) (streamUrl string, err error) {

	t.logger.Trace().Str("torrentId", opts.ID).Str("fileId", opts.FileId).Msg("torbox: Retrieving stream link")

	doneCh := make(chan struct{})

	go func(ctx context.Context) {
		defer func() {
			close(doneCh)
		}()

		for {
			select {
			case <-ctx.Done():
				err = ctx.Err()
				return
			case <-time.After(4 * time.Second):
				torrent, _err := t.GetTorrent(opts.ID)
				if _err != nil {
					t.logger.Error().Err(_err).Msg("torbox: Failed to get torrent")
					err = fmt.Errorf("torbox: Failed to get torrent: %w", _err)
					return
				}

				itemCh <- *torrent

				// Check if the torrent is ready
				if torrent.IsReady {
					time.Sleep(1 * time.Second)
					downloadUrl, err := t.GetTorrentDownloadUrl(debrid.DownloadTorrentOptions{
						ID:     opts.ID,
						FileId: opts.FileId, // Filename
					})
					if err != nil {
						t.logger.Error().Err(err).Msg("torbox: Failed to get download URL")
						return
					}

					streamUrl = downloadUrl
					return
				}
			}
		}
	}(ctx)

	<-doneCh

	return
}

func (t *TorBox) GetTorrentDownloadUrl(opts debrid.DownloadTorrentOptions) (downloadUrl string, err error) {

	t.logger.Trace().Str("torrentId", opts.ID).Msg("torbox: Retrieving download link")

	apiKey, found := t.apiKey.Get()
	if !found {
		return "", fmt.Errorf("torbox: Failed to get download URL: %w", debrid.ErrNotAuthenticated)
	}

	url := t.baseUrl + fmt.Sprintf("/torrents/requestdl?token=%s&torrent_id=%s&zip_link=true", apiKey, opts.ID)
	if opts.FileId != "" {
		// Get the actual file ID
		torrent, err := t.getTorrent(opts.ID)
		if err != nil {
			return "", fmt.Errorf("torbox: Failed to get download URL: %w", err)
		}
		var fId string
		for _, f := range torrent.Files {
			if f.ShortName == opts.FileId {
				fId = strconv.Itoa(f.ID)
				break
			}
		}
		if fId == "" {
			return "", fmt.Errorf("torbox: Failed to get download URL, file not found")
		}
		url = t.baseUrl + fmt.Sprintf("/torrents/requestdl?token=%s&torrent_id=%s&file_id=%s", apiKey, opts.ID, fId)
	}

	resp, err := t.doQuery("GET", url, nil, "application/json")
	if err != nil {
		return "", fmt.Errorf("torbox: Failed to get download URL: %w", err)
	}

	marshaledData, _ := json.Marshal(resp.Data)

	var d string
	err = json.Unmarshal(marshaledData, &d)
	if err != nil {
		return "", fmt.Errorf("torbox: Failed to get download URL: %w", err)
	}

	t.logger.Debug().Str("downloadUrl", d).Msg("torbox: Download link retrieved")

	return d, nil
}

func (t *TorBox) GetTorrent(id string) (ret *debrid.TorrentItem, err error) {
	torrent, err := t.getTorrent(id)
	if err != nil {
		return nil, err
	}

	ret = toDebridTorrent(torrent)

	return ret, nil
}

func (t *TorBox) getTorrent(id string) (ret *Torrent, err error) {

	resp, err := t.doQuery("GET", t.baseUrl+fmt.Sprintf("/torrents/mylist?bypass_cache=true&id=%s", id), nil, "application/json")
	if err != nil {
		return nil, fmt.Errorf("torbox: Failed to get torrent: %w", err)
	}

	marshaledData, _ := json.Marshal(resp.Data)

	err = json.Unmarshal(marshaledData, &ret)
	if err != nil {
		return nil, fmt.Errorf("torbox: Failed to parse torrent: %w", err)
	}

	return ret, nil
}

// GetTorrentInfo uses the info hash to return the torrent's data.
// For cached torrents, it uses the /checkcached endpoint for faster response.
// For uncached torrents, it falls back to /torrentinfo endpoint.
func (t *TorBox) GetTorrentInfo(opts debrid.GetTorrentInfoOptions) (ret *debrid.TorrentInfo, err error) {

	if opts.InfoHash == "" {
		return nil, fmt.Errorf("torbox: No info hash provided")
	}

	resp, err := t.doQuery("GET", t.baseUrl+fmt.Sprintf("/torrents/checkcached?hash=%s&format=object&list_files=true", opts.InfoHash), nil, "application/json")
	if err != nil {
		return nil, fmt.Errorf("torbox: Failed to check cached torrent: %w", err)
	}

	// If the torrent is cached
	if resp.Data != nil {
		data := resp.Data.(map[string]interface{})

		if torrentData, exists := data[opts.InfoHash]; exists {
			marshaledData, _ := json.Marshal(torrentData)

			var torrent TorrentInfo
			err = json.Unmarshal(marshaledData, &torrent)
			if err != nil {
				return nil, fmt.Errorf("torbox: Failed to parse cached torrent: %w", err)
			}

			ret = toDebridTorrentInfo(&torrent)
			return ret, nil
		}
	}

	// If not cached, fall back
	resp, err = t.doQuery("GET", t.baseUrl+fmt.Sprintf("/torrents/torrentinfo?hash=%s&timeout=15", opts.InfoHash), nil, "application/json")
	if err != nil {
		return nil, fmt.Errorf("torbox: Failed to get torrent info: %w", err)
	}

	// DEVNOTE: Handle incorrect TorBox API response
	data := resp.Data.(map[string]interface{})
	if _, ok := data["data"]; ok {
		if _, ok := data["data"].(map[string]interface{}); ok {
			data = data["data"].(map[string]interface{})
		} else {
			return nil, fmt.Errorf("torbox: Failed to parse response")
		}
	}

	marshaledData, _ := json.Marshal(data)

	var torrent TorrentInfo
	err = json.Unmarshal(marshaledData, &torrent)
	if err != nil {
		return nil, fmt.Errorf("torbox: Failed to parse torrent: %w", err)
	}

	ret = toDebridTorrentInfo(&torrent)

	return ret, nil
}

func (t *TorBox) GetTorrents() (ret []*debrid.TorrentItem, err error) {

	torrents, err := t.getTorrents()
	if err != nil {
		return nil, fmt.Errorf("torbox: Failed to get torrents: %w", err)
	}

	// Limit the number of torrents to 500
	if len(torrents) > 500 {
		torrents = torrents[:500]
	}

	for _, t := range torrents {
		ret = append(ret, toDebridTorrent(t))
	}

	slices.SortFunc(ret, func(i, j *debrid.TorrentItem) int {
		return cmp.Compare(j.AddedAt, i.AddedAt)
	})

	return ret, nil
}

func (t *TorBox) getTorrents() (ret []*Torrent, err error) {

	resp, err := t.doQuery("GET", t.baseUrl+"/torrents/mylist?bypass_cache=true", nil, "application/json")
	if err != nil {
		return nil, fmt.Errorf("torbox: Failed to get torrents: %w", err)
	}

	marshaledData, _ := json.Marshal(resp.Data)

	err = json.Unmarshal(marshaledData, &ret)
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to parse torrents")
		return nil, fmt.Errorf("torbox: Failed to parse torrents: %w", err)
	}

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
		FormattedSize:        util.Bytes(uint64(t.Size)),
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

func toDebridTorrentInfo(t *TorrentInfo) (ret *debrid.TorrentInfo) {

	var files []*debrid.TorrentItemFile
	for idx, f := range t.Files {
		nameParts := strings.Split(f.Name, "/")
		var name string

		if len(nameParts) == 1 {
			name = nameParts[0]
		} else {
			name = nameParts[len(nameParts)-1]
		}

		files = append(files, &debrid.TorrentItemFile{
			ID:    name, // Set the ID to the og name so GetStreamUrl can use that to get the real file ID
			Index: idx,
			Name:  name,                       // e.g. "Big Buck Bunny.mp4"
			Path:  fmt.Sprintf("/%s", f.Name), // e.g. "/Big Buck Bunny/Big Buck Bunny.mp4"
			Size:  f.Size,
		})
	}

	ret = &debrid.TorrentInfo{
		Name:  t.Name,
		Hash:  t.Hash,
		Size:  t.Size,
		Files: files,
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
