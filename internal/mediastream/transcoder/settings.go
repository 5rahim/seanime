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
	Name        string
	DecodeFlags []string
	EncodeFlags []string
	ScaleFilter string
}
