package torrentstream

import (
	"context"
	"fmt"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
)

type StartStreamOptions struct {
	MediaId       int    `json:"mediaId"`
	EpisodeNumber int    `json:"episodeNumber"` // Episode number to identify the file
	AniDBEpisode  string `json:"aniDBEpisode"`  // Anizip episode
	AutoSelect    bool   `json:"autoSelect"`    // Automatically select the best file to stream
	TorrentID     string `json:"torrentId"`     // Magnet/File when manually selecting
}

// StartStream is called by the client to start streaming a torrent
func (r *Repository) StartStream(opts *StartStreamOptions) error {

	r.logger.Info().Int("mediaId", opts.MediaId).Msgf("torrentstream: Starting stream for episode %s", opts.AniDBEpisode)

	//
	// Get the media info
	//
	media, anizipMedia, err := r.getMediaInfo(opts.MediaId)
	if err != nil {
		return err
	}

	episode, err := r.getEpisodeInfo(anizipMedia, opts.AniDBEpisode)
	if err != nil {
		return err
	}

	//
	// Find the best torrent
	//
	var torrentId string
	switch opts.AutoSelect {
	case true:
		torrentId, err = r.findBestTorrent(media, anizipMedia, episode, opts.EpisodeNumber)
	case false:
		torrentId = opts.TorrentID
		if torrentId == "" {
			err = fmt.Errorf("torrentstream: No magnet link or torrent file provided")
		}
	}
	if err != nil {
		return err
	}

	_, err = r.client.AddTorrent(torrentId)
	if err != nil {
		return err
	}

	//
	//files := t.Files()
	//if len(files) == 0 {
	//	return errors.New("torrentstream: no files found in the torrent")
	//}
	//
	//spew.Dump(files)
	//
	//file := files[0] // TODO change
	//
	//r.playback.currentFile = mo.Some(file)

	return nil
}

func (r *Repository) getMediaInfo(mediaId int) (media *anilist.BaseMedia, anizipMedia *anizip.Media, err error) {
	// Get the media
	var found bool
	media, found = r.baseMediaCache.Get(mediaId)
	if !found {
		media, found = r.animeCollection.FindMedia(mediaId)
	}
	if !found {
		// Fetch the media
		var mediaF *anilist.BaseMediaByID
		mediaF, err = r.anilistClientWrapper.BaseMediaByID(context.Background(), &mediaId)
		if err != nil {
			return nil, nil, fmt.Errorf("torrentstream: failed to fetch media: %w", err)
		}
		media = mediaF.GetMedia()
	}

	// Get the media
	anizipMedia, err = anizip.FetchAniZipMediaC("anilist", mediaId, r.anizipCache)
	if err != nil {
		return nil, nil, fmt.Errorf("torrentstream: AniDB media not found in the cache")
	}

	return
}

func (r *Repository) getEpisodeInfo(anizipMedia *anizip.Media, aniDBEpisode string) (episode *anizip.Episode, err error) {
	// Get the episode
	var found bool
	episode, found = anizipMedia.FindEpisode(aniDBEpisode)
	if !found {
		return nil, fmt.Errorf("torrentstream: Episode not found in the Anizip media")
	}
	return
}
