package scanner

import (
	"github.com/stretchr/testify/assert"
	"seanime/internal/util"
	"testing"
)

func TestGetLocalFilesFromDir(t *testing.T) {
	t.Skip("Skipping test that requires local files")

	var dir = "E:/Anime"

	logger := util.NewLogger()

	localFiles, err := GetLocalFilesFromDir(dir, logger)

	if assert.NoError(t, err) {
		t.Logf("Found %d local files", len(localFiles))
	}
}
