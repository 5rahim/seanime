package main

import (
	"github.com/seanime-app/seanime/internal/docs"
)

const (
	handlersDirPath = "../internal/handlers"
)

func main() {

	ret := docs.ParseRoutes(handlersDirPath)

	_ = ret

}
