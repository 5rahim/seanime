package mobile

import (
	"embed"
	"os"
	"seanime/internal/core"
	"seanime/internal/handlers"
)

//go:embed all:web
var WebFS embed.FS

func StartServer(dataDir string, cacheDir string, port int) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
			}
		}()

		_ = os.Setenv("SEANIME_DATA_DIR", dataDir)
		_ = os.Setenv("SEANIME_WORKING_DIR", dataDir)

		flags := core.SeanimeFlags{
			DataDir:          dataDir,
			Host:             "127.0.0.1",
			Port:             port,
			DisablePassword:  true,
			IsDesktopSidecar: false,
		}

		app := core.NewApp(&core.ConfigOptions{
			Flags: flags,
		}, nil)

		echoApp := core.NewEchoApp(app, &WebFS)

		handlers.InitRoutes(app, echoApp)

		core.RunEchoServer(app, echoApp)
	}()
}
