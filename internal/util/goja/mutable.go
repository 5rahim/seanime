package goja_util

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/dop251/goja"
)

func BindMutable(vm *goja.Runtime) {
	vm.Set("$mutable", vm.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 || goja.IsUndefined(call.Arguments[0]) || goja.IsNull(call.Arguments[0]) {
			return vm.NewObject()
		}

		// Convert the input to a map first
		jsonBytes, err := json.Marshal(call.Arguments[0].Export())
		if err != nil {
			panic(vm.NewTypeError("Failed to marshal input: %v", err))
		}

		var objMap map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &objMap); err != nil {
			panic(vm.NewTypeError("Failed to unmarshal input: %v", err))
		}

		// Create a new object with getters and setters
		obj := vm.NewObject()

		for key, val := range objMap {
			// Capture current key and value
			k, v := key, val

			if mapVal, ok := v.(map[string]interface{}); ok {
				// For nested objects, create a new mutable object
				nestedObj := vm.NewObject()

				// Add get method
				nestedObj.Set("get", vm.ToValue(func() interface{} {
					return mapVal
				}))

				// Add set method
				nestedObj.Set("set", vm.ToValue(func(call goja.FunctionCall) goja.Value {
					if len(call.Arguments) > 0 {
						newVal := call.Arguments[0].Export()
						if newMap, ok := newVal.(map[string]interface{}); ok {
							mapVal = newMap
							objMap[k] = newMap
						}
					}
					return goja.Undefined()
				}))

				// Add direct property access
				for mk, mv := range mapVal {
					// Capture map key and value
					mapKey := mk
					mapValue := mv
					nestedObj.DefineAccessorProperty(mapKey, vm.ToValue(func() interface{} {
						return mapValue
					}), vm.ToValue(func(call goja.FunctionCall) goja.Value {
						if len(call.Arguments) > 0 {
							mapVal[mapKey] = call.Arguments[0].Export()
						}
						return goja.Undefined()
					}), goja.FLAG_FALSE, goja.FLAG_TRUE)
				}

				obj.Set(k, nestedObj)
			} else if arrVal, ok := v.([]interface{}); ok {
				// For arrays, create a proxy object that allows index access
				arrObj := vm.NewObject()
				for i, av := range arrVal {
					idx := i
					val := av
					arrObj.DefineAccessorProperty(fmt.Sprintf("%d", idx), vm.ToValue(func() interface{} {
						return val
					}), vm.ToValue(func(call goja.FunctionCall) goja.Value {
						if len(call.Arguments) > 0 {
							arrVal[idx] = call.Arguments[0].Export()
							objMap[k] = arrVal
						}
						return goja.Undefined()
					}), goja.FLAG_FALSE, goja.FLAG_TRUE)
				}
				arrObj.Set("length", len(arrVal))

				// Add explicit get/set methods for arrays
				arrObj.Set("get", vm.ToValue(func() interface{} {
					return arrVal
				}))
				arrObj.Set("set", vm.ToValue(func(call goja.FunctionCall) goja.Value {
					if len(call.Arguments) > 0 {
						newVal := call.Arguments[0].Export()
						if newArr, ok := newVal.([]interface{}); ok {
							arrVal = newArr
							objMap[k] = newArr
							arrObj.Set("length", len(newArr))
						}
					}
					return goja.Undefined()
				}))
				obj.Set(k, arrObj)
			} else {
				// For primitive values, create simple getter/setter
				obj.DefineAccessorProperty(k, vm.ToValue(func() interface{} {
					return objMap[k]
				}), vm.ToValue(func(call goja.FunctionCall) goja.Value {
					if len(call.Arguments) > 0 {
						objMap[k] = call.Arguments[0].Export()
					}
					return goja.Undefined()
				}), goja.FLAG_FALSE, goja.FLAG_TRUE)
			}
		}

		// Add a toJSON method that creates a fresh copy
		obj.Set("toJSON", vm.ToValue(func() interface{} {
			// Convert to JSON and back to create a fresh copy with no shared references
			jsonBytes, err := json.Marshal(objMap)
			if err != nil {
				panic(vm.NewTypeError("Failed to marshal to JSON: %v", err))
			}

			var freshCopy interface{}
			if err := json.Unmarshal(jsonBytes, &freshCopy); err != nil {
				panic(vm.NewTypeError("Failed to unmarshal from JSON: %v", err))
			}

			return freshCopy
		}))

		// Add a replace method to completely replace a Go struct's contents.
		// Usage in JS: mutableAnime.replace(e.anime)
		obj.Set("replace", vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 1 {
				panic(vm.NewTypeError("replace requires one argument: target"))
			}

			// Use the current internal state.
			jsonBytes, err := json.Marshal(objMap)
			if err != nil {
				panic(vm.NewTypeError("Failed to marshal state: %v", err))
			}

			// Get the reflect.Value of the target pointer
			target := call.Arguments[0].Export()
			targetVal := reflect.ValueOf(target)
			if targetVal.Kind() != reflect.Ptr {
				// panic(vm.NewTypeError("Target must be a pointer"))
				return goja.Undefined()
			}

			// Create a new instance of the target type and unmarshal into it
			newVal := reflect.New(targetVal.Elem().Type())
			if err := json.Unmarshal(jsonBytes, newVal.Interface()); err != nil {
				panic(vm.NewTypeError("Failed to unmarshal into target: %v", err))
			}

			// Replace the contents of the target with the new value
			targetVal.Elem().Set(newVal.Elem())

			return goja.Undefined()
		}))

		return obj
	}))

	// Add replace function to completely replace a Go struct's contents
	vm.Set("$replace", vm.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.NewTypeError("replace requires two arguments: target and source"))
		}

		target := call.Arguments[0].Export()
		source := call.Arguments[1].Export()

		// Marshal source to JSON
		sourceJSON, err := json.Marshal(source)
		if err != nil {
			panic(vm.NewTypeError("Failed to marshal source: %v", err))
		}

		// Get the reflect.Value of the target pointer
		targetVal := reflect.ValueOf(target)
		if targetVal.Kind() != reflect.Ptr {
			// panic(vm.NewTypeError("Target must be a pointer"))
			// TODO: Handle non-pointer targets
			return goja.Undefined()
		}

		// Create a new instance of the target type
		newVal := reflect.New(targetVal.Elem().Type())

		// Unmarshal JSON into the new instance
		if err := json.Unmarshal(sourceJSON, newVal.Interface()); err != nil {
			panic(vm.NewTypeError("Failed to unmarshal into target: %v", err))
		}

		// Replace the contents of the target with the new value
		targetVal.Elem().Set(newVal.Elem())

		return goja.Undefined()
	}))

	vm.Set("$clone", vm.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Undefined()
		}

		// First convert to JSON to strip all pointers and references
		jsonBytes, err := json.Marshal(call.Arguments[0].Export())
		if err != nil {
			panic(vm.NewTypeError("Failed to marshal input: %v", err))
		}

		// Then unmarshal into a fresh interface{} to get a completely new object
		var newObj interface{}
		if err := json.Unmarshal(jsonBytes, &newObj); err != nil {
			panic(vm.NewTypeError("Failed to unmarshal input: %v", err))
		}

		// Convert back to a goja value
		return vm.ToValue(newObj)
	}))
}
