package plugin

import (
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	discordrpc_presence "seanime/internal/discordrpc/presence"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/manga"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/platforms/platform"
	goja_util "seanime/internal/util/goja"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

type AppContextModules struct {
	Database                        *db.Database
	AnimeLibraryPaths               []string
	AnilistPlatform                 platform.Platform
	PlaybackManager                 *playbackmanager.PlaybackManager
	MediaPlayerRepository           *mediaplayer.Repository
	MangaRepository                 *manga.Repository
	WSEventManager                  events.WSEventManagerInterface
	DiscordPresence                 *discordrpc_presence.Presence
	OnRefreshAnilistAnimeCollection func()
	OnRefreshAnilistMangaCollection func()
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
	MediaPlayerRepository() mo.Option[*mediaplayer.Repository]
	AnilistPlatform() mo.Option[platform.Platform]
	WSEventManager() mo.Option[events.WSEventManagerInterface]

	BindApp(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension)
	// BindStorage binds $storage to the Goja runtime
	BindStorage(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) *Storage
	// BindAnilist binds $anilist to the Goja runtime
	BindAnilist(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension)
	// BindDatabase binds $database to the Goja runtime
	BindDatabase(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension)

	// BindSystem binds $system to the Goja runtime
	BindSystem(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler)

	// BindPlaybackToContextObj binds 'playback' to the UI context object
	BindPlaybackToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler)

	// BindCronToContextObj binds 'cron' to the UI context object
	BindCronToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) *Cron

	// BindDownloaderToContextObj binds 'downloader' to the UI context object
	BindDownloaderToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler)

	// BindMangaToContextObj binds 'manga' to the UI context object
	BindMangaToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler)

	// BindDiscordToContextObj binds 'discord' to the UI context object
	BindDiscordToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler)

	DropPluginData(extId string)
}

var GlobalAppContext = NewAppContext()

////////////////////////////////////////////////////////////////////////////

type AppContextImpl struct {
	logger *zerolog.Logger

	animeLibraryPaths mo.Option[[]string]

	wsEventManager  mo.Option[events.WSEventManagerInterface]
	database        mo.Option[*db.Database]
	playbackManager mo.Option[*playbackmanager.PlaybackManager]
	mediaplayerRepo mo.Option[*mediaplayer.Repository]
	mangaRepository mo.Option[*manga.Repository]
	anilistPlatform mo.Option[platform.Platform]
	discordPresence mo.Option[*discordrpc_presence.Presence]

	onRefreshAnilistAnimeCollection mo.Option[func()]
	onRefreshAnilistMangaCollection mo.Option[func()]
}

func NewAppContext() AppContext {
	nopLogger := zerolog.Nop()
	appCtx := &AppContextImpl{
		logger:          &nopLogger,
		database:        mo.None[*db.Database](),
		playbackManager: mo.None[*playbackmanager.PlaybackManager](),
		mediaplayerRepo: mo.None[*mediaplayer.Repository](),
		anilistPlatform: mo.None[platform.Platform](),
		mangaRepository: mo.None[*manga.Repository](),
	}

	return appCtx
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

func (a *AppContextImpl) MediaPlayerRepository() mo.Option[*mediaplayer.Repository] {
	return a.mediaplayerRepo
}

func (a *AppContextImpl) AnilistPlatform() mo.Option[platform.Platform] {
	return a.anilistPlatform
}

func (a *AppContextImpl) WSEventManager() mo.Option[events.WSEventManagerInterface] {
	return a.wsEventManager
}

func (a *AppContextImpl) SetModulesPartial(modules AppContextModules) {
	if modules.Database != nil {
		a.database = mo.Some(modules.Database)
	}

	if modules.AnimeLibraryPaths != nil {
		a.animeLibraryPaths = mo.Some(modules.AnimeLibraryPaths)
	}

	if modules.PlaybackManager != nil {
		a.playbackManager = mo.Some(modules.PlaybackManager)
	}

	if modules.AnilistPlatform != nil {
		a.anilistPlatform = mo.Some(modules.AnilistPlatform)
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
