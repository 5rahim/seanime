package realdebrid

import (
	"fmt"
	"seanime/internal/debrid/debrid"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTorBox_GetTorrents(t *testing.T) {
	test_utils.InitTestProvider(t)
	logger := util.NewLogger()

	rd := NewRealDebrid(logger)

	err := rd.Authenticate(test_utils.ConfigData.Provider.RealDebridApiKey)
	require.NoError(t, err)

	fmt.Println("=== All torrents ===")

	torrents, err := rd.GetTorrents()
	require.NoError(t, err)

	util.Spew(torrents)
}

func TestTorBox_AddTorrent(t *testing.T) {
	t.Skip("Skipping test that adds a torrent to RealDebrid")

	test_utils.InitTestProvider(t)

	// Already added
	magnet := ""

	logger := util.NewLogger()

	rd := NewRealDebrid(logger)

	err := rd.Authenticate(test_utils.ConfigData.Provider.RealDebridApiKey)
	require.NoError(t, err)

	torrentId, err := rd.AddTorrent(debrid.AddTorrentOptions{
		MagnetLink: magnet,
		InfoHash:   "TODO",
	})
	require.NoError(t, err)

	torrentId2, err := rd.AddTorrent(debrid.AddTorrentOptions{
		MagnetLink: magnet,
		InfoHash:   "TODO",
	})
	require.NoError(t, err)

	require.Equal(t, torrentId, torrentId2)

	fmt.Println(torrentId)
}

func TestTorBox_getTorrentInfo(t *testing.T) {

	test_utils.InitTestProvider(t)

	logger := util.NewLogger()

	rd := NewRealDebridT(logger)

	err := rd.Authenticate(test_utils.ConfigData.Provider.RealDebridApiKey)
	require.NoError(t, err)

	ti, err := rd.getTorrentInfo("W3IWF5TX3AE6G")
	require.NoError(t, err)

	util.Spew(ti)
}

func TestTorBox_GetDownloadUrl(t *testing.T) {

	test_utils.InitTestProvider(t)

	logger := util.NewLogger()

	rd := NewRealDebridT(logger)

	err := rd.Authenticate(test_utils.ConfigData.Provider.RealDebridApiKey)
	require.NoError(t, err)

	urls, err := rd.GetTorrentDownloadUrl(debrid.DownloadTorrentOptions{
		ID:     "W3IWF5TX3AE6G",
		FileId: "11",
	})
	require.NoError(t, err)

	util.Spew(strings.Split(urls, ","))
}

func TestTorBox_InstantAvailability(t *testing.T) {

	test_utils.InitTestProvider(t)

	logger := util.NewLogger()

	rd := NewRealDebridT(logger)

	err := rd.Authenticate(test_utils.ConfigData.Provider.RealDebridApiKey)
	require.NoError(t, err)
	avail := rd.GetInstantAvailability([]string{"9f4961a9c71eeb53abce2ef2afc587b452dee5eb"})
	require.NoError(t, err)

	util.Spew(avail)
}

func TestTorBox_ChooseFileAndDownload(t *testing.T) {
	//t.Skip("Skipping test that adds a torrent to RealDebrid")

	test_utils.InitTestProvider(t)

	magnet := ""

	logger := util.NewLogger()

	rd := NewRealDebrid(logger)

	err := rd.Authenticate(test_utils.ConfigData.Provider.RealDebridApiKey)
	require.NoError(t, err)

	// Should add the torrent and get the torrent info
	torrentInfo, err := rd.GetTorrentInfo(debrid.GetTorrentInfoOptions{
		MagnetLink: magnet,
		InfoHash:   "TODO",
	})
	require.NoError(t, err)

	// The torrent should have one file
	require.Len(t, torrentInfo.Files, 1)

	file := torrentInfo.Files[0]

	// Download the file
	resp, err := rd.AddTorrent(debrid.AddTorrentOptions{
		MagnetLink:   magnet,
		InfoHash:     "TODO",
		SelectFileId: file.ID,
	})
	require.NoError(t, err)

	util.Spew(resp)
}
