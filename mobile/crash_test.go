package mobile

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFatalReason(t *testing.T) {
	logs := "2026-10-31 15:49:41 |INF| app > Starting\n" +
		"2026-10-31 15:49:42 |FTL| app > Failed to initialize config error=\"bad toml\"\n"

	if got := fatalReason(logs); got != `app > Failed to initialize config error="bad toml"` {
		t.Fatalf("unexpected reason %q", got)
	}
	if got := fatalReason("no fatal line"); got != "The server stopped during startup" {
		t.Fatalf("unexpected fallback %q", got)
	}
}

func TestWriteCrashReport(t *testing.T) {
	dir := t.TempDir()
	crashDir.Store(dir)

	writeCrashReport("fatal", "the logs")

	matches, err := filepath.Glob(filepath.Join(dir, "fatal-*.log"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("expected one report, got %v (%v)", matches, err)
	}
	content, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(content), "type=fatal\n") || !strings.Contains(string(content), "the logs") {
		t.Fatalf("unexpected report %q", content)
	}
}

func TestPruneCrashReports(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()

	for i := range maxCrashReports + 2 {
		path := filepath.Join(dir, fmt.Sprintf("go-%d.log", i))
		if err := os.WriteFile(path, []byte("crash"), 0600); err != nil {
			t.Fatal(err)
		}
		modified := now.Add(-time.Duration(i) * time.Hour)
		if err := os.Chtimes(path, modified, modified); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "go-empty.log"), nil, 0600); err != nil {
		t.Fatal(err)
	}

	pruneCrashReports(dir)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != maxCrashReports {
		t.Fatalf("expected %d reports, got %d", maxCrashReports, len(entries))
	}
	if _, err := os.Stat(filepath.Join(dir, "go-0.log")); err != nil {
		t.Fatalf("newest report was removed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "go-empty.log")); !os.IsNotExist(err) {
		t.Fatalf("empty report was kept: %v", err)
	}
}

func TestFailureSignal(t *testing.T) {
	clearFailure()
	reportFailure("first")
	reportFailure("second")

	if got := WaitForFailure(); got != "first" {
		t.Fatalf("unexpected failure %q", got)
	}

	reportFailure("stale")
	clearFailure()
	select {
	case got := <-failures:
		t.Fatalf("stale failure survived: %q", got)
	default:
	}
}

func TestPortInUse(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	if !portInUse(port) {
		t.Fatal("expected the listening port to be reported as in use")
	}

	_ = listener.Close()

	if portInUse(port) {
		t.Fatal("expected the closed port to be free")
	}
	if portInUse(0) {
		t.Fatal("expected no check without a requested port")
	}
}
