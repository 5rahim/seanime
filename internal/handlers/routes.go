package handlers

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"runtime"
	"seanime/internal/core"
	"seanime/internal/util"
	"seanime/internal/util/fiberlogger"
	util2 "seanime/internal/util/proxies"
	"strings"
	"sync"
	"time"
)

func InitRoutes(app *core.App, fiberApp *fiber.App) {

	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Set up a custom logger for fiber.
	// This is not instantiated in `core.NewFiberApp` because we do not want to log requests for the static file server.
	fiberLogger := fiberlogger.New(fiberlogger.Config{
		Logger: app.Logger,
		SkipURIs: []string{
			"/internal/metrics",
			"/_next",
			"/icons",
			"/events",
			"/api/v1/image-proxy",
			"/api/v1/mediastream/transcode/",
			"/api/v1/torrent-client/list",
		},
		Fields:   []string{"method", "error", "url", "latency"},
		Messages: []string{"api: Error", "api: Client error", "api: Success"},
		Levels:   []zerolog.Level{zerolog.ErrorLevel, zerolog.WarnLevel, zerolog.InfoLevel},
	})
	fiberApp.Use(fiberLogger)

	fiberApp.Use(func(c *fiber.Ctx) error {
		// Check if the client has a UUID cookie
		cookie := c.Cookies("Seanime-Client-Id")

		if cookie == "" {
			// Generate a new UUID for the client
			u := uuid.New().String()

			// Create a cookie with the UUID
			cookie := new(fiber.Cookie)
			cookie.Name = "Seanime-Client-Id"
			cookie.Value = u
			cookie.HTTPOnly = false // Make the cookie accessible via JS
			cookie.Expires = time.Now().Add(24 * time.Hour)

			// Set the cookie
			c.Cookie(cookie)

			// Store the UUID in the context for use in the request
			c.Locals("Seanime-Client-Id", u)
		} else {
			// Store the existing UUID in the context for use in the request
			c.Locals("Seanime-Client-Id", cookie)
		}

		return c.Next()
	})

	api := fiberApp.Group("/api")
	v1 := api.Group("/v1")

	if app.IsOffline() {
		v1.Use(func(c *fiber.Ctx) error {
			uriS := strings.Split(c.Request().URI().String(), "v1")
			if len(uriS) > 1 {
				if strings.HasPrefix(uriS[1], "/offline") ||
					strings.HasPrefix(uriS[1], "/settings") ||
					strings.HasPrefix(uriS[1], "/theme") ||
					strings.HasPrefix(uriS[1], "/status") ||
					strings.HasPrefix(uriS[1], "/media-player") ||
					strings.HasPrefix(uriS[1], "/filecache") ||
					strings.HasPrefix(uriS[1], "/playback-manager") ||
					strings.HasPrefix(uriS[1], "/playlists") ||
					strings.HasPrefix(uriS[1], "/directory-selector") ||
					strings.HasPrefix(uriS[1], "/manga") ||
					strings.HasPrefix(uriS[1], "/mediastream") ||
					strings.HasPrefix(uriS[1], "/torrentstream") ||
					strings.HasPrefix(uriS[1], "/extensions") ||
					strings.HasPrefix(uriS[1], "/continuity") ||
					strings.HasPrefix(uriS[1], "/open-in-explorer") {
					return c.Next()
				} else {
					return c.Status(200).SendString("offline")
				}
			}
			return c.Next()
		})
	}

	//fiberApp.Use(pprof.New(pprof.Config{
	//	Prefix: "/api/v1",
	//}))

	v1.Get("/internal/docs", makeHandler(app, HandleGetDocs))

	// Image Proxy
	imageProxy := &util2.ImageProxy{}
	v1.Get("/image-proxy", imageProxy.ProxyImage)

	v1.Get("/proxy", util2.Proxy)

	//
	// General
	//
	v1.Get("/status", makeHandler(app, HandleGetStatus))
	v1.Get("/log/*", makeHandler(app, HandleGetLogContent))
	v1.Get("/logs/filenames", makeHandler(app, HandleGetLogFilenames))
	v1.Delete("/logs", makeHandler(app, HandleDeleteLogs))

	// Auth
	v1.Post("/auth/login", makeHandler(app, HandleLogin))
	v1.Post("/auth/logout", makeHandler(app, HandleLogout))

	// Settings
	v1.Get("/settings", makeHandler(app, HandleGetSettings))
	v1.Patch("/settings", makeHandler(app, HandleSaveSettings))
	v1.Post("/start", makeHandler(app, HandleGettingStarted))
	v1.Patch("/settings/auto-downloader", makeHandler(app, HandleSaveAutoDownloaderSettings))

	// Auto Downloader
	v1.Post("/auto-downloader/run", makeHandler(app, HandleRunAutoDownloader))
	v1.Get("/auto-downloader/rule/:id", makeHandler(app, HandleGetAutoDownloaderRule))
	v1.Get("/auto-downloader/rules", makeHandler(app, HandleGetAutoDownloaderRules))
	v1.Post("/auto-downloader/rule", makeHandler(app, HandleCreateAutoDownloaderRule))
	v1.Patch("/auto-downloader/rule", makeHandler(app, HandleUpdateAutoDownloaderRule))
	v1.Delete("/auto-downloader/rule/:id", makeHandler(app, HandleDeleteAutoDownloaderRule))

	v1.Get("/auto-downloader/items", makeHandler(app, HandleGetAutoDownloaderItems))
	v1.Delete("/auto-downloader/item", makeHandler(app, HandleDeleteAutoDownloaderItem))

	// Other
	v1.Post("/test-dump", makeHandler(app, HandleTestDump))

	v1.Post("/directory-selector", makeHandler(app, HandleDirectorySelector))

	v1.Post("/open-in-explorer", makeHandler(app, HandleOpenInExplorer))

	v1.Post("/media-player/start", makeHandler(app, HandleStartDefaultMediaPlayer))

	//
	// AniList
	//

	v1Anilist := v1.Group("/anilist")

	v1Anilist.Get("/collection", makeHandler(app, HandleGetAnimeCollection))
	v1Anilist.Post("/collection", makeHandler(app, HandleGetAnimeCollection))

	v1Anilist.Get("/collection/raw", makeHandler(app, HandleGetRawAnimeCollection))
	v1Anilist.Post("/collection/raw", makeHandler(app, HandleGetRawAnimeCollection))

	v1Anilist.Get("/media-details/:id", makeHandler(app, HandleGetAnilistAnimeDetails))

	v1Anilist.Get("/studio-details/:id", makeHandler(app, HandleGetAnilistStudioDetails))

	v1Anilist.Post("/list-entry", makeHandler(app, HandleEditAnilistListEntry))

	v1Anilist.Delete("/list-entry", makeHandler(app, HandleDeleteAnilistListEntry))

	v1Anilist.Post("/list-anime", makeHandler(app, HandleAnilistListAnime))

	v1Anilist.Post("/list-recent-anime", makeHandler(app, HandleAnilistListRecentAiringAnime))

	v1Anilist.Get("/list-missed-sequels", makeHandler(app, HandleAnilistListMissedSequels))

	v1Anilist.Get("/stats", makeHandler(app, HandleGetAniListStats))

	//
	// MAL
	//

	v1.Post("/mal/auth", makeHandler(app, HandleMALAuth))

	v1.Post("/mal/logout", makeHandler(app, HandleMALLogout))

	//
	// Library
	//

	v1Library := v1.Group("/library")

	v1Library.Post("/scan", makeHandler(app, HandleScanLocalFiles))

	v1Library.Delete("/empty-directories", makeHandler(app, HandleRemoveEmptyDirectories))

	v1Library.Get("/local-files", makeHandler(app, HandleGetLocalFiles))
	v1Library.Post("/local-files", makeHandler(app, HandleLocalFileBulkAction))
	v1Library.Patch("/local-files", makeHandler(app, HandleUpdateLocalFiles))
	v1Library.Delete("/local-files", makeHandler(app, HandleDeleteLocalFiles))
	v1Library.Get("/local-files/dump", makeHandler(app, HandleDumpLocalFilesToFile))
	v1Library.Post("/local-files/import", makeHandler(app, HandleImportLocalFiles))
	v1Library.Patch("/local-file", makeHandler(app, HandleUpdateLocalFileData))

	v1Library.Get("/collection", makeHandler(app, HandleGetLibraryCollection))

	v1Library.Get("/scan-summaries", makeHandler(app, HandleGetScanSummaries))

	v1Library.Get("/missing-episodes", makeHandler(app, HandleGetMissingEpisodes))

	v1Library.Get("/anime-entry/:id", makeHandler(app, HandleGetAnimeEntry))
	v1Library.Post("/anime-entry/suggestions", makeHandler(app, HandleFetchAnimeEntrySuggestions))
	v1Library.Post("/anime-entry/manual-match", makeHandler(app, HandleAnimeEntryManualMatch))
	v1Library.Patch("/anime-entry/bulk-action", makeHandler(app, HandleAnimeEntryBulkAction))
	v1Library.Post("/anime-entry/open-in-explorer", makeHandler(app, HandleOpenAnimeEntryInExplorer))
	v1Library.Post("/anime-entry/update-progress", makeHandler(app, HandleUpdateAnimeEntryProgress))
	v1Library.Get("/anime-entry/silence/:id", makeHandler(app, HandleGetAnimeEntrySilenceStatus))
	v1Library.Post("/anime-entry/silence", makeHandler(app, HandleToggleAnimeEntrySilenceStatus))

	v1Library.Post("/unknown-media", makeHandler(app, HandleAddUnknownMedia))

	//
	// Torrent / Torrent Client
	//

	v1.Post("/torrent/search", makeHandler(app, HandleSearchTorrent))
	v1.Post("/torrent-client/download", makeHandler(app, HandleTorrentClientDownload))
	v1.Get("/torrent-client/list", makeHandler(app, HandleGetActiveTorrentList))
	v1.Post("/torrent-client/action", makeHandler(app, HandleTorrentClientAction))
	v1.Post("/torrent-client/rule-magnet", makeHandler(app, HandleTorrentClientAddMagnetFromRule))

	//
	// Download
	//

	v1.Post("/download-torrent-file", makeHandler(app, HandleDownloadTorrentFile))

	//
	// Updates
	//

	v1.Get("/latest-update", makeHandler(app, HandleGetLatestUpdate))
	v1.Post("/install-update", makeHandler(app, HandleInstallLatestUpdate))
	v1.Post("/download-release", makeHandler(app, HandleDownloadRelease))

	//
	// Theme
	//

	v1.Get("/theme", makeHandler(app, HandleGetTheme))
	v1.Patch("/theme", makeHandler(app, HandleUpdateTheme))

	//
	// Playback Manager
	//

	v1.Post("/playback-manager/sync-current-progress", makeHandler(app, HandlePlaybackSyncCurrentProgress))
	v1.Post("/playback-manager/start-playlist", makeHandler(app, HandlePlaybackStartPlaylist))
	v1.Post("/playback-manager/playlist-next", makeHandler(app, HandlePlaybackPlaylistNext))
	v1.Post("/playback-manager/cancel-playlist", makeHandler(app, HandlePlaybackCancelCurrentPlaylist))
	v1.Post("/playback-manager/next-episode", makeHandler(app, HandlePlaybackPlayNextEpisode))
	v1.Get("/playback-manager/next-episode", makeHandler(app, HandlePlaybackGetNextEpisode))
	v1.Post("/playback-manager/autoplay-next-episode", makeHandler(app, HandlePlaybackAutoPlayNextEpisode))
	v1.Post("/playback-manager/play", makeHandler(app, HandlePlaybackPlayVideo))
	v1.Post("/playback-manager/play-random", makeHandler(app, HandlePlaybackPlayRandomVideo))
	//------------
	v1.Post("/playback-manager/manual-tracking/start", makeHandler(app, HandlePlaybackStartManualTracking))
	v1.Post("/playback-manager/manual-tracking/cancel", makeHandler(app, HandlePlaybackCancelManualTracking))

	//
	// Playlists
	//

	v1.Get("/playlists", makeHandler(app, HandleGetPlaylists))
	v1.Post("/playlist", makeHandler(app, HandleCreatePlaylist))
	v1.Patch("/playlist", makeHandler(app, HandleUpdatePlaylist))
	v1.Delete("/playlist", makeHandler(app, HandleDeletePlaylist))
	v1.Get("/playlist/episodes/:id/:progress", makeHandler(app, HandleGetPlaylistEpisodes))

	//
	// Onlinestream
	//

	v1.Post("/onlinestream/episode-source", makeHandler(app, HandleGetOnlineStreamEpisodeSource))
	v1.Post("/onlinestream/episode-list", makeHandler(app, HandleGetOnlineStreamEpisodeList))
	v1.Delete("/onlinestream/cache", makeHandler(app, HandleOnlineStreamEmptyCache))

	v1.Post("/onlinestream/search", makeHandler(app, HandleOnlinestreamManualSearch))
	v1.Post("/onlinestream/manual-mapping", makeHandler(app, HandleOnlinestreamManualMapping))
	v1.Post("/onlinestream/get-mapping", makeHandler(app, HandleGetOnlinestreamMapping))
	v1.Post("/onlinestream/remove-mapping", makeHandler(app, HandleRemoveOnlinestreamMapping))

	//
	// Metadata Provider
	//

	v1.Post("/metadata-provider/tvdb-episodes", makeHandler(app, HandlePopulateTVDBEpisodes))
	v1.Delete("/metadata-provider/tvdb-episodes", makeHandler(app, HandleEmptyTVDBEpisodes))

	v1.Post("/metadata-provider/filler", makeHandler(app, HandlePopulateFillerData))
	v1.Delete("/metadata-provider/filler", makeHandler(app, HandleRemoveFillerData))

	//
	// Manga
	//

	v1Manga := v1.Group("/manga")
	v1Manga.Post("/anilist/collection", makeHandler(app, HandleGetAnilistMangaCollection))
	v1Manga.Get("/anilist/collection/raw", makeHandler(app, HandleGetRawAnilistMangaCollection))
	v1Manga.Post("/anilist/collection/raw", makeHandler(app, HandleGetRawAnilistMangaCollection))
	v1Manga.Post("/anilist/list", makeHandler(app, HandleAnilistListManga))
	v1Manga.Get("/collection", makeHandler(app, HandleGetMangaCollection))
	v1Manga.Get("/entry/:id", makeHandler(app, HandleGetMangaEntry))
	v1Manga.Get("/entry/:id/details", makeHandler(app, HandleGetMangaEntryDetails))
	v1Manga.Delete("/entry/cache", makeHandler(app, HandleEmptyMangaEntryCache))
	v1Manga.Post("/chapters", makeHandler(app, HandleGetMangaEntryChapters))
	v1Manga.Post("/pages", makeHandler(app, HandleGetMangaEntryPages))
	v1Manga.Post("/update-progress", makeHandler(app, HandleUpdateMangaProgress))

	v1Manga.Get("/downloads", makeHandler(app, HandleGetMangaDownloadsList))
	v1Manga.Post("/download-chapters", makeHandler(app, HandleDownloadMangaChapters))
	v1Manga.Post("/download-data", makeHandler(app, HandleGetMangaDownloadData))
	v1Manga.Delete("/download-chapter", makeHandler(app, HandleDeleteMangaDownloadedChapters))
	v1Manga.Get("/download-queue", makeHandler(app, HandleGetMangaDownloadQueue))
	v1Manga.Post("/download-queue/start", makeHandler(app, HandleStartMangaDownloadQueue))
	v1Manga.Post("/download-queue/stop", makeHandler(app, HandleStopMangaDownloadQueue))
	v1Manga.Delete("/download-queue", makeHandler(app, HandleClearAllChapterDownloadQueue))
	v1Manga.Post("/download-queue/reset-errored", makeHandler(app, HandleResetErroredChapterDownloadQueue))

	v1Manga.Post("/search", makeHandler(app, HandleMangaManualSearch))
	v1Manga.Post("/manual-mapping", makeHandler(app, HandleMangaManualMapping))
	v1Manga.Post("/get-mapping", makeHandler(app, HandleGetMangaMapping))
	v1Manga.Post("/remove-mapping", makeHandler(app, HandleRemoveMangaMapping))

	//
	// File Cache
	//

	v1FileCache := v1.Group("/filecache")
	v1FileCache.Get("/total-size", makeHandler(app, HandleGetFileCacheTotalSize))
	v1FileCache.Delete("/bucket", makeHandler(app, HandleRemoveFileCacheBucket))
	v1FileCache.Get("/mediastream/videofiles/total-size", makeHandler(app, HandleGetFileCacheMediastreamVideoFilesTotalSize))
	v1FileCache.Delete("/mediastream/videofiles", makeHandler(app, HandleClearFileCacheMediastreamVideoFiles))

	//
	// Discord
	//

	v1Discord := v1.Group("/discord")
	v1Discord.Post("/presence/manga", makeHandler(app, HandleSetDiscordMangaActivity))
	v1Discord.Post("/presence/cancel", makeHandler(app, HandleCancelDiscordActivity))

	//
	// Offline
	//

	v1.Get("/offline/snapshot", makeHandler(app, HandleGetOfflineSnapshot))
	v1.Get("/offline/snapshot-entry", makeHandler(app, HandleGetOfflineSnapshotEntry))
	v1.Post("/offline/snapshot", makeHandler(app, HandleCreateOfflineSnapshot))
	v1.Patch("/offline/snapshot-entry", makeHandler(app, HandleUpdateOfflineEntryListData))
	v1.Post("/offline/sync", makeHandler(app, HandleSyncOfflineData))

	//
	// Media Stream
	//
	v1.Get("/mediastream/settings", makeHandler(app, HandleGetMediastreamSettings))
	v1.Patch("/mediastream/settings", makeHandler(app, HandleSaveMediastreamSettings))
	v1.Post("/mediastream/request", makeHandler(app, HandleRequestMediastreamMediaContainer))
	v1.Post("/mediastream/preload", makeHandler(app, HandlePreloadMediastreamMediaContainer))
	// Transcode
	v1.Post("/mediastream/shutdown-transcode", makeHandler(app, HandleMediastreamShutdownTranscodeStream))
	v1.Get("/mediastream/transcode/*", makeHandler(app, HandleMediastreamTranscode))
	v1.Get("/mediastream/subs/*", makeHandler(app, HandleMediastreamGetSubtitles))
	v1.Get("/mediastream/att/*", makeHandler(app, HandleMediastreamGetAttachments))
	v1.Get("/mediastream/direct", makeHandler(app, HandleMediastreamDirectPlay))
	v1.Get("/mediastream/file/*", makeHandler(app, HandleMediastreamFile))

	//
	// Torrent stream
	//
	v1.Get("/torrentstream/episodes/:id", makeHandler(app, HandleGetTorrentstreamEpisodeCollection))
	v1.Get("/torrentstream/settings", makeHandler(app, HandleGetTorrentstreamSettings))
	v1.Patch("/torrentstream/settings", makeHandler(app, HandleSaveTorrentstreamSettings))
	v1.Post("/torrentstream/start", makeHandler(app, HandleTorrentstreamStartStream))
	v1.Post("/torrentstream/stop", makeHandler(app, HandleTorrentstreamStopStream))
	v1.Post("/torrentstream/drop", makeHandler(app, HandleTorrentstreamDropTorrent))
	v1.Post("/torrentstream/torrent-file-previews", makeHandler(app, HandleGetTorrentstreamTorrentFilePreviews))
	v1.Post("/torrentstream/batch-history", makeHandler(app, HandleGetTorrentstreamBatchHistory))

	//
	// Extensions
	//

	v1Extensions := v1.Group("/extensions")
	v1Extensions.Post("/playground/run", makeHandler(app, HandleRunExtensionPlaygroundCode))
	v1Extensions.Post("/external/fetch", makeHandler(app, HandleFetchExternalExtensionData))
	v1Extensions.Post("/external/install", makeHandler(app, HandleInstallExternalExtension))
	v1Extensions.Post("/external/uninstall", makeHandler(app, HandleUninstallExternalExtension))
	v1Extensions.Post("/external/edit-payload", makeHandler(app, HandleUpdateExtensionCode))
	v1Extensions.Post("/external/reload", makeHandler(app, HandleReloadExternalExtensions))
	v1Extensions.Post("/all", makeHandler(app, HandleGetAllExtensions))
	v1Extensions.Get("/list", makeHandler(app, HandleListExtensionData))
	v1Extensions.Get("/list/manga-provider", makeHandler(app, HandleListMangaProviderExtensions))
	v1Extensions.Get("/list/onlinestream-provider", makeHandler(app, HandleListOnlinestreamProviderExtensions))
	v1Extensions.Get("/list/anime-torrent-provider", makeHandler(app, HandleListAnimeTorrentProviderExtensions))

	//
	// Continuity
	//
	v1Continuity := v1.Group("/continuity")
	v1Continuity.Patch("/item", makeHandler(app, HandleUpdateContinuityWatchHistoryItem))
	v1Continuity.Get("/item/:id", makeHandler(app, HandleGetContinuityWatchHistoryItem))
	v1Continuity.Get("/history", makeHandler(app, HandleGetContinuityWatchHistory))

	//
	// Websocket
	//

	fiberApp.Use("/events", websocketUpgradeMiddleware)
	// Create a new websocket event handler.
	// This will be used to send real-time events to the client.
	// It also attaches the websocket connection to the app instance, so it is available to other handlers.
	fiberApp.Get("/events", newWebSocketEventHandler(app))

}

//----------------------------------------------------------------------------------------------------------------------

// RouteCtx is a context object that is passed to route handlers.
// It contains the App instance and the Fiber context.
type RouteCtx struct {
	App   *core.App
	Fiber *fiber.Ctx
}

// RouteCtx pool
// This is used to avoid allocating memory for each request
var syncPool = sync.Pool{
	New: func() interface{} {
		return &RouteCtx{}
	},
}

// makeHandler creates a new route handler function.
// It takes the App instance and a custom handler function as arguments.
// The custom handler function is similar to a fiber handler, but it takes a RouteCtx as an argument, allowing route handlers to access the app's state.
// We use a sync.Pool to avoid allocating memory for each request.
func makeHandler(app *core.App, handler func(*RouteCtx) error) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) (err error) {
		defer util.HandlePanicInModuleThen("handlers/routes", func() {
			err = errors.New("runtime panic")
		})

		ctx := syncPool.Get().(*RouteCtx)
		defer syncPool.Put(ctx)
		ctx.App = app
		ctx.Fiber = c
		return handler(ctx)
	}
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
	fmt.Printf("\tTotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
	fmt.Printf("\tSys = %v MiB", m.Sys/1024/1024)
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func (c *RouteCtx) AcceptJSON() {
	c.Fiber.Accepts(fiber.MIMEApplicationJSON)
}

// RespondWithData responds with a JSON response containing the given data.
func (c *RouteCtx) RespondWithData(data any) error {
	return c.Fiber.Status(200).JSON(NewDataResponse(data))
}

// RespondWithError responds with a JSON response containing the given error.
func (c *RouteCtx) RespondWithError(err error) error {
	return c.Fiber.Status(500).JSON(NewErrorResponse(err))
}
