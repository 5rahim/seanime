package alldebrid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
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
	AllDebrid struct {
		baseUrl string
		apiKey  mo.Option[string]
		client  *http.Client
		logger  *zerolog.Logger
	}

	Response struct {
		Status string      `json:"status"`
		Data   interface{} `json:"data"`
		Error  *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	AddMagnetResponse struct {
		Magnets []struct {
			Magnet string `json:"magnet"`
			Hash   string `json:"hash"`
			Name   string `json:"name"`
			Size   int64  `json:"size"`
			Ready  bool   `json:"ready"`
			ID     int64  `json:"id"`
			Error  *struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error,omitempty"`
		} `json:"magnets"`
	}

	AddTorrentFileResponse struct {
		Files []struct {
			File  string `json:"file"`
			Name  string `json:"name"`
			Size  int64  `json:"size"`
			Hash  string `json:"hash"`
			Ready bool   `json:"ready"`
			ID    int64  `json:"id"`
			Error *struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error,omitempty"`
		} `json:"files"`
	}

	GetTorrentsResponse struct {
		Magnets []Torrent `json:"magnets"`
	}

	// FileNode represents a node in the AllDebrid file tree
	// It can be either a file (has Link) or a folder (has Entries)
	FileNode struct {
		Name    string     `json:"n"`           // Name of file or folder
		Size    int64      `json:"s,omitempty"` // Size (only for files)
		Link    string     `json:"l,omitempty"` // Download link (only for files)
		Entries []FileNode `json:"e,omitempty"` // Child entries (only for folders)
	}

	GetTorrentFilesResponse struct {
		Magnets []struct {
			ID    string     `json:"id"`
			Files []FileNode `json:"files,omitempty"`
			Error *struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error,omitempty"`
		} `json:"magnets"`
	}

	Torrent struct {
		ID             int64  `json:"id"`
		Filename       string `json:"filename"`
		Size           int64  `json:"size"`
		Hash           string `json:"hash"`
		Status         string `json:"status"` // "Ready", "Downloading", "No peer after 30 minutes", etc.
		StatusCode     int    `json:"statusCode"`
		Downloaded     int64  `json:"downloaded,omitempty"`
		Uploaded       int64  `json:"uploaded,omitempty"`
		Seeders        int    `json:"seeders,omitempty"`
		DownloadSpeed  int64  `json:"downloadSpeed,omitempty"`
		UploadSpeed    int64  `json:"uploadSpeed,omitempty"`
		UploadDate     int64  `json:"uploadDate"`
		CompletionDate int64  `json:"completionDate"`
		Type           string `json:"type,omitempty"`
		Notified       bool   `json:"notified,omitempty"`
		Version        int    `json:"version,omitempty"`
		NbLinks        int    `json:"nbLinks,omitempty"`
	}

	GetTorrentResponse struct {
		Magnets Torrent `json:"magnets"`
	}

	UnrestrictLinkResponse struct {
		Link      string `json:"link"`
		Host      string `json:"host"`
		Filename  string `json:"filename"`
		Streaming bool   `json:"streaming"`
		Filesize  int64  `json:"filesize"`
		ID        string `json:"id"`
	}
)

func NewAllDebrid(logger *zerolog.Logger) debrid.Provider {
	return &AllDebrid{
		baseUrl: "https://api.alldebrid.com/v4.1",
		apiKey:  mo.None[string](),
		client: &http.Client{
			Timeout: time.Second * 30,
		},
		logger: logger,
	}
}

func (a *AllDebrid) GetSettings() debrid.Settings {
	return debrid.Settings{
		ID:   "alldebrid",
		Name: "AllDebrid",
	}
}

func (a *AllDebrid) CheckAuth() error {
	_, err := a.doQuery("GET", "/user", nil, "")
	return err
}

func (a *AllDebrid) Authenticate(apiKey string) error {
	a.apiKey = mo.Some(apiKey)
	return a.CheckAuth()
}

func (a *AllDebrid) doQuery(method, endpoint string, body io.Reader, contentType string) (*Response, error) {
	apiKey, found := a.apiKey.Get()
	if !found {
		return nil, debrid.ErrNotAuthenticated
	}

	var u *url.URL
	var err error
	if strings.HasPrefix(endpoint, "http") {
		u, err = url.Parse(endpoint)
	} else {
		u, err = url.Parse(a.baseUrl + endpoint)
	}

	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("agent", "Seanime")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}
	
	a.logger.Debug().Str("method", method).Str("url", u.String()).Msg("alldebrid: doQuery")

	req.Header.Set("Authorization", "Bearer "+apiKey)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return a.doQueryCtx(context.Background(), method, u.String(), body, contentType)
}

func (a *AllDebrid) doQueryCtx(ctx context.Context, method, endpoint string, body io.Reader, contentType string) (*Response, error) {
	apiKey, found := a.apiKey.Get()
	if !found {
		return nil, debrid.ErrNotAuthenticated
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Add("User-Agent", "Seanime/"+constants.Version)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("api request failed with status: %s", resp.Status)
	}

	var ret Response
	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if ret.Status != "success" {
		msg := "unknown error"
		if ret.Error != nil {
			msg = ret.Error.Message
		}
		return nil, fmt.Errorf("api returned error: %s", msg)
	}

	return &ret, nil
}

// AddTorrent uploads a magnet link or a torrent file from a URL
func (a *AllDebrid) AddTorrent(opts debrid.AddTorrentOptions) (string, error) {
	a.logger.Debug().Msgf("alldebrid: AddTorrent called with: %s", opts.MagnetLink)

	if strings.HasPrefix(opts.MagnetLink, "http") {
		a.logger.Debug().Msg("alldebrid: detected http link, using addTorrentFile")
		return a.addTorrentFile(opts.MagnetLink)
	}

	// Endpoint: /magnet/upload
	
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	
	err := writer.WriteField("magnets[]", opts.MagnetLink)
	if err != nil {
		return "", err
	}
	writer.Close()
	
	resp, err := a.doQuery("POST", "/magnet/upload", &body, writer.FormDataContentType())
	if err != nil {
		a.logger.Error().Err(err).Msgf("alldebrid: AddTorrent failed. URL: %s/magnet/upload", a.baseUrl)
		return "", err
	}

	var data AddMagnetResponse
	b, _ := json.Marshal(resp.Data)
	if err := json.Unmarshal(b, &data); err != nil {
		return "", err
	}
	
	if len(data.Magnets) == 0 {
		return "", fmt.Errorf("no magnet added")
	}

	if data.Magnets[0].Error != nil {
		return "", fmt.Errorf("api error: %s", data.Magnets[0].Error.Message)
	}
	
	return strconv.FormatInt(data.Magnets[0].ID, 10), nil
}

func (a *AllDebrid) addTorrentFile(urlStr string) (string, error) {
	// Download the .torrent file
	resp, err := http.Get(urlStr)
	if err != nil {
		return "", fmt.Errorf("failed to download torrent file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to download torrent file, status: %s", resp.Status)
	}

	// Read content
	fileContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Check if it's a magnet link in a text file (or HTML)
	// If content-type is text/html, scan for magnet link
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		// Try to find regex magnet
		// Find in fileContent
		// We limit the search to avoid memory issues if file is huge, but HTML shouldn't be too huge
		sContent := string(fileContent)
		// Find magnet
		// Using a simplified approach
		if idx := strings.Index(sContent, "magnet:?xt=urn:btih:"); idx != -1 {
			// Extract till quote or whitespace
			endIdx := strings.IndexAny(sContent[idx:], "\"'\n\r\t <")
			if endIdx != -1 {
				magnet := sContent[idx : idx+endIdx]
				// Decode html entities if any?
				magnet = strings.ReplaceAll(magnet, "&amp;", "&")
				a.logger.Debug().Msgf("alldebrid: found magnet link in HTML: %s", magnet)
				return a.AddTorrent(debrid.AddTorrentOptions{MagnetLink: magnet})
			}
		}
		// If no magnet found, fail
		return "", fmt.Errorf("invalid torrent file (html content) and no magnet found")
	}

	// Prepare upload
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	
	part, err := writer.CreateFormFile("files[]", "torrent.torrent")
	if err != nil {
		return "", err
	}
	if _, err := part.Write(fileContent); err != nil {
		return "", err
	}
	writer.Close()

	// Upload to /magnet/upload/file
	apiResp, err := a.doQuery("POST", "/magnet/upload/file", &body, writer.FormDataContentType())
	if err != nil {
		return "", err
	}

	var data AddTorrentFileResponse
	b, _ := json.Marshal(apiResp.Data)
	if err := json.Unmarshal(b, &data); err != nil {
		return "", err
	}

	if len(data.Files) == 0 {
		return "", fmt.Errorf("no file added")
	}

	if data.Files[0].Error != nil {
		return "", fmt.Errorf("api error: %s", data.Files[0].Error.Message)
	}

	return strconv.FormatInt(data.Files[0].ID, 10), nil
}

func (a *AllDebrid) GetTorrentStreamUrl(ctx context.Context, opts debrid.StreamTorrentOptions, itemCh chan debrid.TorrentItem) (streamUrl string, err error) {
	
	doneCh := make(chan struct{})
	
	go func(ctx context.Context) {
		defer close(doneCh)
		
		for {
			select {
			case <-ctx.Done():
				err = ctx.Err()
				return
			case <-time.After(time.Second * 5):
				// Check status
				tInfo, sErr := a.GetTorrent(opts.ID)
				if sErr != nil {
					// Logic to retry or fail?
					a.logger.Error().Err(sErr).Msg("alldebrid: Failed to get torrent status")
					continue // Retry
				}
				
				itemCh <- *tInfo
				
				if tInfo.IsReady {
					// Get download link
					// We need to find the link that matches the file selected
					// 'opts.FileId' should correspond to a file index or ID in the torrent.
					// AllDebrid links are usually just a list.
					// We need 'GetTorrentInfo' which returns files list and match?
					// Or 'GetTorrent' logic.
					
					// Let's call GetTorrentDownloadUrl
					url, dErr := a.GetTorrentDownloadUrl(debrid.DownloadTorrentOptions{
						ID: opts.ID,
						FileId: opts.FileId,
					})
					if dErr != nil {
						a.logger.Error().Err(dErr).Msg("alldebrid: failed to get download url")
						return
					}
					
					streamUrl = url
					return
				}
			}
		}
	}(ctx)
	
	<-doneCh
	return
}

// flattenFileTree recursively traverses the AllDebrid file tree and extracts all files with links
func flattenFileTree(nodes []FileNode, basePath string) []struct {
	Name string
	Path string
	Size int64
	Link string
} {
	var files []struct {
		Name string
		Path string
		Size int64
		Link string
	}

	for _, node := range nodes {
		currentPath := basePath
		if currentPath != "" {
			currentPath += "/"
		}
		currentPath += node.Name

		// If it has a link, it's a file
		if node.Link != "" {
			files = append(files, struct {
				Name string
				Path string
				Size int64
				Link string
			}{
				Name: node.Name,
				Path: currentPath,
				Size: node.Size,
				Link: node.Link,
			})
		}

		// If it has entries, it's a folder - recurse
		if len(node.Entries) > 0 {
			files = append(files, flattenFileTree(node.Entries, currentPath)...)
		}
	}

	return files
}

func (a *AllDebrid) GetTorrentDownloadUrl(opts debrid.DownloadTorrentOptions) (string, error) {
	// 1. Get files/links
	filesResp, err := a.getTorrentFiles(opts.ID)
	if err != nil {
		return "", err
	}
	
	if len(filesResp.Magnets) == 0 {
		return "", fmt.Errorf("magnet not found")
	}
	
	info := filesResp.Magnets[0]
	if info.Error != nil {
		return "", fmt.Errorf("api error: %s", info.Error.Message)
	}

	// Flatten the hierarchical file tree
	flatFiles := flattenFileTree(info.Files, "")
	
	if len(flatFiles) == 0 {
		return "", fmt.Errorf("no files found in torrent")
	}

	// If FileId is provided, return only that file's download URL
	if opts.FileId != "" {
		idx, err := strconv.Atoi(opts.FileId)
		if err != nil {
			return "", fmt.Errorf("invalid file id: %s", opts.FileId)
		}
		
		if idx < 0 || idx >= len(flatFiles) {
			return "", fmt.Errorf("file index out of range")
		}
		
		// Unlock/Unrestrict the link
		return a.unlockLink(flatFiles[idx].Link)
	}

	// If no FileId is provided, return all file download URLs as comma-separated string
	downloadUrls := make([]string, 0, len(flatFiles))
	for _, file := range flatFiles {
		unlockedUrl, err := a.unlockLink(file.Link)
		if err != nil {
			a.logger.Error().Err(err).Str("fileName", file.Name).Msg("alldebrid: Failed to unlock link for file")
			continue
		}
		
		downloadUrls = append(downloadUrls, unlockedUrl)
	}

	if len(downloadUrls) == 0 {
		return "", fmt.Errorf("no download URLs available")
	}

	return strings.Join(downloadUrls, ","), nil
}



func (a *AllDebrid) GetInstantAvailability(hashes []string) map[string]debrid.TorrentItemInstantAvailability {
	// AllDebrid does not have a dedicated instant availability endpoint that checks for cached torrents without adding them.
	// We return an empty map to indicate no instant availability check is performed.
	return make(map[string]debrid.TorrentItemInstantAvailability)
}

func (a *AllDebrid) GetTorrent(id string) (*debrid.TorrentItem, error) {
	st, err := a.getTorrent(id)
	if err != nil {
		return nil, err
	}
	return toDebridTorrent(st), nil
}

func (a *AllDebrid) GetTorrentInfo(opts debrid.GetTorrentInfoOptions) (*debrid.TorrentInfo, error) {
	// Similar to RealDebrid approach: Add -> Get Info -> Delete
	
	if opts.MagnetLink == "" {
		return nil, fmt.Errorf("magnet link required")
	}
	
	id, err := a.AddTorrent(debrid.AddTorrentOptions{MagnetLink: opts.MagnetLink})
	if err != nil {
		return nil, fmt.Errorf("failed to add torrent for info: %w", err)
	}
	
	// Fetch info
	status, err := a.getTorrent(id)
	if err != nil {
		a.DeleteTorrent(id)
		return nil, err
	}
	
	// Get files to list them
	filesResp, err := a.getTorrentFiles(id)
	if err != nil {
		a.DeleteTorrent(id)
		return nil, err
	}
	
	if len(filesResp.Magnets) == 0 {
		a.DeleteTorrent(id)
		return nil, fmt.Errorf("magnet files not found")
	}
	
	filesInfo := filesResp.Magnets[0]

	// Create info
	ret := &debrid.TorrentInfo{
		ID:   &id, // we return the temporary ID
		Name: status.Filename,
		Hash: status.Hash,
		Size: status.Size,
	}
	
	if filesInfo.Files != nil {
		for i, l := range filesInfo.Files {
			ret.Files = append(ret.Files, &debrid.TorrentItemFile{
				ID:    strconv.Itoa(i),
				Index: i,
				Name:  l.Name,
				Path:  l.Name, 
				Size:  l.Size,
			})
		}
	}

	// Delete
	a.DeleteTorrent(id)
	
	return ret, nil
}

func (a *AllDebrid) GetTorrents() ([]*debrid.TorrentItem, error) {
	endpoint := "/magnet/status"
	
	// v4.1 API requires POST, not GET
	resp, err := a.doQuery("POST", endpoint, nil, "")
	if err != nil {
		return nil, err
	}
	
	var data GetTorrentsResponse
	b, _ := json.Marshal(resp.Data)
	json.Unmarshal(b, &data)

	var ret []*debrid.TorrentItem
	for _, m := range data.Magnets {
		ret = append(ret, toDebridTorrent(&m))
	}
	
	// Sort by ID desc
	slices.SortFunc(ret, func(i, j *debrid.TorrentItem) int {
		return strings.Compare(j.ID, i.ID)
	})

	return ret, nil
}

func (a *AllDebrid) DeleteTorrent(id string) error {
	u, _ := url.Parse(a.baseUrl + "/magnet/delete")
	q := u.Query()
	apiKey, _ := a.apiKey.Get()
	q.Set("agent", "Seanime")
	q.Set("apikey", apiKey)
	q.Set("id", id)
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest("GET", u.String(), nil)
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Helpers

func (a *AllDebrid) getTorrent(id string) (*Torrent, error) {
	endpoint := "/magnet/status"
	
	var body io.Reader
	var contentType string
	
	if id != "" {
		// v4.1 API requires POST with form data, not GET with query params
		var formBody bytes.Buffer
		writer := multipart.NewWriter(&formBody)
		writer.WriteField("id", id)
		writer.Close()
		
		body = &formBody
		contentType = writer.FormDataContentType()
	}
	
	resp, err := a.doQuery("POST", endpoint, body, contentType)
	if err != nil {
		return nil, err
	}
	
	if id != "" {
		var data GetTorrentResponse
		b, _ := json.Marshal(resp.Data)
		json.Unmarshal(b, &data)
		
		if data.Magnets.ID == 0 {
			a.logger.Error().Any("data", data).Msg("alldebrid: getTorrent - magnet not found in response")
			return nil, fmt.Errorf("magnet not found")
		}
		return &data.Magnets, nil
	}
	
	// This branch should mostly not be used by this helper as it's typically called with ID
	// But if it is...
	var data GetTorrentsResponse
	b, _ := json.Marshal(resp.Data)
	json.Unmarshal(b, &data)
	
	if len(data.Magnets) == 0 {
		return nil, fmt.Errorf("magnet not found")
	}
	
	return &data.Magnets[0], nil
}

func (a *AllDebrid) getTorrentFiles(id string) (*GetTorrentFilesResponse, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	err := writer.WriteField("id[]", id)
	if err != nil {
		return nil, err
	}
	writer.Close()

	resp, err := a.doQuery("POST", "/magnet/files", &body, writer.FormDataContentType())
	if err != nil {
		return nil, err
	}
	
	var data GetTorrentFilesResponse
	b, _ := json.Marshal(resp.Data)
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	
	return &data, nil
}

func (a *AllDebrid) unlockLink(link string) (string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	err := writer.WriteField("link", link)
	if err != nil {
		return "", err
	}
	writer.Close()

	resp, err := a.doQuery("POST", "/link/unlock", &body, writer.FormDataContentType())
	if err != nil {
		return "", err
	}
	
	var data UnrestrictLinkResponse
	b, _ := json.Marshal(resp.Data)
	json.Unmarshal(b, &data)
	
	return data.Link, nil
}

func toDebridTorrent(t *Torrent) *debrid.TorrentItem {
	status := toDebridTorrentStatus(t)
	
	// Convert Unix timestamp to RFC3339 format
	addedAt := ""
	if t.UploadDate > 0 {
		addedAt = time.Unix(t.UploadDate, 0).Format(time.RFC3339)
	}
	
	// Calculate completion percentage
	completionPercentage := 0
	if t.Size > 0 && t.Downloaded > 0 {
		completionPercentage = int((float64(t.Downloaded) / float64(t.Size)) * 100)
	} else if status == debrid.TorrentItemStatusCompleted {
		completionPercentage = 100
	}

	ret := &debrid.TorrentItem{
		ID:                   strconv.FormatInt(t.ID, 10),
		Name:                 t.Filename,
		Hash:                 t.Hash,
		Size:                 t.Size,
		FormattedSize:        util.Bytes(uint64(t.Size)),
		CompletionPercentage: completionPercentage,
		ETA: "",
		Status:               status,
		AddedAt:              addedAt,
		Speed:                util.ToHumanReadableSpeed(int(t.DownloadSpeed)),
		Seeders:              t.Seeders,
		IsReady:              status == debrid.TorrentItemStatusCompleted,
	}

	return ret
}

func toDebridTorrentStatus(t *Torrent) debrid.TorrentItemStatus {
	switch t.StatusCode {
	case 4:
		return debrid.TorrentItemStatusCompleted
	case 0, 1, 2, 3:
		return debrid.TorrentItemStatusDownloading
	case 7:
		return debrid.TorrentItemStatusError
	default:
		if strings.Contains(strings.ToLower(t.Status), "error") {
			return debrid.TorrentItemStatusError
		}
		return debrid.TorrentItemStatusOther
	}
}