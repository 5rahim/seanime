package core

import (
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type (
	FeatureKey     string
	FeatureManager struct {
		disabledFeatures map[FeatureKey]bool
		DisabledFeatures []FeatureKey
		logger           *zerolog.Logger
	}
)

const (
	// ManageOfflineMode allows switching to offline mode.
	ManageOfflineMode FeatureKey = "ManageOfflineMode"
	// ViewSettings allows viewing the settings page.
	ViewSettings FeatureKey = "ViewSettings"
	// ViewLogs allows viewing the logs.
	ViewLogs FeatureKey = "ViewLogs"
	// UpdateSettings allows updating the settings.
	UpdateSettings FeatureKey = "UpdateSettings"
	ManagePlaylist FeatureKey = "ManagePlaylist"
	// ManageLocalAnimeLibrary encompasses all updates to the local anime library.
	//	- Refreshing library, update local files, opening directory etc.
	ManageLocalAnimeLibrary FeatureKey = "ManageLocalAnimeLibrary"
	// ManageAccount allows logging in and out of the user account.
	ManageAccount FeatureKey = "ManageAccount"
	// ViewAccount allows viewing the user account info.
	ViewAccount FeatureKey = "ViewAccount"
	// ManageLists allows managing AniList/Custom source/Local lists.
	//	- Adding/updating/removing entries
	ManageLists FeatureKey = "ManageLists"
	// RefreshMetadata allows refreshing anime/manga metadata from Anilist and custom sources.
	RefreshMetadata      FeatureKey = "RefreshMetadata"
	ManageMangaDownloads FeatureKey = "ManageMangaDownloads"
	WatchingLocalAnime   FeatureKey = "WatchingLocalAnime"
	TorrentStreaming     FeatureKey = "TorrentStreaming"
	DebridStreaming      FeatureKey = "DebridStreaming"
	OnlineStreaming      FeatureKey = "OnlineStreaming"
	Transcode            FeatureKey = "Transcode"
	Reading              FeatureKey = "Reading"
	// ViewAutoDownloader allows viewing the auto downloader page.
	ViewAutoDownloader FeatureKey = "ViewAutoDownloader"
	// ManageAutoDownloader allows performing actions in the auto downloader.
	ManageAutoDownloader FeatureKey = "ManageAutoDownloader"
	// ViewScanSummaries allows viewing the scan summaries.
	ViewScanSummaries FeatureKey = "ViewScanSummaries"
	ViewExtensions    FeatureKey = "ViewExtensions"
	ManageExtensions  FeatureKey = "ManageExtensions"
	ManageHomeScreen  FeatureKey = "ManageHomeScreen"
	OpenInExplorer    FeatureKey = "OpenInExplorer"
	PluginTray        FeatureKey = "PluginTray"
	ManageNakama      FeatureKey = "ManageNakama"
	ManageDebrid      FeatureKey = "ManageDebrid"
	Proxy             FeatureKey = "Proxy"
	ManageMangaSource FeatureKey = "ManageMangaSource"
	PushRequests      FeatureKey = "PushRequests"
)

func NewFeatureManager(logger *zerolog.Logger, flags SeanimeFlags) *FeatureManager {
	ret := &FeatureManager{
		disabledFeatures: make(map[FeatureKey]bool),
		DisabledFeatures: flags.DisableFeatures,
		logger:           logger,
	}

	if flags.LockDown {
		ret.DisabledFeatures = []FeatureKey{
			ManageOfflineMode,
			ViewSettings,
			ViewLogs,
			UpdateSettings,
			ManagePlaylist,
			ManageLocalAnimeLibrary,
			ManageAccount,
			ViewAccount,
			ManageLists,
			RefreshMetadata,
			ManageMangaDownloads,
			WatchingLocalAnime,
			TorrentStreaming,
			DebridStreaming,
			OnlineStreaming,
			Reading,
			ViewAutoDownloader,
			ManageAutoDownloader,
			ViewScanSummaries,
			ViewExtensions,
			ManageExtensions,
			ManageHomeScreen,
			OpenInExplorer,
			PluginTray,
			ManageNakama,
			ManageDebrid,
			Proxy,
			Transcode,
			ManageMangaSource,
			PushRequests,
		}
	}

	for _, key := range ret.DisabledFeatures {
		ret.disabledFeatures[key] = true
	}

	if len(ret.DisabledFeatures) > 0 {
		logger.Warn().Msgf("app: Disabled features: %s", ret.DisabledFeatures)
	}

	return ret
}

func (m *FeatureManager) IsEnabled(key FeatureKey) bool {
	_, ok := m.disabledFeatures[key]
	return !ok
}

func (m *FeatureManager) IsDisabled(key FeatureKey) bool {
	_, ok := m.disabledFeatures[key]
	return ok
}

func (m *FeatureManager) HasDisabledFeatures() bool {
	return len(m.DisabledFeatures) > 0
}

func (m *FeatureManager) GetDisabledFeatureMap() map[FeatureKey]bool {
	return m.disabledFeatures
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	FeatureFlags struct {
		MainServerTorrentStreaming bool
	}

	ExperimentalFeatureFlags struct {
	}
)

// NewFeatureFlags initializes the feature flags
func NewFeatureFlags(cfg *Config, logger *zerolog.Logger) FeatureFlags {
	ff := FeatureFlags{
		MainServerTorrentStreaming: viper.GetBool("experimental.mainServerTorrentStreaming"),
	}

	checkExperimentalFeatureFlags(&ff, cfg, logger)

	return ff
}

func checkExperimentalFeatureFlags(ff *FeatureFlags, cfg *Config, logger *zerolog.Logger) {
	if ff.MainServerTorrentStreaming {
		logger.Warn().Msg("app: [Feature flag] 'Main Server Torrent Streaming' experimental feature is enabled")
	}
}

func (ff *FeatureFlags) IsMainServerTorrentStreamingEnabled() bool {
	return ff.MainServerTorrentStreaming
}
