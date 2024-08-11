package mal

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"net/url"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"strings"
	"time"
)

const (
	ApiBaseURL string = "https://api.myanimelist.net/v2"
)

type (
	Wrapper struct {
		AccessToken string
		client      *http.Client
		logger      *zerolog.Logger
	}
)

func NewWrapper(accessToken string, logger *zerolog.Logger) *Wrapper {
	return &Wrapper{
		AccessToken: accessToken,
		client:      &http.Client{},
		logger:      logger,
	}
}

func (w *Wrapper) doQuery(method, uri string, body io.Reader, contentType string, data interface{}) error {
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Authorization", "Bearer "+w.AccessToken)

	// Make the HTTP request
	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !((resp.StatusCode >= 200) && (resp.StatusCode <= 299)) {
		return fmt.Errorf("invalid response status %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return err
	}
	return nil
}

func (w *Wrapper) doMutation(method, uri, encodedParams string) error {
	var reader io.Reader
	reader = nil
	if encodedParams != "" {
		reader = strings.NewReader(encodedParams)
	}

	req, err := http.NewRequest(method, uri, reader)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+w.AccessToken)

	// Make the HTTP request
	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !((resp.StatusCode >= 200) && (resp.StatusCode <= 299)) {
		return fmt.Errorf("invalid response status %s", resp.Status)
	}

	return nil
}

func VerifyMALAuth(malInfo *models.Mal, db *db.Database, logger *zerolog.Logger) (*models.Mal, error) {

	// Token has not expired
	if malInfo.TokenExpiresAt.After(time.Now()) {
		logger.Debug().Msg("mal: Token is still valid")
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
		logger.Error().Err(err).Msg("mal: Failed to create request")
		return malInfo, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+malInfo.AccessToken)

	// Response
	res, err := client.Do(req)
	if err != nil {
		logger.Error().Err(err).Msg("mal: Failed to refresh token")
		return malInfo, err
	}
	defer res.Body.Close()

	type malAuthResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int32  `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}

	ret := malAuthResponse{}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return malInfo, err
	}

	if ret.AccessToken == "" {
		logger.Error().Msgf("mal: Failed to refresh token %s", res.Status)
		return malInfo, fmt.Errorf("mal: Failed to refresh token %s", res.Status)
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

	_, err = db.UpsertMalInfo(&updatedMalInfo)
	if err != nil {
		logger.Error().Err(err).Msg("mal: Failed to save updated MAL info")
		return malInfo, err
	}

	logger.Info().Msg("mal: Refreshed token")

	return &updatedMalInfo, nil
}
