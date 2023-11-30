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
	//output.FormatFieldValue = func(i interface{}) string {
	//	return fmt.Sprintf("\"%s\"", i)
	//}
	logger := zerolog.New(output).With().Timestamp().Logger()
	return &logger
}
