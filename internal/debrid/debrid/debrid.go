package debrid

import (
	"context"
	"fmt"
	"path/filepath"
	"seanime/internal/util"
)

var (
	ErrNotAuthenticated     = fmt.Errorf("not authenticated")
	ErrFailedToAuthenticate = fmt.Errorf("failed to authenticate")
	ErrStreamInterrupted    = fmt.Errorf("stream interrupted")
)

type (
	Provider interface {
		GetSettings() Settings
		Authenticate(apiKey string) error
		AddTorrent(opts AddTorrentOptions) (string, error)
		// GetTorrentStreamUrl returns the stream URL for the torrent file. It should block until the stream URL is available.
		GetTorrentStreamUrl(ctx context.Context, opts StreamTorrentOptions, itemCh chan TorrentItem) (streamUrl string, err error)
		// GetTorrentDownloadUrl returns the download URL for the torrent. It should return an error if the torrent is not ready.
		GetTorrentDownloadUrl(opts DownloadTorrentOptions) (downloadUrl string, err error)
		// GetInstantAvailability returns a map where the key is the torrent's info hash
		GetInstantAvailability(hashes []string) map[string]TorrentItemInstantAvailability
		GetTorrent(id string) (*TorrentItem, error)
		GetTorrentInfo(opts GetTorrentInfoOptions) (*TorrentInfo, error)
		GetTorrents() ([]*TorrentItem, error)
		DeleteTorrent(id string) error
	}

	AddTorrentOptions struct {
		MagnetLink   string `json:"magnetLink"`
		InfoHash     string `json:"infoHash"`
		SelectFileId string `json:"selectFileId"` // Real-Debrid only, ID, IDs, or "all"
	}

	StreamTorrentOptions struct {
		ID     string `json:"id"`
		FileId string `json:"fileId"` // ID or index of the file to stream
	}

	GetTorrentInfoOptions struct {
		MagnetLink string `json:"magnetLink"`
		InfoHash   string `json:"infoHash"`
	}

	DownloadTorrentOptions struct {
		ID     string `json:"id"`
		FileId string `json:"fileId"` // ID or index of the file to download
	}

	// TorrentItem represents a torrent added to a Debrid service
	TorrentItem struct {
		ID                   string             `json:"id"`
		Name                 string             `json:"name"`                 // Name of the torrent or file
		Hash                 string             `json:"hash"`                 // SHA1 hash of the torrent
		Size                 int64              `json:"size"`                 // Size of the selected files (size in bytes)
		FormattedSize        string             `json:"formattedSize"`        // Formatted size of the selected files
		CompletionPercentage int                `json:"completionPercentage"` // Progress percentage (0 to 100)
		ETA                  string             `json:"eta"`                  // Formatted estimated time remaining
		Status               TorrentItemStatus  `json:"status"`               // Current download status
		AddedAt              string             `json:"added"`                // Date when the torrent was added, RFC3339 format
		Speed                string             `json:"speed,omitempty"`      // Current download speed (optional, present in downloading state)
		Seeders              int                `json:"seeders,omitempty"`    // Number of seeders (optional, present in downloading state)
		IsReady              bool               `json:"isReady"`              // Whether the torrent is ready to be downloaded
		Files                []*TorrentItemFile `json:"files,omitempty"`      // List of files in the torrent (optional)
	}

	TorrentItemFile struct {
		ID    string `json:"id"` // ID of the file, usually the index
		Index int    `json:"index"`
		Name  string `json:"name"`
		Path  string `json:"path"`
		Size  int64  `json:"size"`
	}

	TorrentItemStatus string

	TorrentItemInstantAvailability struct {
		CachedFiles map[string]*CachedFile `json:"cachedFiles"` // Key is the file ID (or index)
	}

	//------------------------------------------------------------------

	TorrentInfo struct {
		ID    *string            `json:"id"` // ID of the torrent if added to the debrid service
		Name  string             `json:"name"`
		Hash  string             `json:"hash"`
		Size  int64              `json:"size"`
		Files []*TorrentItemFile `json:"files"`
	}

	CachedFile struct {
		Size int64  `json:"size"`
		Name string `json:"name"`
	}
	////////////////////////////////////////////////////////////////////

	Settings struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
)

const (
	TorrentItemStatusDownloading TorrentItemStatus = "downloading"
	TorrentItemStatusCompleted   TorrentItemStatus = "completed"
	TorrentItemStatusSeeding     TorrentItemStatus = "seeding"
	TorrentItemStatusError       TorrentItemStatus = "error"
	TorrentItemStatusStalled     TorrentItemStatus = "stalled"
	TorrentItemStatusPaused      TorrentItemStatus = "paused"
	TorrentItemStatusOther       TorrentItemStatus = "other"
)

func FilterVideoFiles(files []*TorrentItemFile) []*TorrentItemFile {
	var filtered []*TorrentItemFile
	for _, file := range files {
		ext := filepath.Ext(file.Name)
		if util.IsValidVideoExtension(ext) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}
