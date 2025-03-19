package anime

import (
	"seanime/internal/hook"
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

	reqEvent := new(AnimeEntryLibraryDataRequestedEvent)
	reqEvent.EntryLocalFiles = opts.EntryLocalFiles
	reqEvent.MediaId = opts.MediaId
	reqEvent.CurrentProgress = opts.CurrentProgress

	err := hook.GlobalHookManager.OnAnimeEntryLibraryDataRequested().Trigger(reqEvent)
	if err != nil {
		return nil, false
	}

	if reqEvent.EntryLocalFiles == nil || len(reqEvent.EntryLocalFiles) == 0 {
		return nil, false
	}
	sharedPath := strings.Replace(reqEvent.EntryLocalFiles[0].Path, reqEvent.EntryLocalFiles[0].Name, "", 1)
	sharedPath = strings.TrimSuffix(strings.TrimSuffix(sharedPath, "\\"), "/")

	ret = &EntryLibraryData{
		AllFilesLocked: lo.EveryBy(reqEvent.EntryLocalFiles, func(item *LocalFile) bool { return item.Locked }),
		SharedPath:     sharedPath,
	}
	ok = true

	lfw := NewLocalFileWrapper(reqEvent.EntryLocalFiles)
	lfwe, ok := lfw.GetLocalEntryById(reqEvent.MediaId)
	if !ok {
		return ret, true
	}

	ret.UnwatchedCount = len(lfwe.GetUnwatchedLocalFiles(reqEvent.CurrentProgress))

	mainLfs, ok := lfwe.GetMainLocalFiles()
	if !ok {
		return ret, true
	}
	ret.MainFileCount = len(mainLfs)

	event := new(AnimeEntryLibraryDataEvent)
	event.EntryLibraryData = ret
	err = hook.GlobalHookManager.OnAnimeEntryLibraryData().Trigger(event)
	if err != nil {
		return nil, false
	}
	return event.EntryLibraryData, true
}
