package extension_repo

import (
	"context"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"

	"seanime/internal/api/anilist"
	"seanime/internal/extension"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/hook"
	"seanime/internal/platforms/anilist_platform"
)

func TestNewGojaPlugin(t *testing.T) {

	// create a dummy extension with a simple JS payload
	ext := &extension.Extension{
		ID:      "dummy-plugin",
		Payload: "onGetBaseAnime((e) => { console.log(e.Anime); e.next(); })", // log the anime object
	}

	// use 'javascript' as the language
	lang := extension.Language("javascript")
	// Create a pointer logger
	nLogger := zerolog.Nop()
	logger := &nLogger

	// Create a runtime manager with a pool size of 1
	runtimeManager := goja_runtime.NewManager(logger, 1)

	// Create a hook manager with the no-op logger
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})

	// Create the GojaPlugin first so hooks are registered
	plugin, err := NewGojaPlugin(ext, lang, logger, runtimeManager, hm)
	if err != nil {
		t.Fatalf("NewGojaPlugin returned error: %v", err)
	}
	if plugin == nil {
		t.Fatal("NewGojaPlugin returned nil plugin")
	}

	// Create AniList platform after plugin is setup
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilist.NewAnilistClient(""), logger, hm)

	// Retrieve a runtime from the plugin's pool
	vm, err := plugin.pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Failed to get runtime from pool: %v", err)
	}
	defer plugin.pool.Put(vm)

	// Create a channel to wait for the hook to be triggered
	done := make(chan struct{})

	// Add a hook handler that will close the channel when called
	hm.OnGetBaseAnime().BindFunc(func(e hook.Resolver) error {
		close(done)
		return e.Next()
	})

	// Call GetAnime which should trigger the hook
	go anilistPlatform.GetAnime(21)

	// Wait for the hook to be triggered or timeout
	select {
	case <-done:
		// Hook was triggered successfully
	case <-time.After(2 * time.Second):
		t.Fatal("Hook was not triggered within timeout")
	}

	// Check that hook functions were bound.
	// The hooksBinds function registers methods from the hook manager (e.g. OnGetBaseAnime).
	// Depending on FieldMapper implementation, the property name might be 'OnGetBaseAnime' or 'onGetBaseAnime'.
	var hookFn goja.Value
	hookFn = vm.Get("OnGetBaseAnime")
	if goja.IsUndefined(hookFn) || hookFn == nil {
		hookFn = vm.Get("onGetBaseAnime")
	}
	if goja.IsUndefined(hookFn) || hookFn == nil {
		t.Error("Expected hook function for OnGetBaseAnime to be defined in the runtime")
	}
}
