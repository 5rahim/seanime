package handlers

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func HandleOpenInExplorer(c *RouteCtx) error {

	type body struct {
		Path string `json:"path"`
	}

	p := new(body)
	if err := c.Fiber.BodyParser(p); err != nil {
		return c.RespondWithError(err)
	}

	openDirInExplorer(p.Path)

	return c.RespondWithData(true)

}

func openDirInExplorer(dir string) {
	cmd := ""
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "explorer"
		args = []string{strings.ReplaceAll(dir, "/", "\\")}
	case "darwin":
		cmd = "open"
		args = []string{dir}
	case "linux":
		cmd = "xdg-open"
		args = []string{dir}
	default:
		return
	}
	cmdObj := exec.Command(cmd, args...)
	cmdObj.Stdout = os.Stdout
	cmdObj.Stderr = os.Stderr
	_ = cmdObj.Run()
}
