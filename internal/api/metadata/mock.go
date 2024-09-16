package metadata

import (
	"github.com/stretchr/testify/require"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"testing"
)

func GetMockProvider(t *testing.T) Provider {
	filecacher, err := filecache.NewCacher(t.TempDir())
	require.NoError(t, err)
	return NewProvider(&NewProviderImplOptions{
		Logger:     util.NewLogger(),
		FileCacher: filecacher,
	})
}
