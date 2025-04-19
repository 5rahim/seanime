package realdebrid

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
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
	RealDebrid struct {
		baseUrl string
		apiKey  mo.Option[string]
		client  *http.Client
		logger  *zerolog.Logger
	}

	ErrorResponse struct {
		Error        string `json:"error"`
		ErrorDetails string `json:"error_details"`
		ErrorCode    int    `json:"error_code"`
	}

	Torrent struct {
		ID       string   `json:"id"`
		Filename string   `json:"filename"`
		Hash     string   `json:"hash"`
		Bytes    int64    `json:"bytes"`
		Host     string   `json:"host"`
		Split    int      `json:"split"`
		Progress float64  `json:"progress"`
		Status   string   `json:"status"`
		Added    string   `json:"added"`
		Links    []string `json:"links"`
		Ended    string   `json:"ended"`
		Speed    int64    `json:"speed"`
		Seeders  int      `json:"seeders"`
	}

	TorrentInfo struct {
		ID               string             `json:"id"`
		Filename         string             `json:"filename"`
		OriginalFilename string             `json:"original_filename"`
		Hash             string             `json:"hash"`
		Bytes            int64              `json:"bytes"`          // Size of selected files
		OriginalBytes    int64              `json:"original_bytes"` // Size of the torrent
		Host             string             `json:"host"`
		Split            int                `json:"split"`
		Progress         float64            `json:"progress"`
		Status           string             `json:"status"`
		Added            string             `json:"added"`
		Files            []*TorrentInfoFile `json:"files"`
		Links            []string           `json:"links"`
		Ended            string             `json:"ended"`
		Speed            int64              `json:"speed"`
		Seeders          int                `json:"seeders"`
	}

	TorrentInfoFile struct {
		ID       int    `json:"id"`
		Path     string `json:"path"` // e.g. "/Big Buck Bunny/Big Buck Bunny.mp4"
		Bytes    int64  `json:"bytes"`
		Selected int    `json:"selected"` // 1 if selected, 0 if not
	}

	InstantAvailabilityItem struct {
		Hash  string `json:"hash"`
		Files []struct {
			Filename string `json:"filename"`
			Filesize int    `json:"filesize"`
		} `json:"files"`
	}
)

func NewRealDebrid(logger *zerolog.Logger) debrid.Provider {
	return &RealDebrid{
		baseUrl: "https://api.real-debrid.com/rest/1.0",
		apiKey:  mo.None[string](),
		client: &http.Client{
			Timeout: time.Second * 10,
		},
		logger: logger,
	}
}

func NewRealDebridT(logger *zerolog.Logger) *RealDebrid {
	return &RealDebrid{
		baseUrl: "https://api.real-debrid.com/rest/1.0",
		apiKey:  mo.None[string](),
		client: &http.Client{
			Timeout: time.Second * 30,
		},
		logger: logger,
	}
}

func (t *RealDebrid) GetSettings() debrid.Settings {
	return debrid.Settings{
		ID:   "realdebrid",
		Name: "RealDebrid",
	}
}

func (t *RealDebrid) doQuery(method, uri string, body io.Reader, contentType string) (ret []byte, err error) {
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ErrorResponse

		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			t.logger.Error().Err(err).Msg("realdebrid: Failed to decode response")
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		// If the error details are empty, we'll just return the response body
		if errResp.ErrorDetails == "" && errResp.ErrorCode == 0 {
			content, _ := io.ReadAll(resp.Body)
			return content, nil
		}

		return nil, fmt.Errorf("failed to query API: %s, %s", resp.Status, errResp.ErrorDetails)
	}

	content, _ := io.ReadAll(resp.Body)
	//fmt.Println(string(content))

	return content, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (t *RealDebrid) Authenticate(apiKey string) error {
	t.apiKey = mo.Some(apiKey)
	return nil
}

//	{
//	   "string": { // First hash
//	       "string": [ // hoster, ex: "rd"
//	           // All file IDs variants
//	           {
//	               "int": { // file ID, you must ask all file IDs from this array on /selectFiles to get instant downloading
//	                   "filename": "string",
//	                   "filesize": int
//	               },
//	           },
type instantAvailabilityResponse map[string]map[string][]map[int]instantAvailabilityFile
type instantAvailabilityFile struct {
	Filename string `json:"filename"`
	Filesize int64  `json:"filesize"`
}

func (t *RealDebrid) GetInstantAvailability(hashes []string) map[string]debrid.TorrentItemInstantAvailability {

	t.logger.Trace().Strs("hashes", hashes).Msg("realdebrid: Checking instant availability")

	availability := make(map[string]debrid.TorrentItemInstantAvailability)

	if len(hashes) == 0 {
		return availability
	}

	return t.getInstantAvailabilityT(hashes, 3, 100)
}

func (t *RealDebrid) getInstantAvailabilityT(hashes []string, retries int, limit int) (ret map[string]debrid.TorrentItemInstantAvailability) {

	ret = make(map[string]debrid.TorrentItemInstantAvailability)

	var hashBatches [][]string

	for i := 0; i < len(hashes); i += limit {
		end := i + limit
		if end > len(hashes) {
			end = len(hashes)
		}
		hashBatches = append(hashBatches, hashes[i:end])
	}

	for _, batch := range hashBatches {

		hashParams := ""
		for _, hash := range batch {
			hashParams += "/" + hash
		}

		resp, err := t.doQuery("GET", t.baseUrl+"/torrents/instantAvailability"+hashParams, nil, "application/json")
		if err != nil {
			t.logger.Error().Err(err).Msg("realdebrid: Failed to get instant availability")
			return
		}

		//fmt.Println(string(resp))

		var instantAvailability instantAvailabilityResponse
		err = json.Unmarshal(resp, &instantAvailability)
		if err != nil {
			if limit != 1 && retries > 0 {
				t.logger.Warn().Msg("realdebrid: Retrying instant availability request")
				return t.getInstantAvailabilityT(hashes, retries-1, int(math.Ceil(float64(limit)/10)))
			} else {
				t.logger.Error().Err(err).Msg("realdebrid: Failed to parse instant availability")
				return
			}
		}

		for hash, hosters := range instantAvailability {
			currentHash := ""
			for _, _hash := range hashes {
				if strings.EqualFold(hash, _hash) {
					currentHash = _hash
					break
				}
			}

			if currentHash == "" {
				continue
			}

			avail := debrid.TorrentItemInstantAvailability{
				CachedFiles: make(map[string]*debrid.CachedFile),
			}

			for hoster, hosterI := range hosters {
				if hoster != "rd" {
					continue
				}

				for _, hosterFiles := range hosterI {
					for fileId, file := range hosterFiles {
						avail.CachedFiles[strconv.Itoa(fileId)] = &debrid.CachedFile{
							Name: file.Filename,
							Size: file.Filesize,
						}
					}
				}
			}

			if len(avail.CachedFiles) > 0 {
				ret[currentHash] = avail
			}
		}

	}

	return
}

func (t *RealDebrid) AddTorrent(opts debrid.AddTorrentOptions) (string, error) {

	// Check if the torrent is already added
	// If it is, return the torrent ID
	torrentId := ""
	if opts.InfoHash != "" {
		torrents, err := t.getTorrents(false)
		if err == nil {
			for _, torrent := range torrents {
				if torrent.Hash == opts.InfoHash {
					t.logger.Debug().Str("torrentId", torrent.ID).Msg("realdebrid: Torrent already added")
					torrentId = torrent.ID
					break
				}
			}
		}
		time.Sleep(1 * time.Second)
	}

	// If the torrent wasn't already added, add it
	if torrentId == "" {
		resp, err := t.addMagnet(opts.MagnetLink)
		if err != nil {
			return "", err
		}
		torrentId = resp.ID
	}

	// If a file ID is provided, select the file to start downloading it
	if opts.SelectFileId != "" {
		// Select the file to download
		err := t.selectCachedFiles(torrentId, opts.SelectFileId)
		if err != nil {
			return "", err
		}
	}

	return torrentId, nil
}

// GetTorrentStreamUrl blocks until the torrent is downloaded and returns the stream URL for the torrent file by calling GetTorrentDownloadUrl.
func (t *RealDebrid) GetTorrentStreamUrl(ctx context.Context, opts debrid.StreamTorrentOptions, itemCh chan debrid.TorrentItem) (streamUrl string, err error) {

	t.logger.Trace().Str("torrentId", opts.ID).Str("fileId", opts.FileId).Msg("realdebrid: Retrieving stream link")

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
				ti, _err := t.getTorrentInfo(opts.ID)
				if _err != nil {
					t.logger.Error().Err(_err).Msg("realdebrid: Failed to get torrent")
					err = fmt.Errorf("realdebrid: Failed to get torrent: %w", _err)
					return
				}

				dt := toDebridTorrent(&Torrent{
					ID:       ti.ID,
					Filename: ti.Filename,
					Hash:     ti.Hash,
					Bytes:    ti.Bytes,
					Host:     ti.Host,
					Split:    ti.Split,
					Progress: ti.Progress,
					Status:   ti.Status,
					Added:    ti.Added,
					Links:    ti.Links,
					Ended:    ti.Ended,
					Speed:    ti.Speed,
					Seeders:  ti.Seeders,
				})
				itemCh <- *dt

				// Check if the torrent is ready
				if dt.IsReady {
					time.Sleep(1 * time.Second)

					files := make([]*TorrentInfoFile, 0)
					for _, f := range ti.Files {
						if f.Selected == 1 {
							files = append(files, f)
						}
					}

					if len(files) == 0 {
						err = fmt.Errorf("realdebrid: No files downloaded")
						return
					}

					for idx, f := range files {
						if strconv.Itoa(f.ID) == opts.FileId {
							resp, err := t.unrestrictLink(ti.Links[idx])
							if err != nil {
								t.logger.Error().Err(err).Msg("realdebrid: Failed to get download URL")
								return
							}

							streamUrl = resp.Download
							return
						}
					}
					err = fmt.Errorf("realdebrid: File not found")
					return
				}
			}
		}
	}(ctx)

	<-doneCh

	return
}

type unrestrictLinkResponse struct {
	ID         string `json:"id"`
	Filename   string `json:"filename"`
	MimeType   string `json:"mimeType"`
	Filesize   int64  `json:"filesize"`
	Link       string `json:"link"`
	Host       string `json:"host"`
	Chunks     int    `json:"chunks"`
	Crc        int    `json:"crc"`
	Download   string `json:"download"` // Generated download link
	Streamable int    `json:"streamable"`
}

// GetTorrentDownloadUrl returns the download URL for the torrent file.
// If no opts.FileId is provided, it will return a comma-separated list of download URLs for all selected files in the torrent.
func (t *RealDebrid) GetTorrentDownloadUrl(opts debrid.DownloadTorrentOptions) (downloadUrl string, err error) {

	t.logger.Trace().Str("torrentId", opts.ID).Msg("realdebrid: Retrieving download link")

	torrentInfo, err := t.getTorrentInfo(opts.ID)
	if err != nil {
		return "", fmt.Errorf("realdebrid: Failed to get download URL: %w", err)
	}

	files := make([]*TorrentInfoFile, 0)
	for _, f := range torrentInfo.Files {
		if f.Selected == 1 {
			files = append(files, f)
		}
	}

	downloadUrl = ""

	if opts.FileId != "" {
		var file *TorrentInfoFile
		var link string
		for idx, f := range files {
			if strconv.Itoa(f.ID) == opts.FileId {
				file = f
				link = torrentInfo.Links[idx]
				break
			}
		}

		if file == nil || link == "" {
			return "", fmt.Errorf("realdebrid: File not found")
		}

		unrestrictLink, err := t.unrestrictLink(link)
		if err != nil {
			return "", fmt.Errorf("realdebrid: Failed to get download URL: %w", err)
		}

		return unrestrictLink.Download, nil
	}

	for idx := range files {
		link := torrentInfo.Links[idx]
		unrestrictLink, err := t.unrestrictLink(link)
		if err != nil {
			return "", fmt.Errorf("realdebrid: Failed to get download URL: %w", err)
		}
		if downloadUrl != "" {
			downloadUrl += ","
		}
		downloadUrl += unrestrictLink.Download
	}

	return downloadUrl, nil
}

func (t *RealDebrid) GetTorrent(id string) (ret *debrid.TorrentItem, err error) {
	torrent, err := t.getTorrent(id)
	if err != nil {
		return nil, err
	}

	ret = toDebridTorrent(torrent)

	return ret, nil
}

// GetTorrentInfo uses the info hash to return the torrent's data.
// This adds the torrent to the user's account without downloading it and removes it after getting the info.
func (t *RealDebrid) GetTorrentInfo(opts debrid.GetTorrentInfoOptions) (ret *debrid.TorrentInfo, err error) {

	if opts.MagnetLink == "" {
		return nil, fmt.Errorf("realdebrid: Magnet link is required")
	}

	// Add the torrent to the user's account without downloading it
	resp, err := t.addMagnet(opts.MagnetLink)
	if err != nil {
		return nil, fmt.Errorf("realdebrid: Failed to get info: %w", err)
	}

	torrent, err := t.getTorrentInfo(resp.ID)
	if err != nil {
		return nil, err
	}

	go func() {
		// Remove the torrent
		err = t.DeleteTorrent(torrent.ID)
		if err != nil {
			t.logger.Error().Err(err).Msg("realdebrid: Failed to delete torrent")
		}
	}()

	ret = toDebridTorrentInfo(torrent)

	return ret, nil
}

func (t *RealDebrid) GetTorrents() (ret []*debrid.TorrentItem, err error) {

	torrents, err := t.getTorrents(true)
	if err != nil {
		return nil, fmt.Errorf("realdebrid: Failed to get torrents: %w", err)
	}

	for _, t := range torrents {
		ret = append(ret, toDebridTorrent(t))
	}

	slices.SortFunc(ret, func(i, j *debrid.TorrentItem) int {
		return cmp.Compare(j.AddedAt, i.AddedAt)
	})

	return ret, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// selectCachedFiles
// Real Debrid will re-download cached torrent if we select only a few files from the torrent.
// To avoid this, we'll select all *cached* files in the torrent if the file we want to download is cached.
func (t *RealDebrid) selectCachedFiles(id string, idStr string) (err error) {

	t.logger.Trace().Str("torrentId", id).Str("fileId", "all").Msg("realdebrid: Selecting all files")

	return t._selectFiles(id, "all")

	//t.logger.Trace().Str("torrentId", id).Str("fileId", idStr).Msg("realdebrid: Selecting cached files")
	//// If the file ID is "all" or a list of IDs, just call selectFiles
	//if idStr == "all" || strings.Contains(idStr, ",") {
	//	return t._selectFiles(id, idStr)
	//}
	//
	//// Get the torrent info
	//torrent, err := t.getTorrent(id)
	//if err != nil {
	//	return err
	//}
	//
	//// Get the instant availability
	//avail := t.GetInstantAvailability([]string{torrent.Hash})
	//if _, ok := avail[torrent.Hash]; !ok {
	//	return t._selectFiles(id, idStr)
	//}
	//
	//// Get all cached file IDs
	//ids := make([]string, 0)
	//for fileIdStr := range avail[torrent.Hash].CachedFiles {
	//	if fileIdStr != "" {
	//		ids = append(ids, fileIdStr)
	//	}
	//}
	//
	//// If the selected file isn't cached, we'll just download it alone
	//if !slices.Contains(ids, idStr) {
	//	return t._selectFiles(id, idStr)
	//}
	//
	//// Download all cached files
	//return t._selectFiles(id, strings.Join(ids, ","))
}

func (t *RealDebrid) _selectFiles(id string, idStr string) (err error) {

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	t.logger.Trace().Str("torrentId", id).Str("fileId", idStr).Msg("realdebrid: Selecting files")

	err = writer.WriteField("files", idStr)
	if err != nil {
		t.logger.Error().Err(err).Msg("realdebrid: Failed to write field 'files'")
		return fmt.Errorf("realdebrid: Failed to select files: %w", err)
	}

	_, err = t.doQuery("POST", t.baseUrl+fmt.Sprintf("/torrents/selectFiles/%s", id), &body, writer.FormDataContentType())
	if err != nil {
		t.logger.Error().Err(err).Msg("realdebrid: Failed to select files")
		return fmt.Errorf("realdebrid: Failed to select files: %w", err)
	}

	return nil
}

type addMagnetResponse struct {
	ID  string `json:"id"`
	URI string `json:"uri"`
}

func (t *RealDebrid) addMagnet(magnet string) (ret *addMagnetResponse, err error) {

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	t.logger.Trace().Str("magnetLink", magnet).Msg("realdebrid: Adding torrent")

	err = writer.WriteField("magnet", magnet)
	if err != nil {
		t.logger.Error().Err(err).Msg("realdebrid: Failed to write field 'magnet'")
		return nil, fmt.Errorf("torbox: Failed to add torrent: %w", err)
	}

	resp, err := t.doQuery("POST", t.baseUrl+"/torrents/addMagnet", &body, writer.FormDataContentType())
	if err != nil {
		t.logger.Error().Err(err).Msg("realdebrid: Failed to add torrent")
		return nil, fmt.Errorf("realdebrid: Failed to add torrent: %w", err)
	}

	err = json.Unmarshal(resp, &ret)
	if err != nil {
		t.logger.Error().Err(err).Msg("realdebrid: Failed to parse torrent")
		return nil, fmt.Errorf("realdebrid: Failed to parse torrent: %w", err)
	}

	return ret, nil
}

func (t *RealDebrid) unrestrictLink(link string) (ret *unrestrictLinkResponse, err error) {

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	err = writer.WriteField("link", link)
	if err != nil {
		t.logger.Error().Err(err).Msg("realdebrid: Failed to write field 'link'")
		return nil, fmt.Errorf("realdebrid: Failed to unrestrict link: %w", err)
	}

	resp, err := t.doQuery("POST", t.baseUrl+"/unrestrict/link", &body, writer.FormDataContentType())
	if err != nil {
		t.logger.Error().Err(err).Msg("realdebrid: Failed to unrestrict link")
		return nil, fmt.Errorf("realdebrid: Failed to unrestrict link: %w", err)
	}

	err = json.Unmarshal(resp, &ret)
	if err != nil {
		t.logger.Error().Err(err).Msg("realdebrid: Failed to parse unrestrict link")
		return nil, fmt.Errorf("realdebrid: Failed to parse unrestrict link: %w", err)
	}

	return ret, nil
}

func (t *RealDebrid) getTorrents(activeOnly bool) (ret []*Torrent, err error) {
	_url, _ := url.Parse(t.baseUrl + "/torrents")
	q := _url.Query()
	if activeOnly {
		q.Set("filter", "active")
	} else {
		q.Set("limit", "500")
	}

	resp, err := t.doQuery("GET", _url.String(), nil, "application/json")
	if err != nil {
		t.logger.Error().Err(err).Msg("realdebrid: Failed to get torrents")
		return nil, fmt.Errorf("realdebrid: Failed to get torrents: %w", err)
	}

	err = json.Unmarshal(resp, &ret)
	if err != nil {
		t.logger.Error().Err(err).Msg("realdebrid: Failed to parse torrents")
		return nil, fmt.Errorf("realdebrid: Failed to parse torrents: %w", err)
	}

	return ret, nil
}

func (t *RealDebrid) getTorrent(id string) (ret *Torrent, err error) {

	resp, err := t.doQuery("GET", t.baseUrl+fmt.Sprintf("/torrents/info/%s", id), nil, "application/json")
	if err != nil {
		t.logger.Error().Err(err).Msg("realdebrid: Failed to get torrent")
		return nil, fmt.Errorf("realdebrid: Failed to get torrent: %w", err)
	}

	var ti TorrentInfo

	err = json.Unmarshal(resp, &ti)
	if err != nil {
		t.logger.Error().Err(err).Msg("realdebrid: Failed to parse torrent")
		return nil, fmt.Errorf("realdebrid: Failed to parse torrent: %w", err)
	}

	ret = &Torrent{
		ID:       ti.ID,
		Filename: ti.Filename,
		Hash:     ti.Hash,
		Bytes:    ti.Bytes,
		Host:     ti.Host,
		Split:    ti.Split,
		Progress: ti.Progress,
		Status:   ti.Status,
		Added:    ti.Added,
		Links:    ti.Links,
		Ended:    ti.Ended,
		Speed:    ti.Speed,
		Seeders:  ti.Seeders,
	}

	return ret, nil
}

func (t *RealDebrid) getTorrentInfo(id string) (ret *TorrentInfo, err error) {

	resp, err := t.doQuery("GET", t.baseUrl+fmt.Sprintf("/torrents/info/%s", id), nil, "application/json")
	if err != nil {
		t.logger.Error().Err(err).Msg("realdebrid: Failed to get torrent")
		return nil, fmt.Errorf("realdebrid: Failed to get torrent: %w", err)
	}

	err = json.Unmarshal(resp, &ret)
	if err != nil {
		t.logger.Error().Err(err).Msg("realdebrid: Failed to parse torrent")
		return nil, fmt.Errorf("realdebrid: Failed to parse torrent: %w", err)
	}

	return ret, nil
}

func toDebridTorrent(t *Torrent) (ret *debrid.TorrentItem) {

	status := toDebridTorrentStatus(t)

	ret = &debrid.TorrentItem{
		ID:                   t.ID,
		Name:                 t.Filename,
		Hash:                 t.Hash,
		Size:                 t.Bytes,
		FormattedSize:        util.Bytes(uint64(t.Bytes)),
		CompletionPercentage: int(t.Progress),
		ETA:                  "",
		Status:               status,
		AddedAt:              t.Added,
		Speed:                util.ToHumanReadableSpeed(int(t.Speed)),
		Seeders:              t.Seeders,
		IsReady:              status == debrid.TorrentItemStatusCompleted,
	}

	return
}

func toDebridTorrentInfo(t *TorrentInfo) (ret *debrid.TorrentInfo) {

	var files []*debrid.TorrentItemFile
	for _, f := range t.Files {
		name := filepath.Base(f.Path)

		files = append(files, &debrid.TorrentItemFile{
			ID:    strconv.Itoa(f.ID),
			Index: f.ID,
			Name:  name,   // e.g. "Big Buck Bunny.mp4"
			Path:  f.Path, // e.g. "/Big Buck Bunny/Big Buck Bunny.mp4"
			Size:  f.Bytes,
		})
	}

	ret = &debrid.TorrentInfo{
		ID:    &t.ID,
		Name:  t.Filename,
		Hash:  t.Hash,
		Size:  t.OriginalBytes,
		Files: files,
	}

	return
}

func toDebridTorrentStatus(t *Torrent) debrid.TorrentItemStatus {
	switch t.Status {
	case "downloading", "queued":
		return debrid.TorrentItemStatusDownloading
	case "waiting_files_selection", "magnet_conversion":
		return debrid.TorrentItemStatusStalled
	case "downloaded", "dead":
		return debrid.TorrentItemStatusCompleted
	case "uploading":
		return debrid.TorrentItemStatusSeeding
	case "paused":
		return debrid.TorrentItemStatusPaused
	default:
		return debrid.TorrentItemStatusOther
	}
}

func (t *RealDebrid) DeleteTorrent(id string) error {

	_, err := t.doQuery("DELETE", t.baseUrl+fmt.Sprintf("/torrents/delete/%s", id), nil, "application/json")
	if err != nil {
		t.logger.Error().Err(err).Msg("realdebrid: Failed to delete torrent")
		return fmt.Errorf("realdebrid: Failed to delete torrent: %w", err)
	}

	return nil
}
