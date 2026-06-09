package core

import (
	"fmt"
	golog "log"
	"os"
	"path/filepath"
	"seanime/internal/util"
	"seanime/internal/util/crashlog"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (a *App) InitLogging(updateMode bool) {
	// Create logs directory if it doesn't exist
	if util.IsMobile() {
		if err := os.MkdirAll(a.Config.Logs.Dir, 0700); err != nil {
			fmt.Fprintf(os.Stderr, "[InitLogging] os.MkdirAll failed: %v\n", err)
		}
	} else {
		_ = os.MkdirAll(a.Config.Logs.Dir, 0755)
	}

	// Create log file
	logFilePath := filepath.Join(a.Config.Logs.Dir, fmt.Sprintf("seanime-%s.log", time.Now().Format("2006-01-02_15-04-05")))

	var perm os.FileMode = 0664
	if util.IsMobile() {
		fmt.Fprintf(os.Stderr, "[InitLogging] logFilePath: %q\n", logFilePath)
		perm = 0600
	}

	// Open the log file
	logFile, err := os.OpenFile(
		logFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		perm,
	)
	if err != nil {
		if util.IsMobile() {
			fmt.Fprintf(os.Stderr, "[InitLogging] os.OpenFile failed: %v\n", err)
		}
		a.Logger.Error().Err(err).Msg("app: Failed to open log file")
		return
	}

	log.Logger = *a.Logger
	golog.SetOutput(a.Logger)
	util.SetupLoggerSignalHandling(logFile)
	crashlog.GlobalCrashLogger.SetLogDir(a.Config.Logs.Dir)

	a.OnFlushLogs = func() {
		util.WriteGlobalLogBufferToFile(logFile)
		_ = logFile.Sync()
	}

	if !updateMode {
		go func() {
			for {
				util.WriteGlobalLogBufferToFile(logFile)
				time.Sleep(5 * time.Second)
			}
		}()
	}
}

func TrimLogEntries(dir string, logger *zerolog.Logger) {
	// Get all log files in the directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		logger.Error().Err(err).Msg("core: Failed to read log directory")
		return
	}

	// Get the total size of all log entries
	var totalSize int64
	for _, file := range entries {
		if file.IsDir() {
			continue
		}
		info, err := file.Info()
		if err != nil {
			continue
		}
		totalSize += info.Size()
	}

	var files []os.FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, info)
	}

	var serverLogFiles []os.FileInfo
	var scanLogFiles []os.FileInfo

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "seanime-") {
			serverLogFiles = append(serverLogFiles, file)
		} else if strings.Contains(file.Name(), "-scan") {
			scanLogFiles = append(scanLogFiles, file)
		}
	}

	isMobile := util.IsMobile()

	for _, _files := range [][]os.FileInfo{serverLogFiles, scanLogFiles} {
		files := _files

		// Sort from newest to oldest
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModTime().After(files[j].ModTime())
		})

		deleted := 0
		if isMobile {
			// On mobile, retain at most 2 log files
			for i := 2; i < len(files); i++ {
				err := os.Remove(filepath.Join(dir, files[i].Name()))
				if err != nil {
					continue
				}
				deleted++
			}
		} else {
			// Delete all log files older than 14 days
			for i := 1; i < len(files); i++ {
				if time.Since(files[i].ModTime()) > 14*24*time.Hour {
					err := os.Remove(filepath.Join(dir, files[i].Name()))
					if err != nil {
						continue
					}
					deleted++
				}
			}
		}
		if deleted > 0 {
			if isMobile {
				logger.Info().Msgf("app: Deleted %d old log files to retain at most 2", deleted)
			} else {
				logger.Info().Msgf("app: Deleted %d log files older than 14 days", deleted)
			}
		}
	}
}
