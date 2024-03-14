package entities

func (e *MediaEntryEpisode) GetEpisodeNumber() int {
	if e == nil {
		return -1
	}
	return e.EpisodeNumber
}
func (e *MediaEntryEpisode) GetProgressNumber() int {
	if e == nil {
		return -1
	}
	return e.ProgressNumber
}

func (e *MediaEntryEpisode) IsMain() bool {
	if e == nil || e.LocalFile == nil {
		return false
	}
	return e.LocalFile.IsMain()
}

func (e *MediaEntryEpisode) GetLocalFile() *LocalFile {
	if e == nil {
		return nil
	}
	return e.LocalFile
}
