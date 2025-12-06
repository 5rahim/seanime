package handlers

import (
	"net/http"
	"seanime/internal/database/db_bridge"
	"seanime/internal/directstream"

	"github.com/labstack/echo/v4"
)

// HandleDirectstreamPlayLocalFile
//
//	@summary request local file stream.
//	@desc This requests a local file stream and returns the media container to start the playback.
//	@returns mediastream.MediaContainer
//	@route /api/v1/directstream/play/localfile [POST]
func (h *Handler) HandleDirectstreamPlayLocalFile(c echo.Context) error {
	type body struct {
		Path     string `json:"path"`     // The path of the file.
		ClientId string `json:"clientId"` // The session id
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	lfs, _, err := db_bridge.GetLocalFiles(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.App.DirectStreamManager.PlayLocalFile(c.Request().Context(), directstream.PlayLocalFileOptions{
		ClientId:   b.ClientId,
		Path:       b.Path,
		LocalFiles: lfs,
	})
}

// HandleDirectstreamFetchAndConvertToASS
//
//	@summary converts subtitles to ASS.
//	@desc Subtitles will be fetched and converted to ASS.
//	@returns string
//	@route /api/v1/directstream/subs/convert-to-ass [POST]
func (h *Handler) HandleDirectstreamFetchAndConvertToASS(c echo.Context) error {
	type body struct {
		Url string `json:"url"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	ret, err := h.App.DirectStreamManager.FetchAndConvertToASS(b.Url)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, ret)
}

func (h *Handler) HandleDirectstreamGetStream() http.Handler {
	return h.App.DirectStreamManager.ServeEchoStream()
}

func (h *Handler) HandleDirectstreamGetAttachments(c echo.Context) error {
	return h.App.DirectStreamManager.ServeEchoAttachments(c)
}
