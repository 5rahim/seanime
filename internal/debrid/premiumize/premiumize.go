package premiumize

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"seanime/internal/constants"
	"seanime/internal/debrid/debrid"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

type (
	// HashStore persists the info hash a transfer was created from, keyed by transfer ID.
	// The Premiumize API never returns the info hash of a transfer (only id, name, status,
	// progress, message, folder_id and file_id are exposed by /transfer/list), so this is the
	// only way to know a transfer's hash after it has been created, and the only way for that
	// knowledge to survive a process restart (needed e.g. for auto-downloader hash matching).
	// Implementations are expected to persist durably, e.g. to a database.
	HashStore interface {
		// LoadAll returns every persisted transfer id -> hash mapping.
		LoadAll() (map[string]string, error)
		// Save persists a single transfer id -> hash mapping.
		Save(transferId, hash string)
		// Delete removes a persisted mapping, e.g. once its transfer is deleted.
		Delete(transferId string)
	}

	Premiumize struct {
		baseUrl   string
		apiKey    mo.Option[string]
		client    *http.Client
		logger    *zerolog.Logger
		hashStore HashStore
		// hashCache is an in-memory, write-through cache in front of hashStore so lookups
		// don't hit the store on every call. Seeded from hashStore.LoadAll() at construction.
		hashCache *result.Map[string, string]
	}

	statusResponse struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Code    string `json:"code"`
	}

	transfer struct {
		ID       string  `json:"id"`
		Name     string  `json:"name"`
		Status   string  `json:"status"` // queued, running, finished, seeding, error
		Progress float64 `json:"progress"`
		Message  string  `json:"message"`
		FolderID string  `json:"folder_id"`
		FileID   string  `json:"file_id"`
	}

	transferListResponse struct {
		Transfers []*transfer `json:"transfers"`
	}

	transferCreateResponse struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	folderItem struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"` // "file" or "folder"
		Size int64  `json:"size"`
		Link string `json:"link"`
	}

	folderListResponse struct {
		Content []*folderItem `json:"content"`
	}

	itemDetailsResponse struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Size int64  `json:"size"`
		Link string `json:"link"`
	}

	directDLContent struct {
		Path string `json:"path"`
		Size int64  `json:"size"`
		Link string `json:"link"`
	}

	directDLResponse struct {
		Content []*directDLContent `json:"content"`
	}

	cacheCheckResponse struct {
		Response []bool        `json:"response"`
		Filename []string      `json:"filename"`
		Filesize []flexibleInt `json:"filesize"`
	}

	// flexibleInt decodes a JSON number that the Premiumize API sometimes returns as a
	// quoted string (e.g. "12345678901") instead of a bare number.
	flexibleInt int64

	// flatFile is a single downloadable file resolved from a finished transfer,
	// either directly (single file transfer) or recursively from its folder.
	flatFile struct {
		Name string
		Path string
		Size int64
		Link string
	}
)

func (f *flexibleInt) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*f = 0
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		*f = 0
		return nil
	}
	*f = flexibleInt(n)
	return nil
}

var magnetHashRegex = regexp.MustCompile(`(?i)urn:btih:([a-zA-Z0-9]+)`)

func extractInfoHash(magnet string) string {
	m := magnetHashRegex.FindStringSubmatch(magnet)
	if len(m) < 2 {
		return ""
	}
	return strings.ToLower(m[1])
}

// NewPremiumize creates a Premiumize provider. hashStore may be nil, in which case transfer
// hashes are only kept in memory and won't survive a restart (see HashStore).
func NewPremiumize(logger *zerolog.Logger, hashStore HashStore) debrid.Provider {
	p := &Premiumize{
		baseUrl: "https://www.premiumize.me/api",
		apiKey:  mo.None[string](),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:    logger,
		hashStore: hashStore,
		hashCache: result.NewMap[string, string](),
	}

	if hashStore != nil {
		hashes, err := hashStore.LoadAll()
		if err != nil {
			logger.Warn().Err(err).Msg("premiumize: Failed to load persisted transfer hashes")
		} else {
			for id, hash := range hashes {
				p.hashCache.Set(id, hash)
			}
		}
	}

	return p
}

// rememberHash records the info hash a transfer was created from, both in memory and (if
// configured) in the persistent hashStore.
func (p *Premiumize) rememberHash(transferId, hash string) {
	if hash == "" {
		return
	}
	p.hashCache.Set(transferId, hash)
	if p.hashStore != nil {
		p.hashStore.Save(transferId, hash)
	}
}

// forgetHash removes a transfer's recorded hash, both in memory and (if configured) in the
// persistent hashStore.
func (p *Premiumize) forgetHash(transferId string) {
	p.hashCache.Delete(transferId)
	if p.hashStore != nil {
		p.hashStore.Delete(transferId)
	}
}

func (p *Premiumize) GetSettings() debrid.Settings {
	return debrid.Settings{
		ID:   "premiumize",
		Name: "Premiumize",
	}
}

func (p *Premiumize) Authenticate(apiKey string) error {
	p.apiKey = mo.Some(apiKey)

	if _, err := p.doQuery("GET", "/account/info", nil); err != nil {
		return fmt.Errorf("%w: %v", debrid.ErrFailedToAuthenticate, err)
	}

	return nil
}

// doQuery calls the Premiumize API and returns the raw JSON body on success.
// GET requests send params as a query string, non-GET requests send them as a
// form-encoded body. Every Premiumize response carries a top-level "status" field
// which is checked here regardless of the HTTP status code.
func (p *Premiumize) doQuery(method, endpoint string, params url.Values) ([]byte, error) {
	return p.doQueryCtx(context.Background(), method, endpoint, params)
}

func (p *Premiumize) doQueryCtx(ctx context.Context, method, endpoint string, params url.Values) ([]byte, error) {
	apiKey, found := p.apiKey.Get()
	if !found {
		return nil, debrid.ErrNotAuthenticated
	}

	fullUrl := p.baseUrl + endpoint

	var body io.Reader
	if method == http.MethodGet {
		if len(params) > 0 {
			fullUrl += "?" + params.Encode()
		}
	} else {
		body = strings.NewReader(params.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, fullUrl, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("User-Agent", "Seanime/"+constants.Version)
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var status statusResponse
	if err := json.Unmarshal(content, &status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if status.Status != "success" {
		msg := status.Message
		if msg == "" {
			msg = "unknown error"
		}
		return nil, fmt.Errorf("api returned error: %s", msg)
	}

	return content, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// AddTorrent creates a new transfer from a magnet link (or a direct URL to a .torrent file, which
// Premiumize also accepts as "src"). Since the Premiumize API never exposes the info hash of a
// transfer (see hashCache), the only way to avoid creating duplicate transfers for a torrent we
// already added is to check our own in-memory hashCache first.
func (p *Premiumize) AddTorrent(opts debrid.AddTorrentOptions) (string, error) {
	p.logger.Debug().Str("src", opts.MagnetLink).Msg("premiumize: AddTorrent called")

	hash := opts.InfoHash
	if hash == "" {
		hash = extractInfoHash(opts.MagnetLink)
	}

	if id, found := p.findExistingTransferId(hash); found {
		p.logger.Debug().Str("torrentId", id).Msg("premiumize: Torrent already added")
		return id, nil
	}

	params := url.Values{}
	params.Set("src", opts.MagnetLink)

	resp, err := p.doQuery("POST", "/transfer/create", params)
	if err != nil {
		return "", fmt.Errorf("premiumize: failed to add torrent: %w", err)
	}

	var data transferCreateResponse
	if err := json.Unmarshal(resp, &data); err != nil {
		return "", fmt.Errorf("premiumize: failed to parse response: %w", err)
	}

	if data.ID == "" {
		return "", fmt.Errorf("premiumize: no transfer id returned")
	}

	p.rememberHash(data.ID, hash)

	p.logger.Debug().Str("torrentId", data.ID).Str("torrentName", data.Name).Msg("premiumize: Torrent added")

	return data.ID, nil
}

// findExistingTransferId looks up a transfer ID we previously created for the given hash.
func (p *Premiumize) findExistingTransferId(hash string) (string, bool) {
	if hash == "" {
		return "", false
	}

	var id string
	found := false
	p.hashCache.Range(func(tID string, h string) bool {
		if h == hash {
			id = tID
			found = true
			return false
		}
		return true
	})

	return id, found
}

// GetTorrentStreamUrl blocks until the torrent is downloaded and returns the stream URL for the torrent file by calling GetTorrentDownloadUrl.
func (p *Premiumize) GetTorrentStreamUrl(ctx context.Context, opts debrid.StreamTorrentOptions, itemCh chan debrid.TorrentItem) (streamUrl string, err error) {
	p.logger.Trace().Str("torrentId", opts.ID).Str("fileId", opts.FileId).Msg("premiumize: Retrieving stream link")

	doneCh := make(chan struct{})

	go func(ctx context.Context) {
		defer close(doneCh)

		for {
			select {
			case <-ctx.Done():
				err = ctx.Err()
				return
			case <-time.After(5 * time.Second):
				item, sErr := p.GetTorrent(opts.ID)
				if sErr != nil {
					p.logger.Error().Err(sErr).Msg("premiumize: Failed to get torrent status")
					continue
				}

				select {
				case itemCh <- *item:
				default:
				}

				if item.IsReady {
					time.Sleep(1 * time.Second)

					url, dErr := p.GetTorrentDownloadUrl(debrid.DownloadTorrentOptions{
						ID:     opts.ID,
						FileId: opts.FileId,
					})
					if dErr != nil {
						p.logger.Error().Err(dErr).Msg("premiumize: Failed to get download URL")
						err = dErr
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

// GetTorrentDownloadUrl returns the download URL for the torrent.
// If no opts.FileId is provided, it returns a comma-separated list of download URLs for all files.
func (p *Premiumize) GetTorrentDownloadUrl(opts debrid.DownloadTorrentOptions) (string, error) {
	p.logger.Trace().Str("torrentId", opts.ID).Msg("premiumize: Retrieving download link")

	tr, err := p.getTransfer(opts.ID)
	if err != nil {
		return "", fmt.Errorf("premiumize: failed to get download url: %w", err)
	}

	if !isReady(tr.Status) {
		return "", fmt.Errorf("premiumize: torrent is not ready")
	}

	files, _, err := p.resolveFiles(tr)
	if err != nil {
		return "", fmt.Errorf("premiumize: failed to get download url: %w", err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("premiumize: no files found")
	}

	if opts.FileId != "" {
		idx, err := strconv.Atoi(opts.FileId)
		if err != nil || idx < 0 || idx >= len(files) {
			return "", fmt.Errorf("premiumize: invalid file id: %s", opts.FileId)
		}
		return files[idx].Link, nil
	}

	links := make([]string, 0, len(files))
	for _, f := range files {
		links = append(links, f.Link)
	}

	return strings.Join(links, ","), nil
}

// GetInstantAvailability checks whether torrents are cached on Premiumize using /cache/check.
// Premiumize only reports a single aggregate filename/filesize per hash (not a per-file
// breakdown), so cached hashes are reported with one synthetic "0" entry in CachedFiles.
func (p *Premiumize) GetInstantAvailability(hashes []string) map[string]debrid.TorrentItemInstantAvailability {
	p.logger.Trace().Strs("hashes", hashes).Msg("premiumize: Checking instant availability")

	availability := make(map[string]debrid.TorrentItemInstantAvailability)

	if len(hashes) == 0 {
		return availability
	}

	const batchSize = 100
	for i := 0; i < len(hashes); i += batchSize {
		end := i + batchSize
		if end > len(hashes) {
			end = len(hashes)
		}
		batch := hashes[i:end]

		params := url.Values{}
		for _, h := range batch {
			params.Add("items[]", "magnet:?xt=urn:btih:"+h)
		}

		resp, err := p.doQuery("POST", "/cache/check", params)
		if err != nil {
			p.logger.Error().Err(err).Msg("premiumize: Failed to check cache")
			continue
		}

		var data cacheCheckResponse
		if err := json.Unmarshal(resp, &data); err != nil {
			p.logger.Error().Err(err).Msg("premiumize: Failed to parse cache check response")
			continue
		}

		for idx, hash := range batch {
			if idx >= len(data.Response) || !data.Response[idx] {
				continue
			}

			var name string
			if idx < len(data.Filename) {
				name = data.Filename[idx]
			}
			var size int64
			if idx < len(data.Filesize) {
				size = int64(data.Filesize[idx])
			}

			availability[hash] = debrid.TorrentItemInstantAvailability{
				CachedFiles: map[string]*debrid.CachedFile{
					"0": {Name: name, Size: size},
				},
			}
		}
	}

	return availability
}

func (p *Premiumize) GetTorrent(id string) (*debrid.TorrentItem, error) {
	tr, err := p.getTransfer(id)
	if err != nil {
		return nil, err
	}

	item := p.toDebridTorrent(tr)

	// Resolve files (and total size) once the transfer's content is available in the cloud.
	if item.IsReady {
		files, size, err := p.resolveFiles(tr)
		if err != nil {
			p.logger.Error().Err(err).Str("transferId", id).Msg("premiumize: Failed to resolve files")
		} else {
			item.Size = size
			item.FormattedSize = util.Bytes(uint64(size))
			for idx, f := range files {
				item.Files = append(item.Files, &debrid.TorrentItemFile{
					ID:    strconv.Itoa(idx),
					Index: idx,
					Name:  f.Name,
					Path:  f.Path,
					Size:  f.Size,
				})
			}
		}
	}

	return item, nil
}

// GetTorrentInfo resolves a magnet link's contents. Premiumize has no metadata-only resolution
// step like other providers (RealDebrid/AllDebrid can add a torrent and read its file list before
// any data is downloaded); a Premiumize transfer only exposes its files once it has *finished*
// downloading into the user's cloud storage (see the "transfer" type and resolveFiles).
//
// This never adds/starts a transfer as a side effect of a lookup: this is also called by
// autoselect to evaluate multiple candidate torrents per episode, and silently kicking off real
// downloads for candidates that end up rejected would be wasteful and surprising. So it only
// ever reads existing state:
//  1. If we previously started downloading this exact torrent (tracked via hashCache), check its
//     current status: if finished, return its real files; if still in progress, report progress.
//  2. Otherwise, try /transfer/directdl, which resolves files instantly but only succeeds if the
//     torrent is already cached by Premiumize (shared cache, independent of the user's account).
//  3. Otherwise, the torrent isn't cached and there is no way to preview it without downloading
//     it first, so this fails. Callers that want to actually download an uncached torrent should
//     go through AddTorrent instead.
//
// (Reached out to Premiumize about exposing the file list before a transfer finishes, like RD/
// TorBox/AllDebrid do - if/when that lands this whole function gets a lot simpler.)
func (p *Premiumize) GetTorrentInfo(opts debrid.GetTorrentInfoOptions) (*debrid.TorrentInfo, error) {
	if opts.MagnetLink == "" {
		return nil, fmt.Errorf("premiumize: magnet link required")
	}

	hash := opts.InfoHash
	if hash == "" {
		hash = extractInfoHash(opts.MagnetLink)
	}

	if id, found := p.findExistingTransferId(hash); found {
		tr, err := p.getTransfer(id)
		if err == nil {
			if isReady(tr.Status) {
				files, size, fErr := p.resolveFiles(tr)
				if fErr == nil {
					return filesToTorrentInfo(&id, tr.Name, hash, size, toTorrentItemFiles(files)), nil
				}
			} else if tr.Status != "error" {
				return nil, fmt.Errorf("premiumize: torrent is downloading (%.0f%%%s), it must finish before its files can be previewed", tr.Progress*100, statusSuffix(tr.Message))
			}
			// If the transfer errored out, fall through and try the cache check below.
		}
	}

	params := url.Values{}
	params.Set("src", opts.MagnetLink)

	resp, err := p.doQuery("POST", "/transfer/directdl", params)
	if err == nil {
		var data directDLResponse
		if jErr := json.Unmarshal(resp, &data); jErr == nil && len(data.Content) > 0 {
			var files []*debrid.TorrentItemFile
			var totalSize int64
			for idx, c := range data.Content {
				name := c.Path
				if i := strings.LastIndex(c.Path, "/"); i != -1 {
					name = c.Path[i+1:]
				}
				files = append(files, &debrid.TorrentItemFile{
					ID:    strconv.Itoa(idx),
					Index: idx,
					Name:  name,
					Path:  c.Path,
					Size:  c.Size,
				})
				totalSize += c.Size
			}
			return filesToTorrentInfo(nil, files[0].Name, hash, totalSize, files), nil
		}
	}

	return nil, fmt.Errorf("premiumize: torrent is not cached and cannot be previewed without downloading it first")
}

func statusSuffix(message string) string {
	if message == "" {
		return ""
	}
	return ", " + message
}

func toTorrentItemFiles(files []*flatFile) []*debrid.TorrentItemFile {
	ret := make([]*debrid.TorrentItemFile, 0, len(files))
	for idx, f := range files {
		ret = append(ret, &debrid.TorrentItemFile{
			ID:    strconv.Itoa(idx),
			Index: idx,
			Name:  f.Name,
			Path:  f.Path,
			Size:  f.Size,
		})
	}
	return ret
}

// filesToTorrentInfo builds a debrid.TorrentInfo. When id is non-nil, the torrent has actually
// been added to the user's cloud storage (id is set per the debrid.TorrentInfo.ID contract).
func filesToTorrentInfo(id *string, name, hash string, size int64, files []*debrid.TorrentItemFile) *debrid.TorrentInfo {
	return &debrid.TorrentInfo{
		ID:    id,
		Name:  name,
		Hash:  hash,
		Size:  size,
		Files: files,
	}
}

func (p *Premiumize) GetTorrents() ([]*debrid.TorrentItem, error) {
	transfers, err := p.getTransfers()
	if err != nil {
		return nil, fmt.Errorf("premiumize: failed to get torrents: %w", err)
	}

	ret := make([]*debrid.TorrentItem, 0, len(transfers))
	for _, tr := range transfers {
		ret = append(ret, p.toDebridTorrent(tr))
	}

	return ret, nil
}

func (p *Premiumize) DeleteTorrent(id string) error {
	params := url.Values{}
	params.Set("id", id)

	if _, err := p.doQuery("POST", "/transfer/delete", params); err != nil {
		return fmt.Errorf("premiumize: failed to delete torrent: %w", err)
	}

	p.forgetHash(id)

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (p *Premiumize) getTransfers() ([]*transfer, error) {
	resp, err := p.doQuery("GET", "/transfer/list", nil)
	if err != nil {
		return nil, fmt.Errorf("premiumize: failed to get transfers: %w", err)
	}

	var data transferListResponse
	if err := json.Unmarshal(resp, &data); err != nil {
		return nil, fmt.Errorf("premiumize: failed to parse transfers: %w", err)
	}

	return data.Transfers, nil
}

func (p *Premiumize) getTransfer(id string) (*transfer, error) {
	transfers, err := p.getTransfers()
	if err != nil {
		return nil, err
	}

	for _, tr := range transfers {
		if tr.ID == id {
			return tr, nil
		}
	}

	return nil, fmt.Errorf("premiumize: transfer not found")
}

func (p *Premiumize) getItemDetails(id string) (*itemDetailsResponse, error) {
	resp, err := p.doQuery("GET", "/item/details", url.Values{"id": {id}})
	if err != nil {
		return nil, err
	}

	var data itemDetailsResponse
	if err := json.Unmarshal(resp, &data); err != nil {
		return nil, fmt.Errorf("premiumize: failed to parse item details: %w", err)
	}

	return &data, nil
}

// resolveFiles returns the flattened list of downloadable files for a finished transfer,
// along with their total size. A transfer resolves to either a single file (FileID) or a
// folder (FolderID) that may itself contain subfolders, which are traversed recursively.
func (p *Premiumize) resolveFiles(tr *transfer) ([]*flatFile, int64, error) {
	if tr.FileID != "" {
		item, err := p.getItemDetails(tr.FileID)
		if err != nil {
			return nil, 0, err
		}
		return []*flatFile{{Name: item.Name, Path: item.Name, Size: item.Size, Link: item.Link}}, item.Size, nil
	}

	if tr.FolderID != "" {
		files, err := p.listFolderFlat(tr.FolderID, "")
		if err != nil {
			return nil, 0, err
		}

		var total int64
		for _, f := range files {
			total += f.Size
		}

		return files, total, nil
	}

	return nil, 0, nil
}

func (p *Premiumize) listFolderFlat(folderId, basePath string) ([]*flatFile, error) {
	resp, err := p.doQuery("GET", "/folder/list", url.Values{"id": {folderId}})
	if err != nil {
		return nil, err
	}

	var data folderListResponse
	if err := json.Unmarshal(resp, &data); err != nil {
		return nil, fmt.Errorf("premiumize: failed to parse folder list: %w", err)
	}

	var files []*flatFile
	for _, item := range data.Content {
		path := item.Name
		if basePath != "" {
			path = basePath + "/" + item.Name
		}

		if item.Type == "folder" {
			sub, err := p.listFolderFlat(item.ID, path)
			if err != nil {
				return nil, err
			}
			files = append(files, sub...)
			continue
		}

		files = append(files, &flatFile{
			Name: item.Name,
			Path: path,
			Size: item.Size,
			Link: item.Link,
		})
	}

	return files, nil
}

func isReady(status string) bool {
	return status == "finished" || status == "seeding"
}

func (p *Premiumize) toDebridTorrent(tr *transfer) *debrid.TorrentItem {
	status := toDebridTorrentStatus(tr.Status)
	hash, _ := p.hashCache.Get(tr.ID)

	return &debrid.TorrentItem{
		ID:                   tr.ID,
		Name:                 tr.Name,
		Hash:                 hash,
		CompletionPercentage: int(tr.Progress * 100),
		Status:               status,
		IsReady:              isReady(tr.Status),
	}
}

func toDebridTorrentStatus(status string) debrid.TorrentItemStatus {
	switch status {
	case "queued":
		return debrid.TorrentItemStatusStalled
	case "running":
		return debrid.TorrentItemStatusDownloading
	case "finished":
		return debrid.TorrentItemStatusCompleted
	case "seeding":
		return debrid.TorrentItemStatusSeeding
	case "error":
		return debrid.TorrentItemStatusError
	default:
		return debrid.TorrentItemStatusOther
	}
}
