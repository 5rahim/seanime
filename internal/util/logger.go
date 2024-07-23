package util

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
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

var logBuffer bytes.Buffer
var logBufferMutex = &sync.Mutex{}

func NewLogger() *zerolog.Logger {

	timeFormat := fmt.Sprintf("|%s|", time.DateTime)
	formatLevel := func(i interface{}) string {
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
	}
	formatLevelF := func(i interface{}) string {
		if ll, ok := i.(string); ok {
			s := strings.ToLower(ll)
			switch s {
			case "debug":
				s = "|DBG|"
			case "info":
				s = "|" + fmt.Sprint("INF") + "|"
			case "warn":
				s = "|WRN|"
			case "trace":
				s = "|TRC|"
			case "error":
				s = "|ERR|"
			case "fatal":
				s = "|FTL|"
			case "panic":
				s = "|PNC|"
			}
			return fmt.Sprint(s)
		}
		return ""
	}
	formatMessage := func(i interface{}) string {
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
	}
	formatMessageF := func(i interface{}) string {
		if msg, ok := i.(string); ok {
			if bytes.ContainsRune([]byte(msg), ':') {
				parts := strings.SplitN(msg, ":", 2)
				if len(parts) > 1 && len(parts[0]) < len(parts[1]) {
					return parts[0] + " >" + parts[1]
				}
			}
			return msg
		}
		return ""
	}

	// Set up logger
	consoleOutput := zerolog.ConsoleWriter{
		Out:           os.Stdout,
		TimeFormat:    timeFormat,
		FormatLevel:   formatLevel,
		FormatMessage: formatMessage,
	}

	fileOutput := zerolog.ConsoleWriter{
		Out:           &logBuffer,
		NoColor:       true,
		TimeFormat:    timeFormat,
		FormatMessage: formatMessageF,
		FormatLevel:   formatLevelF,
	}

	multi := zerolog.MultiLevelWriter(consoleOutput, fileOutput)
	logger := zerolog.New(multi).With().Timestamp().Logger()
	return &logger
}

func WriteGlobalLogBufferToFile(file *os.File) {
	if file == nil {
		return
	}
	logBufferMutex.Lock()
	defer logBufferMutex.Unlock()
	if _, err := logBuffer.WriteTo(file); err != nil {
		fmt.Print("Failed to write log buffer to file")
	}
	logBuffer.Reset()
}

func colorizeb(s interface{}, c int) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}
