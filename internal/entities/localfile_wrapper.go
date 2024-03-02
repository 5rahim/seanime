package entities

type (
	// LocalFileWrapper takes a slice of LocalFiles and provides helper methods.
	LocalFileWrapper struct {
		LocalFiles          []*LocalFile             `json:"localFiles"`
		LocalEntries        []*LocalFileWrapperEntry `json:"mediaEntries"`
		UnmatchedLocalFiles []*LocalFile             `json:"unmatchedLocalFiles"`
	}

	LocalFileWrapperEntry struct {
		MediaId    int          `json:"mediaId"`
		LocalFiles []*LocalFile `json:"localFiles"`
	}
)

// NewLocalFileWrapper creates and returns a reference to a new LocalFileWrapper
func NewLocalFileWrapper(lfs []*LocalFile) *LocalFileWrapper {
	lfw := &LocalFileWrapper{
		LocalFiles:          lfs,
		LocalEntries:        make([]*LocalFileWrapperEntry, 0),
		UnmatchedLocalFiles: make([]*LocalFile, 0),
	}

	// Group local files by media id
	groupedLfs := GroupLocalFilesByMediaID(lfs)
	for mId, gLfs := range groupedLfs {
		if mId == 0 {
			lfw.UnmatchedLocalFiles = gLfs
			continue
		}
		lfw.LocalEntries = append(lfw.LocalEntries, &LocalFileWrapperEntry{
			MediaId:    mId,
			LocalFiles: gLfs,
		})
	}

	return lfw
}

func (lfw *LocalFileWrapper) GetLocalEntryById(mId int) (*LocalFileWrapperEntry, bool) {
	for _, me := range lfw.LocalEntries {
		if me.MediaId == mId {
			return me, true
		}
	}
	return nil, false
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetMainLocalFiles returns the *main* local files.
func (e *LocalFileWrapperEntry) GetMainLocalFiles() ([]*LocalFile, bool) {
	lfs := make([]*LocalFile, 0)
	for _, lf := range e.LocalFiles {
		if lf.IsMain() {
			lfs = append(lfs, lf)
		}
	}
	if len(lfs) == 0 {
		return nil, false
	}
	return lfs, true
}

// HasMainLocalFiles returns true if there are any *main* local files.
func (e *LocalFileWrapperEntry) HasMainLocalFiles() bool {
	for _, lf := range e.LocalFiles {
		if lf.IsMain() {
			return true
		}
	}
	return false
}

// FindLocalFileWithEpisodeNumber returns the *main* local file with the given episode number.
func (e *LocalFileWrapperEntry) FindLocalFileWithEpisodeNumber(ep int) (*LocalFile, bool) {
	for _, lf := range e.LocalFiles {
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
