package plugin

import (
	"path/filepath"
	"seanime/internal/extension"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

// BindFilepath binds the filepath module to the Goja runtime.
// Permissions needed: filepath
func (a *AppContextImpl) BindFilepath(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension) {
	filepathObj := vm.NewObject()

	filepathObj.Set("base", filepath.Base)
	filepathObj.Set("clean", filepath.Clean)
	filepathObj.Set("dir", filepath.Dir)
	filepathObj.Set("ext", filepath.Ext)
	filepathObj.Set("fromSlash", filepath.FromSlash)
	filepathObj.Set("glob", filepath.Glob)
	filepathObj.Set("isAbs", filepath.IsAbs)
	filepathObj.Set("join", filepath.Join)
	filepathObj.Set("match", filepath.Match)
	filepathObj.Set("rel", filepath.Rel)
	filepathObj.Set("split", filepath.Split)
	filepathObj.Set("splitList", filepath.SplitList)
	filepathObj.Set("toSlash", filepath.ToSlash)
	filepathObj.Set("walk", filepath.Walk)
	filepathObj.Set("walkDir", filepath.WalkDir)

	_ = vm.Set("$filepath", filepathObj)
}
