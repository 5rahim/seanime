//go:generate go run main.go --skipHandlers=false --skipStructs=true --skipTypes=true --skipEndpoints=false

package main

import (
	"flag"
	"github.com/seanime-app/seanime/codegen/internal"
)

func main() {

	var skipHandlers bool
	flag.BoolVar(&skipHandlers, "skipHandlers", false, "Skip generating docs")

	var skipStructs bool
	flag.BoolVar(&skipStructs, "skipStructs", false, "Skip generating structs")

	var skipTypes bool
	flag.BoolVar(&skipTypes, "skipTypes", false, "Skip generating types")

	var skipEndpoints bool
	flag.BoolVar(&skipEndpoints, "skipEndpoints", false, "Skip generating endpoints")

	flag.Parse()

	if !skipHandlers {
		codegen.GenerateHandlers("../internal/handlers", "./generated")
	}

	if !skipStructs {
		codegen.ExtractStructs("../internal", "./generated")
	}

	if !skipTypes {
		codegen.GenerateTypescriptFile("./generated/handlers.json", "./generated/public_structs.json", "../seanime-web/src/api/generated")
	}

	if !skipEndpoints {
		codegen.GenerateTypescriptEndpointsFile("./generated/handlers.json", "./generated/public_structs.json", "../seanime-web/src/api/generated")
	}

}
