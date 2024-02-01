package handlers

import (
	"errors"
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime/internal/constants"
	"github.com/seanime-app/seanime/internal/mal"
	"github.com/seanime-app/seanime/internal/models"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type malAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int32  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// HandleMALAuth
// POST /mal/auth
// Fetches the access and refresh token
func HandleMALAuth(c *RouteCtx) error {

	type body struct {
		Code         string
		State        string
		CodeVerifier string `json:"code_verifier"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
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
		return c.RespondWithError(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(urlData.Encode())))

	// Response
	res, err := client.Do(req)
	if err != nil {
		return c.RespondWithError(err)
	}
	defer res.Body.Close()

	ret := malAuthResponse{}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return c.RespondWithError(err)
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

	_, err = c.App.Database.UpsertMalInfo(&malInfo)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(ret)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandleEditMALListEntryProgress
// POST /v1/mal/list-entry/progress
func HandleEditMALListEntryProgress(c *RouteCtx) error {

	type body struct {
		MediaId  *int `json:"mediaId"`
		Progress *int `json:"progress"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	if b.MediaId == nil || b.Progress == nil {
		return c.RespondWithError(errors.New("mediaId and progress is required"))
	}

	// Get MAL info
	_malInfo, err := c.App.Database.GetMalInfo()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Verify MAL auth
	malInfo, err := verifyMALAuth(_malInfo, c)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Get anime details
	anime, err := mal.GetAnimeDetails(malInfo.AccessToken, *b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	status := mal.MediaListStatusWatching
	if anime.Status == mal.MediaStatusFinishedAiring && anime.NumEpisodes == *b.Progress {
		status = mal.MediaListStatusCompleted
	}

	// Update MAL list entry
	err = mal.UpdateAnimeListStatus(malInfo.AccessToken, &mal.AnimeListStatusParams{
		Status:             &status,
		NumWatchedEpisodes: b.Progress,
	}, *b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.Logger.Debug().Msgf("mal: Updated MAL list entry for mediaId %d", *b.MediaId)

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandleMALLogout
// POST /mal/logout
func HandleMALLogout(c *RouteCtx) error {

	err := c.App.Database.DeleteMalInfo()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// VerifyMALAuth will check if the MAL token has expired and refresh it if necessary.
// It will return the updated MAL info if a refresh was necessary.
func verifyMALAuth(malInfo *models.Mal, c *RouteCtx) (*models.Mal, error) {

	// Token has not expired
	if malInfo.TokenExpiresAt.After(time.Now()) {
		return malInfo, nil
	}

	// Token is expired, refresh it

	client := &http.Client{}

	// Build URL
	urlData := url.Values{}
	urlData.Set("grant_type", "refresh_token")
	urlData.Set("refresh_token", malInfo.RefreshToken)
	encodedData := urlData.Encode()

	req, err := http.NewRequest("POST", "https://myanimelist.net/v1/oauth2/token", strings.NewReader(encodedData))
	if err != nil {
		return malInfo, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+malInfo.AccessToken)

	// Response
	res, err := client.Do(req)
	if err != nil {
		return malInfo, err
	}
	defer res.Body.Close()

	ret := malAuthResponse{}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return malInfo, err
	}

	// Save
	updatedMalInfo := models.Mal{
		BaseModel: models.BaseModel{
			ID:        1,
			UpdatedAt: time.Now(),
		},
		Username:       "",
		AccessToken:    ret.AccessToken,
		RefreshToken:   ret.RefreshToken,
		TokenExpiresAt: time.Now().Add(time.Duration(ret.ExpiresIn) * time.Second),
	}

	_, err = c.App.Database.UpsertMalInfo(&updatedMalInfo)
	if err != nil {
		return malInfo, err
	}

	return &updatedMalInfo, nil
}
