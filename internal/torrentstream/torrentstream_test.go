package torrentstream

import (
	"github.com/seanime-app/seanime/internal/mediaplayers/mediaplayer"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"testing"
)

//func TestSomething(t *testing.T) {
//	test_utils.SetTwoLevelDeep()
//	test_utils.InitTestProvider(t, test_utils.MediaPlayer())
//
//	repo := NewRepository(&NewRepositoryOptions{
//		Logger:                util.NewLogger(),
//		MediaPlayerRepository: mediaplayer.NewTestRepository(t, "mpv"),
//	})
//	defer repo.Cleanup()
//
//	err := repo.InitModules(&Settings{
//		Port: 3002,
//	})
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	fmt.Println(repo.settings.MustGet().DownloadDir)
//
//	app := fiber.New()
//
//	err = repo.StartStream(magnetLink)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	app.Get("/stream", func(c *fiber.Ctx) error {
//		c.Set("Content-Type", "video/mp4")
//
//		if repo.playback.currentFile.IsAbsent() {
//			t.Fatal("No file to stream")
//			return nil
//		}
//
//		// Send the content of the file
//		_, err := io.Copy(c, repo.playback.currentFile.MustGet().NewReader())
//		if err != nil {
//			// Handle error
//			return err
//		}
//
//		return nil
//	})
//
//	app.Get("/hello", func(c *fiber.Ctx) error {
//		fmt.Println(c.BaseURL())
//		return c.SendString("Hello, World!")
//	})
//
//	go func() {
//		app.Listen("127.0.0.1:3002")
//	}()
//	defer app.Shutdown()
//	time.Sleep(1 * time.Second)
//
//	err = repo.mediaPlayerRepository.Play("http://127.0.0.1:3002/stream")
//	if err != nil {
//		repo.logger.Error().Err(err).Msg("Failed to play the stream")
//	}
//	defer repo.mediaPlayerRepository.Stop()
//
//	select {}
//
//}

func TestSomething2(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	repo := NewRepository(&NewRepositoryOptions{
		Logger:                util.NewLogger(),
		MediaPlayerRepository: mediaplayer.NewTestRepository(t, "mpv"),
	})
	defer repo.Cleanup()

	repo.Test()

}
