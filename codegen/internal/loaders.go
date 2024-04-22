package codegen

import (
	"encoding/json"
	"os"
)

func LoadHandlers(path string) []*RouteHandler {
	var handlers []*RouteHandler
	docsContent, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(docsContent, &handlers)
	if err != nil {
		panic(err)
	}
	return handlers
}

func LoadPublicStructs(path string) []*GoStruct {
	var goStructs []*GoStruct
	structsContent, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(structsContent, &goStructs)
	if err != nil {
		panic(err)
	}

	return goStructs
}
