package goja_bindings

import (
	"errors"

	"github.com/dop251/goja"
)

func gojaValueIsDefined(v goja.Value) bool {
	return v != nil && !goja.IsUndefined(v) && !goja.IsNull(v)
}

func NewErrorString(vm *goja.Runtime, err string) goja.Value {
	return vm.ToValue(vm.NewGoError(errors.New(err)))
}

func NewError(vm *goja.Runtime, err error) goja.Value {
	return vm.ToValue(vm.NewGoError(err))
}

func PanicThrowError(vm *goja.Runtime, err error) {
	goError := vm.NewGoError(err)
	panic(vm.ToValue(goError))
}

func PanicThrowErrorString(vm *goja.Runtime, err string) {
	goError := vm.NewGoError(errors.New(err))
	panic(vm.ToValue(goError))
}

func PanicThrowTypeError(vm *goja.Runtime, err string) {
	panic(vm.NewTypeError(err))
}
