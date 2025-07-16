package handlers

import (
	"net/http"
	"path/filepath"
	"seanime/internal/core"
	util "seanime/internal/util/proxies"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
)

type Handler struct {
	App *core.App
}

func InitRoutes(app *core.App, e *echo.Echo) {
	// CORS middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Cookie", "Authorization",
			"X-Seanime-Token", "X-Seanime-Nakama-Token", "X-Seanime-Nakama-Username", "X-Seanime-Nakama-Server-Version", "X-Seanime-Nakama-Peer-Id"},
		AllowCredentials: true,
	}))

	lechoLogger := lecho.From(*app.Logger)

	urisToSkip := []string{
		"/internal/metrics",
		"/_next",
		"/icons",
		"/events",
		"/api/v1/image-proxy",
		"/api/v1/mediastream/transcode/",
		"/api/v1/torrent-client/list",
		"/api/v1/proxy",
	}

	// Logging middleware
	e.Use(lecho.Middleware(lecho.Config{
		Logger: lechoLogger,
		Skipper: func(c echo.Context) bool {
			path := c.Request().URL.RequestURI()
			if filepath.Ext(c.Request().URL.Path) == ".txt" ||
				filepath.Ext(c.Request().URL.Path) == ".png" ||
				filepath.Ext(c.Request().URL.Path) == ".ico" {
				return true
			}
			for _, uri := range urisToSkip {
				if uri == path || strings.HasPrefix(path, uri) {
					return true
				}
			}
			return false
		},
		Enricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
			// Add which file the request came from
			return logger.Str("file", c.Path())
		},
	}))

	// Recovery middleware
	e.Use(middleware.Recover())

	// Client ID middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if the client has a UUID cookie
			cookie, err := c.Cookie("Seanime-Client-Id")

			if err != nil || cookie.Value == "" {
				// Generate a new UUID for the client
				u := uuid.New().String()

				// Create a cookie with the UUID
				newCookie := new(http.Cookie)
				newCookie.Name = "Seanime-Client-Id"
				newCookie.Value = u
				newCookie.HttpOnly = false // Make the cookie accessible via JS
				newCookie.Expires = time.Now().Add(24 * time.Hour)
				newCookie.Path = "/"
				newCookie.Domain = ""
				newCookie.SameSite = http.SameSiteDefaultMode
				newCookie.Secure = false

				// Set the cookie
				c.SetCookie(newCookie)

				// Store the UUID in the context for use in the request
				c.Set("Seanime-Client-Id", u)
			} else {
				// Store the existing UUID in the context for use in the request
				c.Set("Seanime-Client-Id", cookie.Value)
			}

			return next(c)
		}
	})

	e.Use(headMethodMiddleware)

	h := &Handler{App: app}

	e.GET("/events", h.webSocketEventHandler)

	v1 := e.Group("/api").Group("/v1") // Commented out for now, will be used later

	//
	// Auth middleware
	//
	v1.Use(h.OptionalAuthMiddleware)

	imageProxy := &util.ImageProxy{}
	v1.GET("/image-proxy", imageProxy.ProxyImage)

	v1.GET("/proxy", util.VideoProxy)
	v1.HEAD("/proxy", util.VideoProxy)

	v1.GET("/status", h.HandleGetStatus)
	v1.GET("/log/*", h.HandleGetLogContent)
	v1.GET("/logs/filenames", h.HandleGetLogFilenames)
	v1.DELETE("/logs", h.HandleDeleteLogs)
	v1.GET("/logs/latest", h.HandleGetLatestLogContent)

	v1.POST("/announcements", h.HandleGetAnnouncements)

	// Auth
	v1.POST("/auth/login", h.HandleLogin)
	v1.POST("/auth/logout", h.HandleLogout)

	// Settings
	v1.GET("/settings", h.HandleGetSettings)
	v1.PATCH("/settings", h.HandleSaveSettings)
	v1.POST("/start", h.HandleGettingStarted)
	v1.PATCH("/settings/auto-downloader", h.HandleSaveAutoDownloaderSettings)

	// Auto Downloader
	v1.POST("/auto-downloader/run", h.HandleRunAutoDownloader)
	v1.GET("/auto-downloader/rule/:id", h.HandleGetAutoDownloaderRule)
	v1.GET("/auto-downloader/rule/anime/:id", h.HandleGetAutoDownloaderRulesByAnime)
	v1.GET("/auto-downloader/rules", h.HandleGetAutoDownloaderRules)
	v1.POST("/auto-downloader/rule", h.HandleCreateAutoDownloaderRule)
	v1.PATCH("/auto-downloader/rule", h.HandleUpdateAutoDownloaderRule)
	v1.DELETE("/auto-downloader/rule/:id", h.HandleDeleteAutoDownloaderRule)

	v1.GET("/auto-downloader/items", h.HandleGetAutoDownloaderItems)
	v1.DELETE("/auto-downloader/item", h.HandleDeleteAutoDownloaderItem)

	// Other
	v1.POST("/test-dump", h.HandleTestDump)

	v1.POST("/directory-selector", h.HandleDirectorySelector)

	v1.POST("/open-in-explorer", h.HandleOpenInExplorer)

	v1.POST("/media-player/start", h.HandleStartDefaultMediaPlayer)

	//
	// AniList
	//

	v1Anilist := v1.Group("/anilist")

	v1Anilist.GET("/collection", h.HandleGetAnimeCollection)
	v1Anilist.POST("/collection", h.HandleGetAnimeCollection)

	v1Anilist.GET("/collection/raw", h.HandleGetRawAnimeCollection)
	v1Anilist.POST("/collection/raw", h.HandleGetRawAnimeCollection)

	v1Anilist.GET("/media-details/:id", h.HandleGetAnilistAnimeDetails)

	v1Anilist.GET("/studio-details/:id", h.HandleGetAnilistStudioDetails)

	v1Anilist.POST("/list-entry", h.HandleEditAnilistListEntry)

	v1Anilist.DELETE("/list-entry", h.HandleDeleteAnilistListEntry)

	v1Anilist.POST("/list-anime", h.HandleAnilistListAnime)

	v1Anilist.POST("/list-recent-anime", h.HandleAnilistListRecentAiringAnime)

	v1Anilist.GET("/list-missed-sequels", h.HandleAnilistListMissedSequels)

	v1Anilist.GET("/stats", h.HandleGetAniListStats)

	//
	// MAL
	//

	v1.POST("/mal/auth", h.HandleMALAuth)

	v1.POST("/mal/logout", h.HandleMALLogout)

	//
	// Library
	//

	v1Library := v1.Group("/library")

	v1Library.POST("/scan", h.HandleScanLocalFiles)

	v1Library.DELETE("/empty-directories", h.HandleRemoveEmptyDirectories)

	v1Library.GET("/local-files", h.HandleGetLocalFiles)
	v1Library.POST("/local-files", h.HandleLocalFileBulkAction)
	v1Library.PATCH("/local-files", h.HandleUpdateLocalFiles)
	v1Library.DELETE("/local-files", h.HandleDeleteLocalFiles)
	v1Library.GET("/local-files/dump", h.HandleDumpLocalFilesToFile)
	v1Library.POST("/local-files/import", h.HandleImportLocalFiles)
	v1Library.PATCH("/local-file", h.HandleUpdateLocalFileData)

	v1Library.GET("/collection", h.HandleGetLibraryCollection)
	v1Library.GET("/schedule", h.HandleGetAnimeCollectionSchedule)

	v1Library.GET("/scan-summaries", h.HandleGetScanSummaries)

	v1Library.GET("/missing-episodes", h.HandleGetMissingEpisodes)

	v1Library.GET("/anime-entry/:id", h.HandleGetAnimeEntry)
	v1Library.POST("/anime-entry/suggestions", h.HandleFetchAnimeEntrySuggestions)
	v1Library.POST("/anime-entry/manual-match", h.HandleAnimeEntryManualMatch)
	v1Library.PATCH("/anime-entry/bulk-action", h.HandleAnimeEntryBulkAction)
	v1Library.POST("/anime-entry/open-in-explorer", h.HandleOpenAnimeEntryInExplorer)
	v1Library.POST("/anime-entry/update-progress", h.HandleUpdateAnimeEntryProgress)
	v1Library.POST("/anime-entry/update-repeat", h.HandleUpdateAnimeEntryRepeat)
	v1Library.GET("/anime-entry/silence/:id", h.HandleGetAnimeEntrySilenceStatus)
	v1Library.POST("/anime-entry/silence", h.HandleToggleAnimeEntrySilenceStatus)

	v1Library.POST("/unknown-media", h.HandleAddUnknownMedia)

	//
	// Anime
	//
	v1.GET("/anime/episode-collection/:id", h.HandleGetAnimeEpisodeCollection)

	//
	// Torrent / Torrent Client
	//

	v1.POST("/torrent/search", h.HandleSearchTorrent)
	v1.POST("/torrent-client/download", h.HandleTorrentClientDownload)
	v1.GET("/torrent-client/list", h.HandleGetActiveTorrentList)
	v1.POST("/torrent-client/action", h.HandleTorrentClientAction)
	v1.POST("/torrent-client/rule-magnet", h.HandleTorrentClientAddMagnetFromRule)

	//
	// Download
	//

	v1.POST("/download-torrent-file", h.HandleDownloadTorrentFile)

	//
	// Updates
	//

	v1.GET("/latest-update", h.HandleGetLatestUpdate)
	v1.GET("/changelog", h.HandleGetChangelog)
	v1.POST("/install-update", h.HandleInstallLatestUpdate)
	v1.POST("/download-release", h.HandleDownloadRelease)

	//
	// Theme
	//

	v1.GET("/theme", h.HandleGetTheme)
	v1.PATCH("/theme", h.HandleUpdateTheme)

	//
	// Playback Manager
	//

	v1.POST("/playback-manager/sync-current-progress", h.HandlePlaybackSyncCurrentProgress)
	v1.POST("/playback-manager/start-playlist", h.HandlePlaybackStartPlaylist)
	v1.POST("/playback-manager/playlist-next", h.HandlePlaybackPlaylistNext)
	v1.POST("/playback-manager/cancel-playlist", h.HandlePlaybackCancelCurrentPlaylist)
	v1.POST("/playback-manager/next-episode", h.HandlePlaybackPlayNextEpisode)
	v1.GET("/playback-manager/next-episode", h.HandlePlaybackGetNextEpisode)
	v1.POST("/playback-manager/autoplay-next-episode", h.HandlePlaybackAutoPlayNextEpisode)
	v1.POST("/playback-manager/play", h.HandlePlaybackPlayVideo)
	v1.POST("/playback-manager/play-random", h.HandlePlaybackPlayRandomVideo)
	//------------
	v1.POST("/playback-manager/manual-tracking/start", h.HandlePlaybackStartManualTracking)
	v1.POST("/playback-manager/manual-tracking/cancel", h.HandlePlaybackCancelManualTracking)

	//
	// Playlists
	//

	v1.GET("/playlists", h.HandleGetPlaylists)
	v1.POST("/playlist", h.HandleCreatePlaylist)
	v1.PATCH("/playlist", h.HandleUpdatePlaylist)
	v1.DELETE("/playlist", h.HandleDeletePlaylist)
	v1.GET("/playlist/episodes/:id/:progress", h.HandleGetPlaylistEpisodes)

	//
	// Onlinestream
	//

	v1.POST("/onlinestream/episode-source", h.HandleGetOnlineStreamEpisodeSource)
	v1.POST("/onlinestream/episode-list", h.HandleGetOnlineStreamEpisodeList)
	v1.DELETE("/onlinestream/cache", h.HandleOnlineStreamEmptyCache)

	v1.POST("/onlinestream/search", h.HandleOnlinestreamManualSearch)
	v1.POST("/onlinestream/manual-mapping", h.HandleOnlinestreamManualMapping)
	v1.POST("/onlinestream/get-mapping", h.HandleGetOnlinestreamMapping)
	v1.POST("/onlinestream/remove-mapping", h.HandleRemoveOnlinestreamMapping)

	//
	// Metadata Provider
	//

	v1.POST("/metadata-provider/tvdb-episodes", h.HandlePopulateTVDBEpisodes)
	v1.DELETE("/metadata-provider/tvdb-episodes", h.HandleEmptyTVDBEpisodes)

	v1.POST("/metadata-provider/filler", h.HandlePopulateFillerData)
	v1.DELETE("/metadata-provider/filler", h.HandleRemoveFillerData)

	//
	// Manga
	//

	v1Manga := v1.Group("/manga")
	v1Manga.POST("/anilist/collection", h.HandleGetAnilistMangaCollection)
	v1Manga.GET("/anilist/collection/raw", h.HandleGetRawAnilistMangaCollection)
	v1Manga.POST("/anilist/collection/raw", h.HandleGetRawAnilistMangaCollection)
	v1Manga.POST("/anilist/list", h.HandleAnilistListManga)
	v1Manga.GET("/collection", h.HandleGetMangaCollection)
	v1Manga.GET("/latest-chapter-numbers", h.HandleGetMangaLatestChapterNumbersMap)
	v1Manga.POST("/refetch-chapter-containers", h.HandleRefetchMangaChapterContainers)
	v1Manga.GET("/entry/:id", h.HandleGetMangaEntry)
	v1Manga.GET("/entry/:id/details", h.HandleGetMangaEntryDetails)
	v1Manga.DELETE("/entry/cache", h.HandleEmptyMangaEntryCache)
	v1Manga.POST("/chapters", h.HandleGetMangaEntryChapters)
	v1Manga.POST("/pages", h.HandleGetMangaEntryPages)
	v1Manga.POST("/update-progress", h.HandleUpdateMangaProgress)

	v1Manga.GET("/downloaded-chapters/:id", h.HandleGetMangaEntryDownloadedChapters)
	v1Manga.GET("/downloads", h.HandleGetMangaDownloadsList)
	v1Manga.POST("/download-chapters", h.HandleDownloadMangaChapters)
	v1Manga.POST("/download-data", h.HandleGetMangaDownloadData)
	v1Manga.DELETE("/download-chapter", h.HandleDeleteMangaDownloadedChapters)
	v1Manga.GET("/download-queue", h.HandleGetMangaDownloadQueue)
	v1Manga.POST("/download-queue/start", h.HandleStartMangaDownloadQueue)
	v1Manga.POST("/download-queue/stop", h.HandleStopMangaDownloadQueue)
	v1Manga.DELETE("/download-queue", h.HandleClearAllChapterDownloadQueue)
	v1Manga.POST("/download-queue/reset-errored", h.HandleResetErroredChapterDownloadQueue)

	v1Manga.POST("/search", h.HandleMangaManualSearch)
	v1Manga.POST("/manual-mapping", h.HandleMangaManualMapping)
	v1Manga.POST("/get-mapping", h.HandleGetMangaMapping)
	v1Manga.POST("/remove-mapping", h.HandleRemoveMangaMapping)

	v1Manga.GET("/local-page/:path", h.HandleGetLocalMangaPage)

	//
	// File Cache
	//

	v1FileCache := v1.Group("/filecache")
	v1FileCache.GET("/total-size", h.HandleGetFileCacheTotalSize)
	v1FileCache.DELETE("/bucket", h.HandleRemoveFileCacheBucket)
	v1FileCache.GET("/mediastream/videofiles/total-size", h.HandleGetFileCacheMediastreamVideoFilesTotalSize)
	v1FileCache.DELETE("/mediastream/videofiles", h.HandleClearFileCacheMediastreamVideoFiles)

	//
	// Discord
	//

	v1Discord := v1.Group("/discord")
	v1Discord.POST("/presence/manga", h.HandleSetDiscordMangaActivity)
	v1Discord.POST("/presence/legacy-anime", h.HandleSetDiscordLegacyAnimeActivity)
	v1Discord.POST("/presence/anime", h.HandleSetDiscordAnimeActivityWithProgress)
	v1Discord.POST("/presence/anime-update", h.HandleUpdateDiscordAnimeActivityWithProgress)
	v1Discord.POST("/presence/cancel", h.HandleCancelDiscordActivity)

	//
	// Media Stream
	//
	v1.GET("/mediastream/settings", h.HandleGetMediastreamSettings)
	v1.PATCH("/mediastream/settings", h.HandleSaveMediastreamSettings)
	v1.POST("/mediastream/request", h.HandleRequestMediastreamMediaContainer)
	v1.POST("/mediastream/preload", h.HandlePreloadMediastreamMediaContainer)
	// Transcode
	v1.POST("/mediastream/shutdown-transcode", h.HandleMediastreamShutdownTranscodeStream)
	v1.GET("/mediastream/transcode/*", h.HandleMediastreamTranscode)
	v1.GET("/mediastream/subs/*", h.HandleMediastreamGetSubtitles)
	v1.GET("/mediastream/att/*", h.HandleMediastreamGetAttachments)
	v1.GET("/mediastream/direct", h.HandleMediastreamDirectPlay)
	v1.HEAD("/mediastream/direct", h.HandleMediastreamDirectPlay)
	v1.GET("/mediastream/file", h.HandleMediastreamFile)

	//
	// Direct Stream
	//
	v1.POST("/directstream/play/localfile", h.HandleDirectstreamPlayLocalFile)
	v1.GET("/directstream/stream", echo.WrapHandler(h.HandleDirectstreamGetStream()))
	v1.HEAD("/directstream/stream", echo.WrapHandler(h.HandleDirectstreamGetStream()))
	v1.GET("/directstream/att/*", h.HandleDirectstreamGetAttachments)

	//
	// Torrent stream
	//
	v1.GET("/torrentstream/settings", h.HandleGetTorrentstreamSettings)
	v1.PATCH("/torrentstream/settings", h.HandleSaveTorrentstreamSettings)
	v1.POST("/torrentstream/start", h.HandleTorrentstreamStartStream)
	v1.POST("/torrentstream/stop", h.HandleTorrentstreamStopStream)
	v1.POST("/torrentstream/drop", h.HandleTorrentstreamDropTorrent)
	v1.POST("/torrentstream/torrent-file-previews", h.HandleGetTorrentstreamTorrentFilePreviews)
	v1.POST("/torrentstream/batch-history", h.HandleGetTorrentstreamBatchHistory)
	v1.GET("/torrentstream/stream/*", h.HandleTorrentstreamServeStream)

	//
	// Extensions
	//

	v1Extensions := v1.Group("/extensions")
	v1Extensions.POST("/playground/run", h.HandleRunExtensionPlaygroundCode)
	v1Extensions.POST("/external/fetch", h.HandleFetchExternalExtensionData)
	v1Extensions.POST("/external/install", h.HandleInstallExternalExtension)
	v1Extensions.POST("/external/uninstall", h.HandleUninstallExternalExtension)
	v1Extensions.POST("/external/edit-payload", h.HandleUpdateExtensionCode)
	v1Extensions.POST("/external/reload", h.HandleReloadExternalExtensions)
	v1Extensions.POST("/external/reload", h.HandleReloadExternalExtension)
	v1Extensions.POST("/all", h.HandleGetAllExtensions)
	v1Extensions.GET("/updates", h.HandleGetExtensionUpdateData)
	v1Extensions.GET("/list", h.HandleListExtensionData)
	v1Extensions.GET("/payload/:id", h.HandleGetExtensionPayload)
	v1Extensions.GET("/list/development", h.HandleListDevelopmentModeExtensions)
	v1Extensions.GET("/list/manga-provider", h.HandleListMangaProviderExtensions)
	v1Extensions.GET("/list/onlinestream-provider", h.HandleListOnlinestreamProviderExtensions)
	v1Extensions.GET("/list/anime-torrent-provider", h.HandleListAnimeTorrentProviderExtensions)
	v1Extensions.GET("/user-config/:id", h.HandleGetExtensionUserConfig)
	v1Extensions.POST("/user-config", h.HandleSaveExtensionUserConfig)
	v1Extensions.GET("/marketplace", h.HandleGetMarketplaceExtensions)
	v1Extensions.GET("/plugin-settings", h.HandleGetPluginSettings)
	v1Extensions.POST("/plugin-settings/pinned-trays", h.HandleSetPluginSettingsPinnedTrays)
	v1Extensions.POST("/plugin-permissions/grant", h.HandleGrantPluginPermissions)

	//
	// Continuity
	//
	v1Continuity := v1.Group("/continuity")
	v1Continuity.PATCH("/item", h.HandleUpdateContinuityWatchHistoryItem)
	v1Continuity.GET("/item/:id", h.HandleGetContinuityWatchHistoryItem)
	v1Continuity.GET("/history", h.HandleGetContinuityWatchHistory)

	//
	// Sync
	//
	v1Local := v1.Group("/local")
	v1Local.GET("/track", h.HandleLocalGetTrackedMediaItems)
	v1Local.POST("/track", h.HandleLocalAddTrackedMedia)
	v1Local.DELETE("/track", h.HandleLocalRemoveTrackedMedia)
	v1Local.GET("/track/:id/:type", h.HandleLocalGetIsMediaTracked)
	v1Local.POST("/local", h.HandleLocalSyncData)
	v1Local.GET("/queue", h.HandleLocalGetSyncQueueState)
	v1Local.POST("/anilist", h.HandleLocalSyncAnilistData)
	v1Local.POST("/updated", h.HandleLocalSetHasLocalChanges)
	v1Local.GET("/updated", h.HandleLocalGetHasLocalChanges)
	v1Local.GET("/storage/size", h.HandleLocalGetLocalStorageSize)
	v1Local.POST("/sync-simulated-to-anilist", h.HandleLocalSyncSimulatedDataToAnilist)

	v1Local.POST("/offline", h.HandleSetOfflineMode)

	//
	// Debrid
	//

	v1.GET("/debrid/settings", h.HandleGetDebridSettings)
	v1.PATCH("/debrid/settings", h.HandleSaveDebridSettings)
	v1.POST("/debrid/torrents", h.HandleDebridAddTorrents)
	v1.POST("/debrid/torrents/download", h.HandleDebridDownloadTorrent)
	v1.POST("/debrid/torrents/cancel", h.HandleDebridCancelDownload)
	v1.DELETE("/debrid/torrent", h.HandleDebridDeleteTorrent)
	v1.GET("/debrid/torrents", h.HandleDebridGetTorrents)
	v1.POST("/debrid/torrents/info", h.HandleDebridGetTorrentInfo)
	v1.POST("/debrid/torrents/file-previews", h.HandleDebridGetTorrentFilePreviews)
	v1.POST("/debrid/stream/start", h.HandleDebridStartStream)
	v1.POST("/debrid/stream/cancel", h.HandleDebridCancelStream)

	//
	// Report
	//

	v1.POST("/report/issue", h.HandleSaveIssueReport)
	v1.GET("/report/issue/download", h.HandleDownloadIssueReport)

	//
	// Nakama
	//

	v1Nakama := v1.Group("/nakama")
	v1Nakama.GET("/ws", h.HandleNakamaWebSocket)
	v1Nakama.POST("/message", h.HandleSendNakamaMessage)
	v1Nakama.POST("/reconnect", h.HandleNakamaReconnectToHost)
	v1Nakama.POST("/cleanup", h.HandleNakamaRemoveStaleConnections)
	v1Nakama.GET("/host/anime/library", h.HandleGetNakamaAnimeLibrary)
	v1Nakama.GET("/host/anime/library/collection", h.HandleGetNakamaAnimeLibraryCollection)
	v1Nakama.GET("/host/anime/library/files/:id", h.HandleGetNakamaAnimeLibraryFiles)
	v1Nakama.GET("/host/anime/library/files", h.HandleGetNakamaAnimeAllLibraryFiles)
	v1Nakama.POST("/play", h.HandleNakamaPlayVideo)
	v1Nakama.GET("/host/torrentstream/stream", h.HandleNakamaHostTorrentstreamServeStream)
	v1Nakama.GET("/host/anime/library/stream", h.HandleNakamaHostAnimeLibraryServeStream)
	v1Nakama.GET("/host/debridstream/stream", h.HandleNakamaHostDebridstreamServeStream)
	v1Nakama.GET("/host/debridstream/url", h.HandleNakamaHostGetDebridstreamURL)
	v1Nakama.GET("/stream", h.HandleNakamaProxyStream)
	v1Nakama.POST("/watch-party/create", h.HandleNakamaCreateWatchParty)
	v1Nakama.POST("/watch-party/join", h.HandleNakamaJoinWatchParty)
	v1Nakama.POST("/watch-party/leave", h.HandleNakamaLeaveWatchParty)

}

func (h *Handler) JSON(c echo.Context, code int, i interface{}) error {
	return c.JSON(code, i)
}

func (h *Handler) RespondWithData(c echo.Context, data interface{}) error {
	return c.JSON(200, NewDataResponse(data))
}

func (h *Handler) RespondWithError(c echo.Context, err error) error {
	return c.JSON(500, NewErrorResponse(err))
}

func headMethodMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Skip directstream route
		if strings.Contains(c.Request().URL.Path, "/directstream/stream") {
			return next(c)
		}

		if c.Request().Method == http.MethodHead {
			// Set the method to GET temporarily to reuse the handler
			c.Request().Method = http.MethodGet

			defer func() {
				c.Request().Method = http.MethodHead
			}() // Restore method after

			// Call the next handler and then clear the response body
			if err := next(c); err != nil {
				if err.Error() == echo.ErrMethodNotAllowed.Error() {
					return c.NoContent(http.StatusOK)
				}

				return err
			}
		}

		return next(c)
	}
}
