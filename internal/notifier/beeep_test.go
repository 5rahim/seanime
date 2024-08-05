package notifier

import (
	"github.com/gen2brain/beeep"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBeeep(t *testing.T) {

	err := beeep.Notify("Seanime", "Downloaded 1 episode", "")
	require.NoError(t, err)

}
