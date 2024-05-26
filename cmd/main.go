package main

import (
	"github.com/seanime-app/seanime/internal/core"
	"github.com/seanime-app/seanime/internal/cron"
	"github.com/seanime-app/seanime/internal/handlers"
)

func main() {

	// Print the header
	core.PrintHeader()

	// Get the flags
	flags := core.GetSeanimeFlags()

	// Create the app instance
	app := core.NewApp(&core.ConfigOptions{
		DataDir: flags.DataDir,
	})
	defer app.Cleanup()

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
