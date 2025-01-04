package videofile

import (
	"os"
	"path/filepath"
	"seanime/internal/util"
	"testing"
)

func TestFfprobeGetInfo_1(t *testing.T) {
	t.Skip()

	testFilePath := ""

	mi, err := FfprobeGetInfo("", testFilePath, "1")
	if err != nil {
		t.Fatalf("Error getting media info: %v", err)
	}

	util.Spew(mi)
}

func TestExtractAttachment(t *testing.T) {
	t.Skip()

	testFilePath := ""

	testDir := t.TempDir()

	mi, err := FfprobeGetInfo("", testFilePath, "1")
	if err != nil {
		t.Fatalf("Error getting media info: %v", err)
	}

	util.Spew(mi)

	err = ExtractAttachment("", testFilePath, "1", mi, testDir, util.NewLogger())
	if err != nil {
		t.Fatalf("Error extracting attachment: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(testDir, "videofiles", "1", "att"))
	if err != nil {
		t.Fatalf("Error reading directory: %v", err)
	}

	for _, entry := range entries {
		info, _ := entry.Info()
		t.Logf("Entry: %s, Size: %d\n", entry.Name(), info.Size())
	}
}
