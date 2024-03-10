package entities

import (
	"github.com/samber/lo"
	"strings"
)

type (
	MediaEntryLibraryData struct {
		AllFilesLocked bool   `json:"allFilesLocked"`
		SharedPath     string `json:"sharedPath"`
	}

	NewMediaEntryLibraryDataOptions struct {
		entryLocalFiles []*LocalFile
		mediaId         int
	}
)

// NewMediaEntryLibraryData creates a new MediaEntryLibraryData based on the media id and a list of local files related to the media.
// It will return false if the list of local files is empty.
func NewMediaEntryLibraryData(opts *NewMediaEntryLibraryDataOptions) (*MediaEntryLibraryData, bool) {

	if opts.entryLocalFiles == nil || len(opts.entryLocalFiles) == 0 {
		return nil, false
	}
	sharedPath := strings.Replace(opts.entryLocalFiles[0].Path, opts.entryLocalFiles[0].Name, "", 1)
	sharedPath = strings.TrimSuffix(strings.TrimSuffix(sharedPath, "\\"), "/")

	return &MediaEntryLibraryData{
		AllFilesLocked: lo.EveryBy(opts.entryLocalFiles, func(item *LocalFile) bool { return item.Locked }),
		SharedPath:     sharedPath,
	}, true
}
