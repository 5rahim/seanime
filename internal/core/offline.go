package core

import (
	"seanime/internal/api/metadata"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/platforms/offline_platform"

	"github.com/spf13/viper"
)

// SetOfflineMode changes the offline mode.
// It updates the config and active AniList platform.
func (a *App) SetOfflineMode(enabled bool) {
	// Update the config
	a.Config.Server.Offline = enabled
	viper.Set("server.offline", enabled)
	err := viper.WriteConfig()
	if err != nil {
		a.Logger.Err(err).Msg("app: Failed to write config after setting offline mode")
	}
	a.Logger.Info().Bool("enabled", enabled).Msg("app: Offline mode set")
	a.isOffline = &enabled

	// Update the platform and metadata provider
	if enabled {
		a.AnilistPlatform, _ = offline_platform.NewOfflinePlatform(a.LocalManager, a.AnilistClient, a.Logger)
		a.MetadataProvider = a.LocalManager.GetOfflineMetadataProvider()
	} else {
		// DEVNOTE: We don't handle local platform since the feature doesn't allow offline mode
		a.AnilistPlatform = anilist_platform.NewAnilistPlatform(a.AnilistClient, a.Logger)
		a.MetadataProvider = metadata.NewProvider(&metadata.NewProviderImplOptions{
			Logger:     a.Logger,
			FileCacher: a.FileCacher,
		})
		a.InitOrRefreshAnilistData()
	}

	a.InitOrRefreshModules()
}
