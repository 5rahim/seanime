package scanner

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// ScanLogger is a custom logger struct for scanning operations.
type ScanLogger struct {
	logger  *zerolog.Logger
	logFile *os.File
	buffer  *bytes.Buffer
	mu      sync.Mutex
}

// NewScanLogger creates a new ScanLogger with a log file named based on the current datetime.
// - dir: The directory to save the log file in. This should come from the config.
func NewScanLogger(outputDir string) (*ScanLogger, error) {
	// Generate a log file name with the current datetime
	logFileName := fmt.Sprintf("%s-scan.log", time.Now().Format("2006-01-02_15-04-05"))

	// Create the logs directory if it doesn't exist
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.Mkdir(outputDir, 0755)
		if err != nil {
			return nil, err
		}
	}

	// Open the log file for writing
	logFile, err := os.OpenFile(filepath.Join(outputDir, logFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// Create a buffer for storing log entries
	buffer := new(bytes.Buffer)

	mu := sync.Mutex{}

	// Create an array writer to wrap the JSON encoder
	logger := zerolog.New(ThreadSafeWriteSyncer{buffer, &mu}).With().Logger()

	return &ScanLogger{&logger, logFile, buffer, mu}, nil
}

// NewConsoleScanLogger creates a new mock ScanLogger
func NewConsoleScanLogger() (*ScanLogger, error) {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
	}

	// Create an array writer to wrap the JSON encoder
	logger := zerolog.New(output).With().Logger()

	return &ScanLogger{logger: &logger, logFile: nil, buffer: nil, mu: sync.Mutex{}}, nil
}

func (sl *ScanLogger) LogMediaContainer(level zerolog.Level) *zerolog.Event {
	return sl.logger.WithLevel(level).Str("context", "MediaContainer")
}

func (sl *ScanLogger) LogMatcher(level zerolog.Level) *zerolog.Event {
	return sl.logger.WithLevel(level).Str("context", "Matcher")
}

func (sl *ScanLogger) LogFileHydrator(level zerolog.Level) *zerolog.Event {
	return sl.logger.WithLevel(level).Str("context", "FileHydrator")
}

func (sl *ScanLogger) LogMediaFetcher(level zerolog.Level) *zerolog.Event {
	return sl.logger.WithLevel(level).Str("context", "MediaFetcher")
}

// Done flushes the buffer to the log file and closes the file.
func (sl *ScanLogger) Done() error {
	if sl.logFile == nil {
		return nil
	}

	// Write buffer contents to the log file
	_, err := sl.logFile.Write(sl.buffer.Bytes())
	if err != nil {
		return err
	}

	// Sync and close the log file
	err = sl.logFile.Sync()
	if err != nil {
		return err
	}

	return sl.logFile.Close()
}

func (sl *ScanLogger) Close() {
	if sl.logFile == nil {
		return
	}
	err := sl.logFile.Sync()
	if err != nil {
		return
	}
}

type ThreadSafeWriteSyncer struct {
	w  *bytes.Buffer
	mu *sync.Mutex
}

func (t ThreadSafeWriteSyncer) Write(p []byte) (n int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.w.Write(p)
}
