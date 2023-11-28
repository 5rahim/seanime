package mediaplayer

import (
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/mpchc"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/vlc"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRepository_StartTracking(t *testing.T) {
	logger := util.NewLogger()
	WSEventManager := events.NewMockWSEventManager(logger)

	vlc := &vlc.VLC{
		Host:     "localhost",
		Port:     8080,
		Password: "seanime",
		Logger:   logger,
	}

	mpc := &mpchc.MpcHc{
		Host:   "localhost",
		Port:   13579,
		Logger: logger,
	}

	repo := &Repository{
		Logger:         logger,
		Default:        "vlc",
		VLC:            vlc,
		MpcHc:          mpc,
		WSEventManager: WSEventManager,
	}

	err := repo.Play("E:\\ANIME\\Violet.Evergarden.The.Movie.1080p.Dual.Audio.BDRip.10.bits.DD.x265-EMBER.mkv")
	assert.NoError(t, err)

	repo.StartTracking()

}
