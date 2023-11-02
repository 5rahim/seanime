package main

import (
	"fmt"
	"github.com/common-nighthawk/go-figure"
	"github.com/seanime-app/seanime-server/internal/core"
	"github.com/seanime-app/seanime-server/internal/handlers"
)

func main() {

	fmt.Println()
	myFigure := figure.NewFigure("Seanime", "big", true)
	myFigure.Print()
	fmt.Println()

	app := core.NewApp(&core.ServerOptions{})
	fiberApp := core.NewFiberApp(app)
	//fiberWebApp := core.NewFiberWebApp()

	handlers.InitRoutes(app, fiberApp)

	core.RunServer(app, fiberApp)
	//core.RunWebApp(app, fiberWebApp)

	select {}

}
