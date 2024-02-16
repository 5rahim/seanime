package handlers

import (
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/core"
	"sync"
)

func InitRoutes(app *core.App, fiberApp *fiber.App) {

	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Set up a custom logger for fiber
	fiberLogger := fiberzerolog.New(fiberzerolog.Config{
		Logger: app.Logger,
		SkipURIs: []string{
			"/internal/metrics",
			"/_next",
			"/icons",
		},
		Levels: []zerolog.Level{zerolog.ErrorLevel, zerolog.WarnLevel, zerolog.TraceLevel},
	})
	fiberApp.Use(fiberLogger)

	api := fiberApp.Group("/api")
	v1 := api.Group("/v1")

	//
	// General
	//
	v1.Get("/status", makeHandler(app, HandleStatus))

	// Auth
	v1.Post("/auth/login", makeHandler(app, HandleLogin))
	v1.Post("/auth/logout", makeHandler(app, HandleLogout))

	// Settings
	v1.Get("/settings", makeHandler(app, HandleGetSettings))
	v1.Patch("/settings", makeHandler(app, HandleSaveSettings))
	v1.Patch("/settings/list-sync", makeHandler(app, HandleSaveListSyncSettings))
	v1.Patch("/settings/auto-downloader", makeHandler(app, HandleSaveAutoDownloaderSettings))

	// List Sync
	v1.Get("/list-sync/anime-diffs", makeHandler(app, HandleGetListSyncAnimeDiffs))
	v1.Post("/list-sync/cache", makeHandler(app, HandleDeleteListSyncCache))
	v1.Post("/list-sync/anime", makeHandler(app, HandleSyncAnime))

	// Auto Downloader
	v1.Post("/auto-downloader/run", makeHandler(app, HandleRunAutoDownloaderRule))
	v1.Get("/auto-downloader/rule/:id", makeHandler(app, HandleGetAutoDownloaderRule))
	v1.Get("/auto-downloader/rules", makeHandler(app, HandleGetAutoDownloaderRules))
	v1.Post("/auto-downloader/rule", makeHandler(app, HandleCreateAutoDownloaderRule))
	v1.Patch("/auto-downloader/rule", makeHandler(app, HandleUpdateAutoDownloaderRule))
	v1.Delete("/auto-downloader/rule/:id", makeHandler(app, HandleDeleteAutoDownloaderRule))

	v1.Get("/auto-downloader/items", makeHandler(app, HandleGetAutoDownloaderItems))
	v1.Delete("/auto-downloader/item", makeHandler(app, HandleDeleteAutoDownloaderItem))

	// Other
	v1.Post("/test-dump", makeHandler(app, HandleTestDump))

	// Directory selector input
	// POST /v1/directory-selector
	v1.Post("/directory-selector", makeHandler(app, HandleDirectorySelector))

	// Open directory in explorer
	// POST /v1/open-in-explorer
	v1.Post("/open-in-explorer", makeHandler(app, HandleOpenInExplorer))

	// Open Media Player
	// POST /v1/media-player/start
	v1.Post("/media-player/start", makeHandler(app, HandleStartDefaultMediaPlayer))

	// POST /v1/media-player/play
	v1.Post("/media-player/play", makeHandler(app, HandlePlayVideo))
	// POST /v1/media-player/mpv-detect-playback
	v1.Post("/media-player/mpv-detect-playback", makeHandler(app, HandleMpvDetectPlayback))

	//
	// AniList
	//

	v1Anilist := v1.Group("/anilist")

	// Get "cached" AniList collection
	// GET /v1/anilist/collection
	v1Anilist.Get("/collection", makeHandler(app, HandleGetAnilistCollection))

	// Get (up-to-date) AniList collection
	// This refreshes the collection held by the app
	// POST /v1/anilist/collection
	v1Anilist.Post("/collection", makeHandler(app, HandleGetAnilistCollection))

	// Get details for AniList media
	// GET /v1/anilist/media-details
	v1Anilist.Get("/media-details/:id", makeHandler(app, HandleGetAnilistMediaDetails))

	// Edit AniList List Entry
	// POST /v1/anilist/list-entry
	v1Anilist.Post("/list-entry", makeHandler(app, HandleEditAnilistListEntry))

	// Delete AniList List Entry
	// POST /v1/anilist/list-entry
	v1Anilist.Delete("/list-entry", makeHandler(app, HandleDeleteAnilistListEntry))

	// Edit AniList List Entry's progress
	// POST /v1/anilist/list-entry
	v1Anilist.Post("/list-entry/progress", makeHandler(app, HandleEditAnilistListEntryProgress))

	//
	// MAL
	//

	// Authenticate user with MAL
	// POST /v1/mal/auth
	v1.Post("/mal/auth", makeHandler(app, HandleMALAuth))
	// Logout from MAL
	// POST /v1/mal/logout
	v1.Post("/mal/logout", makeHandler(app, HandleMALLogout))
	// Logout from MAL
	// POST /v1/mal/progress
	v1.Post("/mal/list-entry/progress", makeHandler(app, HandleEditMALListEntryProgress))

	//
	// Library
	//

	v1Library := v1.Group("/library")

	// Scan the library
	v1Library.Post("/scan", makeHandler(app, HandleScanLocalFiles))

	// DELETE /v1/library/empty-directories
	v1Library.Delete("/empty-directories", makeHandler(app, HandleRemoveEmptyDirectories))

	// Get all the local files from the database
	// GET /v1/library/local-files
	v1Library.Get("/local-files", makeHandler(app, HandleGetLocalFiles))

	// POST /v1/library/local-files
	v1Library.Post("/local-files", makeHandler(app, HandleLocalFileBulkAction))

	// DELETE /v1/library/local-files
	v1Library.Delete("/local-files", makeHandler(app, HandleDeleteLocalFiles))

	// Get the library collection
	// GET /v1/library/collection
	v1Library.Get("/collection", makeHandler(app, HandleGetLibraryCollection))

	// Get the latest scan summaries
	// GET /v1/library/scan-summaries
	v1Library.Get("/scan-summaries", makeHandler(app, HandleGetLatestScanSummaries))

	// Get missing episodes
	// GET /v1/library/missing-episodes
	v1Library.Get("/missing-episodes", makeHandler(app, HandleGetMissingEpisodes))

	// Update local file data
	// PATCH /v1/library/local-file
	v1Library.Patch("/local-file", makeHandler(app, HandleUpdateLocalFileData))

	// Retrieve MediaEntry
	// GET /v1/library/media-entry
	v1Library.Get("/media-entry/:id", makeHandler(app, HandleGetMediaEntry))

	// Retrieve SimpleMediaEntry
	// GET /v1/library/simple-media-entry
	v1Library.Get("/simple-media-entry/:id", makeHandler(app, HandleGetSimpleMediaEntry))

	// Get suggestions for a prospective Media Entry
	// POST /v1/library/collection
	v1Library.Post("/media-entry/suggestions", makeHandler(app, HandleFindProspectiveMediaEntrySuggestions))

	// Create Media Entry from directory path and AniList media id
	// POST /v1/library/media-entry/manual-match
	v1Library.Post("/media-entry/manual-match", makeHandler(app, HandleMediaEntryManualMatch))

	// Media Entry Bulk Action
	// PATCH /v1/library/entry/bulk-action
	v1Library.Patch("/media-entry/bulk-action", makeHandler(app, HandleMediaEntryBulkAction))

	// Open Media Entry in File Explorer
	// POST /v1/library/media-entry/open-in-explorer
	v1Library.Post("/media-entry/open-in-explorer", makeHandler(app, HandleOpenMediaEntryInExplorer))

	// Add unknown media by IDs
	// POST /v1/library/unknown-media
	v1Library.Post("/media-entry/unknown-media", makeHandler(app, HandleAddUnknownMedia))

	//
	// Nyaa
	//

	v1.Post("/nyaa/search", makeHandler(app, HandleNyaaSearch))

	//
	// qBittorrent
	//

	v1.Post("/download", makeHandler(app, HandleDownloadNyaaTorrents))
	v1.Get("/torrents", makeHandler(app, HandleGetActiveTorrentList))
	v1.Post("/torrent", makeHandler(app, HandleTorrentAction))

	//
	// Download
	//

	v1.Post("/download-torrent-file", makeHandler(app, HandleDownloadTorrentFile))
	v1.Post("/download-release", makeHandler(app, HandleDownloadRelease))

	//
	// Updates
	//

	v1.Get("/latest-update", makeHandler(app, HandleGetLatestUpdate))

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

func makeHandler(app *core.App, handler func(*RouteCtx) error) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := syncPool.Get().(*RouteCtx)
		defer syncPool.Put(ctx)
		ctx.App = app
		ctx.Fiber = c
		return handler(ctx)
	}
}

func (c *RouteCtx) AcceptJSON() {
	c.Fiber.Accepts(fiber.MIMEApplicationJSON)
}

func (c *RouteCtx) RespondWithData(data any) error {
	return c.Fiber.Status(200).JSON(NewDataResponse(data))
}

func (c *RouteCtx) RespondWithError(err error) error {
	return c.Fiber.Status(500).JSON(NewErrorResponse(err))
}
