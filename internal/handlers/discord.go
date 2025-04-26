package handlers

import (
	discordrpc_presence "seanime/internal/discordrpc/presence"

	"github.com/labstack/echo/v4"
)

// HandleSetDiscordMangaActivity
//
//	@summary sets manga activity for discord rich presence.
//	@route /api/v1/discord/presence/manga [POST]
//	@returns bool
func (h *Handler) HandleSetDiscordMangaActivity(c echo.Context) error {

	type body struct {
		MediaId int    `json:"mediaId"`
		Title   string `json:"title"`
		Image   string `json:"image"`
		Chapter string `json:"chapter"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		h.App.Logger.Error().Err(err).Msg("discord rpc handler: failed to parse request body")
		return h.RespondWithData(c, false)
	}

	h.App.DiscordPresence.SetMangaActivity(&discordrpc_presence.MangaActivity{
		ID:      b.MediaId,
		Title:   b.Title,
		Image:   b.Image,
		Chapter: b.Chapter,
	})

	return h.RespondWithData(c, true)
}

// HandleSetDiscordLegacyAnimeActivity
//
//	@summary sets anime activity for discord rich presence.
//	@route /api/v1/discord/presence/legacy-anime [POST]
//	@returns bool
func (h *Handler) HandleSetDiscordLegacyAnimeActivity(c echo.Context) error {

	type body struct {
		MediaId       int    `json:"mediaId"`
		Title         string `json:"title"`
		Image         string `json:"image"`
		IsMovie       bool   `json:"isMovie"`
		EpisodeNumber int    `json:"episodeNumber"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		h.App.Logger.Error().Err(err).Msg("discord rpc handler: failed to parse request body")
		return h.RespondWithData(c, false)
	}

	h.App.DiscordPresence.LegacySetAnimeActivity(&discordrpc_presence.LegacyAnimeActivity{
		ID:            b.MediaId,
		Title:         b.Title,
		Image:         b.Image,
		IsMovie:       b.IsMovie,
		EpisodeNumber: b.EpisodeNumber,
	})

	return h.RespondWithData(c, true)
}

// HandleSetDiscordAnimeActivityWithProgress
//
//	@summary sets anime activity for discord rich presence with progress.
//	@route /api/v1/discord/presence/anime [POST]
//	@returns bool
func (h *Handler) HandleSetDiscordAnimeActivityWithProgress(c echo.Context) error {

	type body struct {
		MediaId             int     `json:"mediaId"`
		Title               string  `json:"title"`
		Image               string  `json:"image"`
		IsMovie             bool    `json:"isMovie"`
		EpisodeNumber       int     `json:"episodeNumber"`
		Progress            int     `json:"progress"`
		Duration            int     `json:"duration"`
		TotalEpisodes       *int    `json:"totalEpisodes,omitempty"`
		CurrentEpisodeCount *int    `json:"currentEpisodeCount,omitempty"`
		EpisodeTitle        *string `json:"episodeTitle,omitempty"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		h.App.Logger.Error().Err(err).Msg("discord rpc handler: failed to parse request body")
		return h.RespondWithData(c, false)
	}

	h.App.DiscordPresence.SetAnimeActivity(&discordrpc_presence.AnimeActivity{
		ID:                  b.MediaId,
		Title:               b.Title,
		Image:               b.Image,
		IsMovie:             b.IsMovie,
		EpisodeNumber:       b.EpisodeNumber,
		Progress:            b.Progress,
		Duration:            b.Duration,
		TotalEpisodes:       b.TotalEpisodes,
		CurrentEpisodeCount: b.CurrentEpisodeCount,
		EpisodeTitle:        b.EpisodeTitle,
	})

	return h.RespondWithData(c, true)
}

// HandleUpdateDiscordAnimeActivityWithProgress
//
//	@summary updates the anime activity for discord rich presence with progress.
//	@route /api/v1/discord/presence/anime-update [POST]
//	@returns bool
func (h *Handler) HandleUpdateDiscordAnimeActivityWithProgress(c echo.Context) error {

	type body struct {
		Progress int  `json:"progress"`
		Duration int  `json:"duration"`
		Paused   bool `json:"paused"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		h.App.Logger.Error().Err(err).Msg("discord rpc handler: failed to parse request body")
		return h.RespondWithData(c, false)
	}

	h.App.DiscordPresence.UpdateAnimeActivity(b.Progress, b.Duration, b.Paused)
	return h.RespondWithData(c, true)
}

// HandleCancelDiscordActivity
//
//	@summary cancels the current discord rich presence activity.
//	@route /api/v1/discord/presence/cancel [POST]
//	@returns bool
func (h *Handler) HandleCancelDiscordActivity(c echo.Context) error {
	h.App.DiscordPresence.Close()
	return h.RespondWithData(c, true)
}
