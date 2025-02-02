package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"seanime/internal/api/mal"
	"seanime/internal/constants"
	"seanime/internal/database/models"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/labstack/echo/v4"
)

type MalAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int32  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// HandleMALAuth
//
//	@summary fetches the access and refresh tokens for the given code.
//	@desc This is used to authenticate the user with MyAnimeList.
//	@desc It will save the info in the database, effectively logging the user in.
//	@desc The client should re-fetch the server status after this.
//	@route /api/v1/mal/auth [POST]
//	@returns handlers.MalAuthResponse
func (h *Handler) HandleMALAuth(c echo.Context) error {

	type body struct {
		Code         string `json:"code"`
		State        string `json:"state"`
		CodeVerifier string `json:"code_verifier"`
	}

	b := new(body)
	if err := c.Bind(b); err != nil {
		return h.RespondWithError(c, err)
	}

	client := &http.Client{}

	// Build URL
	urlData := url.Values{}
	urlData.Set("client_id", constants.MalClientId)
	urlData.Set("grant_type", "authorization_code")
	urlData.Set("code", b.Code)
	urlData.Set("code_verifier", b.CodeVerifier)
	encodedData := urlData.Encode()

	req, err := http.NewRequest("POST", "https://myanimelist.net/v1/oauth2/token", strings.NewReader(encodedData))
	if err != nil {
		return h.RespondWithError(c, err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(urlData.Encode())))

	// Response
	res, err := client.Do(req)
	if err != nil {
		return h.RespondWithError(c, err)
	}
	defer res.Body.Close()

	ret := MalAuthResponse{}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return h.RespondWithError(c, err)
	}

	// Save
	malInfo := models.Mal{
		BaseModel: models.BaseModel{
			ID:        1,
			UpdatedAt: time.Now(),
		},
		Username:       "",
		AccessToken:    ret.AccessToken,
		RefreshToken:   ret.RefreshToken,
		TokenExpiresAt: time.Now().Add(time.Duration(ret.ExpiresIn) * time.Second),
	}

	_, err = h.App.Database.UpsertMalInfo(&malInfo)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, ret)
}

// HandleEditMALListEntryProgress
//
//	@summary updates the progress of a MAL list entry.
//	@route /api/v1/mal/list-entry/progress [POST]
//	@returns bool
func (h *Handler) HandleEditMALListEntryProgress(c echo.Context) error {

	type body struct {
		MediaId  *int `json:"mediaId"`
		Progress *int `json:"progress"`
	}

	b := new(body)
	if err := c.Bind(b); err != nil {
		return h.RespondWithError(c, err)
	}

	if b.MediaId == nil || b.Progress == nil {
		return h.RespondWithError(c, errors.New("mediaId and progress is required"))
	}

	// Get MAL info
	_malInfo, err := h.App.Database.GetMalInfo()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Verify MAL auth
	malInfo, err := mal.VerifyMALAuth(_malInfo, h.App.Database, h.App.Logger)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Get MAL Wrapper
	malWrapper := mal.NewWrapper(malInfo.AccessToken, h.App.Logger)

	// Update MAL list entry
	err = malWrapper.UpdateAnimeProgress(&mal.AnimeListProgressParams{
		NumEpisodesWatched: b.Progress,
	}, *b.MediaId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	h.App.Logger.Debug().Msgf("mal: Updated MAL list entry for mediaId %d", *b.MediaId)

	return h.RespondWithData(c, true)
}

// HandleMALLogout
//
//	@summary logs the user out of MyAnimeList.
//	@desc This will delete the MAL info from the database, effectively logging the user out.
//	@desc The client should re-fetch the server status after this.
//	@route /api/v1/mal/logout [POST]
//	@returns bool
func (h *Handler) HandleMALLogout(c echo.Context) error {

	err := h.App.Database.DeleteMalInfo()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}
