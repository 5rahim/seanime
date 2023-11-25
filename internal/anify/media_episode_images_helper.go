package anify

func (m *MediaEpisodeImagesEntry) HasEpisode(num int) bool {
	for _, ep := range m.EpisodeImageData {
		if ep.EpisodeNumber == num {
			return true
		}
	}
	return false
}
