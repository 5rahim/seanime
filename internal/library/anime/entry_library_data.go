package anime

import (
	"strings"

	"github.com/samber/lo"
)

type (
	EntryLibraryData struct {
		AllFilesLocked bool   `json:"allFilesLocked"`
		SharedPath     string `json:"sharedPath"`
		UnwatchedCount int    `json:"unwatchedCount"`
		MainFileCount  int    `json:"mainFileCount"`
	}

	NewEntryLibraryDataOptions struct {
		EntryLocalFiles []*LocalFile
		MediaId         int
		CurrentProgress int
	}
)

// NewEntryLibraryData creates a new EntryLibraryData based on the media id and a list of local files related to the media.
// It will return false if the list of local files is empty.
func NewEntryLibraryData(opts *NewEntryLibraryDataOptions) (ret *EntryLibraryData, ok bool) {

	if opts.EntryLocalFiles == nil || len(opts.EntryLocalFiles) == 0 {
		return nil, false
	}
	sharedPath := strings.Replace(opts.EntryLocalFiles[0].Path, opts.EntryLocalFiles[0].Name, "", 1)
	sharedPath = strings.TrimSuffix(strings.TrimSuffix(sharedPath, "\\"), "/")

	ret = &EntryLibraryData{
		AllFilesLocked: lo.EveryBy(opts.EntryLocalFiles, func(item *LocalFile) bool { return item.Locked }),
		SharedPath:     sharedPath,
	}
	ok = true

	lfw := NewLocalFileWrapper(opts.EntryLocalFiles)
	lfwe, ok := lfw.GetLocalEntryById(opts.MediaId)
	if !ok {
		return ret, true
	}

	ret.UnwatchedCount = len(lfwe.GetUnwatchedLocalFiles(opts.CurrentProgress))

	mainLfs, ok := lfwe.GetMainLocalFiles()
	if !ok {
		return ret, true
	}
	ret.MainFileCount = len(mainLfs)

	return ret, true
}
