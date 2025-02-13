package extension_repo_test

import (
	"context"
	"fmt"
	"os"
	"seanime/internal/extension"
	"seanime/internal/extension_repo"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/util"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

func setupTestVM(t *testing.T) *goja.Runtime {
	runtimeManager := goja_runtime.NewManager(util.NewLogger(), 10)
	pool, err := runtimeManager.GetOrCreatePool(func() *goja.Runtime {
		initFn, err := extension_repo.SetupGojaExtensionVM(nil, extension.LanguageTypescript, util.NewLogger())
		if err != nil {
			return nil
		}
		return initFn()
	})
	require.NoError(t, err)

	vm, err := pool.Get(context.Background())
	require.NoError(t, err)
	return vm
}

func TestGojaDocument(t *testing.T) {
	vm := setupTestVM(t)
	defer vm.ClearInterrupt()

	tests := []struct {
		entry string
	}{
		{entry: "./goja_bindings/goja_doc_test/doc-example.ts"},
		{entry: "./goja_bindings/goja_doc_test/doc-example-2.ts"},
	}

	for _, tt := range tests {
		t.Run(tt.entry, func(t *testing.T) {
			fileB, err := os.ReadFile(tt.entry)
			require.NoError(t, err)

			now := time.Now()

			source, err := extension_repo.JSVMTypescriptToJS(string(fileB))
			require.NoError(t, err)

			_, err = vm.RunString(source)
			require.NoError(t, err)

			_, err = vm.RunString(`function NewProvider() { return new Provider() }`)
			require.NoError(t, err)

			newProviderFunc, ok := goja.AssertFunction(vm.Get("NewProvider"))
			require.True(t, ok)

			classObjVal, err := newProviderFunc(goja.Undefined())
			require.NoError(t, err)

			classObj := classObjVal.ToObject(vm)

			testFunc, ok := goja.AssertFunction(classObj.Get("test"))
			require.True(t, ok)

			ret, err := testFunc(classObj)
			require.NoError(t, err)

			promise := ret.Export().(*goja.Promise)

			for promise.State() == goja.PromiseStatePending {
				time.Sleep(10 * time.Millisecond)
			}

			if promise.State() == goja.PromiseStateFulfilled {
				t.Logf("Fulfilled: %v", promise.Result())
			} else {
				t.Fatalf("Rejected: %v", promise.Result())
			}

			fmt.Println(time.Since(now).Seconds())
		})
	}
}

func TestGojaFormData(t *testing.T) {
	vm := setupTestVM(t)
	defer vm.ClearInterrupt()

	_, err := vm.RunString(`
var fd = new FormData();
fd.append("name", "John Doe");
fd.append("age", 30);

console.log("Has 'name':", fd.has("name")); // true
console.log("Get 'name':", fd.get("name")); // John Doe
console.log("GetAll 'name':", fd.getAll("name")); // ["John Doe"]
console.log("Keys:", Array.from(fd.keys())); // ["name", "age"]
console.log("Values:", Array.from(fd.values())); // ["John Doe", 30]

fd.delete("name");
console.log("Has 'name' after delete:", fd.has("name")); // false

console.log("Entries:");
for (let entry of fd.entries()) {
	console.log(entry[0], entry[1]);
}

var contentType = fd.getContentType();
var buffer = fd.getBuffer();
console.log("Content-Type:", contentType);
console.log("Buffer:", buffer);
	`)
	require.NoError(t, err)
}

func TestGojaFormDataAndFetch(t *testing.T) {
	vm := setupTestVM(t)
	defer vm.ClearInterrupt()

	_, err := vm.RunString(`
async function run() {
	const formData = new FormData();
	formData.append("username", "John");
	formData.append("accountnum", 123456);
	
	console.log(formData.get("username")); // John

	const fData = new URLSearchParams();
	for (const pair of formData.entries()) {
		fData.append(pair[0], pair[1]);
	}
	
	const response = await fetch('https://httpbin.org/post', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/x-www-form-urlencoded',
		},
		body: fData
	});

	const data = await response.json();
	console.log(data);

	console.log("Echoed GojaFormData content:");
    if (data.form) {
        for (const key in data.form) {
            console.log(key, data.form[key]);
        }
    } else {
        console.log("No form data echoed in the response.");
    }

	return data;
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

	if promise.State() == goja.PromiseStateFulfilled {
		spew.Dump(promise.Result())
	} else {
		err := promise.Result()
		spew.Dump(err)
	}
}

func TestGojaCrypto(t *testing.T) {
	vm := setupTestVM(t)
	defer vm.ClearInterrupt()

	filepath := "./goja_bindings/goja_crypto_test/crypto-example.ts"
	fileB, err := os.ReadFile(filepath)
	require.NoError(t, err)

	_, err = vm.RunString(string(fileB))
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

func TestGojaTorrentUtils(t *testing.T) {
	vm := setupTestVM(t)
	defer vm.ClearInterrupt()

	filepath := "./goja_bindings/goja_torrent_test/torrent-utils-example.ts"
	fileB, err := os.ReadFile(filepath)
	require.NoError(t, err)

	_, err = vm.RunString(string(fileB))
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
