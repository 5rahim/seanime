package anilist

import (
	"context"
	"errors"
)

// UpdateMediaListEntryProgress is a wrapper around Client.UpdateMediaListEntryProgress.
// It updates the progress of a media list entry. If the progress is equal to the total episodes, the status will be set to "completed".
func (cw *ClientWrapper) UpdateMediaListEntryProgress(ctx context.Context, mediaId *int, progress *int, totalEpisodes *int) error {

	if mediaId == nil || progress == nil || totalEpisodes == nil {
		return errors.New("missing fields")
	}

	status := MediaListStatusCurrent
	if *progress == *totalEpisodes {
		status = MediaListStatusCompleted
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

	return nil
}
