package mediaplayer

import (
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/mediaplayers/mpchc"
	"github.com/seanime-app/seanime/internal/mediaplayers/mpv"
	"github.com/seanime-app/seanime/internal/mediaplayers/vlc"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRepository_StartTracking(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	logger := util.NewLogger()
	WSEventManager := events.NewMockWSEventManager(logger)

	_vlc := &vlc.VLC{
		Host:     "localhost",
		Port:     8080,
		Password: "seanime",
		Logger:   logger,
	}

	_mpc := &mpchc.MpcHc{
		Host:   "localhost",
		Port:   13579,
		Logger: logger,
	}

	repo := &Repository{
		Logger:         logger,
		Default:        "vlc",
		VLC:            _vlc,
		MpcHc:          _mpc,
		Mpv:            mpv.New(logger, "", ""),
		WSEventManager: WSEventManager,
	}

	err := repo.Play("E:\\ANIME\\Violet.Evergarden.The.Movie.1080p.Dual.Audio.BDRip.10.bits.DD.x265-EMBER.mkv")
	assert.NoError(t, err)

	repo.StartTracking()

}
