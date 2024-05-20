package mediaplayer

import (
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRepository_StartTracking(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	repo := NewTestRepository(t, "mpv")

	err := repo.Play("E:\\ANIME\\Sousou no Frieren\\[SubsPlease] Sousou no Frieren - 01 (1080p) [F02B9CEE].mkv")
	assert.NoError(t, err)

	repo.StartTracking()

	go func() {
		time.Sleep(5 * time.Second)
		repo.Stop()
	}()
}
