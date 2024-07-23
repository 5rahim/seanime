package onlinestream_providers

import (
	"errors"
)

var (
	ErrSourceNotFound = errors.New("video source not found")
	ErrServerNotFound = errors.New("server not found")
)

const (
	GogoanimeProvider string = "gogoanime"
	ZoroProvider      string = "zoro"
)
