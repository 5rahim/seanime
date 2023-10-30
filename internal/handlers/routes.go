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
	v1 := api.Group("/v1", anilistTokenMiddleware)

	v1.Post("/auth", makeHandler(app, HandleAuth))
	v1.Post("/settings/save", makeHandler(app, HandleSaveSettings))
	v1.Post("/test-dump", makeHandler(app, HandleManualDump))

	v1Library := v1.Group("/library")

	v1Library.Post("/scan", makeHandler(app, HandleScanLocalFiles))
	v1Library.Get("/localfiles/all", makeHandler(app, HandleGetLocalFiles))
	v1Library.Get("/collection", makeHandler(app, HandleGetLibraryCollection))
	v1Library.Get("/entry", makeHandler(app, HandleGetMediaEntry))

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
