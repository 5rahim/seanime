package builtin_client

import (
	"seanime/internal/database/db"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

const testMagnet = "magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567&dn=Seanime%20Test"

func TestClientPersistsTorrentState(t *testing.T) {
	logger := zerolog.Nop()
	database, err := db.NewDatabase("", "seanime-test", &logger)
	require.NoError(t, err)

	newClient := func() *Client {
		client, clientErr := New(&NewClientOptions{
			Logger:             &logger,
			Database:           database,
			Dir:                t.TempDir(),
			Port:               -1,
			MaxActiveDownloads: 1,
			DisableNetwork:     true,
		})
		if clientErr != nil && strings.Contains(clientErr.Error(), "operation not permitted") {
			t.Skip("environment does not permit the torrent client's localhost listener")
		}
		require.NoError(t, clientErr)
		return client
	}

	client := newClient()
	torrent, err := client.AddMagnet(testMagnet, t.TempDir())
	require.NoError(t, err)
	hash := torrent.InfoHash().HexString()
	require.True(t, client.TorrentExists(hash))

	require.NoError(t, client.PauseTorrent(hash))
	require.NoError(t, client.SetSequential(hash, true))
	require.NoError(t, client.SetForceStart(hash, true))
	snapshots := client.Snapshots()
	require.Len(t, snapshots, 1)
	require.True(t, snapshots[0].ForceStart)
	require.True(t, snapshots[0].Sequential)
	require.False(t, snapshots[0].Paused)
	client.Close()

	restored := newClient()
	t.Cleanup(restored.Close)
	restoredSnapshots := restored.Snapshots()
	require.Len(t, restoredSnapshots, 1)
	require.Equal(t, hash, restoredSnapshots[0].Hash)
	require.True(t, restoredSnapshots[0].ForceStart)
	require.True(t, restoredSnapshots[0].Sequential)

	require.NoError(t, restored.RemoveTorrent(hash, false))
	require.Empty(t, restored.Snapshots())
	persisted, err := database.GetLocalTorrents()
	require.NoError(t, err)
	require.Empty(t, persisted)
}
