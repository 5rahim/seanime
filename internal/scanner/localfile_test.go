package scanner

import (
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetLocalFilesFromDir(t *testing.T) {

	var dir = "E:/Anime"

	logger := util.NewLogger()

	os.Setenv("SEA_LOCAL_DIR", dir)

	localFiles, err := GetLocalFilesFromDir(os.Getenv("SEA_LOCAL_DIR"), logger)

	if assert.NoError(t, err) {
		t.Logf("Found %d local files", len(localFiles))
	}
}
