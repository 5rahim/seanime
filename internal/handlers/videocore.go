package handlers

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/labstack/echo/v4"
)

// HandleVideoCoreInSightGetCharacterDetails
//
//	@summary returns the character details.
//	@param malId - int - true - "The MAL character ID"
//	@returns videocore.InSightCharacterDetails
//	@route /api/v1/videocore/insight/character/{malId} [GET]
func (h *Handler) HandleVideoCoreInSightGetCharacterDetails(c echo.Context) error {
	malId, err := strconv.Atoi(c.Param("malId"))
	if err != nil {
		return h.RespondWithError(c, err)
	}

	ret, err := h.App.VideoCore.InSight().GetCharacterInfo(malId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, ret)
}

// HandleVideoCoreSaveScreenshot
//
//	@summary saves a screenshot to a local directory.
//	@route /api/v1/videocore/screenshot [POST]
//	@returns bool
func (h *Handler) HandleVideoCoreSaveScreenshot(c echo.Context) error {
	type body struct {
		Dir        string `json:"dir"`
		Filename   string `json:"filename"`
		Base64Data string `json:"base64Data"`
	}

	var req body
	if err := c.Bind(&req); err != nil {
		return h.RespondWithError(c, err)
	}

	if req.Dir == "" || req.Filename == "" || req.Base64Data == "" {
		return h.RespondWithError(c, fmt.Errorf("missing required fields"))
	}

	settings, err := h.App.Database.GetSettings()
	if err != nil {
		return h.RespondWithError(c, err)
	}
	if err := h.guardPrivilegedMediaPlayer(c, settings); err != nil {
		return err
	}
	if err := h.guardStrictFilesystemPath(c, req.Dir); err != nil {
		return err
	}

	data, err := base64.StdEncoding.DecodeString(req.Base64Data)
	if err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to decode base64 data: %w", err))
	}

	if err := os.MkdirAll(req.Dir, 0755); err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to create directory: %w", err))
	}

	filePath := filepath.Join(req.Dir, req.Filename)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to write file: %w", err))
	}

	return h.RespondWithData(c, true)
}
