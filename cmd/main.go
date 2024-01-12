package main

import (
	"fmt"
	"github.com/common-nighthawk/go-figure"
	"github.com/seanime-app/seanime/internal/core"
	"github.com/seanime-app/seanime/internal/cron"
	"github.com/seanime-app/seanime/internal/handlers"
)

var development = true

func main() {

	fmt.Println()
	myFigure := figure.NewFigure("Seanime", "big", true)
	myFigure.Print()
	fmt.Println()
	fmt.Println("(alpha version, use at your own risk)")
	fmt.Println()

	app := core.NewApp(&core.DefaultAppOptions)
	if development {
		fiberApp := core.NewFiberApp(app)

		handlers.InitRoutes(app, fiberApp)

		core.RunServer(app, fiberApp)
	} else {
		fiberApp := core.NewFiberApp(app)
		fiberWebApp := core.NewFiberWebApp()

		handlers.InitRoutes(app, fiberApp)

		core.RunServer(app, fiberApp)
		core.RunWebApp(app, fiberWebApp)
	}

	cron.RunJobs(app)

	select {}

}
