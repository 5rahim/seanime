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

	api.Post("*", makeHandler(app, HandleEnforceAnilistToken))
	api.Post("/auth", makeHandler(app, HandleAuth))

	api.Post("/scan", makeHandler(app, HandleScanLocalFiles))
	api.Get("/localfiles/latest", makeHandler(app, HandleGetLocalFiles))
	api.Get("/entries/all", makeHandler(app, HandleGetLibraryEntries))

	api.Post("/settings/save", makeHandler(app, HandleSaveSettings))
	api.Post("/test-dump", makeHandler(app, HandleManualDump))

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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

func (c *RouteCtx) GetAnilistToken() string {
	return c.Fiber.Cookies("anilistToken", "")
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
