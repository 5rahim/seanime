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

        const url = "https://animetosho.org/storage/torrent/da9aad67b6f8bb82757bb3ef95235b42624c34f7/%5BSubsPlease%5D%20Make%20Heroine%20ga%20Oosugiru%21%20-%2011%20%281080p%29%20%5B58B3496A%5D.torrent"

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
