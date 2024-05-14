package util

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	colorBlack = iota + 30
	colorRed
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
	colorCyan
	colorWhite

	colorBold     = 1
	colorDarkGray = 90

	unknownLevel = "???"
)

func NewLogger() *zerolog.Logger {
	// Set up logger
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: fmt.Sprintf("|%s|", time.DateTime),
		FormatLevel: func(i interface{}) string {
			if ll, ok := i.(string); ok {
				s := strings.ToLower(ll)
				switch s {
				case "debug":
					s = "|DBG|"
				case "info":
					s = "|" + fmt.Sprint(colorizeb("INF", colorBold)) + "|"
				case "warn":
					s = colorizeb("|WRN|", colorYellow)
				case "trace":
					s = colorizeb("|TRC|", colorDarkGray)
				case "error":
					s = colorizeb("|ERR|", colorRed)
				case "fatal":
					s = colorizeb("|FTL|", colorRed)
				case "panic":
					s = colorizeb("|PNC|", colorRed)
				}
				return fmt.Sprint(s)
			}
			return ""
		},
		FormatMessage: func(i interface{}) string {
			if msg, ok := i.(string); ok {
				if bytes.ContainsRune([]byte(msg), ':') {
					parts := strings.SplitN(msg, ":", 2)
					if len(parts) > 1 && len(parts[0]) < len(parts[1]) {
						return colorizeb(parts[0]+" >", colorCyan) + parts[1]
					}
				}
				return msg
			}
			return ""
		},
	}
	logger := zerolog.New(output).With().Timestamp().Logger()
	return &logger
}

func colorize(s interface{}, c color.Attribute) string {
	return color.New(c).Sprint(s)
}

func colorizeb(s interface{}, c int) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}
