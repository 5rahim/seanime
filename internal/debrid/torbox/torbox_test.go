package torbox

import (
	"fmt"
	"seanime/internal/debrid/debrid"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
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
	magnet := ""

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
