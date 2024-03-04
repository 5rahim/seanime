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

	col := color.New(color.FgHiBlue)
	bold := color.New(color.FgHiWhite, color.Bold)
	fmt.Println()
	col.Print("           *      .        \n    .            *           \n                      *             \n        *        *      \n                    .     \n")
	col.Print("     ^  +~+~~\n    ^   )`.).\n      )``)``) .~~\n      ).-'.-')|)\n    |-).-).-'_'-/\n^~^~\\ `o-o-o'  /~^~^ ~^~~\n^~~^~'---.____/~^~^ ~^~~^~^\n")
	col.Printf("^~~^= =~^=~^~^= ~^~^~  ~^~^ ~^~`\n")
	col.Print("=-=-_-__=_-=")
	bold.Printf(" SEANIME %s ", constants.Version)
	col.Print("_-= _=_=-=")
	fmt.Println()
	fmt.Println()

	// Create the app instance
	app := core.NewApp(&core.DefaultAppOptions, constants.Version)

	// Create the fiber app instance
	fiberApp := core.NewFiberApp(app)

	// Initialize the routes
	handlers.InitRoutes(app, fiberApp)

	// Run the server
	core.RunServer(app, fiberApp)

	// Run the jobs in the background
	cron.RunJobs(app)

	select {}

}
