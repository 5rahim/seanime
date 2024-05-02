package transcoder

import (
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/mediastream/videofile"
	"io"
	"os"
	"path/filepath"
)

func Map[T, U any](ts []T, f func(T, int) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i], i)
	}
	return us
}

func GetInfo(path string, logger *zerolog.Logger, settings *Settings) (*videofile.MediaInfo, error) {
	defer printExecTime(logger, "mediainfo for %s", path)()

	me, err := videofile.NewMediaInfoExtractor(path, logger)
	if err != nil {
		return nil, err
	}

	ret, err := me.GetInfo(settings.MetadataDir)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

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
	return os.WriteFile(savePath, content, 0o644)
}
