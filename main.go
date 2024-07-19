package main

import (
	"embed"
	"github.com/seanime-app/seanime/internal/app"
)

//go:embed all:web
var WebFS embed.FS

func main() {
	app.StartApp(WebFS)
}
