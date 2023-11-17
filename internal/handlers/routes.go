package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/seanime-app/seanime-server/internal/core"
	"sync"
)

func InitRoutes(app *core.App, fiberApp *fiber.App) {

	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

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
	// Other
	v1.Post("/test-dump", makeHandler(app, HandleManualDump))

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

	// Edit AniList Media List Data
	// POST /v1/anilist/list-entry
	v1Anilist.Post("/list-entry", makeHandler(app, HandleEditAnilistListEntry))

	// Edit AniList Media List Entry's score
	// POST /v1/anilist/list-entry
	v1Anilist.Post("/list-entry/progress", makeHandler(app, HandleEditAnilistListEntryProgress))

	//
	// Library
	//

	v1Library := v1.Group("/library")

	// Scan the library
	v1Library.Post("/scan", makeHandler(app, HandleScanLocalFiles))

	// Get all the local files from the database
	// GET /v1/library/local-files
	v1Library.Get("/local-files", makeHandler(app, HandleGetLocalFiles))

	// Get the library collection
	// GET /v1/library/collection
	v1Library.Get("/collection", makeHandler(app, HandleGetLibraryCollection))

	// Get missing episodes
	// GET /v1/library/missing-episodes
	v1Library.Get("/missing-episodes", makeHandler(app, HandleGetMissingEpisodes))

	// Update local file data
	// PATCH /v1/library/local-file
	v1Library.Patch("/local-file", makeHandler(app, HandleUpdateLocalFileData))

	// Retrive MediaEntry
	// GET /v1/library/media-entry
	v1Library.Get("/media-entry/:id", makeHandler(app, HandleGetMediaEntry))

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

	//
	// Nyaa
	//

	v1.Post("/nyaa/search", makeHandler(app, HandleNyaaSearch))

	//
	// qBittorrent
	//

	v1.Post("/download", makeHandler(app, HandleDownloadNyaaTorrents))

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
