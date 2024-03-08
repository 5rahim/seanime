package playbackmanager_test

import (
	"context"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/db"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/playbackmanager"
	"github.com/seanime-app/seanime/internal/util"
)

func getPlaybackManager() (*playbackmanager.PlaybackManager, *anilist.ClientWrapper, *anilist.AnimeCollection, error) {
	logger := util.NewLogger()
	wsEventManager := events.NewMockWSEventManager(logger)
	databaseInfo := db.GetTestDatabaseInfo()
	database, err := db.NewDatabase(databaseInfo.DataDir, databaseInfo.Name, logger)
	if err != nil {
		return nil, nil, nil, err
	}
	_, anilistClientWrapper, userData := anilist.MockAnilistClientWrappers()

	anilistCollection, err := anilistClientWrapper.Client.AnimeCollection(context.Background(), &userData.Username2)
	if err != nil {
		return nil, nil, nil, err
	}

	return playbackmanager.New(&playbackmanager.NewProgressManagerOptions{
		Logger:               logger,
		WSEventManager:       wsEventManager,
		AnilistClientWrapper: anilistClientWrapper,
		Database:             database,
		AnilistCollection:    anilistCollection,
		RefreshAnilistCollectionFunc: func() {
			// Do nothing
		},
	}), anilistClientWrapper, anilistCollection, nil
}
