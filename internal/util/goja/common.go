package gojautil

import (
	"fmt"

	"github.com/dop251/goja"
)

// ExpectStringArg ensures the argument exists and is strictly a string.
// It panics with a TypeError if validation fails.
// Example:
//
//	func do(call goja.FunctionCall) goja.Value {
//		url := gojautil.ExpectStringArg(vm, call, 0)
func ExpectStringArg(vm *goja.Runtime, call goja.FunctionCall, index int) string {
	arg := call.Argument(index)

	if goja.IsUndefined(arg) {
		panic(vm.NewTypeError(fmt.Sprintf("Argument %d is missing", index)))
	}

	// Export returns the underlying Go value
	if _, ok := arg.Export().(string); !ok {
		panic(vm.NewTypeError(fmt.Sprintf("Argument %d must be a string", index)))
	}

	return arg.String()
}

// ExpectIntArg ensures the argument exists and is strictly a number (int64 compatible).
func ExpectIntArg(vm *goja.Runtime, call goja.FunctionCall, index int) int64 {
	arg := call.Argument(index)

	if goja.IsUndefined(arg) {
		panic(vm.NewTypeError(fmt.Sprintf("Argument %d is missing", index)))
	}

	// Goja stores numbers as int64 or float64
	val := arg.ToInteger()
	// We check if it was actually a number type originally
	if _, ok := arg.Export().(int64); !ok {
		if _, ok := arg.Export().(float64); !ok {
			panic(vm.NewTypeError(fmt.Sprintf("Argument %d must be a number", index)))
		}
	}

	return val
}

// ExpectObjectArg ensures the value is a non-null object and returns the *goja.Object wrapper.
func ExpectObjectArg(vm *goja.Runtime, val goja.Value, argName string) *goja.Object {
	if val == nil || goja.IsUndefined(val) || goja.IsNull(val) {
		panic(vm.NewTypeError(fmt.Sprintf("%s must be an object", argName)))
	}

	// Export check ensures it's not just a primitive wrapped as an object
	// (e.g. strict validation against "new String('a')")
	if _, ok := val.Export().(map[string]interface{}); !ok {
		panic(vm.NewTypeError(fmt.Sprintf("%s must be a valid object/map", argName)))
	}

	return val.ToObject(vm)
}

// ExpectBoolArg ensures the argument is strictly a boolean.
func ExpectBoolArg(vm *goja.Runtime, call goja.FunctionCall, index int) bool {
	arg := call.Argument(index)

	if goja.IsUndefined(arg) {
		panic(vm.NewTypeError(fmt.Sprintf("Argument %d is missing", index)))
	}

	if _, ok := arg.Export().(bool); !ok {
		panic(vm.NewTypeError(fmt.Sprintf("Argument %d must be a boolean", index)))
	}

	return arg.ToBoolean()
}

// ExpectArrayArg ensures the argument is an array and returns the Object wrapper.
// You can then iterate over it using .Export().([]interface{}) or key access.
func ExpectArrayArg(vm *goja.Runtime, call goja.FunctionCall, index int) *goja.Object {
	arg := call.Argument(index)

	if goja.IsUndefined(arg) {
		panic(vm.NewTypeError(fmt.Sprintf("Argument %d is missing", index)))
	}

	// Export() of an array returns []interface{}
	if _, ok := arg.Export().([]interface{}); !ok {
		panic(vm.NewTypeError(fmt.Sprintf("Argument %d must be an array", index)))
	}

	return arg.ToObject(vm)
}

// ExpectFunctionArg ensures the argument is a callable function.
// It returns the goja.Callable which you can invoke directly in Go.
func ExpectFunctionArg(vm *goja.Runtime, call goja.FunctionCall, index int) goja.Callable {
	arg := call.Argument(index)

	if goja.IsUndefined(arg) {
		panic(vm.NewTypeError(fmt.Sprintf("Argument %d is missing", index)))
	}

	fn, ok := goja.AssertFunction(arg)
	if !ok {
		panic(vm.NewTypeError(fmt.Sprintf("Argument %d must be a function", index)))
	}

	return fn
}

// GetStringField extracts a string from an object.
// If 'required' is true, it panics on missing keys or wrong types.
func GetStringField(vm *goja.Runtime, obj *goja.Object, key string, required bool) string {
	val := obj.Get(key)

	if goja.IsUndefined(val) || goja.IsNull(val) {
		if required {
			panic(vm.NewTypeError(fmt.Sprintf("Missing required field: '%s'", key)))
		}
		return ""
	}

	strVal := val.String()
	// Strict type check
	if _, ok := val.Export().(string); !ok {
		panic(vm.NewTypeError(fmt.Sprintf("Field '%s' must be a string", key)))
	}

	return strVal
}

// GetIntField extracts an int from an object.
func GetIntField(vm *goja.Runtime, obj *goja.Object, key string, required bool) int64 {
	val := obj.Get(key)

	if goja.IsUndefined(val) || goja.IsNull(val) {
		if required {
			panic(vm.NewTypeError(fmt.Sprintf("Missing required field: '%s'", key)))
		}
		return 0
	}

	return val.ToInteger()
}
