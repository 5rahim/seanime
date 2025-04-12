//go:generate go run main.go --skipHandlers=false --skipStructs=false --skipTypes=false --skipPluginEvents=false --skipHookEvents=false
package main

import (
	"flag"
	codegen "seanime/codegen/internal"
)

func main() {

	var skipHandlers bool
	flag.BoolVar(&skipHandlers, "skipHandlers", false, "Skip generating docs")

	var skipStructs bool
	flag.BoolVar(&skipStructs, "skipStructs", false, "Skip generating structs")

	var skipTypes bool
	flag.BoolVar(&skipTypes, "skipTypes", false, "Skip generating types")

	var skipPluginEvents bool
	flag.BoolVar(&skipPluginEvents, "skipPluginEvents", false, "Skip generating plugin events")

	var skipHookEvents bool
	flag.BoolVar(&skipHookEvents, "skipHookEvents", false, "Skip generating hook events")

	flag.Parse()

	if !skipHandlers {
		codegen.GenerateHandlers("../internal/handlers", "./generated")
	}

	if !skipStructs {
		codegen.ExtractStructs("../internal", "./generated")
	}

	if !skipTypes {
		goStructStrs := codegen.GenerateTypescriptEndpointsFile("./generated/handlers.json", "./generated/public_structs.json", "../seanime-web/src/api/generated", "../internal/events")
		codegen.GenerateTypescriptFile("./generated/handlers.json", "./generated/public_structs.json", "../seanime-web/src/api/generated", goStructStrs)
	}

	if !skipPluginEvents {
		codegen.GeneratePluginEventFile("../internal/plugin/ui/events.go", "../seanime-web/src/app/(main)/_features/plugin/generated")
	}

	if !skipHookEvents {
		codegen.GeneratePluginHooksDefinitionFile("../internal/extension_repo/goja_plugin_types", "./generated/public_structs.json", "./generated")
	}

}
