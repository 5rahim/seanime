package main

import (
	"embed"
	"seanime/internal/app"
)

//go:embed all:web
var WebFS embed.FS

//go:embed internal/icon/logo.png
var embeddedLogo []byte

func main() {
	app.StartApp(WebFS, embeddedLogo)
}
