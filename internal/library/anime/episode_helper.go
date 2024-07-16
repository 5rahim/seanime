package anime

func (e *AnimeEntryEpisode) GetEpisodeNumber() int {
	if e == nil {
		return -1
	}
	return e.EpisodeNumber
}
func (e *AnimeEntryEpisode) GetProgressNumber() int {
	if e == nil {
		return -1
	}
	return e.ProgressNumber
}

func (e *AnimeEntryEpisode) IsMain() bool {
	if e == nil || e.LocalFile == nil {
		return false
	}
	return e.LocalFile.IsMain()
}

func (e *AnimeEntryEpisode) GetLocalFile() *LocalFile {
	if e == nil {
		return nil
	}
	return e.LocalFile
}
