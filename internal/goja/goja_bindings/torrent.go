package goja_bindings

import (
	"seanime/internal/torrents/torrent"

	"github.com/dop251/goja"
)

func BindTorrentUtils(vm *goja.Runtime) error {
	torrentUtils := vm.NewObject()
	torrentUtils.Set("getMagnetLinkFromTorrentData", getMagnetLinkFromTorrentDataFunc(vm))
	vm.Set("$torrentUtils", torrentUtils)

	return nil
}

func getMagnetLinkFromTorrentDataFunc(vm *goja.Runtime) (ret func(c goja.FunctionCall) goja.Value) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	return func(call goja.FunctionCall) goja.Value {
		defer func() {
			if r := recover(); r != nil {
				panic(vm.ToValue("selection is nil"))
			}
		}()

		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: getMagnetLinkFromTorrentData requires at least 1 argument"))
		}

		str, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue(vm.NewTypeError("argument is not a string")))
		}

		magnet, err := torrent.StrDataToMagnetLink(str)
		if err != nil {
			return vm.ToValue("")
		}

		return vm.ToValue(magnet)
	}
}
