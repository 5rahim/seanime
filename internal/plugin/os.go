package plugin

import (
	"os"
	"os/exec"
	"runtime"
	"seanime/internal/extension"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

// BindOS binds the os module to the Goja runtime.
// Permissions needed: os
func (a *AppContextImpl) BindOS(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension) {
	osObj := vm.NewObject()

	_ = osObj.Set("platform", runtime.GOOS)
	osObj.Set("args", os.Args)
	osObj.Set("cmd", exec.Command)
	osObj.Set("exit", os.Exit)
	osObj.Set("getenv", os.Getenv)
	osObj.Set("dirFS", os.DirFS)
	osObj.Set("readFile", os.ReadFile)
	osObj.Set("writeFile", os.WriteFile)
	osObj.Set("readDir", os.ReadDir)
	osObj.Set("tempDir", os.TempDir)
	osObj.Set("truncate", os.Truncate)
	osObj.Set("getwd", os.Getwd)
	osObj.Set("mkdir", os.Mkdir)
	osObj.Set("mkdirAll", os.MkdirAll)
	osObj.Set("rename", os.Rename)
	osObj.Set("remove", os.Remove)
	osObj.Set("removeAll", os.RemoveAll)

	_ = vm.Set("$os", osObj)
}
