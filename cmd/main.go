package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/inconshreveable/mousetrap"
	"github.com/seanime-app/seanime/internal/constants"
	"github.com/seanime-app/seanime/internal/core"
	"github.com/seanime-app/seanime/internal/cron"
	"github.com/seanime-app/seanime/internal/handlers"
	"github.com/spf13/cobra"
	"os"
	"runtime"
)

var rootArgs = struct {
	DataDir string
}{}

func init() {
	// Add flags
	rootCmd.Flags().StringVar(&rootArgs.DataDir, "datadir", "", "Directory that contains all Seanime data")
}

var rootCmd = &cobra.Command{
	Use:   "seanime",
	Short: "Self-hosted, user-friendly, media server for anime and manga enthusiasts.",
	Long:  "Self-hosted, user-friendly, media server for anime and manga enthusiasts.",
	Run: func(cmd *cobra.Command, args []string) {
		// Create the app instance
		app := core.NewApp(&core.ConfigOptions{
			DataDir: rootArgs.DataDir,
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
	},
}

func main() {
	col := color.New(color.FgHiBlue)
	bold := color.New(color.FgHiWhite, color.Bold)
	fmt.Println()
	col.Printf("\n .-----.    \n/    _ /  \n\\_..`--.  \n.-._)   \\ \n\\ ")
	bold.Print(constants.Version)
	col.Printf(" / \n `-----'  \n")
	bold.Print(" SEANIME")
	fmt.Println()
	fmt.Println()

	if runtime.GOOS == "windows" && mousetrap.StartedByExplorer() {
		app := core.NewApp(&core.ConfigOptions{
			DataDir: "",
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
	} else {

		if err := rootCmd.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
