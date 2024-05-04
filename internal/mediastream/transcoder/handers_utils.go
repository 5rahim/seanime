package transcoder

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"strings"
	"time"
)

var client = &http.Client{Timeout: 10 * time.Second}

type Item struct {
	Path string `json:"path"`
}

func GetPath(c *fiber.Ctx) (string, error) {
	vals, ok := c.GetReqHeaders()["X-Route"]
	if !ok || vals[0] == "" {
		return "", errors.New("missing route. Please specify the X-Route header to a valid route")
	}
	return vals[0], nil
}

func GetRoute(c *fiber.Ctx) string {
	vals, ok := c.GetReqHeaders()["X-Route"]
	if !ok {
		return ""
	}
	return vals[0]
}

func SanitizePath(path string) error {
	if strings.Contains(path, "/") || strings.Contains(path, "..") {
		return errors.New("invalid parameter. Can't contains path delimiters or …")
	}
	return nil
}

func GetClientId(c *fiber.Ctx) (string, error) {
	return "1", nil
	//vals, ok := c.GetReqHeaders()["X-CLIENT-ID"]
	//if !ok || vals[0] == "" {
	//	return "", errors.New("missing client id. Please specify the X-CLIENT-ID header to a guid constant for the lifetime of the player (but unique per instance)")
	//}
	//return vals[0], nil
}

func ParseSegment(segment string) (int32, error) {
	var ret int32
	_, err := fmt.Sscanf(segment, "segment-%d.ts", &ret)
	if err != nil {
		return 0, errors.New("could not parse segment")
	}
	return ret, nil
}
