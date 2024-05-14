package torrentstream

import (
	"context"
	"fmt"
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	itorrent "github.com/seanime-app/seanime/internal/torrents/torrent"
)

type StartStreamOptions struct {
	MediaId       int                   `json:"mediaId"`
	EpisodeNumber int                   `json:"episodeNumber"` // RELATIVE Episode number to identify the file
	AniDBEpisode  string                `json:"aniDBEpisode"`  // Anizip episode
	AutoSelect    bool                  `json:"autoSelect"`    // Automatically select the best file to stream
	Torrent       itorrent.AnimeTorrent `json:"torrent"`       // Selected torrent
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

	episodeNumber := opts.EpisodeNumber

	//
	// Find the best torrent / Select the torrent
	//
	var torrentToStream *playbackTorrent
	switch opts.AutoSelect {
	case true:
		torrentToStream, err = r.findBestTorrent(media, anizipMedia, episode, episodeNumber)
		if err != nil {
			return err
		}
	case false:
		torrentToStream, err = r.findBestTorrentFromManualSelection(opts.Torrent.Link, media, episodeNumber)
		if err != nil {
			return err
		}
	}

	//
	// Set current file
	//
	r.playback.currentFile = mo.Some(torrentToStream.File)
	r.playback.currentTorrent = mo.Some(torrentToStream.Torrent)

	//
	// Start the server
	//
	r.serverManager.StartServer()

	//
	// Start the stream
	//
	err = r.playbackManager.StartStreamingUsingMediaPlayer(r.client.GetStreamingUrl())
	if err != nil {
		r.logger.Error().Err(err).Msg("torrentstream: Failed to start the stream")
		return err
	}

	r.logger.Info().Msg("torrentstream: Stream started")

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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
		return nil, nil, fmt.Errorf("torrentstream: Could not fetch AniDB media: %w", err)
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
