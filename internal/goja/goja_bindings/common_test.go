package goja_bindings_test

import (
	"errors"
	"fmt"
	"os"
	"seanime/internal/extension_repo"
	"seanime/internal/util"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestVM(t *testing.T) *goja.Runtime {
	vm := goja.New()
	vm.SetParserOptions(parser.WithDisableSourceMaps)
	// Bind the shared bindings
	extension_repo.ShareBinds(vm, util.NewLogger())
	fm := extension_repo.FieldMapper{}
	vm.SetFieldNameMapper(fm)
	return vm
}

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

func TestDivideFunction(t *testing.T) {
	vm := goja.New()
	vm.Set("divide", divide)

	// Case 1: Successful division
	result, err := vm.RunString("divide(10, 3);")
	assert.NoError(t, err)
	assert.Equal(t, 3.3333333333333335, result.Export())

	// Case 2: Division by zero should throw an exception
	_, err = vm.RunString("divide(10, 0);")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "division by zero")

	// Case 3: Handling error with try-catch in JS
	result, err = vm.RunString(`
		try {
			divide(10, 0);
		} catch (e) {
			e.toString();
		}
	`)
	assert.NoError(t, err)
	assert.Equal(t, "GoError: division by zero", result.Export())
}

func multipleReturns() (int, string, float64) {
	return 42, "hello", 3.14
}

func TestMultipleReturns(t *testing.T) {
	vm := goja.New()
	vm.Set("multiReturn", multipleReturns)

	v, err := vm.RunString("multiReturn();")
	assert.NoError(t, err)
	util.Spew(v.Export())
}

func TestByteSliceToUint8Array(t *testing.T) {
	// Initialize a new Goja VM
	vm := goja.New()

	// Create a Go byte slice
	data := []byte("hello")

	// Set the byte slice in the Goja VM
	vm.Set("data", data)

	extension_repo.ShareBinds(vm, util.NewLogger())

	// JavaScript code to verify the type and contents of 'data'
	jsCode := `
        console.log(typeof data, data);

		const dataArrayBuffer = new ArrayBuffer(5);
		const uint8Array = new Uint8Array(dataArrayBuffer);
		uint8Array[0] = 104;
		uint8Array[1] = 101;
		uint8Array[2] = 108;
		uint8Array[3] = 108;
		uint8Array[4] = 111;
		console.log(typeof uint8Array, uint8Array);

		console.log("toString", $toString(uint8Array));
		console.log("toString", uint8Array.toString());


        true; // Return true if all checks pass
    `

	// Run the JavaScript code in the Goja VM
	result, err := vm.RunString(jsCode)
	if err != nil {
		t.Fatalf("JavaScript error: %v", err)
	}

	// Assert that the result is true
	assert.Equal(t, true, result.Export())
}

func TestGojaDocument(t *testing.T) {
	vm := setupTestVM(t)
	defer vm.ClearInterrupt()

	tests := []struct {
		entry string
	}{
		{entry: "./js/test/doc-example.ts"},
		{entry: "./js/test/doc-example-2.ts"},
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

func TestOptionalParams(t *testing.T) {
	vm := setupTestVM(t)
	defer vm.ClearInterrupt()

	type Options struct {
		Add int `json:"add"`
	}

	vm.Set("test", func(a int, opts Options) int {
		fmt.Println("opts", opts)
		return a + opts.Add
	})

	vm.RunString(`
		const result = test(1);
		console.log(result);

		const result2 = test(1, { add: 10 });
		console.log(result2);
	`)
}
