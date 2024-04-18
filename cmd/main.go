package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/seanime-app/seanime/internal/constants"
	"github.com/seanime-app/seanime/internal/core"
	"github.com/seanime-app/seanime/internal/cron"
	"github.com/seanime-app/seanime/internal/handlers"
	"strings"
)

// @title API
// @version 1.0
// @BasePath /api
func main() {
	col := color.New(color.FgHiMagenta)
	bold := color.New(color.FgHiWhite, color.Bold)
	fmt.Println()
	col.Printf("\n .-----.    \n/    _ /  \n\\_..`--.  \n.-._)   \\ \n\\ ")
	bold.Print(constants.Version)
	col.Printf(" / \n `-----'  \n")
	bold.Print(" SEANIME")
	fmt.Println()
	fmt.Println()

	// Help flag
	flag.Usage = func() {
		fmt.Printf("Self-hosted, user-friendly, media server for anime and manga enthusiasts.\n\n")
		fmt.Printf("Usage:\n  seanime [flags]\n\n")
		fmt.Printf("Flags:\n")
		fmt.Printf("  -datadir, --datadir string")
		fmt.Printf("   directory that contains all Seanime data\n")
		fmt.Printf("  -h                           show this help message\n")
	}
	// Parse flags
	var dataDir string
	flag.StringVar(&dataDir, "datadir", "", "Directory that contains all Seanime data")
	flag.Parse()

	// Create the app instance
	app := core.NewApp(&core.ConfigOptions{
		DataDir: strings.TrimSpace(dataDir),
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
