package anilist

import (
	"context"
	"errors"
)

// UpdateMediaListEntryProgress is a wrapper around Client.UpdateMediaListEntryProgress.
// It updates the progress of a media list entry. If the progress is equal to the total episodes, the status will be set to "completed".
func (cw *ClientWrapper) UpdateMediaListEntryProgress(ctx context.Context, mediaId *int, progress *int, totalEpisodes *int) error {

	if mediaId == nil || progress == nil {
		return errors.New("missing fields")
	}

	totalEp := 0
	if totalEpisodes != nil && *totalEpisodes > 0 {
		totalEp = *totalEpisodes
	}

	status := MediaListStatusCurrent
	if totalEp > 0 && *progress >= totalEp {
		status = MediaListStatusCompleted
	}

	if totalEp > 0 && *progress > totalEp {
		*progress = totalEp
	}

	// Update the progress
	_, err := cw.Client.UpdateMediaListEntryProgress(
		ctx,
		mediaId,
		progress,
		&status,
	)
	if err != nil {
		return err
	}

	cw.logger.Debug().Msgf("anilist: Updated media list entry for mediaId %d", *mediaId)

	return nil
}
