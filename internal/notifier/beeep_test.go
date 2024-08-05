package notifier

import (
	"github.com/gen2brain/beeep"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"seanime/internal/test_utils"
	"testing"
)

func TestBeeep(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t)

	err := beeep.Notify("Seanime", "Downloaded 1 episode", filepath.Join(test_utils.ConfigData.Path.DataDir, "logo.png"))
	require.NoError(t, err)

}
