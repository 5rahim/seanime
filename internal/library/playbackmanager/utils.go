package playbackmanager

import (
	"errors"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db_bridge"
	"seanime/internal/hook"
	"seanime/internal/library/anime"
	"seanime/internal/util"
	"strings"

	"github.com/samber/mo"
)

// GetCurrentMediaID returns the media id of the currently playing media
func (pm *PlaybackManager) GetCurrentMediaID() (int, error) {
	if pm.currentLocalFile.IsAbsent() {
		return 0, errors.New("no media is currently playing")
	}
	return pm.currentLocalFile.MustGet().MediaId, nil
}

// GetLocalFilePlaybackDetails is called once everytime a new video is played. It returns the anilist entry, local file and local file wrapper entry.
func (pm *PlaybackManager) getLocalFilePlaybackDetails(path string) (*anilist.AnimeListEntry, *anime.LocalFile, *anime.LocalFileWrapperEntry, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	// Normalize path
	path = util.NormalizePath(path)

	pm.Logger.Debug().Str("path", path).Msg("playback manager: Getting local file playback details")

	// Find the local file from the path
	lfs, _, err := db_bridge.GetLocalFiles(pm.Database)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting local files: %s", err.Error())
	}

	reqEvent := &PlaybackLocalFileDetailsRequestedEvent{
		Path:                  path,
		LocalFiles:            lfs,
		AnimeListEntry:        &anilist.AnimeListEntry{},
		LocalFile:             &anime.LocalFile{},
		LocalFileWrapperEntry: &anime.LocalFileWrapperEntry{},
	}
	err = hook.GlobalHookManager.OnPlaybackLocalFileDetailsRequested().Trigger(reqEvent)
	if err != nil {
		return nil, nil, nil, err
	}
	lfs = reqEvent.LocalFiles // Override the local files

	// Default prevented, use the hook's details
	if reqEvent.DefaultPrevented {
		pm.Logger.Debug().Msg("playback manager: Local file details processing prevented by hook")
		if reqEvent.AnimeListEntry == nil || reqEvent.LocalFile == nil || reqEvent.LocalFileWrapperEntry == nil {
			return nil, nil, nil, errors.New("local file details not found")
		}
		return reqEvent.AnimeListEntry, reqEvent.LocalFile, reqEvent.LocalFileWrapperEntry, nil
	}

	var lf *anime.LocalFile
	// Find the local file from the path
	for _, l := range lfs {
		if l.GetNormalizedPath() == path {
			lf = l
			pm.Logger.Debug().Msg("playback manager: Local file found by path")
			break
		}
	}

	// If the local file is not found, the path might be a filename (in the case of VLC)
	if lf == nil {
		for _, l := range lfs {
			if strings.ToLower(l.Name) == path {
				pm.Logger.Debug().Msg("playback manager: Local file found by name")
				lf = l
				break
			}
		}
	}

	if lf == nil {
		return nil, nil, nil, errors.New("local file not found")
	}
	if lf.MediaId == 0 {
		return nil, nil, nil, errors.New("local file has not been matched")
	}

	if pm.animeCollection.IsAbsent() {
		return nil, nil, nil, fmt.Errorf("error getting anime collection: %w", err)
	}

	ret, ok := pm.animeCollection.MustGet().GetListEntryFromAnimeId(lf.MediaId)
	if !ok {
		return nil, nil, nil, errors.New("anilist list entry not found")
	}

	// Create local file wrapper
	lfw := anime.NewLocalFileWrapper(lfs)
	lfe, ok := lfw.GetLocalEntryById(lf.MediaId)
	if !ok {
		return nil, nil, nil, errors.New("local file wrapper entry not found")
	}

	return ret, lf, lfe, nil
}

// GetStreamPlaybackDetails is called once everytime a new video is played.
func (pm *PlaybackManager) getStreamPlaybackDetails(mId int) mo.Option[*anilist.AnimeListEntry] {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.animeCollection.IsAbsent() {
		return mo.None[*anilist.AnimeListEntry]()
	}

	reqEvent := &PlaybackStreamDetailsRequestedEvent{
		AnimeCollection: pm.animeCollection.MustGet(),
		MediaId:         mId,
		AnimeListEntry:  &anilist.AnimeListEntry{},
	}
	err := hook.GlobalHookManager.OnPlaybackStreamDetailsRequested().Trigger(reqEvent)
	if err != nil {
		return mo.None[*anilist.AnimeListEntry]()
	}

	if reqEvent.DefaultPrevented {
		pm.Logger.Debug().Msg("playback manager: Stream details processing prevented by hook")
		if reqEvent.AnimeListEntry == nil {
			return mo.None[*anilist.AnimeListEntry]()
		}
		return mo.Some(reqEvent.AnimeListEntry)
	}

	ret, ok := pm.animeCollection.MustGet().GetListEntryFromAnimeId(mId)
	if !ok {
		return mo.None[*anilist.AnimeListEntry]()
	}

	return mo.Some(ret)
}
