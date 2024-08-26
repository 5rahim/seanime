package goja_bindings

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"github.com/dop251/goja"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
)

// MD5 Hash
func cryptoMD5Func(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.MD5 requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue("TypeError: argument is not a string"))
		}
		hash := md5.Sum([]byte(val))
		return vm.ToValue(hash[:])
	}
}

// SHA1 Hash
func cryptoSHA1Func(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.SHA1 requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue("TypeError: argument is not a string"))
		}
		hash := sha1.Sum([]byte(val))
		return vm.ToValue(hash[:])
	}
}

// SHA256 Hash
func cryptoSHA256Func(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.SHA256 requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue("TypeError: argument is not a string"))
		}
		hash := sha256.Sum256([]byte(val))
		return vm.ToValue(hash[:])
	}
}

// SHA512 Hash
func cryptoSHA512Func(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.SHA512 requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue("TypeError: argument is not a string"))
		}
		hash := sha512.Sum512([]byte(val))
		return vm.ToValue(hash[:])
	}
}

// SHA3 Hash
func cryptoSHA3Func(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.SHA3 requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue("TypeError: argument is not a string"))
		}
		hash := sha3.Sum256([]byte(val))
		return vm.ToValue(hash[:])
	}
}

// RIPEMD-160 Hash
func cryptoRIPEMD160Func(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: CryptoJS.RIPEMD160 requires at least 1 argument"))
		}
		if !gojaValueIsDefined(call.Arguments[0]) {
			return vm.ToValue("")
		}
		val, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue("TypeError: argument is not a string"))
		}
		hasher := ripemd160.New()
		hasher.Write([]byte(val))
		return vm.ToValue(hasher.Sum(nil))
	}
}
