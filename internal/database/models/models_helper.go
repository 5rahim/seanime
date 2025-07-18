package models

func (s *Settings) GetMediaPlayer() *MediaPlayerSettings {
	if s == nil || s.MediaPlayer == nil {
		return &MediaPlayerSettings{}
	}
	return s.MediaPlayer
}

func (s *Settings) GetTorrent() *TorrentSettings {
	if s == nil || s.Torrent == nil {
		return &TorrentSettings{}
	}
	return s.Torrent
}

func (s *Settings) GetAnilist() *AnilistSettings {
	if s == nil || s.Anilist == nil {
		return &AnilistSettings{}
	}
	return s.Anilist
}

func (s *Settings) GetManga() *MangaSettings {
	if s == nil || s.Manga == nil {
		return &MangaSettings{}
	}
	return s.Manga
}

func (s *Settings) GetLibrary() *LibrarySettings {
	if s == nil || s.Library == nil {
		return &LibrarySettings{}
	}
	return s.Library
}

func (s *Settings) GetListSync() *ListSyncSettings {
	if s == nil || s.ListSync == nil {
		return &ListSyncSettings{}
	}
	return s.ListSync
}

func (s *Settings) GetAutoDownloader() *AutoDownloaderSettings {
	if s == nil || s.AutoDownloader == nil {
		return &AutoDownloaderSettings{}
	}
	return s.AutoDownloader
}

func (s *Settings) GetDiscord() *DiscordSettings {
	if s == nil || s.Discord == nil {
		return &DiscordSettings{}
	}
	return s.Discord
}

func (s *Settings) GetNotifications() *NotificationSettings {
	if s == nil || s.Notifications == nil {
		return &NotificationSettings{}
	}
	return s.Notifications
}

func (s *Settings) GetNakama() *NakamaSettings {
	if s == nil || s.Nakama == nil {
		return &NakamaSettings{}
	}
	return s.Nakama
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *Settings) GetSensitiveValues() []string {
	if s == nil {
		return []string{}
	}
	return []string{
		s.GetMediaPlayer().VlcPassword,
		s.GetTorrent().QBittorrentPassword,
		s.GetTorrent().TransmissionPassword,
	}
}

func (s *DebridSettings) GetSensitiveValues() []string {
	if s == nil {
		return []string{}
	}
	return []string{
		s.ApiKey,
	}
}
