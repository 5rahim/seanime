package plugin

import (
	"seanime/internal/database/db"
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
	OnRefreshAnilistAnimeCollection func()
	OnRefreshAnilistMangaCollection func()
}

// AppContext contains all the modules that are available to the plugin.
// It is used to bind JS APIs to the Goja runtimes.
type AppContext interface {
	Database() mo.Option[*db.Database]
	PlaybackManager() mo.Option[*playbackmanager.PlaybackManager]
	AnilistPlatform() mo.Option[platform.Platform]
	SetModules(AppContextModules)

	BindStorage(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension)
}

var GlobalAppContext = NewAppContext()

////////////////////////////////////////////////////////////////////////////

type AppContextImpl struct {
	database        mo.Option[*db.Database]
	playbackManager mo.Option[*playbackmanager.PlaybackManager]
	anilistPlatform mo.Option[platform.Platform]

	onRefreshAnilistAnimeCollection mo.Option[func()]
	onRefreshAnilistMangaCollection mo.Option[func()]
}

func NewAppContext() AppContext {
	appCtx := &AppContextImpl{
		database:        mo.None[*db.Database](),
		playbackManager: mo.None[*playbackmanager.PlaybackManager](),
		anilistPlatform: mo.None[platform.Platform](),
	}

	return appCtx
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

// SetModules sets modules individually if they are not nil
func (a *AppContextImpl) SetModules(modules AppContextModules) {
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
}
