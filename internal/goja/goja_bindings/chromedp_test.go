package goja_bindings

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"seanime/internal/util"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupChromeVMWithServer(t *testing.T) (*goja.Runtime, *ChromeDP, *httptest.Server) {
	vm := goja.New()
	chrome := BindChromeDP(vm)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `
<!DOCTYPE html>
<html>
<head>
	<title>Test Page</title>
	<style>
		.loading { display: block; }
		.loaded { display: none; }
		.dynamic-content.ready .loading { display: none; }
		.dynamic-content.ready .loaded { display: block; }
	</style>
</head>
<body>
	<h1 id="title">Hello World</h1>
	<div class="content">
		<p id="description">This is a test page</p>
		<ul id="list">
			<li>Item 1</li>
			<li>Item 2</li>
			<li>Item 3</li>
		</ul>
	</div>
	
	<div id="dynamic-content" class="dynamic-content">
		<div class="loading">Loading...</div>
		<div class="loaded">
			<h2 id="dynamic-title">Dynamically loaded content</h2>
			<p id="dynamic-text">This content was loaded by JavaScript</p>
		</div>
	</div>
	
	<div id="counter">0</div>
	<button id="increment-btn">Increment</button>
	
	<div id="ajax-result"></div>
	
	<script>
		// Static data
		window.testData = {
			name: "Test",
			value: 42,
			items: [1, 2, 3]
		};
		
		// Simulate dynamic content loading
		setTimeout(function() {
			document.getElementById('dynamic-content').classList.add('ready');
			
			// Add more dynamic items
			var list = document.getElementById('list');
			var newItem = document.createElement('li');
			newItem.textContent = 'Dynamic Item 4';
			newItem.id = 'dynamic-item';
			list.appendChild(newItem);
		}, 100);
		
		// Counter functionality
		var counter = 0;
		var counterEl = document.getElementById('counter');
		var btnEl = document.getElementById('increment-btn');
		
		btnEl.addEventListener('click', function() {
			counter++;
			counterEl.textContent = counter;
		});
		
		// Simulate AJAX call
		setTimeout(function() {
			var ajaxResult = document.getElementById('ajax-result');
			ajaxResult.innerHTML = '<span id="ajax-data">API Data Loaded</span>';
		}, 150);
		
		// Make data available after a delay
		setTimeout(function() {
			window.loadedData = {
				status: 'ready',
				timestamp: Date.now(),
				items: ['A', 'B', 'C']
			};
		}, 200);
	</script>
</body>
</html>
`
		fmt.Fprint(w, html)
	}))

	return vm, chrome, server
}

// Helper to wait for promise resolution while processing response channel
func waitForPromise(t *testing.T, promise *goja.Promise, maxWait time.Duration) {
	timeout := time.After(maxWait)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatalf("Promise did not resolve within %v, state: %v", maxWait, promise.State())
		case <-ticker.C:
			if promise.State() != goja.PromiseStatePending {
				return
			}
		}
	}
}

func TestChromeDPScrape_Basic(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const html = await ChromeDP.scrape(%q);
			return html;
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	// Wait for promise to resolve
	promise := result.Export().(*goja.Promise)
	for promise.State() == goja.PromiseStatePending {
		time.Sleep(100 * time.Millisecond)
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())
	html := promise.Result().String()

	assert.Contains(t, html, "Hello World")
	assert.Contains(t, html, "Test Page")
	assert.Contains(t, html, "This is a test page")
}

func TestChromeDPScrape_WithOptions(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const html = await ChromeDP.scrape(%q, {
				timeout: 10,
				waitSelector: "#title",
				waitDuration: 100,
				userAgent: "TestBot/1.0"
			});
			return html;
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	for promise.State() == goja.PromiseStatePending {
		time.Sleep(100 * time.Millisecond)
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())
	html := promise.Result().String()

	assert.Contains(t, html, "Hello World")
}

func TestChromeDPScrape_WaitForDynamicContent(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const html = await ChromeDP.scrape(%q, {
				timeout: 10,
				waitSelector: "#dynamic-item",
				waitDuration: 200
			});
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

	assert.Contains(t, html, "Dynamic Item 4", "Should contain dynamically added item")
	assert.Contains(t, html, "dynamic-item", "Should contain the dynamic item ID")
}

func TestChromeDPScrape_InvalidURL(t *testing.T) {
	vm, chrome, _ := setupChromeVMWithServer(t)
	defer chrome.Close()

	jsCode := `
		(async () => {
			try {
				await ChromeDP.scrape("http://invalid-url-that-doesnt-exist-12345.com");
				return "should not reach here";
			} catch (e) {
				return e.toString();
			}
		})();
	`

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	for promise.State() == goja.PromiseStatePending {
		time.Sleep(100 * time.Millisecond)
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())
	errMsg := promise.Result().String()

	// Should contain some error indication
	assert.NotEqual(t, "should not reach here", errMsg)
	assert.Contains(t, errMsg, "GoError")
}

func TestChromeDPScrape_MissingArgument(t *testing.T) {
	vm, chrome, _ := setupChromeVMWithServer(t)
	defer chrome.Close()

	jsCode := `
		try {
			ChromeDP.scrape();
		} catch (e) {
			e.toString();
		}
	`

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	errMsg := result.String()
	assert.Contains(t, errMsg, "scrape requires at least a URL argument")
}

func TestChromeDPEvaluate_Basic(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const result = await ChromeDP.evaluate(%q, "document.title");
			return result;
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	for promise.State() == goja.PromiseStatePending {
		time.Sleep(100 * time.Millisecond)
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())
	title := promise.Result().String()

	assert.Equal(t, "Test Page", title)
}

func TestChromeDPEvaluate_ComplexData(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const result = await ChromeDP.evaluate(%q, "window.testData");
			return result;
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	for promise.State() == goja.PromiseStatePending {
		time.Sleep(100 * time.Millisecond)
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())

	data := promise.Result().ToObject(vm)
	assert.NotNil(t, data)

	name := data.Get("name")
	assert.Equal(t, "Test", name.String())

	value := data.Get("value")
	assert.Equal(t, int64(42), value.ToInteger())
}

func TestChromeDPEvaluate_DOMQuery(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const result = await ChromeDP.evaluate(%q, 
				"document.querySelector('#description').textContent"
			);
			return result;
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	for promise.State() == goja.PromiseStatePending {
		time.Sleep(100 * time.Millisecond)
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())
	text := promise.Result().String()

	assert.Equal(t, "This is a test page", text)
}

func TestChromeDPEvaluate_DOMManipulation(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	// Test that we can click a button and read the updated value
	jsCode := fmt.Sprintf(`
		(async () => {
			const result = await ChromeDP.evaluate(%q, 
				"document.getElementById('increment-btn').click(); document.getElementById('counter').textContent"
			);
			return result;
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	waitForPromise(t, promise, 10*time.Second)

	if promise.State() == goja.PromiseStateRejected {
		t.Fatalf("Promise was rejected: %v", promise.Result())
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())
	counter := promise.Result().String()

	assert.Equal(t, "1", counter, "Counter should be incremented to 1")
}

func TestChromeDPEvaluate_WaitForDynamicData(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	// Wait for dynamically loaded data
	jsCode := fmt.Sprintf(`
		(async () => {
			const result = await ChromeDP.evaluate(%q, 
				"window.loadedData",
				{
					waitDuration: 250
				}
			);
			return result;
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	waitForPromise(t, promise, 10*time.Second)

	if promise.State() == goja.PromiseStateRejected {
		t.Fatalf("Promise was rejected: %v", promise.Result())
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())

	data := promise.Result().ToObject(vm)
	assert.NotNil(t, data)

	status := data.Get("status")
	assert.Equal(t, "ready", status.String())

	items := data.Get("items").Export().([]interface{})
	assert.Len(t, items, 3)
	assert.Equal(t, "A", items[0])
	assert.Equal(t, "B", items[1])
	assert.Equal(t, "C", items[2])
}

func TestChromeDPEvaluate_AjaxContent(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	// Check for AJAX loaded content
	jsCode := fmt.Sprintf(`
		(async () => {
			const result = await ChromeDP.evaluate(%q, 
				"document.querySelector('#ajax-data')?.textContent",
				{
					waitDuration: 200
				}
			);
			return result;
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	waitForPromise(t, promise, 10*time.Second)

	if promise.State() == goja.PromiseStateRejected {
		t.Fatalf("Promise was rejected: %v", promise.Result())
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())
	text := promise.Result().String()

	assert.Equal(t, "API Data Loaded", text)
}

func TestChromeDPEvaluate_MissingArguments(t *testing.T) {
	vm, chrome, _ := setupChromeVMWithServer(t)
	defer chrome.Close()

	jsCode := `
		try {
			ChromeDP.evaluate("http://example.com");
		} catch (e) {
			e.toString();
		}
	`

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	errMsg := result.String()
	assert.Contains(t, errMsg, "evaluate requires a URL and a JavaScript expression")
}

func TestChromeDPScreenshot_Basic(t *testing.T) {
	t.Skip("Skipping screenshot test - requires headless browser environment")

	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const screenshot = await ChromeDP.screenshot(%q);
			return screenshot;
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	for promise.State() == goja.PromiseStatePending {
		time.Sleep(100 * time.Millisecond)
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())

	// Screenshot should return a byte array
	screenshot := promise.Result().Export()
	assert.NotNil(t, screenshot)
}

func TestChromeDPConcurrency(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const promises = [];
			for (let i = 0; i < 3; i++) {
				promises.push(ChromeDP.scrape(%q));
			}
			const results = await Promise.all(promises);
			return results.length;
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	for promise.State() == goja.PromiseStatePending {
		time.Sleep(100 * time.Millisecond)
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())
	count := promise.Result().ToInteger()

	assert.Equal(t, int64(3), count)
}

func TestChromeDPIntegration_ScrapeAndParse(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	// Also bind the Document parser
	err := BindDocument(vm)
	require.NoError(t, err)

	url := server.URL

	jsCode := fmt.Sprintf(`
		(async () => {
			const html = await ChromeDP.scrape(%q);
			const $ = LoadDoc(html);
			const title = $("#title").text();
			const items = [];
			$("#list li").each(function() {
				items.push(this.text());
			});
			return { title, items };
		})();
	`, url)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	waitForPromise(t, promise, 10*time.Second)

	if promise.State() == goja.PromiseStateRejected {
		t.Fatalf("Promise was rejected: %v", promise.Result())
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())

	data := promise.Result().ToObject(vm)
	assert.NotNil(t, data)

	title := data.Get("title").String()
	assert.Equal(t, "Hello World", title)

	items := data.Get("items").Export().([]interface{})
	assert.GreaterOrEqual(t, len(items), 3, "Should have at least 3 items")
	assert.Equal(t, "Item 1", items[0])
	assert.Equal(t, "Item 2", items[1])
	assert.Equal(t, "Item 3", items[2])
}

func TestChromeDPIntegration_ScrapeWithDynamicContent(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	err := BindDocument(vm)
	require.NoError(t, err)

	url := server.URL

	// Wait for dynamic content to be added
	jsCode := fmt.Sprintf(`
		(async () => {
			const html = await ChromeDP.scrape(%q, {
				waitSelector: "#dynamic-item",
				waitDuration: 250
			});
			const $ = LoadDoc(html);
			
			const title = $("#title").text();
			const items = [];
			$("#list li").each(function() {
				items.push(this.text());
			});
			
			const ajaxContent = $("#ajax-data").text();
			const dynamicTitle = $("#dynamic-title").text();
			
			return { title, items, ajaxContent, dynamicTitle };
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

	title := data.Get("title").String()
	assert.Equal(t, "Hello World", title)

	items := data.Get("items").Export().([]interface{})
	assert.Len(t, items, 4, "Should have 4 items including the dynamic one")
	assert.Equal(t, "Item 1", items[0])
	assert.Equal(t, "Item 2", items[1])
	assert.Equal(t, "Item 3", items[2])
	assert.Equal(t, "Dynamic Item 4", items[3])

	ajaxContent := data.Get("ajaxContent").String()
	assert.Equal(t, "API Data Loaded", ajaxContent)

	dynamicTitle := data.Get("dynamicTitle").String()
	assert.Equal(t, "Dynamically loaded content", dynamicTitle)
}

func TestChromeDPIntegration_EvaluateAndModifyDOM(t *testing.T) {
	vm, chrome, server := setupChromeVMWithServer(t)
	defer server.Close()
	defer chrome.Close()

	url := server.URL

	// Evaluate JavaScript that modifies the DOM and returns the result
	evalScript := `(function() {
		document.getElementById('title').textContent = 'Modified Title';
		var newDiv = document.createElement('div');
		newDiv.id = 'test-element';
		newDiv.textContent = 'Test Element Created';
		document.body.appendChild(newDiv);
		var btn = document.getElementById('increment-btn');
		for (var i = 0; i < 5; i++) {
			btn.click();
		}
		return {
			modifiedTitle: document.getElementById('title').textContent,
			counterValue: document.getElementById('counter').textContent,
			newElement: document.getElementById('test-element').textContent,
			listLength: document.querySelectorAll('#list li').length
		};
	})()`

	jsCode := fmt.Sprintf(`
		(async () => {
			const result = await ChromeDP.evaluate(%q, %q);
			return result;
		})();
	`, url, evalScript)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)
	waitForPromise(t, promise, 10*time.Second)

	if promise.State() == goja.PromiseStateRejected {
		t.Fatalf("Promise was rejected: %v", promise.Result())
	}

	require.Equal(t, goja.PromiseStateFulfilled, promise.State())

	data := promise.Result().ToObject(vm)
	assert.NotNil(t, data)

	modifiedTitle := data.Get("modifiedTitle").String()
	assert.Equal(t, "Modified Title", modifiedTitle)

	counterValue := data.Get("counterValue").String()
	assert.Equal(t, "5", counterValue)

	newElement := data.Get("newElement").String()
	assert.Equal(t, "Test Element Created", newElement)

	listLength := data.Get("listLength").ToInteger()
	assert.Equal(t, int64(3), listLength, "Should have 3 initial list items")
}

func TestChromeDPTimeout(t *testing.T) {
	vm, chrome, _ := setupChromeVMWithServer(t)
	defer chrome.Close()

	// Create a server that delays response
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		fmt.Fprint(w, "<html><body>Slow</body></html>")
	}))
	defer slowServer.Close()

	jsCode := fmt.Sprintf(`
		(async () => {
			try {
				await ChromeDP.scrape(%q, { timeout: 1 });
				return "should timeout";
			} catch (e) {
				return "timeout error";
			}
		})();
	`, slowServer.URL)

	result, err := vm.RunString(jsCode)
	require.NoError(t, err)

	promise := result.Export().(*goja.Promise)

	// Wait for timeout
	timeout := time.After(3 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Test timeout")
		case <-ticker.C:
			if promise.State() != goja.PromiseStatePending {
				goto done
			}
		}
	}

done:
	require.Equal(t, goja.PromiseStateFulfilled, promise.State())
	msg := promise.Result().String()
	assert.Equal(t, "timeout error", msg)
}

func BenchmarkChromeDPScrape(b *testing.B) {
	vm := goja.New()
	chrome := BindChromeDP(vm)
	defer chrome.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "<html><body><h1>Benchmark</h1></body></html>")
	}))
	defer server.Close()

	url := server.URL
	logger := util.NewLogger()
	logger.Info().Msgf("Running benchmark with URL: %s", url)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsCode := fmt.Sprintf(`
			(async () => {
				return await ChromeDP.scrape(%q);
			})();
		`, url)

		result, err := vm.RunString(jsCode)
		if err != nil {
			b.Fatal(err)
		}

		promise := result.Export().(*goja.Promise)
		for promise.State() == goja.PromiseStatePending {
			time.Sleep(10 * time.Millisecond)
		}
	}
}
