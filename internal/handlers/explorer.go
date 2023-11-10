package handlers

import (
	"fmt"
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

	dir := p.Path
	cmd := ""
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "explorer"
		// Convert the directory path to lowercase for case-insensitivity
		lowerCasePath := strings.ToLower(dir)
		args = []string{lowerCasePath}
	case "darwin":
		cmd = "open"
		args = []string{dir}
	case "linux":
		cmd = "xdg-open"
		args = []string{dir}
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	cmdObj := exec.Command(cmd, args...)
	cmdObj.Stdout = os.Stdout
	cmdObj.Stderr = os.Stderr
	_ = cmdObj.Run()

	return c.RespondWithData(true)

}
