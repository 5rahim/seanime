package anilist

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"seanime/internal/util/limiter"
	"sync"
)

func (c *Client) AddMediaToPlanning(mIds []int, rateLimiter *limiter.Limiter, logger *zerolog.Logger) error {
	if len(mIds) == 0 {
		logger.Debug().Msg("anilist: No media added to planning list")
		return nil
	}
	if rateLimiter == nil {
		return errors.New("anilist: no rate limiter provided")
	}

	status := MediaListStatusPlanning

	scoreRaw := 0
	progress := 0

	wg := sync.WaitGroup{}
	for _, _id := range mIds {
		wg.Add(1)
		go func(id int) {
			rateLimiter.Wait()
			defer wg.Done()
			_, err := c.UpdateMediaListEntry(
				context.Background(),
				&id,
				&status,
				&scoreRaw,
				&progress,
				nil,
				nil,
			)
			if err != nil {
				logger.Error().Msg("anilist: An error occurred while adding media to planning list: " + err.Error())
			}
		}(_id)
	}
	wg.Wait()

	logger.Debug().Any("count", len(mIds)).Msg("anilist: Media added to planning list")

	return nil
}
