package mpchc

import (
	"github.com/stretchr/testify/assert"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
	"time"
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
	assert.Nil(t, err)

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
	assert.Nil(t, err)

	res, err := mpc.OpenAndPlay("E:\\ANIME\\Violet.Evergarden.The.Movie.1080p.Dual.Audio.BDRip.10.bits.DD.x265-EMBER.mkv")
	assert.Nil(t, err)

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
	assert.Nil(t, err)

	res, err := mpc.GetVariables()
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Logf("%+v", res)

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
	assert.Nil(t, err)

	_, err = mpc.OpenAndPlay("E:\\ANIME\\Violet.Evergarden.The.Movie.1080p.Dual.Audio.BDRip.10.bits.DD.x265-EMBER.mkv")
	assert.Nil(t, err)

	err = mpc.Pause()

	time.Sleep(400 * time.Millisecond)

	err = mpc.Seek(100000)
	assert.Nil(t, err)

	time.Sleep(400 * time.Millisecond)

	err = mpc.Pause()

	vars, err := mpc.GetVariables()
	assert.Nil(t, err)

	t.Logf("%+v", vars)

}
