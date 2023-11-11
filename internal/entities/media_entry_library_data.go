package entities

import (
	"github.com/samber/lo"
)

type (
	MediaEntryLibraryData struct {
		AllFilesLocked bool `json:"allFilesLocked"`
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
	return &MediaEntryLibraryData{
		AllFilesLocked: lo.EveryBy(opts.entryLocalFiles, func(item *LocalFile) bool { return item.Locked }),
	}, true
}
