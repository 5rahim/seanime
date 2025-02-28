package plugin

import (
	"seanime/internal/database/db"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/platforms/platform"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

type AppContextModules struct {
	Database                        *db.Database
	AnilistPlatform                 platform.Platform
	PlaybackManager                 *playbackmanager.PlaybackManager
	WSEventManager                  events.WSEventManagerInterface
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
	AnilistPlatform() mo.Option[platform.Platform]
	WSEventManager() mo.Option[events.WSEventManagerInterface]

	// BindStorage binds $storage to the Goja runtime
	BindStorage(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension)
	// BindAnilist binds $anilist to the Goja runtime
	BindAnilist(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension)
	// BindDatabase binds $database to the Goja runtime
	BindDatabase(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension)
	// BindFilepath binds $filepath to the Goja runtime
	BindFilepath(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension)
	// BindOS binds $os to the Goja runtime
	BindOS(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension)
	// BindFilesystem binds $filesystem to the Goja runtime
	BindFilesystem(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension)
}

var GlobalAppContext = NewAppContext()

////////////////////////////////////////////////////////////////////////////

type AppContextImpl struct {
	logger *zerolog.Logger

	wsEventManager  mo.Option[events.WSEventManagerInterface]
	database        mo.Option[*db.Database]
	playbackManager mo.Option[*playbackmanager.PlaybackManager]
	anilistPlatform mo.Option[platform.Platform]

	onRefreshAnilistAnimeCollection mo.Option[func()]
	onRefreshAnilistMangaCollection mo.Option[func()]
}

func NewAppContext() AppContext {
	nopLogger := zerolog.Nop()
	appCtx := &AppContextImpl{
		logger:          &nopLogger,
		database:        mo.None[*db.Database](),
		playbackManager: mo.None[*playbackmanager.PlaybackManager](),
		anilistPlatform: mo.None[platform.Platform](),
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

	if modules.WSEventManager != nil {
		a.wsEventManager = mo.Some(modules.WSEventManager)
	}
}
