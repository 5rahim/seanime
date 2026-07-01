package goja_bindings

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrowserAPI_NavigateAndGetHTML(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const browser = await ChromeDP.newBrowser();
			await browser.navigate(%q);
			const html = await browser.outerHTML("html");
			await browser.close();
			return html;
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	waitForPromise(t, promise, 15*time.Second)

	if promise.State() == goja.PromiseStateRejected {
		t.Fatalf("Promise was rejected: %v", promise.Result())
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())
	html := promise.Result().String()

	assert.Contains(t, html, "Hello World")
	assert.Contains(t, html, "Test Page")
}

func TestBrowserAPI_ClickAndRead(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const browser = await ChromeDP.newBrowser();
			await browser.navigate(%q);
			
			// Click the increment button 3 times
			await browser.click("#increment-btn");
			await browser.click("#increment-btn");
			await browser.click("#increment-btn");
			
			// Read the counter value
			const counterValue = await browser.text("#counter");
			await browser.close();
			
			return counterValue;
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	waitForPromise(t, promise, 15*time.Second)

	if promise.State() == goja.PromiseStateRejected {
		t.Fatalf("Promise was rejected: %v", promise.Result())
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())
	counter := promise.Result().String()

	assert.Equal(t, "3", counter)
}

func TestBrowserAPI_WaitForDynamicContent(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const browser = await ChromeDP.newBrowser();
			await browser.navigate(%q);
			
			// Wait for the dynamic item to appear
			await browser.waitVisible("#dynamic-item");
			
			// Get the text of the dynamic item
			const itemText = await browser.text("#dynamic-item");
			await browser.close();
			
			return itemText;
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	waitForPromise(t, promise, 15*time.Second)

	if promise.State() == goja.PromiseStateRejected {
		t.Fatalf("Promise was rejected: %v", promise.Result())
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())
	text := promise.Result().String()

	assert.Equal(t, "Dynamic Item 4", text)
}

func TestBrowserAPI_EvaluateJS(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const browser = await ChromeDP.newBrowser();
			await browser.navigate(%q);
			
			// Wait for dynamic data to load
			await browser.sleep(250);
			
			// Evaluate JavaScript to get the window object
			const data = await browser.evaluate("window.loadedData");
			await browser.close();
			
			return data;
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	waitForPromise(t, promise, 15*time.Second)

	if promise.State() == goja.PromiseStateRejected {
		t.Fatalf("Promise was rejected: %v", promise.Result())
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())

	data := promise.Result().ToObject(vm)
	assert.NotNil(t, data)

	status := data.Get("status").String()
	assert.Equal(t, "ready", status)

	items := data.Get("items").Export().([]interface{})
	assert.Len(t, items, 3)
	assert.Equal(t, "A", items[0])
}

func TestBrowserAPI_GetAttribute(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const browser = await ChromeDP.newBrowser();
			await browser.navigate(%q);
			
			// Wait for dynamic item
			await browser.waitVisible("#dynamic-item");
			
			// Get the ID attribute
			const id = await browser.attribute("#dynamic-item", "id");
			await browser.close();
			
			return id;
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	waitForPromise(t, promise, 15*time.Second)

	if promise.State() == goja.PromiseStateRejected {
		t.Fatalf("Promise was rejected: %v", promise.Result())
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())
	id := promise.Result().String()

	assert.Equal(t, "dynamic-item", id)
}

func TestBrowserAPI_ComplexInteraction(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	formServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `
<!DOCTYPE html>
<html>
<head><title>Form Test</title></head>
<body>
	<form id="test-form">
		<input type="text" id="username" name="username" />
		<input type="password" id="password" name="password" />
		<button type="submit" id="submit-btn">Submit</button>
	</form>
	<div id="result"></div>
	<script>
		document.getElementById('test-form').addEventListener('submit', function(e) {
			e.preventDefault();
			const username = document.getElementById('username').value;
			const password = document.getElementById('password').value;
			document.getElementById('result').textContent = 'Submitted: ' + username + ' / ' + password;
		});
	</script>
</body>
</html>
`
		fmt.Fprint(w, html)
	}))
	defer formServer.Close()

	url := formServer.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const browser = await ChromeDP.newBrowser();
			await browser.navigate(%q);
			
			// Wait for form to be ready
			await browser.waitReady("#test-form");
			
			// Fill in the form
			await browser.sendKeys("#username", "testuser");
			await browser.sendKeys("#password", "testpass");
			
			// Click submit
			await browser.click("#submit-btn");
			
			// Wait a bit for the form handler to run
			await browser.sleep(100);
			
			// Read the result
			const result = await browser.text("#result");
			await browser.close();
			
			return result;
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	waitForPromise(t, promise, 15*time.Second)

	if promise.State() == goja.PromiseStateRejected {
		t.Fatalf("Promise was rejected: %v", promise.Result())
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())
	resultText := promise.Result().String()

	assert.Contains(t, resultText, "testuser")
	assert.Contains(t, resultText, "testpass")
}

func TestBrowserAPI_MultipleBrowsers(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			// Create two browsers
			const browser1 = await ChromeDP.newBrowser();
			const browser2 = await ChromeDP.newBrowser();
			
			// Navigate both
			await browser1.navigate(%q);
			await browser2.navigate(%q);
			
			// Click button different numbers of times
			await browser1.click("#increment-btn");
			await browser1.click("#increment-btn");
			
			await browser2.click("#increment-btn");
			await browser2.click("#increment-btn");
			await browser2.click("#increment-btn");
			await browser2.click("#increment-btn");
			
			// Read counters
			const counter1 = await browser1.text("#counter");
			const counter2 = await browser2.text("#counter");
			
			await browser1.close();
			await browser2.close();
			
			return { counter1, counter2 };
		})();
	`, url, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	waitForPromise(t, promise, 20*time.Second)

	if promise.State() == goja.PromiseStateRejected {
		t.Fatalf("Promise was rejected: %v", promise.Result())
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())

	data := promise.Result().ToObject(vm)
	counter1 := data.Get("counter1").String()
	counter2 := data.Get("counter2").String()

	assert.Equal(t, "2", counter1)
	assert.Equal(t, "4", counter2)
}

func TestBrowserAPI_ReuseAndNavigate(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const browser = await ChromeDP.newBrowser();
			
			// Navigate to page and click once
			await browser.navigate(%q);
			await browser.click("#increment-btn");
			const counter1 = await browser.text("#counter");
			
			// Navigate to page again (fresh state)
			await browser.navigate(%q);
			await browser.click("#increment-btn");
			await browser.click("#increment-btn");
			const counter2 = await browser.text("#counter");
			
			await browser.close();
			
			return { counter1, counter2 };
		})();
	`, url, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	waitForPromise(t, promise, 20*time.Second)

	if promise.State() == goja.PromiseStateRejected {
		t.Fatalf("Promise was rejected: %v", promise.Result())
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())

	data := promise.Result().ToObject(vm)
	counter1 := data.Get("counter1").String()
	counter2 := data.Get("counter2").String()

	assert.Equal(t, "1", counter1)
	assert.Equal(t, "2", counter2)
}

func TestBrowserAPI_ExecuteCDP(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const browser = await ChromeDP.newBrowser();
			await browser.navigate(%q);
			
			// Execute a CDP command to enable network domain
			await browser.executeCDP("Network.enable", {});
			
			await browser.close();
			return "success";
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	waitForPromise(t, promise, 15*time.Second)

	if promise.State() == goja.PromiseStateRejected {
		t.Fatalf("Promise was rejected: %v", promise.Result())
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())
	val := promise.Result().String()
	assert.Equal(t, "success", val)
}

func TestBrowserAPI_ListenTarget_Network(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const browser = await ChromeDP.newBrowser();
			
			const events = [];
			
			// Setup listener before enabling network and navigating
			browser.listenTarget((ev) => {
				if (ev.method === "Network.requestWillBeSent" || ev.method === "Network.responseReceived") {
					events.push(ev);
				}
			});
			
			await browser.executeCDP("Network.enable", {});
			await browser.navigate(%q);
			
			// Sleep a bit to allow asynchronous events to be processed
			await browser.sleep(500);
			await browser.close();
			
			return events;
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	waitForPromise(t, promise, 15*time.Second)

	if promise.State() == goja.PromiseStateRejected {
		t.Fatalf("Promise was rejected: %v", promise.Result())
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())

	eventsObj := promise.Result().ToObject(vm)
	require.NotNil(t, eventsObj)

	// Convert array of events to Go slice
	var events []map[string]interface{}
	err = vm.ExportTo(eventsObj, &events)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(events), 1, "Should capture at least 1 network event")

	foundRequest := false
	foundResponse := false
	for _, ev := range events {
		method := ev["method"].(string)
		params := ev["params"].(map[string]interface{})

		if method == "Network.requestWillBeSent" {
			foundRequest = true
			req := params["request"].(map[string]interface{})
			assert.Contains(t, req["url"].(string), url)
		} else if method == "Network.responseReceived" {
			foundResponse = true
			resp := params["response"].(map[string]interface{})
			assert.Contains(t, resp["url"].(string), url)
		}
	}

	assert.True(t, foundRequest, "Should have received Network.requestWillBeSent event")
	assert.True(t, foundResponse, "Should have received Network.responseReceived event")
}
