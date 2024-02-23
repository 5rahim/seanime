package handlers

import (
	"errors"
	"gorm.io/gorm"
	"strconv"
)

// HandleGetMediaEntrySilenceStatus will return the silence status of a media entry.
//
//	GET /v1/library/media-entry/silence/:id
func HandleGetMediaEntrySilenceStatus(c *RouteCtx) error {
	mId, err := strconv.Atoi(c.Fiber.Params("id"))
	if err != nil {
		return c.RespondWithError(errors.New("invalid id"))
	}

	mediaEntry, err := c.App.Database.GetSilencedMediaEntry(uint(mId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.RespondWithData(false)
		} else {
			return c.RespondWithError(err)
		}
	}

	return c.RespondWithData(mediaEntry)
}

// HandleToggleMediaEntrySilenceStatus will toggle the silence status of a media entry.
//
// The status should be re-fetched after this.
//
//	POST /v1/library/media-entry/silence
func HandleToggleMediaEntrySilenceStatus(c *RouteCtx) error {

	type body struct {
		MediaId int `json:"mediaId"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	mediaEntry, err := c.App.Database.GetSilencedMediaEntry(uint(b.MediaId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = c.App.Database.InsertSilencedMediaEntry(uint(b.MediaId))
			if err != nil {
				return c.RespondWithError(err)
			}
			return c.RespondWithData(true)
		} else {
			return c.RespondWithError(err)
		}
	}

	err = c.App.Database.DeleteSilencedMediaEntry(mediaEntry.ID)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}
