package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/seanime-app/seanime-server/internal/core"
	"sync"
)

func InitRoutes(app *core.App, fiberApp *fiber.App) {

	api := fiberApp.Group("/api")

	api.Post("*", makeHandler(app, EnforceAnilistToken))
	api.Post("/auth", makeHandler(app, Auth))
	api.Post("/scan", makeHandler(app, ScanLocalFiles))

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
