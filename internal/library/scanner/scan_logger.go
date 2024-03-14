package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

// ScanLogger is a custom logger struct for scanning operations.
type ScanLogger struct {
	logger  *zerolog.Logger
	logFile *os.File
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

	// Create an array writer to wrap the JSON encoder

	logger := zerolog.New(logFile).With().Logger()

	return &ScanLogger{&logger, logFile}, nil
}

// NewConsoleScanLogger creates a new mock ScanLogger
func NewConsoleScanLogger() (*ScanLogger, error) {

	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
	}

	// Create an array writer to wrap the JSON encoder

	logger := zerolog.New(output).With().Logger()

	return &ScanLogger{logger: &logger, logFile: nil}, nil
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

func (sl *ScanLogger) Close() {
	if sl.logFile == nil {
		return
	}
	err := sl.logFile.Sync()
	if err != nil {
		return
	}
}
