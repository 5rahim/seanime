package filecache

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStoreSetAndGet(t *testing.T) {
	store, _ := NewStore(DefaultOptions)
	key := "key"
	value := struct {
		Name string
	}{
		Name: "value",
	}
	store.Set(key, value)

	var out struct {
		Name string
	}
	found, _ := store.Get(key, &out)
	if !found || !assert.Equal(t, value, out) {
		t.Errorf("Failed to get the correct value. Expected %v, got %v", value, out)
	}

	spew.Dump(out)

	store.Close()

	store2, _ := NewStore(DefaultOptions)
	found, _ = store2.Get(key, &out)
	if !found || out != value {
		t.Errorf("Failed to get the correct value. Expected %v, got %v", value, out)
	}

	store2.Close()

	// Output:
	spew.Dump(out)
}
