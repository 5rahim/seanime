//go:build windows && nosystray

package server

import (
	"embed"
)

func StartServer(webFS embed.FS, embeddedLogo []byte) {

	app, flags, selfupdater := startApp(embeddedLogo)

	startAppLoop(&webFS, app, flags, selfupdater)
}