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

func NewMediaEntryLibraryData(opts *NewMediaEntryLibraryDataOptions) (*MediaEntryLibraryData, bool) {
	if opts.entryLocalFiles == nil {
		return nil, false
	}
	return &MediaEntryLibraryData{
		AllFilesLocked: lo.EveryBy(opts.entryLocalFiles, func(item *LocalFile) bool { return item.Locked }),
	}, true
}
