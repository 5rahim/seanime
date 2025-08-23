package mpchc

import "strconv"

func (api *MpcHc) Play() (err error) {
	_, err = api.Execute(playCmd, nil)
	return
}

func (api *MpcHc) Pause() (err error) {
	_, err = api.Execute(pauseCmd, nil)
	return
}

func (api *MpcHc) TogglePlay() (err error) {
	_, err = api.Execute(playPauseCmd, nil)
	return
}

func (api *MpcHc) Stop() (err error) {
	_, err = api.Execute(stopCmd, nil)
	return
}

func (api *MpcHc) ToggleFullScreen() (err error) {
	_, err = api.Execute(fullscreenCmd, nil)
	return
}

// SeekTo position in ms
func (api *MpcHc) SeekTo(pos int) (err error) {
	_, err = api.Execute(seekCmd, map[string]interface{}{"position": millisecondsToDuration(pos)})
	return
}

//----------------------------------------------------------------------------------------------------------------------

func millisecondsToDuration(ms int) string {
	if ms <= 0 {
		return "00:00:00"
	}

	duration := ms / 1000
	hours := duration / 3600
	duration %= 3600

	minutes := duration / 60
	duration %= 60

	return padStart(strconv.Itoa(hours), 2, "0") + ":" + padStart(strconv.Itoa(minutes), 2, "0") + ":" + padStart(strconv.Itoa(duration), 2, "0")
}

func padStart(s string, length int, pad string) string {
	for len(s) < length {
		s = pad + s
	}
	return s
}
