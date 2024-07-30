package crashlog

import (
	"bytes"
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"io"
	"os"
	"path/filepath"
	"seanime/internal/util"
	"sync"
	"time"
)

// Global variable that continuously records logs from specific programs and writes them to a file when something unexpected happens.

type CrashLogger struct {
	//logger    *zerolog.Logger
	//logBuffer *bytes.Buffer
	//mu        sync.Mutex
	logDir mo.Option[string]
}

type CrashLoggerArea struct {
	name       string
	logger     *zerolog.Logger
	logBuffer  *bytes.Buffer
	mu         sync.Mutex
	ctx        context.Context
	cancelFunc context.CancelFunc
}

var GlobalCrashLogger = NewCrashLogger()

// NewCrashLogger creates a new CrashLogger instance.
func NewCrashLogger() *CrashLogger {

	//var logBuffer bytes.Buffer
	//
	//fileOutput := zerolog.ConsoleWriter{
	//	Out:           &logBuffer,
	//	TimeFormat:    time.DateTime,
	//	FormatMessage: util.ZerologFormatMessageSimple,
	//	FormatLevel:   util.ZerologFormatLevelSimple,
	//	NoColor:       true,
	//}
	//
	//multi := zerolog.MultiLevelWriter(fileOutput)
	//logger := zerolog.New(multi).With().Timestamp().Logger()

	return &CrashLogger{
		//logger:    &logger,
		//logBuffer: &logBuffer,
		//mu:        sync.Mutex{},
		logDir: mo.None[string](),
	}
}

func (c *CrashLogger) SetLogDir(dir string) {
	c.logDir = mo.Some(dir)
}

// InitArea creates a new CrashLoggerArea instance.
// This instance can be used to log crashes in a specific area.
func (c *CrashLogger) InitArea(area string) *CrashLoggerArea {

	var logBuffer bytes.Buffer

	fileOutput := zerolog.ConsoleWriter{
		Out:         &logBuffer,
		TimeFormat:  time.DateTime,
		FormatLevel: util.ZerologFormatLevelSimple,
		NoColor:     true,
	}

	multi := zerolog.MultiLevelWriter(fileOutput)
	logger := zerolog.New(multi).With().Timestamp().Logger()

	//ctx, cancelFunc := context.WithCancel(context.Background())

	return &CrashLoggerArea{
		logger:    &logger,
		name:      area,
		logBuffer: &logBuffer,
		mu:        sync.Mutex{},
		//ctx:        ctx,
		//cancelFunc: cancelFunc,
	}
}

// Stdout returns the CrashLoggerArea's log buffer so that it can be used as a writer.
//
//	Example:
//		crashLogger := crashlog.GlobalCrashLogger.InitArea("ffmpeg")
//		defer crashLogger.Close()
//
//		cmd.Stdout = crashLogger.Stdout()
func (a *CrashLoggerArea) Stdout() io.Writer {
	return a.logBuffer
}

func (a *CrashLoggerArea) LogError(msg string) {
	a.logger.Error().Msg(msg)
}

func (a *CrashLoggerArea) LogErrorf(format string, args ...interface{}) {
	a.logger.Error().Msgf(format, args...)
}

func (a *CrashLoggerArea) LogInfof(format string, args ...interface{}) {
	a.logger.Info().Msgf(format, args...)
}

// Close should be always called using defer when a new area is created
//
//	logArea := crashlog.GlobalCrashLogger.InitArea("ffmpeg")
//	defer logArea.Close()
func (a *CrashLoggerArea) Close() {
	a.logBuffer.Reset()
	//a.cancelFunc()
}

func (c *CrashLogger) WriteAreaLogToFile(area *CrashLoggerArea) {
	logDir, found := c.logDir.Get()
	if !found {
		return
	}

	// e.g. crash-ffmpeg-2021-09-01_15-04-05.log
	logFilePath := filepath.Join(logDir, fmt.Sprintf("crash-%s-%s.log", area.name, time.Now().Format("2006-01-02_15-04-05")))

	// Create file
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		fmt.Printf("Failed to open log file: %s\n", logFilePath)
		return
	}
	defer logFile.Close()

	area.mu.Lock()
	defer area.mu.Unlock()
	if _, err := area.logBuffer.WriteTo(logFile); err != nil {
		fmt.Printf("Failed to write crash log buffer to file for %s\n", area.name)
	}
	area.logBuffer.Reset()
}
