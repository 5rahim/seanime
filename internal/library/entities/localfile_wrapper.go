package entities

type (
	// LocalFileWrapper takes a slice of LocalFiles and provides helper methods.
	LocalFileWrapper struct {
		localFiles          []*LocalFile
		localEntries        []*LocalFileWrapperEntry
		unmatchedLocalFiles []*LocalFile
	}

	LocalFileWrapperEntry struct {
		mediaId    int
		localFiles []*LocalFile
	}
)

// NewLocalFileWrapper creates and returns a reference to a new LocalFileWrapper
func NewLocalFileWrapper(lfs []*LocalFile) *LocalFileWrapper {
	lfw := &LocalFileWrapper{
		localFiles:          lfs,
		localEntries:        make([]*LocalFileWrapperEntry, 0),
		unmatchedLocalFiles: make([]*LocalFile, 0),
	}

	// Group local files by media id
	groupedLfs := GroupLocalFilesByMediaID(lfs)
	for mId, gLfs := range groupedLfs {
		if mId == 0 {
			lfw.unmatchedLocalFiles = gLfs
			continue
		}
		lfw.localEntries = append(lfw.localEntries, &LocalFileWrapperEntry{
			mediaId:    mId,
			localFiles: gLfs,
		})
	}

	return lfw
}

func (lfw *LocalFileWrapper) GetLocalEntryById(mId int) (*LocalFileWrapperEntry, bool) {
	for _, me := range lfw.localEntries {
		if me.mediaId == mId {
			return me, true
		}
	}
	return nil, false
}

// GetMainLocalFiles returns the *main* local files.
func (e *LocalFileWrapperEntry) GetMainLocalFiles() ([]*LocalFile, bool) {
	lfs := make([]*LocalFile, 0)
	for _, lf := range e.localFiles {
		if lf.IsMain() {
			lfs = append(lfs, lf)
		}
	}
	if len(lfs) == 0 {
		return nil, false
	}
	return lfs, true
}

// GetUnwatchedLocalFiles returns the *main* local files that have not been watched.
// It returns an empty slice if all local files have been watched.
//
// /!\ IF Episode 0 is present, progress will be decremented by 1. This is because we assume AniList includes the episode 0 in the total count.
func (e *LocalFileWrapperEntry) GetUnwatchedLocalFiles(progress int) []*LocalFile {
	ret := make([]*LocalFile, 0)
	lfs, ok := e.GetMainLocalFiles()
	if !ok {
		return ret
	}

	for _, lf := range lfs {
		if lf.GetEpisodeNumber() == 0 {
			progress = progress - 1
			break
		}
	}

	for _, lf := range lfs {
		if lf.GetEpisodeNumber() > progress {
			ret = append(ret, lf)
		}
	}

	return ret
}

// HasMainLocalFiles returns true if there are any *main* local files.
func (e *LocalFileWrapperEntry) HasMainLocalFiles() bool {
	for _, lf := range e.localFiles {
		if lf.IsMain() {
			return true
		}
	}
	return false
}

// FindLocalFileWithEpisodeNumber returns the *main* local file with the given episode number.
func (e *LocalFileWrapperEntry) FindLocalFileWithEpisodeNumber(ep int) (*LocalFile, bool) {
	for _, lf := range e.localFiles {
		if !lf.IsMain() {
			continue
		}
		if lf.GetEpisodeNumber() == ep {
			return lf, true
		}
	}
	return nil, false
}

// FindLatestLocalFile returns the *main* local file with the highest episode number.
func (e *LocalFileWrapperEntry) FindLatestLocalFile() (*LocalFile, bool) {
	lfs, ok := e.GetMainLocalFiles()
	if !ok {
		return nil, false
	}
	// Get the local file with the highest episode number
	latest := lfs[0]
	for _, lf := range lfs {
		if lf.GetEpisodeNumber() > latest.GetEpisodeNumber() {
			latest = lf
		}
	}
	return latest, true
}

// FindNextEpisode returns the *main* local file whose episode number is after the given local file.
func (e *LocalFileWrapperEntry) FindNextEpisode(lf *LocalFile) (*LocalFile, bool) {
	lfs, ok := e.GetMainLocalFiles()
	if !ok {
		return nil, false
	}
	// Get the local file whose episode number is after the given local file
	var next *LocalFile
	for _, l := range lfs {
		if l.GetEpisodeNumber() == lf.GetEpisodeNumber()+1 {
			next = l
			break
		}
	}
	return next, true
}

func (lfw *LocalFileWrapper) GetUnmatchedLocalFiles() []*LocalFile {
	return lfw.unmatchedLocalFiles
}

func (lfw *LocalFileWrapper) GetLocalEntries() []*LocalFileWrapperEntry {
	return lfw.localEntries
}

func (lfw *LocalFileWrapper) GetLocalFiles() []*LocalFile {
	return lfw.localFiles
}

func (e *LocalFileWrapperEntry) GetLocalFiles() []*LocalFile {
	return e.localFiles
}

func (e *LocalFileWrapperEntry) GetMediaId() int {
	return e.mediaId
}
