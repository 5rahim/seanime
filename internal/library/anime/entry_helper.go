package anime

import "github.com/samber/lo"

// HasWatchedAll returns true if all episodes have been watched.
// Returns false if there are no downloaded episodes.
func (e *Entry) HasWatchedAll() bool {
	// If there are no episodes, return nil
	latestEp, ok := e.FindLatestEpisode()
	if !ok {
		return false
	}

	return e.GetCurrentProgress() >= latestEp.GetProgressNumber()

}

// FindNextEpisode returns the episode whose episode number is the same as the progress number + 1.
// Returns false if there are no episodes or if there is no next episode.
func (e *Entry) FindNextEpisode() (*Episode, bool) {
	eps, ok := e.FindMainEpisodes()
	if !ok {
		return nil, false
	}
	ep, ok := lo.Find(eps, func(ep *Episode) bool {
		return ep.GetProgressNumber() == e.GetCurrentProgress()+1
	})
	if !ok {
		return nil, false
	}
	return ep, true
}

// FindLatestEpisode returns the *main* episode with the highest episode number.
// Returns false if there are no episodes.
func (e *Entry) FindLatestEpisode() (*Episode, bool) {
	// If there are no episodes, return nil
	eps, ok := e.FindMainEpisodes()
	if !ok {
		return nil, false
	}
	// Get the episode with the highest progress number
	latest := eps[0]
	for _, ep := range eps {
		if ep.GetProgressNumber() > latest.GetProgressNumber() {
			latest = ep
		}
	}
	return latest, true
}

// FindLatestLocalFile returns the *main* local file with the highest episode number.
// Returns false if there are no local files.
func (e *Entry) FindLatestLocalFile() (*LocalFile, bool) {
	lfs, ok := e.FindMainLocalFiles()
	// If there are no local files, return nil
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

//----------------------------------------------------------------------------------------------------------------------

// GetCurrentProgress returns the progress number.
// If the media entry is not in any AniList list, returns 0.
func (e *Entry) GetCurrentProgress() int {
	listData, ok := e.FindListData()
	if !ok {
		return 0
	}
	return listData.Progress
}

// FindEpisodes returns the episodes.
// Returns false if there are no episodes.
func (e *Entry) FindEpisodes() ([]*Episode, bool) {
	if e.Episodes == nil {
		return nil, false
	}
	return e.Episodes, true
}

// FindMainEpisodes returns the main episodes.
// Returns false if there are no main episodes.
func (e *Entry) FindMainEpisodes() ([]*Episode, bool) {
	if e.Episodes == nil {
		return nil, false
	}
	eps := make([]*Episode, 0)
	for _, ep := range e.Episodes {
		if ep.IsMain() {
			eps = append(eps, ep)
		}
	}
	return e.Episodes, true
}

// FindLocalFiles returns the local files.
// Returns false if there are no local files.
func (e *Entry) FindLocalFiles() ([]*LocalFile, bool) {
	if !e.IsDownloaded() {
		return nil, false
	}
	return e.LocalFiles, true
}

// FindMainLocalFiles returns *main* local files.
// Returns false if there are no local files.
func (e *Entry) FindMainLocalFiles() ([]*LocalFile, bool) {
	if !e.IsDownloaded() {
		return nil, false
	}
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

// IsDownloaded returns true if there are local files.
func (e *Entry) IsDownloaded() bool {
	if e.LocalFiles == nil {
		return false
	}
	return len(e.LocalFiles) > 0
}

func (e *Entry) FindListData() (*EntryListData, bool) {
	if e.EntryListData == nil {
		return nil, false
	}
	return e.EntryListData, true
}

func (e *Entry) IsInAnimeCollection() bool {
	_, ok := e.FindListData()
	return ok
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (e *SimpleEntry) GetCurrentProgress() int {
	listData, ok := e.FindListData()
	if !ok {
		return 0
	}
	return listData.Progress
}

func (e *SimpleEntry) FindMainEpisodes() ([]*Episode, bool) {
	if e.Episodes == nil {
		return nil, false
	}
	eps := make([]*Episode, 0)
	for _, ep := range e.Episodes {
		if ep.IsMain() {
			eps = append(eps, ep)
		}
	}
	return e.Episodes, true
}

func (e *SimpleEntry) FindNextEpisode() (*Episode, bool) {
	eps, ok := e.FindMainEpisodes()
	if !ok {
		return nil, false
	}
	ep, ok := lo.Find(eps, func(ep *Episode) bool {
		return ep.GetProgressNumber() == e.GetCurrentProgress()+1
	})
	if !ok {
		return nil, false
	}
	return ep, true
}

func (e *SimpleEntry) FindLatestEpisode() (*Episode, bool) {
	// If there are no episodes, return nil
	eps, ok := e.FindMainEpisodes()
	if !ok {
		return nil, false
	}
	// Get the episode with the highest progress number
	latest := eps[0]
	for _, ep := range eps {
		if ep.GetProgressNumber() > latest.GetProgressNumber() {
			latest = ep
		}
	}
	return latest, true
}

func (e *SimpleEntry) FindLatestLocalFile() (*LocalFile, bool) {
	lfs, ok := e.FindMainLocalFiles()
	// If there are no local files, return nil
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

func (e *SimpleEntry) FindMainLocalFiles() ([]*LocalFile, bool) {
	if e.LocalFiles == nil {
		return nil, false
	}
	if len(e.LocalFiles) == 0 {
		return nil, false
	}
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

func (e *SimpleEntry) FindListData() (*EntryListData, bool) {
	if e.EntryListData == nil {
		return nil, false
	}
	return e.EntryListData, true
}

func (e *SimpleEntry) IsInAnimeCollection() bool {
	_, ok := e.FindListData()
	return ok
}
