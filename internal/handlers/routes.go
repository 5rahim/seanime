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
	v1.Post("/settings/save", makeHandler(app, HandleSaveSettings))
	// Other
	v1.Post("/test-dump", makeHandler(app, HandleManualDump))
	v1.Post("/directory-selector", makeHandler(app, HandleDirectorySelector))

	//
	// AniList
	//

	v1Anilist := v1.Group("/anilist")
	v1Anilist.Get("/collection", makeHandler(app, HandleGetAnilistCollection))

	//
	// Library
	//

	v1Library := v1.Group("/library")
	v1Library.Post("/scan", makeHandler(app, HandleScanLocalFiles))
	v1Library.Get("/localfiles/all", makeHandler(app, HandleGetLocalFiles))
	v1Library.Get("/collection", makeHandler(app, HandleGetLibraryCollection))
	v1Library.Get("/entry", makeHandler(app, HandleGetMediaEntry))

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
var pool = sync.Pool{
	New: func() interface{} {
		return &RouteCtx{}
	},
}

func makeHandler(app *core.App, handler func(*RouteCtx) error) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := pool.Get().(*RouteCtx)
		defer pool.Put(ctx)
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
	return c.Fiber.Status(404).JSON(NewErrorResponse(err))
}
