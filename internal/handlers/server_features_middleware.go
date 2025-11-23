package handlers

import (
	"errors"
	"seanime/internal/core"
	"slices"
	"strings"

	"github.com/labstack/echo/v4"
)

func (h *Handler) FeaturesMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !h.App.FeatureManager.HasDisabledFeatures() {
			return next(c)
		}

		var ErrFeatureDisabled = errors.New("feature disabled")

		type pathFeatureConfig struct {
			PathStartsWith string
			ShouldReject   bool
			Methods        []string
			ExcludePaths   []string
		}

		var UpdateMethods = []string{"POST", "PUT", "DELETE", "PATCH"}
		var Empty []string

		path := c.Request().URL.Path
		method := strings.ToUpper(c.Request().Method)

		var pathFeatureConfigs = []pathFeatureConfig{
			// offline mode
			{"/api/v1/local", h.App.FeatureManager.IsDisabled(core.ManageOfflineMode), UpdateMethods, Empty},
			// settings
			{"/api/v1/start", h.App.FeatureManager.IsDisabled(core.UpdateSettings), UpdateMethods, Empty},
			{"/api/v1/settings", h.App.FeatureManager.IsDisabled(core.UpdateSettings), UpdateMethods, Empty},
			{"/api/v1/torrentstream/settings", h.App.FeatureManager.IsDisabled(core.UpdateSettings), UpdateMethods, Empty},
			{"/api/v1/debrid/settings", h.App.FeatureManager.IsDisabled(core.UpdateSettings), UpdateMethods, Empty},
			{"/api/v1/mediastream/settings", h.App.FeatureManager.IsDisabled(core.UpdateSettings), UpdateMethods, Empty},
			{"/api/v1/report", h.App.FeatureManager.IsDisabled(core.UpdateSettings), UpdateMethods, Empty},
			{"/api/v1/theme", h.App.FeatureManager.IsDisabled(core.UpdateSettings), UpdateMethods, Empty},
			{"/api/v1/memory", h.App.FeatureManager.IsDisabled(core.UpdateSettings), Empty, Empty},
			{"/api/v1/filecache", h.App.FeatureManager.IsDisabled(core.UpdateSettings), Empty, Empty},
			// account
			{"/api/v1/auth", h.App.FeatureManager.IsDisabled(core.ManageAccount), UpdateMethods, Empty},
			{"/api/v1/mal/auth", h.App.FeatureManager.IsDisabled(core.ManageAccount), UpdateMethods, Empty},
			{"/api/v1/mal/logout", h.App.FeatureManager.IsDisabled(core.ManageAccount), UpdateMethods, Empty},
			// lists
			{"/api/v1/anilist/list-entry", h.App.FeatureManager.IsDisabled(core.ManageLists), UpdateMethods, Empty},
			{"/api/v1/library/anime-entry/update-progress", h.App.FeatureManager.IsDisabled(core.ManageLists), UpdateMethods, Empty},
			{"/api/v1/library/anime-entry/update-repeat", h.App.FeatureManager.IsDisabled(core.ManageLists), UpdateMethods, Empty},
			{"/api/v1/manga/update-progress", h.App.FeatureManager.IsDisabled(core.ManageLists), UpdateMethods, Empty},
			// refresh metadata
			{"/api/v1/anilist/cache-layer/status", h.App.FeatureManager.IsDisabled(core.RefreshMetadata), UpdateMethods, Empty},
			{"/api/v1/library/scan", h.App.FeatureManager.IsDisabled(core.RefreshMetadata), UpdateMethods, Empty},
			{"/api/v1/manga/refetch-chapter-containers", h.App.FeatureManager.IsDisabled(core.RefreshMetadata), UpdateMethods, Empty},
			// playlists
			{"/api/v1/playlist", h.App.FeatureManager.IsDisabled(core.ManagePlaylist), UpdateMethods, Empty},
			{"/api/v1/playback-manager/start-playlist", h.App.FeatureManager.IsDisabled(core.ManagePlaylist), UpdateMethods, Empty},
			{"/api/v1/playback-manager/playlist-next", h.App.FeatureManager.IsDisabled(core.ManagePlaylist), UpdateMethods, Empty},
			{"/api/v1/playback-manager/cancel-playlist", h.App.FeatureManager.IsDisabled(core.ManagePlaylist), UpdateMethods, Empty},
			// playback
			{"/api/v1/playback-manager", h.App.FeatureManager.IsDisabled(core.WatchingLocalAnime), UpdateMethods, []string{"/api/v1/playback-manager/start-playlist", "/api/v1/playback-manager/playlist-next", "/api/v1/playback-manager/cancel-playlist"}},
			{"/api/v1/media-player/start", h.App.FeatureManager.IsDisabled(core.WatchingLocalAnime), UpdateMethods, Empty},
			// torrent client / auto downloader
			{"/api/v1/torrent/search", h.App.FeatureManager.IsDisabled(core.ManageAutoDownloader), UpdateMethods, Empty},
			{"/api/v1/torrent-client", h.App.FeatureManager.IsDisabled(core.ManageAutoDownloader), UpdateMethods, Empty},
			{"/api/v1/download-torrent-file", h.App.FeatureManager.IsDisabled(core.ManageAutoDownloader), UpdateMethods, Empty},
			{"/api/v1/auto-downloader", h.App.FeatureManager.IsDisabled(core.ManageAutoDownloader), UpdateMethods, Empty},
			// onlinestream
			{"/api/v1/onlinestream", h.App.FeatureManager.IsDisabled(core.OnlineStreaming), UpdateMethods, []string{"/api/v1/onlinestream/search", "/api/v1/onlinestream/manual-mapping", "/api/v1/onlinestream/get-mapping", "/api/v1/onlinestream/remove-mapping"}},
			{"/api/v1/onlinestream/search", h.App.FeatureManager.IsDisabled(core.ManageMangaSource), UpdateMethods, Empty},
			{"/api/v1/onlinestream/manual-mapping", h.App.FeatureManager.IsDisabled(core.ManageMangaSource), UpdateMethods, Empty},
			{"/api/v1/onlinestream/get-mapping", h.App.FeatureManager.IsDisabled(core.ManageMangaSource), UpdateMethods, Empty},
			{"/api/v1/onlinestream/remove-mapping", h.App.FeatureManager.IsDisabled(core.ManageMangaSource), UpdateMethods, Empty},
			// custom source
			//{"/api/v1/custom-source", h.App.FeatureManager.IsDisabled(core.ManageMangaSource), UpdateMethods, Empty},
			// nakama
			{"/api/v1/nakama", h.App.FeatureManager.IsDisabled(core.ManageNakama), UpdateMethods, Empty},
			// open in explorer
			{"/api/v1/open-in-explorer", h.App.FeatureManager.IsDisabled(core.OpenInExplorer), Empty, Empty},
			{"/api/v1/library/anime-entry/open-in-explorer", h.App.FeatureManager.IsDisabled(core.OpenInExplorer), UpdateMethods, Empty},
			// debrid
			{"/api/v1/debrid", h.App.FeatureManager.IsDisabled(core.ManageDebrid), UpdateMethods, []string{"/api/v1/debrid/settings", "/api/v1/debrid/torrents/info", "/api/v1/debrid/torrents/file-previews"}},
			{"/api/v1/debrid/stream", h.App.FeatureManager.IsDisabled(core.DebridStreaming), UpdateMethods, Empty},
			// home items
			{"/api/v1/status/home-items", h.App.FeatureManager.IsDisabled(core.ManageHomeScreen), UpdateMethods, Empty},
			// extensions
			{"/api/v1/extensions", h.App.FeatureManager.IsDisabled(core.ManageExtensions), UpdateMethods, []string{"/api/v1/extensions/all"}},
			{"/api/v1/extensions/updates", h.App.FeatureManager.IsDisabled(core.ManageExtensions), Empty, Empty},
			// proxy
			{"/api/v1/proxy", h.App.FeatureManager.IsDisabled(core.Proxy), Empty, Empty},
			{"/api/v1/image-proxy", h.App.FeatureManager.IsDisabled(core.Proxy), Empty, Empty},
			// logs
			{"/api/v1/log", h.App.FeatureManager.IsDisabled(core.ViewLogs), Empty, Empty},
			{"/api/v1/logs", h.App.FeatureManager.IsDisabled(core.ViewLogs), Empty, Empty},
			{"/api/v1/logs", h.App.FeatureManager.IsDisabled(core.UpdateSettings), []string{"DELETE"}, Empty},
			// torrent stream
			{"/api/v1/torrentstream", h.App.FeatureManager.IsDisabled(core.TorrentStreaming), UpdateMethods, []string{"/api/v1/torrentstream/settings"}},
			// transcode
			{"/api/v1/mediastream", h.App.FeatureManager.IsDisabled(core.Transcode), UpdateMethods, []string{"/api/v1/mediastream/settings"}},
			{"/api/v1/directstream", h.App.FeatureManager.IsDisabled(core.WatchingLocalAnime), UpdateMethods, Empty},
			{"/api/v1/mediastream/file", h.App.FeatureManager.IsDisabled(core.WatchingLocalAnime), Empty, Empty},
			{"/api/v1/mediastream", h.App.FeatureManager.IsDisabled(core.WatchingLocalAnime), Empty, Empty},
			// manga
			{"/api/v1/manga", h.App.FeatureManager.IsDisabled(core.ManageMangaSource), UpdateMethods, []string{"/api/v1/manga/pages", "/api/v1/manga/chapters"}},
			{"/api/v1/manga", h.App.FeatureManager.IsDisabled(core.Reading), UpdateMethods, Empty},
			// manga downloads
			{"/api/v1/manga/download", h.App.FeatureManager.IsDisabled(core.ManageMangaDownloads), UpdateMethods, Empty},
			// local anime library
			{"/api/v1/metadata-provider", h.App.FeatureManager.IsDisabled(core.ManageLocalAnimeLibrary), UpdateMethods, Empty},
			{"/api/v1/library", h.App.FeatureManager.IsDisabled(core.ManageLocalAnimeLibrary), UpdateMethods, []string{"/api/v1/library/anime-entry/update-progress", "/api/v1/library/anime-entry/update-repeat"}},
			{"/api/v1/library/explorer", h.App.FeatureManager.IsDisabled(core.ManageLocalAnimeLibrary), UpdateMethods, Empty},
		}

		pathPrefixes := make([]string, 0, len(pathFeatureConfigs))
		for _, config := range pathFeatureConfigs {
			pathPrefixes = append(pathPrefixes, config.PathStartsWith)
			if config.ShouldReject &&
				strings.HasPrefix(path, config.PathStartsWith) &&
				!slices.ContainsFunc(config.ExcludePaths, func(i string) bool { return path == i }) {
				if len(config.Methods) == 0 || strings.Contains(strings.Join(config.Methods, ","), strings.ToUpper(method)) {
					return h.RespondWithError(c, ErrFeatureDisabled)
				}
			}
		}

		if h.App.FeatureManager.IsDisabled(core.PushRequests) {
			pathPrefixes = append(pathPrefixes, "/api/v1/anilist/list-anime", "/api/v1/anilist/list-manga", "/api/v1/anilist/list-recent-anime", "/api/v1/manga/anilist/list", "/api/v1/announcements")
			if !slices.ContainsFunc(pathPrefixes, func(i string) bool { return strings.HasPrefix(path, i) }) {
				if strings.Contains(strings.Join(UpdateMethods, ","), strings.ToUpper(method)) {
					//return h.RespondWithError(c, ErrFeatureDisabled)
					return h.RespondWithData(c, nil)
				}
			}
		}

		return next(c)
	}
}
