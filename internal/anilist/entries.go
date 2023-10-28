package anilist

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime-server/internal/limiter"
)

func (c *Client) AddMediaToPlanning(mId []int, rateLimiter *limiter.Limiter, logger *zerolog.Logger) error {
	if len(mId) == 0 {
		logger.Info().Msg("[anilist] no media added to planning list")
	}
	if rateLimiter == nil {
		return errors.New("[anilist] no rate limiter provided")
	}

	status := MediaListStatusPlanning

	lo.ForEach(mId, func(id int, index int) {
		rateLimiter.Wait()
		_, err := c.UpdateEntry(
			context.Background(),
			&id,
			&status,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
		)
		if err != nil {
			logger.Error().Msg("[anilist] error while adding media to plannig list: " + err.Error())
		}
	})

	return nil
}
