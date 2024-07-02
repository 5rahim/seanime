package playbackmanager

import (
	"errors"
	"fmt"
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/library/anime"
	"path/filepath"
	"strings"
)

// GetCurrentMediaID returns the media id of the currently playing media
func (pm *PlaybackManager) GetCurrentMediaID() (int, error) {
	if pm.currentLocalFile.IsAbsent() {
		return 0, errors.New("no media is currently playing")
	}
	return pm.currentLocalFile.MustGet().MediaId, nil
}

// getLocalFilePlaybackDetails is called once everytime a new video is played. It returns the anilist entry, local file and local file wrapper entry.
func (pm *PlaybackManager) getLocalFilePlaybackDetails(path string) (*anilist.MediaListEntry, *anime.LocalFile, *anime.LocalFileWrapperEntry, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	// Normalize path
	path = filepath.ToSlash(strings.ToLower(path))

	// Find the local file from the path
	lfs, _, err := pm.Database.GetLocalFiles()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting local files: %s", err.Error())
	}

	var lf *anime.LocalFile
	// Find the local file from the path
	for _, l := range lfs {
		if l.GetNormalizedPath() == path {
			lf = l
			break
		}
	}
	// If the local file is not found, the path might be a filename (in the case of VLC)
	if lf == nil {
		for _, l := range lfs {
			if strings.ToLower(l.Name) == path {
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

	ret, ok := pm.animeCollection.GetListEntryFromMediaId(lf.MediaId)
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

// getStreamPlaybackDetails is called once everytime a new video is played.
func (pm *PlaybackManager) getStreamPlaybackDetails(mId int) mo.Option[*anilist.MediaListEntry] {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	ret, ok := pm.animeCollection.GetListEntryFromMediaId(mId)
	if !ok {
		return mo.None[*anilist.MediaListEntry]()
	}

	return mo.Some(ret)
}
