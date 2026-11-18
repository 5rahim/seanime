package mobile

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"seanime/internal/constants"
	"seanime/internal/util"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
)

const (
	maxCrashReports  = 5
	maxFatalLogBytes = 32 * 1024
	fatalLevelMarker = "|FTL| "
)

var (
	crashDir   atomic.Value
	startingUp atomic.Bool
	failures   = make(chan string, 1)
)

func init() {
	zerolog.FatalExitFunc = handleFatal
}

func EnableCrashRecording(dir string) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return
	}
	crashDir.Store(dir)
	pruneCrashReports(dir)

	path := filepath.Join(dir, fmt.Sprintf("go-%d.log", time.Now().UnixNano()))
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return
	}
	defer file.Close()

	_ = debug.SetCrashOutput(file, debug.CrashOptions{})
}

func WaitForFailure() string {
	return <-failures
}

func clearFailure() {
	select {
	case <-failures:
	default:
	}
}

func reportFailure(reason string) {
	select {
	case failures <- reason:
	default:
	}
}

func handleFatal() {
	logs := string(util.GlobalLogBufferTail(maxFatalLogBytes))
	writeCrashReport("fatal", logs)

	if startingUp.Load() {
		reportFailure(fatalReason(logs))
		runtime.Goexit()
	}
	os.Exit(1)
}

func fatalReason(logs string) string {
	if index := strings.LastIndex(logs, fatalLevelMarker); index >= 0 {
		line, _, _ := strings.Cut(logs[index+len(fatalLevelMarker):], "\n")
		return strings.TrimSpace(line)
	}
	return "The server stopped during startup"
}

func writeCrashReport(kind string, body string) {
	dir, ok := crashDir.Load().(string)
	if !ok {
		return
	}

	now := time.Now()
	report := fmt.Sprintf("type=%s\ntime=%s\nserver=%s\nplatform=%s/%s\n\n%s\n",
		kind, now.Format(time.RFC3339), constants.Version, runtime.GOOS, runtime.GOARCH, body)

	path := filepath.Join(dir, fmt.Sprintf("%s-%d.log", kind, now.UnixNano()))
	temp := path + ".tmp"
	if err := os.WriteFile(temp, []byte(report), 0600); err != nil {
		return
	}
	_ = os.Rename(temp, path)
}

func pruneCrashReports(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var reports []os.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil || info.IsDir() {
			continue
		}
		if info.Size() == 0 {
			_ = os.Remove(filepath.Join(dir, info.Name()))
			continue
		}
		reports = append(reports, info)
	}

	sort.Slice(reports, func(i, j int) bool {
		return reports[i].ModTime().After(reports[j].ModTime())
	})
	for i := maxCrashReports; i < len(reports); i++ {
		_ = os.Remove(filepath.Join(dir, reports[i].Name()))
	}
}
