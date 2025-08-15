package metadata

import (
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"testing"

	"github.com/stretchr/testify/require"
)

func GetMockProvider(t *testing.T) Provider {
	filecacher, err := filecache.NewCacher(t.TempDir())
	require.NoError(t, err)
	return NewProvider(&NewProviderImplOptions{
		Logger:     util.NewLogger(),
		FileCacher: filecacher,
	})
}
