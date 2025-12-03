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
