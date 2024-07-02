package playbackmanager

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/library/anime"
)

// StartRandomVideo starts a random video from the collection.
func (pm *PlaybackManager) StartRandomVideo() error {
	pm.playlistHub.reset()
	if err := pm.checkOrLoadOfflineAnimeCollection(); err != nil {
		return err
	}

	//
	// Retrieve random episode
	//

	// Get lfs
	lfs, _, err := pm.Database.GetLocalFiles()
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
		anilistEntry, ok := pm.animeCollection.GetListEntryFromMediaId(e.GetMediaId())
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

	err = pm.MediaPlayerRepository.Play(lfs[0].GetPath())
	if err != nil {
		return err
	}

	pm.MediaPlayerRepository.StartTracking()

	return nil
}
