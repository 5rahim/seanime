package playbackmanager_test

import (
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/mediaplayer"
	"github.com/seanime-app/seanime/internal/mpchc"
	"github.com/seanime-app/seanime/internal/mpv"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/vlc"
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
	playbackManager, anilistClientWrapper, anilistCollection, err := getPlaybackManager()
	if err != nil {
		t.Fatal(err)
	}

	repo := getRepo()

	playbackManager.SetAnilistClientWrapper(anilistClientWrapper)
	playbackManager.SetAnilistCollection(anilistCollection)
	playbackManager.SetMediaPlayerRepository(repo)

	// Test the playlist hub
	lfs := make([]*entities.LocalFile, 0)
	for _, path := range localFilePaths {
		lf := entities.NewLocalFile(path, "E:/ANIME")
		epNum, _ := strconv.Atoi(lf.ParsedData.Episode)
		lf.MediaId = mediaId
		lf.Metadata.Type = entities.LocalFileTypeMain
		lf.Metadata.Episode = epNum
		lf.Metadata.AniDBEpisode = lf.ParsedData.Episode
		lfs = append(lfs, lf)
	}

	playlist := &entities.Playlist{
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
		Host:     "localhost",
		Port:     8080,
		Password: "seanime",
		Logger:   logger,
	}

	mpc := &mpchc.MpcHc{
		Host:   "localhost",
		Port:   13579,
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
