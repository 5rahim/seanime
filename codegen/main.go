//go:generate go run main.go --skipHandlers=false --skipStructs=false --skipTypes=false

package main

import (
	"flag"
	"seanime/codegen/internal"
)

func main() {

	var skipHandlers bool
	flag.BoolVar(&skipHandlers, "skipHandlers", false, "Skip generating docs")

	var skipStructs bool
	flag.BoolVar(&skipStructs, "skipStructs", false, "Skip generating structs")

	var skipTypes bool
	flag.BoolVar(&skipTypes, "skipTypes", false, "Skip generating types")

	flag.Parse()

	if !skipHandlers {
		codegen.GenerateHandlers("../internal/handlers", "./generated")
	}

	if !skipStructs {
		codegen.ExtractStructs("../internal", "./generated")
	}

	if !skipTypes {
		goStructStrs := codegen.GenerateTypescriptEndpointsFile("./generated/handlers.json", "./generated/public_structs.json", "../seanime-web/src/api/generated")
		codegen.GenerateTypescriptFile("./generated/handlers.json", "./generated/public_structs.json", "../seanime-web/src/api/generated", goStructStrs)
	}

}
