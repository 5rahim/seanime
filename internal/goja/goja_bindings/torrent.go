package goja_bindings

import (
	"github.com/dop251/goja"
	"seanime/internal/torrents/torrent"
)

func BindTorrentUtils(vm *goja.Runtime) error {
	vm.Set("getMagnetLinkFromTorrentData", getMagnetLinkFromTorrentDataFunc(vm))

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
			panic(vm.NewTypeError("argument is not a string").ToString())
		}

		magnet, err := torrent.StrDataToMagnetLink(str)
		if err != nil {
			return vm.ToValue("")
		}

		return vm.ToValue(magnet)
	}
}
