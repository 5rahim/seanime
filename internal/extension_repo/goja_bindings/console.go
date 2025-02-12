package goja_bindings

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Console
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type console struct {
	logger *zerolog.Logger
}

func BindConsole(vm *goja.Runtime, logger *zerolog.Logger) error {
	c := &console{
		logger: logger,
	}
	consoleObj := vm.NewObject()
	consoleObj.Set("log", c.logFunc("log"))
	consoleObj.Set("error", c.logFunc("error"))
	consoleObj.Set("warn", c.logFunc("warn"))
	consoleObj.Set("info", c.logFunc("info"))
	consoleObj.Set("debug", c.logFunc("debug"))

	vm.Set("console", consoleObj)

	return nil
}

func (c *console) logFunc(t string) (ret func(c goja.FunctionCall) goja.Value) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error().Msgf("extension: Panic from console: %v", r)
			ret = func(_ goja.FunctionCall) goja.Value {
				return goja.Undefined()
			}
		}
	}()

	return func(call goja.FunctionCall) goja.Value {
		var ret []string
		for _, arg := range call.Arguments {
			switch v := arg.Export().(type) {
			case nil:
				ret = append(ret, "undefined")
			case bool:
				ret = append(ret, fmt.Sprintf("%t", v))
			case int64, float64:
				ret = append(ret, fmt.Sprintf("%v", v))
			case string:
				if v == "" {
					ret = append(ret, fmt.Sprintf("%q", v))
					break
				}
				ret = append(ret, fmt.Sprintf("%s", v))
			case []byte:
				ret = append(ret, fmt.Sprintf("ArrayBuffer(%d) [%s]", len(v), strings.Trim(fmt.Sprint(v), "[]")))
			case map[string]interface{}:
				// Check toString method
				//if toStringI, ok := v["toString"]; ok {
				//	if toString, ok := toStringI.(func(call goja.FunctionCall) goja.Value); ok {
				//		ret = append(ret, toString(goja.FunctionCall{Arguments: nil}).String())
				//		break
				//	}
				//}
				ret = append(ret, fmt.Sprintf("%+v", v))
			default:
				ret = append(ret, fmt.Sprintf("%+v", v))
			}
		}

		switch t {
		case "log", "warn", "info", "debug":
			c.logger.Debug().Msgf("extension: [console.%s] %s", t, strings.Join(ret, " "))
		case "error":
			c.logger.Error().Msgf("extension: [console.error] %s", strings.Join(ret, " "))
		}
		return goja.Undefined()
	}
}

//func (c *console) logFunc(t string) (ret func(c goja.FunctionCall) goja.Value) {
//	defer func() {
//		if r := recover(); r != nil {
//			c.logger.Error().Msgf("extension: Panic from console: %v", r)
//			ret = func(call goja.FunctionCall) goja.Value {
//				return goja.Undefined()
//			}
//		}
//	}()
//
//	return func(call goja.FunctionCall) goja.Value {
//		var ret []string
//		for _, arg := range call.Arguments {
//			if arg == nil || arg.Export() == nil || arg.ExportType() == nil {
//				ret = append(ret, "undefined")
//				continue
//			}
//			if bytes, ok := arg.Export().([]byte); ok {
//				ret = append(ret, fmt.Sprintf("%s", string(bytes)))
//				continue
//			}
//			if arg.ExportType().Kind() == reflect.Struct || arg.ExportType().Kind() == reflect.Map || arg.ExportType().Kind() == reflect.Slice {
//				ret = append(ret, strings.ReplaceAll(spew.Sprint(arg.Export()), "\n", ""))
//			} else {
//				ret = append(ret, fmt.Sprintf("%+v", arg.Export()))
//			}
//		}
//
//		switch t {
//		case "log", "warn", "info", "debug":
//			c.logger.Debug().Msgf("extension: [console.%s] %s", t, strings.Join(ret, " "))
//		case "error":
//			c.logger.Error().Msgf("extension: [console.error] %s", strings.Join(ret, " "))
//		}
//		return goja.Undefined()
//	}
//}
