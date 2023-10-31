package handlers

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/seanime-app/seanime-server/internal/models"
	"time"
)

type SettingsBody struct {
	Library     models.LibrarySettings     `json:"library"`
	MediaPlayer models.MediaPlayerSettings `json:"mediaPlayer"`
	Torrent     models.TorrentSettings     `json:"torrent"`
}

func HandleSaveSettings(c *RouteCtx) error {

	body := new(SettingsBody)

	if err := c.Fiber.BodyParser(body); err != nil {
		return c.RespondWithError(err)
	}

	settings, err := c.App.Database.UpsertSettings(&models.Settings{
		BaseModel: models.BaseModel{
			ID:        1,
			UpdatedAt: time.Now(),
		},
		Library:     &body.Library,
		MediaPlayer: &body.MediaPlayer,
		Torrent:     &body.Torrent,
	})

	if err != nil {
		return c.RespondWithError(err)
	}

	err = c.App.WebsocketConn.WriteMessage(websocket.TextMessage, []byte("Settings updated"))
	if err != nil {
		c.App.Logger.Error().Err(err).Msg("Failed to send message to websocket")
	}

	// Refresh the settings dependents
	c.App.InitOrRefreshDependencies()

	return c.RespondWithData(settings)
}
