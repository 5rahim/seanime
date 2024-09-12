package filecache

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"seanime/internal/test_utils"
	"sync"
	"testing"
	"time"
)

func TestCacherFunctions(t *testing.T) {
	test_utils.InitTestProvider(t)

	tempDir := t.TempDir()
	t.Log(tempDir)

	cacher, err := NewCacher(filepath.Join(tempDir, "cache"))
	require.NoError(t, err)

	bucket := Bucket{
		name: "test",
		ttl:  10 * time.Second,
	}

	keys := []string{"key1", "key2", "key3"}

	type valStruct = struct {
		Name string
	}

	values := []*valStruct{
		{
			Name: "value1",
		},
		{
			Name: "value2",
		},
		{
			Name: "value3",
		},
	}

	for i, key := range keys {
		err = cacher.Set(bucket, key, values[i])
		if err != nil {
			t.Fatalf("Failed to set the value: %v", err)
		}
	}

	allVals, err := GetAll[*valStruct](cacher, bucket)
	if err != nil {
		t.Fatalf("Failed to get all values: %v", err)
	}

	if len(allVals) != len(keys) {
		t.Fatalf("Failed to get all values: expected %d, got %d", len(keys), len(allVals))
	}

	spew.Dump(allVals)
}

func TestCacherSetAndGet(t *testing.T) {
	test_utils.InitTestProvider(t)

	tempDir := t.TempDir()
	t.Log(tempDir)

	cacher, err := NewCacher(filepath.Join(test_utils.ConfigData.Path.DataDir, "cache"))

	bucket := Bucket{
		name: "test",
		ttl:  4 * time.Second,
	}
	key := "key"
	value := struct {
		Name string
	}{
		Name: "value",
	}
	// Add "key" -> value to the bucket, with a TTL of 4 seconds
	err = cacher.Set(bucket, key, value)
	if err != nil {
		t.Fatalf("Failed to set the value: %v", err)
	}

	var out struct {
		Name string
	}
	// Get the value of "key" from the bucket, it shouldn't be expired
	found, err := cacher.Get(bucket, key, &out)
	if err != nil {
		t.Errorf("Failed to get the value: %v", err)
	}
	if !found || !assert.Equal(t, value, out) {
		t.Errorf("Failed to get the correct value. Expected %v, got %v", value, out)
	}

	spew.Dump(out)

	time.Sleep(3 * time.Second)

	// Get the value of "key" from the bucket again, it shouldn't be expired
	found, err = cacher.Get(bucket, key, &out)
	if !found {
		t.Errorf("Failed to get the value")
	}
	if !found || out != value {
		t.Errorf("Failed to get the correct value. Expected %v, got %v", value, out)
	}

	spew.Dump(out)

	// Spin up a goroutine to set "key2" -> value2 to the bucket, with a TTL of 1 second
	// cacher should be thread-safe
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		key2 := "key2"
		value2 := struct {
			Name string
		}{
			Name: "value2",
		}
		var out2 struct {
			Name string
		}
		err = cacher.Set(bucket, key2, value2)
		if err != nil {
			t.Errorf("Failed to set the value: %v", err)
		}

		found, err = cacher.Get(bucket, key2, &out2)
		if err != nil {
			t.Errorf("Failed to get the value: %v", err)
		}

		if !found || !assert.Equal(t, value2, out2) {
			t.Errorf("Failed to get the correct value. Expected %v, got %v", value2, out2)
		}

		_ = cacher.Delete(bucket, key2)

		spew.Dump(out2)

	}()

	time.Sleep(2 * time.Second)

	// Get the value of "key" from the bucket, it should be expired
	found, _ = cacher.Get(bucket, key, &out)
	if found {
		t.Errorf("Failed to delete the value")
		spew.Dump(out)
	}

	wg.Wait()

}
