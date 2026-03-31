package handlers

import (
	"seanime/internal/core"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

type ServerConfigResponse struct {
	BaseURL string `json:"baseUrl"`
}

func (h *Handler) HandleGetServerConfig(c echo.Context) error {
	return h.RespondWithData(c, &ServerConfigResponse{BaseURL: h.App.Config.GetBaseURLPath()})
}

func (h *Handler) HandleSaveServerConfigBaseURL(c echo.Context) error {
	type body struct {
		BaseURL string `json:"baseUrl"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	normalized := core.NormalizeBaseURLPath(b.BaseURL)
	viper.Set("server.baseUrl", normalized)
	if err := viper.WriteConfig(); err != nil {
		return h.RespondWithError(c, err)
	}

	h.App.Config.Server.BaseURL = normalized

	return h.RespondWithData(c, true)
}
