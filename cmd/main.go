package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/seanime-app/seanime/internal/constants"
	"github.com/seanime-app/seanime/internal/core"
	"github.com/seanime-app/seanime/internal/cron"
	"github.com/seanime-app/seanime/internal/handlers"
)

func main() {

	purple := color.New(color.FgHiMagenta)
	fmt.Println()
	purple.Print("                    ⦿ SEANIME")
	fmt.Printf(" %s ", constants.Version)
	fmt.Println()
	fmt.Println()

	// Create the app instance
	app := core.NewApp(&core.DefaultAppOptions, constants.Version)

	// Create the fiber app instance
	fiberApp := core.NewFiberApp(app)

	// Initialize the routes
	handlers.InitRoutes(app, fiberApp)

	core.RunServer(app, fiberApp)

	cron.RunJobs(app)

	select {}

}
