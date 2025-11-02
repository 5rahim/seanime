package anime

import (
	"seanime/internal/util"
)

type (
	// LegacyPlaylist holds the data from models.PlaylistEntry
	LegacyPlaylist struct {
		DbId       uint         `json:"dbId"`       // DbId is the database ID of the models.PlaylistEntry
		Name       string       `json:"name"`       // Name is the name of the playlist
		LocalFiles []*LocalFile `json:"localFiles"` // LocalFiles is a list of local files in the playlist, in order
	}
)

// NewLegacyPlaylist creates a new Playlist instance
func NewLegacyPlaylist(name string) *LegacyPlaylist {
	return &LegacyPlaylist{
		Name:       name,
		LocalFiles: make([]*LocalFile, 0),
	}
}

func (pd *LegacyPlaylist) SetLocalFiles(lfs []*LocalFile) {
	pd.LocalFiles = lfs
}

// AddLocalFile adds a local file to the playlist
func (pd *LegacyPlaylist) AddLocalFile(localFile *LocalFile) {
	pd.LocalFiles = append(pd.LocalFiles, localFile)
}

// RemoveLocalFile removes a local file from the playlist
func (pd *LegacyPlaylist) RemoveLocalFile(path string) {
	for i, lf := range pd.LocalFiles {
		if lf.GetNormalizedPath() == util.NormalizePath(path) {
			pd.LocalFiles = append(pd.LocalFiles[:i], pd.LocalFiles[i+1:]...)
			return
		}
	}
}

func (pd *LegacyPlaylist) LocalFileExists(path string, lfs []*LocalFile) bool {
	for _, lf := range lfs {
		if lf.GetNormalizedPath() == util.NormalizePath(path) {
			return true
		}
	}
	return false
}
