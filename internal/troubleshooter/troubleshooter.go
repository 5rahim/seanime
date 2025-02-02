package troubleshooter

import (
	"fmt"
	"seanime/internal/database/models"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/onlinestream"
	"seanime/internal/torrentstream"

	"github.com/rs/zerolog"
)

type (
	Troubleshooter struct {
		logsDir       string
		logger        *zerolog.Logger
		rules         []RuleBuilder
		state         *AppState // For accessing app state like settings
		modules       *Modules
		clientParams  ClientParams
		currentResult Result
	}

	Modules struct {
		MediaPlayerRepository   *mediaplayer.Repository
		OnlinestreamRepository  *onlinestream.Repository
		TorrentstreamRepository *torrentstream.Repository
	}

	NewTroubleshooterOptions struct {
		LogsDir string
		Logger  *zerolog.Logger
		State   *AppState
	}

	AppState struct {
		Settings              *models.Settings
		TorrentstreamSettings *models.TorrentstreamSettings
		MediastreamSettings   *models.MediastreamSettings
		DebridSettings        *models.DebridSettings
	}

	Result struct {
		Items []ResultItem `json:"items"`
	}

	ResultItem struct {
		Module         Module   `json:"module"`
		Observation    string   `json:"observation"`
		Recommendation string   `json:"recommendation"`
		Level          Level    `json:"level"`
		Errors         []string `json:"errors"`
		Warnings       []string `json:"warnings"`
		Logs           []string `json:"logs"`
	}
)

type (
	Module string
	Level  string
)

const (
	LevelError   Level = "error"
	LevelWarning Level = "warning"
	LevelInfo    Level = "info"
	LevelDebug   Level = "debug"
)

const (
	ModulePlayback         Module = "Playback"
	ModuleMediaPlayer      Module = "Media player"
	ModuleAnimeLibrary     Module = "Anime library"
	ModuleMediaStreaming   Module = "Media streaming"
	ModuleTorrentStreaming Module = "Torrent streaming"
)

func NewTroubleshooter(opts NewTroubleshooterOptions, modules *Modules) *Troubleshooter {
	return &Troubleshooter{
		logsDir: opts.LogsDir,
		logger:  opts.Logger,
		state:   opts.State,
		modules: modules,
	}
}

////////////////////

type (
	ClientParams struct {
		LibraryPlaybackOption         string `json:"libraryPlaybackOption"`         // "desktop_media_player" or "media_streaming" or "external_player_link"
		TorrentOrDebridPlaybackOption string `json:"torrentOrDebridPlaybackOption"` // "desktop_torrent_player" or "external_player_link"
	}
)

func (t *Troubleshooter) Run(clientParams ClientParams) {
	t.logger.Info().Msg("troubleshooter: Running troubleshooter")
	t.clientParams = clientParams
	t.currentResult = Result{}

	go t.checkModule(ModulePlayback)

}

func (t *Troubleshooter) checkModule(module Module) {
	t.logger.Info().Str("module", string(module)).Msg("troubleshooter: Checking module")

	switch module {
	case ModulePlayback:
		t.checkPlayback()
	}
}

func (t *Troubleshooter) checkPlayback() {
	t.logger.Info().Msg("troubleshooter: Checking playback")

	switch t.clientParams.LibraryPlaybackOption {
	case "desktop_media_player":
		t.currentResult.AddItem(ResultItem{
			Module:      ModulePlayback,
			Observation: "Your downloaded anime files will be played using the desktop media player you have selected on this device.",
			Level:       LevelInfo,
		})
		t.checkDesktopMediaPlayer()
	case "media_streaming":
		t.currentResult.AddItem(ResultItem{
			Module:      ModulePlayback,
			Observation: "Your downloaded anime files will be played using the media streaming (integrated player) on this device.",
			Level:       LevelInfo,
		})
	case "external_player_link":
		t.currentResult.AddItem(ResultItem{
			Module:      ModulePlayback,
			Observation: "Your downloaded anime files will be played using the external player link you have entered on this device.",
			Level:       LevelInfo,
		})
	}
}

func (t *Troubleshooter) checkDesktopMediaPlayer() {
	t.logger.Info().Msg("troubleshooter: Checking desktop media player")

	binaryPath := t.modules.MediaPlayerRepository.GetExecutablePath()
	defaultPlayer := t.modules.MediaPlayerRepository.GetDefault()

	if binaryPath == "" {
		t.currentResult.AddItem(ResultItem{
			Module:         ModuleMediaPlayer,
			Observation:    fmt.Sprintf("You have selected %s as your desktop media player, but haven't set up the application path for it in the settings.", defaultPlayer),
			Recommendation: "Set up the application path for your desktop media player in the settings.",
			Level:          LevelError,
		})
	}

	_, err := IsExecutable(binaryPath)
	if err != nil {
		t.currentResult.AddItem(ResultItem{
			Module:         ModuleMediaPlayer,
			Observation:    fmt.Sprintf("The application path for your desktop media player is not valid"),
			Recommendation: "Set up the application path for your desktop media player in the settings.",
			Level:          LevelError,
			Errors:         []string{err.Error()},
			Logs:           []string{binaryPath},
		})
	}
}

/////////

func (r *Result) AddItem(item ResultItem) {
	r.Items = append(r.Items, item)
}
