package util

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func NewLogger() *zerolog.Logger {
	// Set up logger
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
	}
	logger := zerolog.New(output).With().Timestamp().Logger()
	return &logger
}
