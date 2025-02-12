package goja_bindings

import (
	"encoding/base64"
	"encoding/hex"
	"github.com/dop251/goja"
	"golang.org/x/text/encoding/charmap"
	"unicode/utf16"
)

// UTF-8 Encode
func utf8Parse(input string) []byte {
	return []byte(input)
}

// UTF-8 Decode
func utf8Stringify(input []byte) string {
	return string(input)
}

// Base64 Encode
func base64Parse(input string) []byte {
	data, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil
	}
	return data
}

// Base64 Decode
func base64Stringify(input []byte) string {
	return base64.StdEncoding.EncodeToString(input)
}

// Hex Encode
func hexParse(input string) []byte {
	data, err := hex.DecodeString(input)
	if err != nil {
		return nil
	}
	return data
}

// Hex Decode
func hexStringify(input []byte) string {
	return hex.EncodeToString(input)
}

// Latin1 Encode
func latin1Parse(input string) []byte {
	encoder := charmap.ISO8859_1.NewEncoder()
	data, _ := encoder.Bytes([]byte(input))
	return data
}

// Latin1 Decode
func latin1Stringify(input []byte) string {
	decoder := charmap.ISO8859_1.NewDecoder()
	data, _ := decoder.Bytes(input)
	return string(data)
}

// UTF-16 Encode
func utf16Parse(input string) []byte {
	encoded := utf16.Encode([]rune(input))
	result := make([]byte, len(encoded)*2)
	for i, val := range encoded {
		result[i*2] = byte(val >> 8)
		result[i*2+1] = byte(val)
	}
	return result
}

// UTF-16 Decode
func utf16Stringify(input []byte) string {
	if len(input)%2 != 0 {
		return ""
	}
	decoded := make([]uint16, len(input)/2)
	for i := 0; i < len(decoded); i++ {
		decoded[i] = uint16(input[i*2])<<8 | uint16(input[i*2+1])
	}
	return string(utf16.Decode(decoded))
}

// UTF-16LE Encode
func utf16LEParse(input string) []byte {
	encoded := utf16.Encode([]rune(input))
	result := make([]byte, len(encoded)*2)
	for i, val := range encoded {
		result[i*2] = byte(val)
		result[i*2+1] = byte(val >> 8)
	}
	return result
}

// UTF-16LE Decode
func utf16LEStringify(input []byte) string {
	if len(input)%2 != 0 {
		return ""
	}
	decoded := make([]uint16, len(input)/2)
	for i := 0; i < len(decoded); i++ {
		decoded[i] = uint16(input[i*2]) | uint16(input[i*2+1])<<8
	}
	return string(utf16.Decode(decoded))
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// CryptoJS.enc.Utf8.parse(input: string): WordArray
func cryptoEncUtf8ParseFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.enc.Utf8.parse requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val := call.Argument(0).String()
		return vm.ToValue(utf8Parse(val))
	}
}

// CryptoJS.enc.Utf8.stringify(wordArray: WordArray): string
func cryptoEncUtf8StringifyFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.enc.Utf8.stringify requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val, ok := call.Argument(0).Export().([]byte)
		if !ok {
			return vm.ToValue("")
		}
		return vm.ToValue(utf8Stringify(val))
	}
}

// CryptoJS.enc.Base64.parse(input: string): WordArray
// e.g.
func cryptoEncBase64ParseFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.enc.Base64.parse requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val := call.Argument(0).String()
		return vm.ToValue(base64Parse(val))
	}
}

// CryptoJS.enc.Base64.stringify(wordArray: WordArray): string
func cryptoEncBase64StringifyFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.enc.Base64.stringify requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val, ok := call.Argument(0).Export().([]byte)
		if !ok {
			return vm.ToValue("")
		}
		return vm.ToValue(base64Stringify(val))
	}
}

// CryptoJS.enc.Hex.parse(input: string): WordArray
func cryptoEncHexParseFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.enc.Hex.parse requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val := call.Argument(0).String()
		return vm.ToValue(hexParse(val))
	}
}

// CryptoJS.enc.Hex.stringify(wordArray: WordArray): string
func cryptoEncHexStringifyFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.enc.Hex.stringify requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val, ok := call.Argument(0).Export().([]byte)
		if !ok {
			return vm.ToValue("")
		}
		return vm.ToValue(hexStringify(val))
	}
}

// CryptoJS.enc.Latin1.parse(input: string): WordArray
func cryptoEncLatin1ParseFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.enc.Latin1.parse requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val := call.Argument(0).String()
		return vm.ToValue(latin1Parse(val))
	}
}

// CryptoJS.enc.Latin1.stringify(wordArray: WordArray): string
func cryptoEncLatin1StringifyFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.enc.Latin1.stringify requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val, ok := call.Argument(0).Export().([]byte)
		if !ok {
			return vm.ToValue("")
		}
		return vm.ToValue(latin1Stringify(val))
	}
}

// CryptoJS.enc.Utf16.parse(input: string): WordArray
func cryptoEncUtf16ParseFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.enc.Utf16.parse requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val := call.Argument(0).String()
		return vm.ToValue(utf16Parse(val))
	}
}

// CryptoJS.enc.Utf16.stringify(wordArray: WordArray): string
func cryptoEncUtf16StringifyFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.enc.Utf16.stringify requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val, ok := call.Argument(0).Export().([]byte)
		if !ok {
			return vm.ToValue("")
		}
		return vm.ToValue(utf16Stringify(val))
	}
}

// CryptoJS.enc.Utf16LE.parse(input: string): WordArray
func cryptoEncUtf16LEParseFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.enc.Utf16LE.parse requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val := call.Argument(0).String()
		return vm.ToValue(utf16LEParse(val))
	}
}

// CryptoJS.enc.Utf16LE.stringify(wordArray: WordArray): string
func cryptoEncUtf16LEStringifyFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.enc.Utf16LE.stringify requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val, ok := call.Argument(0).Export().([]byte)
		if !ok {
			return vm.ToValue("")
		}
		return vm.ToValue(utf16LEStringify(val))
	}
}
