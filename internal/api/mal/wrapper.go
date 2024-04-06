package mal

import (
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"net/http"
	"strings"
)

const (
	ApiBaseURL string = "https://api.myanimelist.net/v2"
)

type (
	Wrapper struct {
		AccessToken string
		client      *http.Client
	}
)

func NewWrapper(accessToken string) *Wrapper {
	return &Wrapper{
		AccessToken: accessToken,
		client:      &http.Client{},
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
