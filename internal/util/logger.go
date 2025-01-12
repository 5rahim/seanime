package util

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

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

// Stores logs from all loggers. Used to write logs to a file when WriteGlobalLogBufferToFile is called.
// It is reset after writing to a file.
var logBuffer bytes.Buffer
var logBufferMutex = &sync.Mutex{}

func NewLogger() *zerolog.Logger {

	timeFormat := fmt.Sprintf("%s", time.DateTime)
	fieldsOrder := []string{"method", "status", "error", "uri", "latency_human"}
	fieldsExclude := []string{"host", "latency", "referer", "remote_ip", "user_agent", "bytes_in", "bytes_out", "file"}

	// Set up logger
	consoleOutput := zerolog.ConsoleWriter{
		Out:           os.Stdout,
		TimeFormat:    timeFormat,
		FormatLevel:   ZerologFormatLevelPretty,
		FormatMessage: ZerologFormatMessagePretty,
		FieldsExclude: fieldsExclude,
		FieldsOrder:   fieldsOrder,
	}

	fileOutput := zerolog.ConsoleWriter{
		Out:           &logBuffer,
		TimeFormat:    timeFormat,
		FormatMessage: ZerologFormatMessageSimple,
		FormatLevel:   ZerologFormatLevelSimple,
		NoColor:       true, // Needed to prevent color codes from being written to the file
		FieldsExclude: fieldsExclude,
		FieldsOrder:   fieldsOrder,
	}

	multi := zerolog.MultiLevelWriter(consoleOutput, fileOutput)
	logger := zerolog.New(multi).With().Timestamp().Logger()
	return &logger
}

func WriteGlobalLogBufferToFile(file *os.File) {
	defer HandlePanicInModuleThen("util/WriteGlobalLogBufferToFile", func() {})

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

func SetupLoggerSignalHandling(file *os.File) {
	if file == nil {
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Trace().Msgf("Received signal: %s", sig)
		// Flush log buffer to the log file when the app exits
		WriteGlobalLogBufferToFile(file)
		_ = file.Close()
		os.Exit(0)
	}()
}

func ZerologFormatMessagePretty(i interface{}) string {
	if msg, ok := i.(string); ok {
		if bytes.ContainsRune([]byte(msg), ':') {
			parts := strings.SplitN(msg, ":", 2)
			if len(parts) > 1 {
				return colorizeb(parts[0], colorCyan) + colorizeb(" >", colorDarkGray) + parts[1]
			}
		}
		return msg
	}
	return ""
}

func ZerologFormatMessageSimple(i interface{}) string {
	if msg, ok := i.(string); ok {
		if bytes.ContainsRune([]byte(msg), ':') {
			parts := strings.SplitN(msg, ":", 2)
			if len(parts) > 1 {
				return parts[0] + " >" + parts[1]
			}
		}
		return msg
	}
	return ""
}

func ZerologFormatLevelPretty(i interface{}) string {
	if ll, ok := i.(string); ok {
		s := strings.ToLower(ll)
		switch s {
		case "debug":
			s = "DBG" + colorizeb(" -", colorDarkGray)
		case "info":
			s = fmt.Sprint(colorizeb("INF", colorBold)) + colorizeb(" -", colorDarkGray)
		case "warn":
			s = colorizeb("WRN", colorYellow) + colorizeb(" -", colorDarkGray)
		case "trace":
			s = colorizeb("TRC", colorDarkGray) + colorizeb(" -", colorDarkGray)
		case "error":
			s = colorizeb("ERR", colorRed) + colorizeb(" -", colorDarkGray)
		case "fatal":
			s = colorizeb("FTL", colorRed) + colorizeb(" -", colorDarkGray)
		case "panic":
			s = colorizeb("PNC", colorRed) + colorizeb(" -", colorDarkGray)
		}
		return fmt.Sprint(s)
	}
	return ""
}

func ZerologFormatLevelSimple(i interface{}) string {
	if ll, ok := i.(string); ok {
		s := strings.ToLower(ll)
		switch s {
		case "debug":
			s = "|DBG|"
		case "info":
			s = "|INF|"
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

func colorizeb(s interface{}, c int) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}
