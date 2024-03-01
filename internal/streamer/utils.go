package streamer

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"
)

func printExecTime(message string, args ...any) func() {
	msg := fmt.Sprintf(message, args...)
	start := time.Now()
	log.Printf("Running %s", msg)

	return func() {
		log.Printf("%s finished in %s", msg, time.Since(start))
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
