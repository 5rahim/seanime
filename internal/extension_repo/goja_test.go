package extension_repo_test

import (
	"fmt"
	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
	"github.com/davecgh/go-spew/spew"
	"github.com/dop251/goja"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/require"
	"os"
	"seanime/internal/extension"
	"seanime/internal/extension_repo"
	"seanime/internal/util"
	"testing"
	"time"
)

func TestGojaWithExtension(t *testing.T) {
	// Get the script
	filepath := "./goja_manga_test/my-manga-provider.ts"
	fileB, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatal(err)
	}

	ext := &extension.Extension{
		ID:          "my-manga-provider",
		Name:        "MyMangaProvider",
		Version:     "0.1.0",
		ManifestURI: "",
		Language:    extension.LanguageTypescript,
		Type:        extension.TypeMangaProvider,
		Description: "",
		Author:      "",
		Payload:     string(fileB),
	}

	// Create the provider
	provider, _, err := extension_repo.NewGojaMangaProvider(ext, ext.Language, util.NewLogger())
	require.NoError(t, err)

	// Test the search function
	searchResult, err := provider.Search(hibikemanga.SearchOptions{Query: "dandadan"})
	require.NoError(t, err)

	spew.Dump(searchResult)

	// Should have a result with rating of 1
	var dandadanRes *hibikemanga.SearchResult
	for _, res := range searchResult {
		if res.SearchRating == 1 {
			dandadanRes = res
			break
		}
	}
	require.NotNil(t, dandadanRes)
	spew.Dump(dandadanRes)

	// Test the search function again
	searchResult, err = provider.Search(hibikemanga.SearchOptions{Query: "boku no kokoro no yaibai"})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(searchResult), 1)

	t.Logf("Search results: %d", len(searchResult))

	// Test the findChapters function
	chapters, err := provider.FindChapters("pYN47sZm") // Boku no Kokoro no Yabai Yatsu
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(chapters), 100)

	t.Logf("Chapters: %d", len(chapters))

	// Test the findChapterPages function
	pages, err := provider.FindChapterPages("WLxnx") // Boku no Kokoro no Yabai Yatsu - Chapter 1
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(pages), 10)

	for _, page := range pages {
		t.Logf("Page: %s, Index: %d\n", page.URL, page.Index)
	}
}

func TestGojaCode(t *testing.T) {

	// VM
	vm, err := extension_repo.CreateJSVM(util.NewLogger())
	require.NoError(t, err)

	// Get the script
	filepath := "./goja_manga_test/my-manga-provider.ts"
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

	// Test the search function
	searchFunc, ok := goja.AssertFunction(classObj.Get("search"))
	require.True(t, ok)

	searchOpts := hibikemanga.SearchOptions{
		Query: "dandadan",
		Year:  0,
	}

	marshaledSearchOpts, err := json.Marshal(searchOpts)
	require.NoError(t, err)
	var searchData map[string]interface{}
	err = json.Unmarshal(marshaledSearchOpts, &searchData)
	require.NoError(t, err)

	// Call the search function
	searchResult, err := searchFunc(classObj, vm.ToValue(searchData))
	require.NoError(t, err)

	promise := searchResult.Export().(*goja.Promise)

	for promise.State() == goja.PromiseStatePending {
		time.Sleep(10 * time.Millisecond)
	}

	if promise.State() == goja.PromiseStateFulfilled {
		//spew.Dump(promise.Result())
		var res []*hibikemanga.SearchResult

		retValue := promise.Result()
		retValueCast, ok := retValue.Export().(interface{})
		require.True(t, ok)

		marshaled, err := json.Marshal(retValueCast)
		require.NoError(t, err)

		err = json.Unmarshal(marshaled, &res)
		require.NoError(t, err)

		for _, r := range res {
			t.Logf("Title: %s, Search Rating: %.2f\n", r.Title, r.SearchRating)
		}
	} else {
		err := promise.Result()
		t.Fatal(err)
	}

	fmt.Println(time.Since(now).Seconds())
}

func TestGojaFormData(t *testing.T) {

	// VM
	vm, err := extension_repo.CreateJSVM(util.NewLogger())
	require.NoError(t, err)

	_, err = vm.RunString(`
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

	// VM
	vm, err := extension_repo.CreateJSVM(util.NewLogger())
	require.NoError(t, err)

	_, err = vm.RunString(`
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

	// Get the function
	runFunc, ok := goja.AssertFunction(vm.Get("run"))
	require.True(t, ok)

	// Call the function
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
