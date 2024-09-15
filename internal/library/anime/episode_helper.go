package anime

func (e *Episode) GetEpisodeNumber() int {
	if e == nil {
		return -1
	}
	return e.EpisodeNumber
}
func (e *Episode) GetProgressNumber() int {
	if e == nil {
		return -1
	}
	return e.ProgressNumber
}

func (e *Episode) IsMain() bool {
	if e == nil || e.LocalFile == nil {
		return false
	}
	return e.LocalFile.IsMain()
}

func (e *Episode) GetLocalFile() *LocalFile {
	if e == nil {
		return nil
	}
	return e.LocalFile
}
