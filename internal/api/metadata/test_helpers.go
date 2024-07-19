package metadata

import (
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"testing"
)

func TestGetMockProvider(t *testing.T) *Provider {
	tempDir := t.TempDir()
	fileCacher, err := filecache.NewCacher(tempDir)
	if err != nil {
		t.Fatalf("could not create filecacher: %v", err)
	}

	metadataProvider := NewProvider(&NewProviderOptions{
		Logger:     util.NewLogger(),
		FileCacher: fileCacher,
	})

	return metadataProvider
}
