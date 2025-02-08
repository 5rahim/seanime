package transcoder

import "os"

func GetEnvOr(env string, def string) string {
	out := os.Getenv(env)
	if out == "" {
		return def
	}
	return out
}

type HwAccelSettings struct {
	Name          string   `json:"name"`
	DecodeFlags   []string `json:"decodeFlags"`
	EncodeFlags   []string `json:"encodeFlags"`
	ScaleFilter   string   `json:"scaleFilter"`
	WithForcedIdr bool     `json:"removeForcedIdr"`
}
