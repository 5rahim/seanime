package plugin

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/continuity"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	debrid_client "seanime/internal/debrid/client"
	"seanime/internal/directstream"
	discordrpc_presence "seanime/internal/discordrpc/presence"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/extension_repo/prompt"
	"seanime/internal/library/autodownloader"
	"seanime/internal/library/autoscanner"
	"seanime/internal/library/fillermanager"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/manga"
	"seanime/internal/mediacore"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/mediastream"
	"seanime/internal/onlinestream"
	"seanime/internal/platforms/platform"
	"seanime/internal/torrent_clients/torrent_client"
	"seanime/internal/torrents/autoselect"
	"seanime/internal/torrents/torrent"
	"seanime/internal/torrentstream"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	gojautil "seanime/internal/util/goja"
	"seanime/internal/videocore"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

type AppContextModules struct {
	IsOfflineRef                    *util.Ref[bool]
	Database                        *db.Database
	AnimeLibraryPaths               *[]string
	AnilistPlatformRef              *util.Ref[platform.Platform]
	PlaybackManager                 *playbackmanager.PlaybackManager
	MediaPlayerRepository           *mediaplayer.Repository
	MangaRepository                 *manga.Repository
	MetadataProviderRef             *util.Ref[metadata_provider.Provider]
	WSEventManager                  events.WSEventManagerInterface
	DiscordPresence                 *discordrpc_presence.Presence
	TorrentRepository               *torrent.Repository
	TorrentClientRepository         *torrent_client.Repository
	DebridClientRepository          *debrid_client.Repository
	ContinuityManager               *continuity.Manager
	AutoScanner                     *autoscanner.AutoScanner
	AutoDownloader                  *autodownloader.AutoDownloader
	FileCacher                      *filecache.Cacher
	OnlinestreamRepository          *onlinestream.Repository
	MediastreamRepository           *mediastream.Repository
	TorrentstreamRepository         *torrentstream.Repository
	FillerManager                   *fillermanager.FillerManager
	VideoCore                       *videocore.VideoCore
	MediacoreCoordinator            *mediacore.Coordinator
	DirectStreamManager             *directstream.Manager
	AutoSelect                      *autoselect.AutoSelect
	OnRefreshAnilistAnimeCollection func()
	OnRefreshAnilistMangaCollection func()
	PromptManager                   *prompt.Manager
	Auth                            AuthActions
	Anilist                         AnilistActions
	Settings                        SettingsActions
	Extensions                      ExtensionActions
}

type AuthActions struct {
	Login  func(token string) error
	Logout func() error
}

type AnilistActions struct {
	UseOfficialClient func() error
	UseCustomClient   func(config anilist.CustomClientConfig) error
}

type SettingsActions struct {
	OnSaved              func(settings *models.Settings)
	OnMediastreamSaved   func(settings *models.MediastreamSettings)
	OnTorrentstreamSaved func(settings *models.TorrentstreamSettings)
	OnDebridSaved        func(settings *models.DebridSettings)
}

type ExtensionActions struct {
	SetDisabled func(id string, disabled bool) error
	GetName     func(id string) string
}

// AppContext allows plugins to interact with core modules.
// It binds JS APIs to the Goja runtimes for that purpose.
type AppContext interface {
	// SetModulesPartial sets modules if they are not nil
	SetModulesPartial(AppContextModules)
	// SetLogger sets the logger for the context
	SetLogger(logger *zerolog.Logger)

	Database() mo.Option[*db.Database]
	PlaybackManager() mo.Option[*playbackmanager.PlaybackManager]
	VideoCore() mo.Option[*videocore.VideoCore]
	MediacoreCoordinator() mo.Option[*mediacore.Coordinator]
	DirectStreamManager() mo.Option[*directstream.Manager]
	MediaPlayerRepository() mo.Option[*mediaplayer.Repository]
	AnilistPlatformRef() mo.Option[*util.Ref[platform.Platform]]
	MetadataProviderRef() mo.Option[*util.Ref[metadata_provider.Provider]]
	WSEventManager() mo.Option[events.WSEventManagerInterface]

	IsOffline() bool

	BindApp(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension)
	// BindStorage binds $storage to the Goja runtime
	BindStorage(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler) *Storage
	// BindAnilist binds $anilist to the Goja runtime
	BindAnilist(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)
	// BindAnilistCustomClient binds runtime client swap APIs to $anilist
	BindAnilistCustomClient(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)
	// BindDatabase binds $database to the Goja runtime
	BindDatabase(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension)
	// BindSystem binds $system to the Goja runtime
	BindSystem(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindPlaybackToContextObj binds 'playback' to the UI context object
	BindPlaybackToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindVideoCoreToContextObj binds 'videoCore' to the UI context object
	BindVideoCoreToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindCronToContextObj binds 'cron' to the UI context object
	BindCronToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler) *Cron

	// BindAuthToContextObj binds 'auth' to the UI context object
	BindAuthToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindAppSettingsToContextObj binds 'appSettings' to the UI context object
	BindAppSettingsToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindExtensionsToContextObj binds 'extensions' to the UI context object
	BindExtensionsToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindDownloaderToContextObj binds 'downloader' to the UI context object
	BindDownloaderToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindMangaToContextObj binds 'manga' to the UI context object
	BindMangaToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindAnimeToContextObj binds 'anime' to the UI context object
	BindAnimeToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindDiscordToContextObj binds 'discord' to the UI context object
	BindDiscordToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindContinuityToContextObj binds 'continuity' to the UI context object
	BindContinuityToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindTorrentClientToContextObj binds 'torrentClient' to the UI context object
	BindTorrentClientToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindTorrentstreamToContextObj binds 'torrentstream' to the UI context object
	BindTorrentstreamToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindDebridToContextObj binds 'debrid' to the UI context object
	BindDebridToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindDebridstreamToContextObj binds 'debridstream' to the UI context object
	BindDebridstreamToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindMediastreamToContextObj binds 'mediastream' to the UI context object
	BindMediastreamToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindOnlinestreamToContextObj binds 'onlinestream' to the UI context object
	BindOnlinestreamToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindFillerManagerToContextObj binds 'fillerManager' to the UI context object
	BindFillerManagerToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindAutoDownloaderToContextObj binds 'autoDownloader' to the UI context object
	BindAutoDownloaderToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindAutoScannerToContextObj binds 'autoScanner' to the UI context object
	BindAutoScannerToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindAutoSelectToContextObj binds 'autoSelect' to the UI context object
	BindAutoSelectToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindScannerToContextObj binds 'scanner' to the UI context object
	BindScannerToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindTorrentSearchToContextObj binds 'torrentSearch' to the UI context object
	BindTorrentSearchToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindFileCacherToContextObj binds 'fileCacher' to the UI context object
	BindFileCacherToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	// BindExternalPlayerLinkToContextObj binds 'externalPlayerLink' to the UI context object
	BindExternalPlayerLinkToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler)

	DropPluginData(extId string)
}

var GlobalAppContext = NewAppContext()

////////////////////////////////////////////////////////////////////////////

type AppContextImpl struct {
	logger *zerolog.Logger

	animeLibraryPaths mo.Option[[]string]

	wsEventManager                  mo.Option[events.WSEventManagerInterface]
	database                        mo.Option[*db.Database]
	playbackManager                 mo.Option[*playbackmanager.PlaybackManager]
	mediaplayerRepo                 mo.Option[*mediaplayer.Repository]
	mangaRepository                 mo.Option[*manga.Repository]
	anilistPlatformRef              mo.Option[*util.Ref[platform.Platform]]
	discordPresence                 mo.Option[*discordrpc_presence.Presence]
	metadataProviderRef             mo.Option[*util.Ref[metadata_provider.Provider]]
	fillerManager                   mo.Option[*fillermanager.FillerManager]
	torrentRepository               mo.Option[*torrent.Repository]
	torrentClientRepository         mo.Option[*torrent_client.Repository]
	debridClientRepository          mo.Option[*debrid_client.Repository]
	torrentstreamRepository         mo.Option[*torrentstream.Repository]
	mediastreamRepository           mo.Option[*mediastream.Repository]
	onlinestreamRepository          mo.Option[*onlinestream.Repository]
	continuityManager               mo.Option[*continuity.Manager]
	autoScanner                     mo.Option[*autoscanner.AutoScanner]
	autoDownloader                  mo.Option[*autodownloader.AutoDownloader]
	fileCacher                      mo.Option[*filecache.Cacher]
	onRefreshAnilistAnimeCollection mo.Option[func()]
	onRefreshAnilistMangaCollection mo.Option[func()]
	videoCore                       mo.Option[*videocore.VideoCore]
	mediacoreCoordinator            mo.Option[*mediacore.Coordinator]
	directStreamManager             mo.Option[*directstream.Manager]
	isOfflineRef                    *util.Ref[bool]
	autoSelect                      mo.Option[*autoselect.AutoSelect]
	promptManager                   mo.Option[*prompt.Manager]
	auth                            AuthActions
	anilist                         AnilistActions
	settings                        SettingsActions
	extensions                      ExtensionActions
}

func NewAppContext() AppContext {
	appCtx := &AppContextImpl{
		logger:                          new(zerolog.Nop()),
		database:                        mo.None[*db.Database](),
		playbackManager:                 mo.None[*playbackmanager.PlaybackManager](),
		mediaplayerRepo:                 mo.None[*mediaplayer.Repository](),
		anilistPlatformRef:              mo.None[*util.Ref[platform.Platform]](),
		mangaRepository:                 mo.None[*manga.Repository](),
		metadataProviderRef:             mo.None[*util.Ref[metadata_provider.Provider]](),
		wsEventManager:                  mo.None[events.WSEventManagerInterface](),
		discordPresence:                 mo.None[*discordrpc_presence.Presence](),
		fillerManager:                   mo.None[*fillermanager.FillerManager](),
		torrentRepository:               mo.None[*torrent.Repository](),
		torrentClientRepository:         mo.None[*torrent_client.Repository](),
		debridClientRepository:          mo.None[*debrid_client.Repository](),
		torrentstreamRepository:         mo.None[*torrentstream.Repository](),
		mediastreamRepository:           mo.None[*mediastream.Repository](),
		onlinestreamRepository:          mo.None[*onlinestream.Repository](),
		continuityManager:               mo.None[*continuity.Manager](),
		autoScanner:                     mo.None[*autoscanner.AutoScanner](),
		autoDownloader:                  mo.None[*autodownloader.AutoDownloader](),
		fileCacher:                      mo.None[*filecache.Cacher](),
		onRefreshAnilistAnimeCollection: mo.None[func()](),
		onRefreshAnilistMangaCollection: mo.None[func()](),
		videoCore:                       mo.None[*videocore.VideoCore](),
		mediacoreCoordinator:            mo.None[*mediacore.Coordinator](),
		directStreamManager:             mo.None[*directstream.Manager](),
		isOfflineRef:                    util.NewRef(false),
		autoSelect:                      mo.None[*autoselect.AutoSelect](),
		promptManager:                   mo.None[*prompt.Manager](),
	}

	return appCtx
}

func (a *AppContextImpl) IsOffline() bool {
	return a.isOfflineRef.Get()
}

func (a *AppContextImpl) SetLogger(logger *zerolog.Logger) {
	a.logger = logger
}

func (a *AppContextImpl) Database() mo.Option[*db.Database] {
	return a.database
}

func (a *AppContextImpl) PlaybackManager() mo.Option[*playbackmanager.PlaybackManager] {
	return a.playbackManager
}

func (a *AppContextImpl) VideoCore() mo.Option[*videocore.VideoCore] {
	return a.videoCore
}

func (a *AppContextImpl) MediacoreCoordinator() mo.Option[*mediacore.Coordinator] {
	return a.mediacoreCoordinator
}

func (a *AppContextImpl) DirectStreamManager() mo.Option[*directstream.Manager] {
	return a.directStreamManager
}

func (a *AppContextImpl) MediaPlayerRepository() mo.Option[*mediaplayer.Repository] {
	return a.mediaplayerRepo
}

func (a *AppContextImpl) AnilistPlatformRef() mo.Option[*util.Ref[platform.Platform]] {
	return a.anilistPlatformRef
}

func (a *AppContextImpl) MetadataProviderRef() mo.Option[*util.Ref[metadata_provider.Provider]] {
	return a.metadataProviderRef
}

func (a *AppContextImpl) WSEventManager() mo.Option[events.WSEventManagerInterface] {
	return a.wsEventManager
}

func (a *AppContextImpl) SetModulesPartial(modules AppContextModules) {
	if modules.IsOfflineRef != nil {
		a.isOfflineRef = modules.IsOfflineRef
	}

	if modules.Database != nil {
		a.database = mo.Some(modules.Database)
	}

	if modules.AnimeLibraryPaths != nil {
		a.animeLibraryPaths = mo.Some(*modules.AnimeLibraryPaths)
	}

	if modules.MetadataProviderRef.IsPresent() {
		a.metadataProviderRef = mo.Some(modules.MetadataProviderRef)
	}

	if modules.PlaybackManager != nil {
		a.playbackManager = mo.Some(modules.PlaybackManager)
	}

	if modules.AnilistPlatformRef.IsPresent() {
		a.anilistPlatformRef = mo.Some(modules.AnilistPlatformRef)
	}

	if modules.MediaPlayerRepository != nil {
		a.mediaplayerRepo = mo.Some(modules.MediaPlayerRepository)
	}

	if modules.FillerManager != nil {
		a.fillerManager = mo.Some(modules.FillerManager)
	}

	if modules.OnRefreshAnilistAnimeCollection != nil {
		a.onRefreshAnilistAnimeCollection = mo.Some(modules.OnRefreshAnilistAnimeCollection)
	}

	if modules.OnRefreshAnilistMangaCollection != nil {
		a.onRefreshAnilistMangaCollection = mo.Some(modules.OnRefreshAnilistMangaCollection)
	}

	if modules.MangaRepository != nil {
		a.mangaRepository = mo.Some(modules.MangaRepository)
	}

	if modules.DiscordPresence != nil {
		a.discordPresence = mo.Some(modules.DiscordPresence)
	}

	if modules.WSEventManager != nil {
		a.wsEventManager = mo.Some(modules.WSEventManager)
	}

	if modules.ContinuityManager != nil {
		a.continuityManager = mo.Some(modules.ContinuityManager)
	}

	if modules.TorrentRepository != nil {
		a.torrentRepository = mo.Some(modules.TorrentRepository)
	}

	if modules.TorrentClientRepository != nil {
		a.torrentClientRepository = mo.Some(modules.TorrentClientRepository)
	}

	if modules.DebridClientRepository != nil {
		a.debridClientRepository = mo.Some(modules.DebridClientRepository)
	}

	if modules.TorrentstreamRepository != nil {
		a.torrentstreamRepository = mo.Some(modules.TorrentstreamRepository)
		a.autoSelect = mo.Some(modules.TorrentstreamRepository.GetAutoSelect())
	}

	if modules.MediastreamRepository != nil {
		a.mediastreamRepository = mo.Some(modules.MediastreamRepository)
	}

	if modules.OnlinestreamRepository != nil {
		a.onlinestreamRepository = mo.Some(modules.OnlinestreamRepository)
	}

	if modules.AutoDownloader != nil {
		a.autoDownloader = mo.Some(modules.AutoDownloader)
	}

	if modules.AutoScanner != nil {
		a.autoScanner = mo.Some(modules.AutoScanner)
	}

	if modules.FileCacher != nil {
		a.fileCacher = mo.Some(modules.FileCacher)
	}

	if modules.VideoCore != nil {
		a.videoCore = mo.Some(modules.VideoCore)
	}

	if modules.MediacoreCoordinator != nil {
		a.mediacoreCoordinator = mo.Some(modules.MediacoreCoordinator)
	}

	if modules.DirectStreamManager != nil {
		a.directStreamManager = mo.Some(modules.DirectStreamManager)
	}

	if modules.PromptManager != nil {
		a.promptManager = mo.Some(modules.PromptManager)
	}

	if modules.Auth.Login != nil {
		a.auth.Login = modules.Auth.Login
	}
	if modules.Auth.Logout != nil {
		a.auth.Logout = modules.Auth.Logout
	}

	if modules.Anilist.UseOfficialClient != nil {
		a.anilist.UseOfficialClient = modules.Anilist.UseOfficialClient
	}
	if modules.Anilist.UseCustomClient != nil {
		a.anilist.UseCustomClient = modules.Anilist.UseCustomClient
	}

	if modules.Settings.OnSaved != nil {
		a.settings.OnSaved = modules.Settings.OnSaved
	}
	if modules.Settings.OnMediastreamSaved != nil {
		a.settings.OnMediastreamSaved = modules.Settings.OnMediastreamSaved
	}
	if modules.Settings.OnTorrentstreamSaved != nil {
		a.settings.OnTorrentstreamSaved = modules.Settings.OnTorrentstreamSaved
	}
	if modules.Settings.OnDebridSaved != nil {
		a.settings.OnDebridSaved = modules.Settings.OnDebridSaved
	}

	if modules.Extensions.SetDisabled != nil {
		a.extensions.SetDisabled = modules.Extensions.SetDisabled
	}
	if modules.Extensions.GetName != nil {
		a.extensions.GetName = modules.Extensions.GetName
	}
}

func (a *AppContextImpl) DropPluginData(extId string) {
	db, ok := a.database.Get()
	if !ok {
		return
	}

	err := db.Gorm().Where("plugin_id = ?", extId).Delete(&models.PluginData{}).Error
	if err != nil {
		a.logger.Error().Err(err).Msg("Failed to drop plugin data")
	}
}
