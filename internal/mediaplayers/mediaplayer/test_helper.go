package mediaplayer

import (
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/mediaplayers/libmpv"
	"github.com/seanime-app/seanime/internal/mediaplayers/mpchc"
	"github.com/seanime-app/seanime/internal/mediaplayers/mpv"
	"github.com/seanime-app/seanime/internal/mediaplayers/vlc"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"testing"
)

func NewTestRepository(t *testing.T, defaultPlayer string) *Repository {
	if defaultPlayer == "" {
		defaultPlayer = "mpv"
	}
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	logger := util.NewLogger()
	WSEventManager := events.NewMockWSEventManager(logger)

	_vlc := &vlc.VLC{
		Host:     test_utils.ConfigData.Provider.VlcHost,
		Port:     test_utils.ConfigData.Provider.VlcPort,
		Password: test_utils.ConfigData.Provider.VlcPassword,
		Logger:   logger,
	}

	_mpc := &mpchc.MpcHc{
		Host:   test_utils.ConfigData.Provider.MpcHost,
		Port:   test_utils.ConfigData.Provider.MpcPort,
		Logger: logger,
	}

	_mpv := mpv.New(logger, "", "")

	repo := NewRepository(&NewRepositoryOptions{
		Logger:         logger,
		Default:        defaultPlayer,
		WSEventManager: WSEventManager,
		Mpv:            _mpv,
		VLC:            _vlc,
		MpcHc:          _mpc,
		LibMpv:         libmpv.New(logger),
	})

	return repo
}
