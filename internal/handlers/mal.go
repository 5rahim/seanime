package handlers

import (
	"errors"
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime/internal/api/mal"
	"github.com/seanime-app/seanime/internal/constants"
	"github.com/seanime-app/seanime/internal/database/models"
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

// HandleMALAuth will fetch the access and refresh tokens for the given code.
//
// This is used to authenticate the user with MyAnimeList.
// It will save the info in the database, effectively logging the user in.
//
// A new Status should be fetched after this.
//
//	POST /v1/mal/auth
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

// HandleEditMALListEntryProgress will update the progress of a MAL list entry.
//
//	POST /v1/mal/list-entry/progress
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

	// Get MAL Wrapper
	malWrapper := mal.NewWrapper(malInfo.AccessToken)

	// Update MAL list entry
	err = malWrapper.UpdateAnimeProgress(&mal.AnimeListProgressParams{
		NumEpisodesWatched: b.Progress,
	}, *b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.Logger.Debug().Msgf("mal: Updated MAL list entry for mediaId %d", *b.MediaId)

	return c.RespondWithData(true)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandleMALLogout will delete the MAL info from the database, effectively logging the user out.
// A new Status should be fetched after this.
//
//	POST /mal/logout
func HandleMALLogout(c *RouteCtx) error {

	err := c.App.Database.DeleteMalInfo()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// verifyMALAuth will check if the MAL token has expired and refresh it if necessary.
// It will return the updated MAL info if a refresh was necessary or the current MAL info if it wasn't.
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
