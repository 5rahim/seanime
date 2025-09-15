package handlers

import (
	"os"
	"path/filepath"
	"runtime"
	"seanime/internal/util"
	"strings"

	"github.com/labstack/echo/v4"
)

// HandleOpenInExplorer
//
//	@summary opens the given directory in the file explorer.
//	@desc It returns 'true' whether the operation was successful or not.
//	@route /api/v1/open-in-explorer [POST]
//	@returns bool
func (h *Handler) HandleOpenInExplorer(c echo.Context) error {

	type body struct {
		Path string `json:"path"`
	}

	p := new(body)
	if err := c.Bind(p); err != nil {
		return h.RespondWithError(c, err)
	}

	stat, err := os.Stat(p.Path)
	if err != nil {
		return h.RespondWithError(c, err)
	}
	if !stat.IsDir() {
		p.Path = filepath.Dir(p.Path)
	}

	OpenDirInExplorer(p.Path)

	return h.RespondWithData(c, true)
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
	cmdObj := util.NewCmd(cmd, args...)
	cmdObj.Stdout = os.Stdout
	cmdObj.Stderr = os.Stderr
	_ = cmdObj.Run()
}
