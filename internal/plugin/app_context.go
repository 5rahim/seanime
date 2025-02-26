package plugin

import (
	"seanime/internal/database/db"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/platforms/platform"

	"github.com/samber/mo"
)

type AppContextModules struct {
	Database        *db.Database
	AnilistPlatform platform.Platform
	PlaybackManager *playbackmanager.PlaybackManager
}

// AppContext contains all the modules that are available to the plugin.
// It is used to bind JS APIs to the Goja runtimes.
type AppContext interface {
	Database() mo.Option[*db.Database]
	PlaybackManager() mo.Option[*playbackmanager.PlaybackManager]
	AnilistPlatform() mo.Option[platform.Platform]
	SetModules(AppContextModules)
}

var GlobalAppContext = NewAppContext()

////////////////////////////////////////////////////////////////////////////

type AppContextImpl struct {
	database        mo.Option[*db.Database]
	playbackManager mo.Option[*playbackmanager.PlaybackManager]
	anilistPlatform mo.Option[platform.Platform]
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
}
