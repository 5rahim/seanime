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

	handlers.InitRoutes(app, fiberApp)

	// Start the server
	core.RunServer(app, fiberApp)

}
