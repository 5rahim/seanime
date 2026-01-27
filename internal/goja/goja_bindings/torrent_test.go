package goja_bindings

import (
	"seanime/internal/util"
	"testing"
	"time"

	"github.com/dop251/goja"
	gojabuffer "github.com/dop251/goja_nodejs/buffer"
	gojarequire "github.com/dop251/goja_nodejs/require"
	"github.com/stretchr/testify/require"
)

func TestGojaTorrentUtils(t *testing.T) {
	t.Skip("Add a real magnet link")
	vm := goja.New()

	registry := new(gojarequire.Registry)
	registry.Enable(vm)
	gojabuffer.Enable(vm)
	BindTorrentUtils(vm)
	BindConsole(vm, util.NewLogger())
	BindFetch(vm)

	_, err := vm.RunString(`
async function run() {
    try {

        console.log("\nTesting torrent file to magnet link")

        const url = ".torrent"

        const data = await (await fetch(url)).text()
        
        const magnetLink = getMagnetLinkFromTorrentData(data)

        console.log("Magnet link:", magnetLink)
    }
    catch (e) {
        console.error(e)
    }
}
`)
	require.NoError(t, err)

	runFunc, ok := goja.AssertFunction(vm.Get("run"))
	require.True(t, ok)

	ret, err := runFunc(goja.Undefined())
	require.NoError(t, err)

	promise := ret.Export().(*goja.Promise)

	for promise.State() == goja.PromiseStatePending {
		time.Sleep(10 * time.Millisecond)
	}

	if promise.State() == goja.PromiseStateRejected {
		err := promise.Result()
		t.Fatal(err)
	}
}
