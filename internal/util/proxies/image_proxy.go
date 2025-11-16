package util

import (
	"encoding/json"
	"io"
	"net/http"
	"seanime/internal/util"

	"github.com/imroc/req/v3"
	"github.com/labstack/echo/v4"
)

type ImageProxy struct{}

func (ip *ImageProxy) GetImage(url string, headers map[string]string) ([]byte, error) {
	request := req.C().NewRequest()

	for key, value := range headers {
		request.SetHeader(key, value)
	}

	resp, err := request.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (ip *ImageProxy) setHeaders(c echo.Context) {
	c.Set("Content-Type", "image/jpeg")
	c.Set("Cache-Control", "public, max-age=31536000")
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Access-Control-Allow-Methods", "GET")
	c.Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	c.Set("Access-Control-Allow-Credentials", "true")
}

func (ip *ImageProxy) ProxyImage(c echo.Context) (err error) {
	defer util.HandlePanicInModuleWithError("util/ImageProxy", &err)

	url := c.QueryParam("url")
	headersJSON := c.QueryParam("headers")

	if url == "" || headersJSON == "" {
		return c.String(echo.ErrBadRequest.Code, "No URL provided")
	}

	headers := make(map[string]string)
	if err := json.Unmarshal([]byte(headersJSON), &headers); err != nil {
		return c.String(echo.ErrBadRequest.Code, "Error parsing headers JSON")
	}

	ip.setHeaders(c)
	imageBuffer, err := ip.GetImage(url, headers)
	if err != nil {
		return c.String(echo.ErrInternalServerError.Code, "Error fetching image")
	}

	return c.Blob(http.StatusOK, c.Response().Header().Get("Content-Type"), imageBuffer)
}
