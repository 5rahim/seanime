package entities

import (
	"github.com/samber/lo"
)

type (
	MediaEntryLibraryData struct {
		AllFilesLocked bool `json:"allFilesLocked"`
	}

	NewMediaEntryLibraryDataOptions struct {
		groupedLocalFiles *map[int][]*LocalFile
		mediaId           int
	}
)

func NewMediaEntryLibraryData(opts *NewMediaEntryLibraryDataOptions) (*MediaEntryLibraryData, bool) {
	entryLfs, found := (*opts.groupedLocalFiles)[opts.mediaId]
	if !found {
		return nil, false
	}
	return &MediaEntryLibraryData{
		AllFilesLocked: lo.EveryBy(entryLfs, func(item *LocalFile) bool { return item.Locked }),
	}, true
}
