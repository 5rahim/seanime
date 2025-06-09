package playbackmanager

import (
	"context"
	"fmt"
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/anime"

	"github.com/samber/lo"
)

type StartRandomVideoOptions struct {
	UserAgent string
	ClientId  string
}

// StartRandomVideo starts a random video from the collection.
// Note that this might now be suited if the user has multiple seasons of the same anime.
func (pm *PlaybackManager) StartRandomVideo(opts *StartRandomVideoOptions) error {
	pm.playlistHub.reset()
	if err := pm.checkOrLoadAnimeCollection(); err != nil {
		return err
	}

	animeCollection, err := pm.platform.GetAnimeCollection(context.Background(), false)
	if err != nil {
		return err
	}

	//
	// Retrieve random episode
	//

	// Get lfs
	lfs, _, err := db_bridge.GetLocalFiles(pm.Database)
	if err != nil {
		return fmt.Errorf("error getting local files: %s", err.Error())
	}

	// Create a local file wrapper
	lfw := anime.NewLocalFileWrapper(lfs)
	// Get entries (grouped by media id)
	lfEntries := lfw.GetLocalEntries()
	lfEntries = lo.Filter(lfEntries, func(e *anime.LocalFileWrapperEntry, _ int) bool {
		return e.HasMainLocalFiles()
	})
	if len(lfEntries) == 0 {
		return fmt.Errorf("no playable media found")
	}

	continueLfs := make([]*anime.LocalFile, 0)
	otherLfs := make([]*anime.LocalFile, 0)
	for _, e := range lfEntries {
		anilistEntry, ok := animeCollection.GetListEntryFromAnimeId(e.GetMediaId())
		if !ok {
			continue
		}
		progress := 0
		if anilistEntry.Progress != nil {
			progress = *anilistEntry.Progress
		}
		if anilistEntry.Status == nil || *anilistEntry.Status == "COMPLETED" {
			continue
		}
		firstUnwatchedFile, found := e.GetFirstUnwatchedLocalFiles(progress)
		if !found {
			continue
		}
		if *anilistEntry.Status == "CURRENT" || *anilistEntry.Status == "REPEATING" {
			continueLfs = append(continueLfs, firstUnwatchedFile)
		} else {
			otherLfs = append(otherLfs, firstUnwatchedFile)
		}
	}

	if len(continueLfs) == 0 && len(otherLfs) == 0 {
		return fmt.Errorf("no playable file found")
	}

	lfs = append(continueLfs, otherLfs...)
	// only choose from continueLfs if there are more than 8 episodes
	if len(continueLfs) > 8 {
		lfs = continueLfs
	}

	lfs = lo.Shuffle(lfs)

	err = pm.StartPlayingUsingMediaPlayer(&StartPlayingOptions{
		Payload:   lfs[0].GetPath(),
		UserAgent: opts.UserAgent,
		ClientId:  opts.ClientId,
	})
	if err != nil {
		return err
	}

	return nil
}
