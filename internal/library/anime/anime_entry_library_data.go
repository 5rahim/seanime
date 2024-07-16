package anime

import (
	"github.com/samber/lo"
	"strings"
)

type (
	AnimeEntryLibraryData struct {
		AllFilesLocked bool   `json:"allFilesLocked"`
		SharedPath     string `json:"sharedPath"`
	}

	NewAnimeEntryLibraryDataOptions struct {
		EntryLocalFiles []*LocalFile
		MediaId         int
	}
)

// NewAnimeEntryLibraryData creates a new AnimeEntryLibraryData based on the media id and a list of local files related to the media.
// It will return false if the list of local files is empty.
func NewAnimeEntryLibraryData(opts *NewAnimeEntryLibraryDataOptions) (*AnimeEntryLibraryData, bool) {

	if opts.EntryLocalFiles == nil || len(opts.EntryLocalFiles) == 0 {
		return nil, false
	}
	sharedPath := strings.Replace(opts.EntryLocalFiles[0].Path, opts.EntryLocalFiles[0].Name, "", 1)
	sharedPath = strings.TrimSuffix(strings.TrimSuffix(sharedPath, "\\"), "/")

	return &AnimeEntryLibraryData{
		AllFilesLocked: lo.EveryBy(opts.EntryLocalFiles, func(item *LocalFile) bool { return item.Locked }),
		SharedPath:     sharedPath,
	}, true
}
