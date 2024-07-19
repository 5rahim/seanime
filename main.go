package main

import (
	"embed"
	"seanime/internal/app"
)

//go:embed all:web
var WebFS embed.FS

func main() {
	app.StartApp(WebFS)
}
