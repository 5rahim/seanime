package torbox

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"seanime/internal/debrid/debrid"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"strconv"
	"testing"
)

func TestTorBox_GetTorrents(t *testing.T) {
	test_utils.InitTestProvider(t)
	logger := util.NewLogger()

	tb := NewTorBox(logger)

	err := tb.Authenticate(test_utils.ConfigData.Provider.TorBoxApiKey)
	require.NoError(t, err)

	fmt.Println("=== All torrents ===")

	torrents, err := tb.GetTorrents()
	require.NoError(t, err)

	util.Spew(torrents)

	fmt.Println("=== Selecting torrent ===")

	torrent, err := tb.GetTorrent(strconv.Itoa(98926))
	require.NoError(t, err)

	util.Spew(torrent)

	fmt.Println("=== Download link ===")

	downloadUrl, err := tb.GetTorrentDownloadUrl(debrid.DownloadTorrentOptions{
		ID: strconv.Itoa(98926),
	})
	require.NoError(t, err)

	fmt.Println(downloadUrl)
}

func TestTorBox_AddTorrent(t *testing.T) {
	t.Skip("Skipping test that adds a torrent to TorBox")

	test_utils.InitTestProvider(t)

	// Already added
	magnet := "magnet:?xt=urn:btih:80431b4f9a12f4e06616062d3d3973b9ef99b5e6&dn=%5BSubsPlease%5D%20Bocchi%20the%20Rock%21%20-%2001%20%281080p%29%20%5BE04F4EFB%5D.mkv&tr=http%3A%2F%2Fnyaa.tracker.wf%3A7777%2Fannounce&tr=udp%3A%2F%2Fopen.stealth.si%3A80%2Fannounce&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337%2Fannounce&tr=udp%3A%2F%2Fexodus.desync.com%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.torrent.eu.org%3A451%2Fannounce"

	logger := util.NewLogger()

	tb := NewTorBox(logger)

	err := tb.Authenticate(test_utils.ConfigData.Provider.TorBoxApiKey)
	require.NoError(t, err)

	torrentId, err := tb.AddTorrent(debrid.AddTorrentOptions{
		MagnetLink: magnet,
	})
	require.NoError(t, err)

	fmt.Println(torrentId)
}
