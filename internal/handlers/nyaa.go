package handlers

import (
	"errors"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/nyaa"
)

func HandleNyaaSearch(c *RouteCtx) error {

	type body struct {
		Query          string             `json:"query"`
		EpisodeNumber  *int               `json:"episodeNumber"`
		Batch          *bool              `json:"batch"`
		Media          *anilist.BaseMedia `json:"media"`
		AbsoluteOffset *int               `json:"absoluteOffset"`
		Quality        *string            `json:"quality"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	ret := make([]nyaa.Torrent, 0)

	if len(b.Query) == 0 {
		query, ok := nyaa.BuildSearchQuery(&nyaa.BuildSearchQueryOptions{
			Media:          b.Media,
			Batch:          b.Batch,
			EpisodeNumber:  b.EpisodeNumber,
			Quality:        b.Quality,
			AbsoluteOffset: b.AbsoluteOffset,
		})
		if !ok {
			return c.RespondWithError(errors.New("could not build search query"))
		}
		res, err := nyaa.Search(nyaa.SearchOptions{
			Provider: "nyaa",
			Query:    query,
			Category: "anime",
			SortBy:   "downloads",
			Filter:   "",
		})
		if err != nil {
			return c.RespondWithError(err)
		}
		ret = res
	} else {
		res, err := nyaa.Search(nyaa.SearchOptions{
			Provider: "nyaa",
			Query:    b.Query,
			Category: "anime",
			SortBy:   "downloads",
			Filter:   "",
		})
		if err != nil {
			return c.RespondWithError(err)
		}
		ret = res
	}

	return c.RespondWithData(ret)

}
