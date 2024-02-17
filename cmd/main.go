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

	app := core.NewApp(&core.DefaultAppOptions, constants.Version)

	fiberApp := core.NewFiberApp(app)

	handlers.InitRoutes(app, fiberApp)

	core.RunServer(app, fiberApp)

	cron.RunJobs(app)

	select {}

}
