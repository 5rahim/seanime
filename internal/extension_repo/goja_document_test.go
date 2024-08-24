package extension_repo_test

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
	"os"
	"seanime/internal/extension_repo"
	"seanime/internal/util"
	"testing"
	"time"
)

func TestGojaDocument(t *testing.T) {

	// VM
	vm, err := extension_repo.CreateJSVM(util.NewLogger())
	require.NoError(t, err)

	// Get the script
	filepath := "./goja_doc_test/doc-example.ts"
	fileB, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()

	// Convert the typescript to javascript
	source, err := extension_repo.JSVMTypescriptToJS(string(fileB))
	require.NoError(t, err)

	// Run the program on the VM
	_, err = vm.RunString(source)
	require.NoError(t, err)

	_, err = vm.RunString(`function NewProvider() {
    return new Provider()
}`)
	require.NoError(t, err)

	newProviderFunc, ok := goja.AssertFunction(vm.Get("NewProvider"))
	require.True(t, ok)

	// Create the provider
	classObjVal, err := newProviderFunc(goja.Undefined())
	require.NoError(t, err)

	classObj := classObjVal.ToObject(vm)

	testFunc, ok := goja.AssertFunction(classObj.Get("test"))
	require.True(t, ok)
	_, err = testFunc(classObj)
	require.NoError(t, err)

	fmt.Println(time.Since(now).Seconds())

}
