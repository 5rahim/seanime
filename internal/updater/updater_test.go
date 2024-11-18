package updater

import (
	"github.com/stretchr/testify/require"
	"seanime/internal/util"
	"testing"
)

func TestUpdater_GetLatestUpdate(t *testing.T) {

	docsUrl = "https://seanime.rahim.app/api/releases" // simulate dead endpoint

	u := New("2.0.2", util.NewLogger(), nil)

	update, err := u.GetLatestUpdate()
	require.NoError(t, err)

	require.NotNil(t, update)

	util.Spew(update)
}
