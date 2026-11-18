package mobile

import (
	"embed"
	"fmt"
	"net"
	"os"
	"runtime/debug"
	"seanime/internal/core"
	"seanime/internal/handlers"
	"strconv"
)

//go:embed all:web
var WebFS embed.FS

func StartServer(dataDir string, cacheDir string, port int) {
	clearFailure()

	go func() {
		defer func() {
			startingUp.Store(false)
			if r := recover(); r != nil {
				writeCrashReport("panic", fmt.Sprintf("panic: %v\n\n%s", r, debug.Stack()))
				reportFailure(fmt.Sprintf("panic: %v", r))
			}
		}()

		if portInUse(port) {
			reason := fmt.Sprintf("Port %d is being used by another app. Close that app or change the port in the configuration, then start again.", port)
			writeCrashReport("fatal", reason)
			reportFailure(reason)
			return
		}

		_ = os.Setenv("SEANIME_DATA_DIR", dataDir)
		_ = os.Setenv("SEANIME_WORKING_DIR", dataDir)

		flags := core.SeanimeFlags{
			DataDir:          dataDir,
			Port:             port,
			IsDesktopSidecar: false,
		}

		startingUp.Store(true)
		app := core.NewApp(&core.ConfigOptions{
			Flags: flags,
		}, nil)
		startingUp.Store(false)

		app.InitLogging(false)

		echoApp := core.NewEchoApp(app, &WebFS)

		handlers.InitRoutes(app, echoApp)

		core.RunEchoServer(app, echoApp)
	}()
}

func portInUse(port int) bool {
	if port <= 0 {
		return false
	}
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return true
	}
	_ = listener.Close()
	return false
}
