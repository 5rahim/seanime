//
// The transcoder package is based on the work of "zoriya/Kyoo" licensed under the GPL-3.0 License.
//

package transcoder

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"io"
	"os"
	"path/filepath"
	"time"
)

func getSavedInfo[T any](savePath string, mi *T) error {
	savedFile, err := os.Open(savePath)
	if err != nil {
		return err
	}
	saved, err := io.ReadAll(savedFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(saved, mi)
	if err != nil {
		return err
	}
	return nil
}

func saveInfo[T any](savePath string, mi *T) error {
	content, err := json.Marshal(*mi)
	if err != nil {
		return err
	}
	// create directory if it doesn't exist
	_ = os.MkdirAll(filepath.Dir(savePath), 0755)
	return os.WriteFile(savePath, content, 0666)
}

func printExecTime(logger *zerolog.Logger, message string, args ...any) func() {
	msg := fmt.Sprintf(message, args...)
	start := time.Now()
	logger.Trace().Msgf("transcoder: Running %s", msg)

	return func() {
		logger.Trace().Msgf("transcoder: %s finished in %s", msg, time.Since(start))
	}
}
