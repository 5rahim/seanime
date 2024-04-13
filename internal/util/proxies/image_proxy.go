package util

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
)

type ImageProxy struct{}

func (ip *ImageProxy) GetImage(url string, headers map[string]string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
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

func (ip *ImageProxy) setHeaders(c *fiber.Ctx) {
	c.Set("Content-Type", "image/jpeg")
	c.Set("Cache-Control", "public, max-age=31536000")
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Access-Control-Allow-Methods", "GET")
	c.Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	c.Set("Access-Control-Allow-Credentials", "true")
}

func (ip *ImageProxy) ProxyImage(c *fiber.Ctx) error {
	url := c.Query("url")
	headersJSON := c.Query("headers")

	if url == "" || headersJSON == "" {
		return c.Status(fiber.StatusBadRequest).SendString("No URL provided")
	}

	headers := make(map[string]string)
	err := json.Unmarshal([]byte(headersJSON), &headers)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Error parsing headers JSON")
	}

	ip.setHeaders(c)
	imageBuffer, err := ip.GetImage(url, headers)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error fetching image")
	}

	return c.Send(imageBuffer)
}
