package goja_bindings

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"seanime/internal/util"
	"sync"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/dop251/goja"
	gojabuffer "github.com/dop251/goja_nodejs/buffer"
	gojarequire "github.com/dop251/goja_nodejs/require"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetch_ThreadSafety(t *testing.T) {
	// Create a test server that simulates different response times
	var serverRequestCount int
	var serverMu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverMu.Lock()
		serverRequestCount++
		currentRequest := serverRequestCount
		serverMu.Unlock()

		// Simulate varying response times to increase chance of race conditions
		time.Sleep(time.Duration(currentRequest%3) * 50 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"request": %d}`, currentRequest)
	}))
	defer server.Close()

	// Create JavaScript test code that makes concurrent fetch calls
	jsCode := fmt.Sprintf(`
		const url = %q;
		const promises = [];
		
		// Function to make a fetch request and verify response
		async function makeFetch(i) {
			const response = await fetch(url);
			const data = await response.json();
			return { index: i, data };
		}

		// Create multiple concurrent requests
		for (let i = 0; i < 50; i++) {
			promises.push(makeFetch(i));
		}

		// Wait for all requests to complete
		Promise.all(promises)
	`, server.URL)

	// Run the code multiple times to increase chance of catching race conditions
	for i := 0; i < 5; i++ {
		t.Run(fmt.Sprintf("Iteration_%d", i), func(t *testing.T) {
			// Create a new VM for each iteration
			vm := goja.New()
			BindFetch(vm)

			// Execute the JavaScript code
			v, err := vm.RunString(jsCode)
			assert.NoError(t, err)

			// Get the Promise
			promise, ok := v.Export().(*goja.Promise)
			assert.True(t, ok)

			// Wait for the Promise to resolve
			for promise.State() == goja.PromiseStatePending {
				time.Sleep(10 * time.Millisecond)
			}

			// Verify the Promise resolved successfully
			assert.Equal(t, goja.PromiseStateFulfilled, promise.State())

			// Verify we got an array of results
			results, ok := promise.Result().Export().([]interface{})
			assert.True(t, ok)
			assert.Len(t, results, 50)

			// Verify each result has the expected structure
			for _, result := range results {
				resultMap, ok := result.(map[string]interface{})
				assert.True(t, ok)
				assert.Contains(t, resultMap, "index")
				assert.Contains(t, resultMap, "data")

				data, ok := resultMap["data"].(map[string]interface{})
				assert.True(t, ok)
				assert.Contains(t, data, "request")
			}
		})
	}
}

func TestFetch_VMIsolation(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"test": "data"}`)
	}))
	defer server.Close()

	// Create multiple VMs and make concurrent requests
	const numVMs = 5
	const requestsPerVM = 40

	var wg sync.WaitGroup
	for i := 0; i < numVMs; i++ {
		wg.Add(1)
		go func(vmIndex int) {
			defer wg.Done()

			// Create a new VM for this goroutine
			vm := goja.New()
			BindFetch(vm)

			// Create JavaScript code that makes multiple requests
			jsCode := fmt.Sprintf(`
				const url = %q;
				const promises = [];
				
				for (let i = 0; i < %d; i++) {
					promises.push(fetch(url).then(r => r.json()));
				}

				Promise.all(promises)
			`, server.URL, requestsPerVM)

			// Execute the code
			v, err := vm.RunString(jsCode)
			assert.NoError(t, err)

			// Get and wait for the Promise
			promise := v.Export().(*goja.Promise)
			for promise.State() == goja.PromiseStatePending {
				time.Sleep(10 * time.Millisecond)
			}

			// Verify the Promise resolved successfully
			assert.Equal(t, goja.PromiseStateFulfilled, promise.State())

			// Verify we got the expected number of results
			results := promise.Result().Export().([]interface{})
			assert.Len(t, results, requestsPerVM)
		}(i)
	}

	wg.Wait()
}

func TestGojaPromiseAll(t *testing.T) {
	vm := goja.New()

	BindFetch(vm)

	registry := new(gojarequire.Registry)
	registry.Enable(vm)
	gojabuffer.Enable(vm)
	BindConsole(vm, util.NewLogger())

	_, err := vm.RunString(`
	async function run() {
		const [a, b, c] = await Promise.all([
			fetch("https://jsonplaceholder.typicode.com/todos/1"),
			fetch("https://jsonplaceholder.typicode.com/todos/2"),
			fetch("https://jsonplaceholder.typicode.com/todos/3"),
			fetch("https://jsonplaceholder.typicode.com/todos/4"),
			fetch("https://jsonplaceholder.typicode.com/todos/5"),
			fetch("https://jsonplaceholder.typicode.com/todos/6"),
			fetch("https://jsonplaceholder.typicode.com/todos/7"),
			fetch("https://jsonplaceholder.typicode.com/todos/8"),
		])

		const dataA = await a.json();
		const dataB = await b.json();
		const dataC = await c.json();

		console.log("Data A:", dataA.title);
		console.log("Data B:", dataB);
		console.log("Data C:", dataC);
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
}

func TestGojaFormDataAndFetch(t *testing.T) {
	vm := goja.New()
	BindFetch(vm)

	registry := new(gojarequire.Registry)
	registry.Enable(vm)
	gojabuffer.Enable(vm)
	BindConsole(vm, util.NewLogger())

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
		body: formData
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

func TestGojaFetchPostJSON(t *testing.T) {
	vm := goja.New()

	BindFetch(vm)

	registry := new(gojarequire.Registry)
	registry.Enable(vm)
	gojabuffer.Enable(vm)
	BindConsole(vm, util.NewLogger())

	_, err := vm.RunString(`
async function run() {
	const response = await fetch('https://httpbin.org/post', {
		method: 'POST',
		body: { name: "John Doe", age: 30 },
	});

	const data = await response.json();
	console.log(data);

	console.log("Echoed content:");
    if (data.json) {
        for (const key in data.json) {
            console.log(key, data.json[key]);
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
