package playbackmanager_test

import (
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/library/anime"
	"github.com/seanime-app/seanime/internal/mediaplayers/mediaplayer"
	"github.com/seanime-app/seanime/internal/mediaplayers/mpchc"
	"github.com/seanime-app/seanime/internal/mediaplayers/mpv"
	"github.com/seanime-app/seanime/internal/mediaplayers/vlc"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"strconv"
	"testing"
)

var defaultPlayer = "vlc"
var localFilePaths = []string{
	"E:/ANIME/Dungeon Meshi/[EMBER] Dungeon Meshi - 04.mkv",
	"E:/ANIME/Dungeon Meshi/[EMBER] Dungeon Meshi - 05.mkv",
	"E:/ANIME/Dungeon Meshi/[EMBER] Dungeon Meshi - 06.mkv",
}
var mediaId = 153518

func TestPlaylists(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist(), test_utils.MediaPlayer())

	playbackManager, anilistClientWrapper, anilistCollection, err := getPlaybackManager(t)
	if err != nil {
		t.Fatal(err)
	}

	repo := getRepo()

	playbackManager.SetAnilistClientWrapper(anilistClientWrapper)
	playbackManager.SetAnilistCollection(anilistCollection)
	playbackManager.SetMediaPlayerRepository(repo)

	// Test the playlist hub
	lfs := make([]*anime.LocalFile, 0)
	for _, path := range localFilePaths {
		lf := anime.NewLocalFile(path, "E:/ANIME")
		epNum, _ := strconv.Atoi(lf.ParsedData.Episode)
		lf.MediaId = mediaId
		lf.Metadata.Type = anime.LocalFileTypeMain
		lf.Metadata.Episode = epNum
		lf.Metadata.AniDBEpisode = lf.ParsedData.Episode
		lfs = append(lfs, lf)
	}

	playlist := &anime.Playlist{
		DbId:       1,
		Name:       "test",
		LocalFiles: lfs,
	}

	err = playbackManager.StartPlaylist(playlist)
	if err != nil {
		t.Fatal(err)
	}

	select {}

}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func getRepo() *mediaplayer.Repository {
	logger := util.NewLogger()
	WSEventManager := events.NewMockWSEventManager(logger)

	vlcI := &vlc.VLC{
		Host:     test_utils.ConfigData.Provider.VlcPath,
		Port:     test_utils.ConfigData.Provider.VlcPort,
		Password: test_utils.ConfigData.Provider.VlcPassword,
		Logger:   logger,
	}

	mpc := &mpchc.MpcHc{
		Host:   test_utils.ConfigData.Provider.MpcHost,
		Path:   test_utils.ConfigData.Provider.MpcPath,
		Port:   test_utils.ConfigData.Provider.MpcPort,
		Logger: logger,
	}

	repo := mediaplayer.NewRepository(&mediaplayer.NewRepositoryOptions{
		Logger:         logger,
		Default:        defaultPlayer,
		VLC:            vlcI,
		MpcHc:          mpc,
		Mpv:            mpv.New(logger, "", ""),
		WSEventManager: WSEventManager,
	})
	return repo
}
