package entities

func (e *MediaEntryEpisode) GetEpisodeNumber() int {
	return e.EpisodeNumber
}
func (e *MediaEntryEpisode) GetProgressNumber() int {
	return e.ProgressNumber
}

func (e *MediaEntryEpisode) GetLocalFile() *LocalFile {
	return e.LocalFile
}
