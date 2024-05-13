package torrentstream

import (
	"errors"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/davecgh/go-spew/spew"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/mediaplayers/mediaplayer"
	"github.com/seanime-app/seanime/internal/util"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type (
	Repository struct {
		client                *Client
		mediaPlayerRepository *mediaplayer.Repository
		settings              mo.Option[Settings] // None by default, set and refreshed by SetSettings
		logger                *zerolog.Logger
		playback              *Playback
	}

	Playback struct {
		currentFile mo.Option[*torrent.File]
	}

	Settings struct {
		Host                string
		TorrentClientPort   int
		StreamingServerHost string
		StreamingServerPort int
		DownloadDir         string
		DisableIPV6         bool
		AddToLibrary        bool
	}

	NewRepositoryOptions struct {
		Logger                *zerolog.Logger
		MediaPlayerRepository *mediaplayer.Repository
	}
)

// NewRepository creates a new injectable Repository instance
func NewRepository(opts *NewRepositoryOptions) *Repository {
	ret := &Repository{
		logger:                opts.Logger,
		mediaPlayerRepository: opts.MediaPlayerRepository,
		playback: &Playback{
			currentFile: mo.None[*torrent.File](),
		},
	}
	ret.client = NewClient(&NewClientOptions{
		Repository: ret,
	})
	return ret
}

func (r *Repository) SetMediaPlayerRepository(mediaPlayerRepository *mediaplayer.Repository) {
	r.mediaPlayerRepository = mediaPlayerRepository
}

// InitModules sets the settings for the torrentstream module
// It should be called before any other method, to ensure the module is active
func (r *Repository) InitModules(settings *models.TorrentstreamSettings, host string) (err error) {
	r.client.Close()

	defer util.HandlePanicInModuleWithError("torrentstream/InitModules", &err)

	if settings == nil {
		r.logger.Error().Msg("torrentstream: Cannot initialize module, no settings provided")
		r.settings = mo.None[Settings]()
		return errors.New("torrentstream: Cannot initialize module, no settings provided")
	}

	s := *settings

	// Set default download directory, which is a temporary directory
	if s.DownloadDir == "" {
		s.DownloadDir = r.getDefaultDownloadPath()
		_ = os.MkdirAll(s.DownloadDir, os.ModePerm) // Create the directory if it doesn't exist
	}

	if s.StreamingServerPort == 0 {
		s.StreamingServerPort = 43212
	}
	if s.TorrentClientPort == 0 {
		s.TorrentClientPort = 43213
	}
	if s.StreamingServerHost == "" {
		s.StreamingServerHost = "0.0.0.0"
	}

	// Set the settings
	r.settings = mo.Some(Settings{
		Host:                host,
		StreamingServerPort: s.StreamingServerPort,
		StreamingServerHost: s.StreamingServerHost,
		TorrentClientPort:   s.TorrentClientPort,
		DownloadDir:         s.DownloadDir,
		DisableIPV6:         s.DisableIPV6,
		AddToLibrary:        s.AddToLibrary,
	})

	// Initialize the torrent client
	err = r.client.createTorrentClient()
	if err != nil {
		return err
	}

	r.logger.Info().Msg("torrentstream: Module initialized")
	return nil
}

func (r *Repository) FailIfNoSettings() error {
	if r.settings.IsAbsent() {
		return errors.New("torrentstream: no settings provided, the module is dormant")
	}
	return nil
}

func (r *Repository) Cleanup() {
	if r.settings.IsAbsent() {
		return
	}
	r.client.Close()
	_ = os.RemoveAll(r.settings.MustGet().DownloadDir)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) StartStream(magnet string) error {

	r.logger.Info().Str("magnet", magnet).Msg("torrentstream: Starting stream")

	t, err := r.client.AddTorrent(magnet)
	if err != nil {
		return err
	}

	files := t.Files()
	if len(files) == 0 {
		return errors.New("torrentstream: no files found in the torrent")
	}

	spew.Dump(files)

	file := files[0] // TODO change

	r.playback.currentFile = mo.Some(file)

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) getDefaultDownloadPath() string {
	tempDir := os.TempDir()
	downloadDirPath := filepath.Join(tempDir, "seanime", "torrentstream")
	return downloadDirPath
}

var magnetLink = "magnet:?xt=urn:btih:O7DBLUBVWSQPLSJXEZXHCRQU5ZF2KKM2&tr=http%3A%2F%2Fnyaa.tracker.wf%3A7777%2Fannounce&tr=udp%3A%2F%2Fopen.stealth.si%3A80%2Fannounce&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337%2Fannounce&tr=udp%3A%2F%2Fexodus.desync.com%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.torrent.eu.org%3A451%2Fannounce&dn=%5BEMBER%5D%20Dungeon%20Meshi%20S01E17%20%5B1080p%5D%20%5BDual%20Audio%20HEVC%20WEBRip%20DD%5D%20%28Delicious%20in%20Dungeon%29"

func (r *Repository) Test() {

	err := r.InitModules(&models.TorrentstreamSettings{}, "127.0.0.1")
	if err != nil {
		panic(err)
	}

	fmt.Println(r.settings.MustGet().DownloadDir)

	err = r.StartStream(magnetLink)
	if err != nil {
		panic(err)
	}

	port := fmt.Sprintf("0.0.0.0:%s", "3002")
	srv := &http.Server{Addr: port}
	http.HandleFunc("/stream", func(w http.ResponseWriter, _r *http.Request) {
		r.logger.Info().Msg("Request received")
		w.Header().Set("Content-Type", "video/mp4")
		// Open the file using the custom reader
		http.ServeContent(w, _r, r.playback.currentFile.MustGet().DisplayPath(), time.Unix(r.playback.currentFile.MustGet().Torrent().Metainfo().CreationDate, 0), r.playback.currentFile.MustGet().NewReader())
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// show hello world
		io.WriteString(w, "Hello, world!")
	})

	go func() {
		r.logger.Info().Msg("Starting server")
		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			} else {
				panic(err)
			}
		}
	}()

	defer srv.Shutdown(nil)

	time.Sleep(1 * time.Second)

	err = r.mediaPlayerRepository.Play("http://127.0.0.1:3002/stream")
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to play the stream")
	}
	defer r.mediaPlayerRepository.Stop()

	select {}

}

//

func (r *Repository) Test2() {

	err := r.InitModules(&models.TorrentstreamSettings{}, "127.0.0.1")
	if err != nil {
		panic(err)
	}

	fmt.Println(r.settings.MustGet().DownloadDir)

	app := fiber.New()

	err = r.StartStream(magnetLink)
	if err != nil {
		panic(err)
	}

	app.Get("/stream", adaptor.HTTPHandlerFunc(func(w http.ResponseWriter, _r *http.Request) {
		r.logger.Info().Msg("Request received")
		w.Header().Set("Content-Type", "video/mp4")
		spew.Dump(_r.Header)
		// Open the file using the custom reader
		http.ServeContent(w, _r, r.playback.currentFile.MustGet().DisplayPath(), time.Unix(r.playback.currentFile.MustGet().Torrent().Metainfo().CreationDate, 0), r.playback.currentFile.MustGet().NewReader())
	}))

	app.Get("/hello", func(c *fiber.Ctx) error {
		fmt.Println(c.BaseURL())
		return c.SendString("Hello, World!")
	})

	go func() {
		app.Listen("127.0.0.1:3002")
	}()
	defer app.Shutdown()

	//time.Sleep(1 * time.Second)
	//
	//err = r.mediaPlayerRepository.Play("http://127.0.0.1:3002/stream")
	//if err != nil {
	//	r.logger.Error().Err(err).Msg("Failed to play the stream")
	//}
	//defer r.mediaPlayerRepository.Stop()

	select {}

}
