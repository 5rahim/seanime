package transcoder

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"time"
)

func printExecTime(logger *zerolog.Logger, message string, args ...any) func() {
	msg := fmt.Sprintf(message, args...)
	start := time.Now()
	logger.Trace().Msgf("transcoder: Running %s", msg)

	return func() {
		logger.Trace().Msgf("transcoder: %s finished in %s", msg, time.Since(start))
	}
}

func GetHash(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	h := sha1.New()
	h.Write([]byte(path))
	h.Write([]byte(info.ModTime().String()))
	sha := hex.EncodeToString(h.Sum(nil))
	return sha, nil
}
