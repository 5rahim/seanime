package handlers

import (
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime/internal/constants"
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

// HandleMalAuth
// POST /mal/auth
// Fetches the access and refresh token
func HandleMalAuth(c *RouteCtx) error {

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

// HandleMalLogout
// POST /mal/logout
func HandleMalLogout(c *RouteCtx) error {

	err := c.App.Database.DeleteMalInfo()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}
