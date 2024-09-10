package util

import (
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
	"strings"
)

func Proxy(c *fiber.Ctx) error {
	url := c.Query("url")
	headers := c.Query("headers")

	client := &http.Client{}

	req, err := http.NewRequest(c.Method(), url, nil)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	if headers != "" {
		headerList := strings.Split(headers, ",")
		for _, header := range headerList {
			parts := strings.SplitN(header, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				req.Header.Set(key, value)
			}
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	defer resp.Body.Close()

	for k, vs := range resp.Header {
		for _, v := range vs {
			c.Set(k, v)
		}
	}

	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.Send(body)
}
