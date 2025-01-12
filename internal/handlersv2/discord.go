package handlersv2

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

// HandleCancelDiscordActivity
//
//	@summary cancels the current discord rich presence activity.
//	@route /api/v1/discord/presence/cancel [POST]
//	@returns bool
func (h *Handler) HandleCancelDiscordActivity(c echo.Context) error {
	h.App.DiscordPresence.Close()
	return h.RespondWithData(c, true)
}
