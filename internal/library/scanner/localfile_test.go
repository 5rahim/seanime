package scanner

import (
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
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
