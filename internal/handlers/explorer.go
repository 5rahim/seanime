package handlers

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// HandleOpenInExplorer
//
//	@summary opens the given directory in the file explorer.
//	@desc It returns 'true' whether the operation was successful or not.
//	@route /api/v1/open-in-explorer [POST]
//	@returns bool
func HandleOpenInExplorer(c *RouteCtx) error {

	type body struct {
		Path string `json:"path"`
	}

	p := new(body)
	if err := c.Fiber.BodyParser(p); err != nil {
		return c.RespondWithError(err)
	}

	OpenDirInExplorer(p.Path)

	return c.RespondWithData(true)

}

func OpenDirInExplorer(dir string) {
	if dir == "" {
		return
	}

	cmd := ""
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "explorer"
		args = []string{strings.ReplaceAll(strings.ToLower(dir), "/", "\\")}
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
