package updater

import (
	"seanime/internal/util"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdater_GetLatestUpdate(t *testing.T) {

	fallbackGithubUrl = "https://seanime.app/api/releases" // simulate dead endpoint

	u := New("2.0.2", util.NewLogger(), nil)

	update, err := u.GetLatestUpdate()
	require.NoError(t, err)

	require.NotNil(t, update)

	util.Spew(update)
}
