package handlers

import (
	"context"
	"github.com/seanime-app/seanime/internal/api/anizip"
)

func HandlePopulateTVDBEpisodes(c *RouteCtx) error {
	type body struct {
		MediaId int `json:"mediaId"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	anizipMedia, err := anizip.FetchAniZipMedia("anilist", b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	mediaF, err := c.App.AnilistClientWrapper.BasicMediaByID(context.Background(), &b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}
	media := mediaF.GetMedia()

	// Create media wrapper
	mw := c.App.MetadataProvider.NewMediaWrapper(media, anizipMedia)

	// Fetch episodes
	episodes, err := mw.GetTVDBEpisodes(true)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Respond
	return c.RespondWithData(episodes)
}

func HandleEmptyTVDBEpisodes(c *RouteCtx) error {
	type body struct {
		MediaId int `json:"mediaId"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	anizipMedia, err := anizip.FetchAniZipMedia("anilist", b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	mediaF, err := c.App.AnilistClientWrapper.BasicMediaByID(context.Background(), &b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}
	media := mediaF.GetMedia()

	// Create media wrapper
	mw := c.App.MetadataProvider.NewMediaWrapper(media, anizipMedia)

	// Empty TVDB episodes bucket
	err = mw.EmptyTVDBEpisodesBucket(b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Respond
	return c.RespondWithData(true)
}
