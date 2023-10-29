package mpchc

import (
	"errors"
	"os/exec"
	"strings"
	"time"
)

func (api *MpcHc) getExecutableName() string {
	if len(api.Path) > 0 {
		if strings.Contains(api.Path, "64") {
			return "mpc-hc64.exe"
		} else {
			return strings.Replace(api.Path, "C:\\Program Files\\MPC-HC\\", "", 1)
		}
	}
	return "mpc-hc64.exe"
}

func (api *MpcHc) getExecutablePath() string {

	if len(api.Path) > 0 {
		return api.Path
	}

	return "C:\\Program Files\\MPC-HC\\mpc-hc64.exe"
}

func (api *MpcHc) isRunning(executable string) bool {
	cmd := exec.Command("tasklist")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), executable)
}

func (api *MpcHc) Start() error {
	name := api.getExecutableName()
	exe := api.getExecutablePath()
	if api.isRunning(name) {
		return nil
	}

	cmd := exec.Command(exe)
	err := cmd.Start()
	if err != nil {
		return errors.New("failed to start MPC-HC")
	}

	time.Sleep(1 * time.Second)

	return nil
}
