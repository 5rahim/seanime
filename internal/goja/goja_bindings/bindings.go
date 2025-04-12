package goja_bindings

import (
	"github.com/dop251/goja"
)

func gojaValueIsDefined(v goja.Value) bool {
	return v != nil && !goja.IsUndefined(v) && !goja.IsNull(v)
}
