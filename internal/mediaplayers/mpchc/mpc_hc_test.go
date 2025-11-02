package mpchc

import (
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestMpcHc_Start(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	mpc := &MpcHc{
		Host:   test_utils.ConfigData.Provider.MpcHost,
		Path:   test_utils.ConfigData.Provider.MpcPath,
		Port:   test_utils.ConfigData.Provider.MpcPort,
		Logger: util.NewLogger(),
	}

	err := mpc.Start()
	assert.NoError(t, err)

}

func TestMpcHc_Play(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	mpc := &MpcHc{
		Host:   test_utils.ConfigData.Provider.MpcHost,
		Path:   test_utils.ConfigData.Provider.MpcPath,
		Port:   test_utils.ConfigData.Provider.MpcPort,
		Logger: util.NewLogger(),
	}

	err := mpc.Start()
	assert.NoError(t, err)

	res, err := mpc.OpenAndPlay("E:\\ANIME\\Violet.Evergarden.The.Movie.1080p.Dual.Audio.BDRip.10.bits.DD.x265-EMBER.mkv")
	assert.NoError(t, err)

	t.Log(res)

}

func TestMpcHc_GetVariables(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	mpc := &MpcHc{
		Host:   test_utils.ConfigData.Provider.MpcHost,
		Path:   test_utils.ConfigData.Provider.MpcPath,
		Port:   test_utils.ConfigData.Provider.MpcPort,
		Logger: util.NewLogger(),
	}

	err := mpc.Start()
	assert.NoError(t, err)

	res, err := mpc.GetVariables()
	if err != nil {
		t.Fatal(err.Error())
	}

	spew.Dump(res)

}

func TestMpcHc_Seek(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	mpc := &MpcHc{
		Host:   test_utils.ConfigData.Provider.MpcHost,
		Path:   test_utils.ConfigData.Provider.MpcPath,
		Port:   test_utils.ConfigData.Provider.MpcPort,
		Logger: util.NewLogger(),
	}

	err := mpc.Start()
	assert.NoError(t, err)

	_, err = mpc.OpenAndPlay("E:\\ANIME\\[SubsPlease] Bocchi the Rock! (01-12) (1080p) [Batch]\\[SubsPlease] Bocchi the Rock! - 01v2 (1080p) [ABDDAE16].mkv")
	assert.NoError(t, err)

	err = mpc.Pause()

	time.Sleep(400 * time.Millisecond)

	err = mpc.SeekTo(100000)
	assert.NoError(t, err)

	time.Sleep(400 * time.Millisecond)

	err = mpc.Pause()

	vars, err := mpc.GetVariables()
	assert.NoError(t, err)

	spew.Dump(vars)

}
