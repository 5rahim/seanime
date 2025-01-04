package models

func (s *Settings) GetSensitiveValues() []string {
	return []string{
		s.MediaPlayer.VlcPassword,
		s.Torrent.QBittorrentPassword,
		s.Torrent.TransmissionPassword,
	}
}

func (s *DebridSettings) GetSensitiveValues() []string {
	return []string{
		s.ApiKey,
	}
}
