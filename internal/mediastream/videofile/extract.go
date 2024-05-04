package videofile

import (
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"os/exec"
	"path/filepath"
)

func GetFileSubsCacheDir(outDir string, hash string) string {
	return filepath.Join(outDir, "videofiles", hash, "/subs")
}

func GetFileAttCacheDir(outDir string, hash string) string {
	return filepath.Join(outDir, "videofiles", hash, "/att")
}

func ExtractAttachment(path string, hash string, mediaInfo *MediaInfo, cacheDir string, logger *zerolog.Logger) (err error) {

	attachmentPath := GetFileAttCacheDir(cacheDir, hash)
	subsPath := GetFileSubsCacheDir(cacheDir, hash)
	_ = os.MkdirAll(attachmentPath, 0755)
	_ = os.MkdirAll(subsPath, 0755)

	subsDir, err := os.ReadDir(subsPath)
	if err == nil {
		if len(subsDir) == len(mediaInfo.Subtitles) {
			return
		}
	}

	cmd := exec.Command(
		"ffmpeg",
		"-dump_attachment:t", "",
		// override old attachments
		"-y",
		"-i", path,
	)
	cmd.Dir = attachmentPath

	for _, sub := range mediaInfo.Subtitles {
		if ext := sub.Extension; ext != nil {
			cmd.Args = append(
				cmd.Args,
				"-map", fmt.Sprintf("0:s:%d", sub.Index),
				"-c:s", "copy",
				fmt.Sprintf("%s/%d.%s", subsPath, sub.Index, *ext),
			)
		}
	}
	logger.Trace().Msgf("transcoder: Starting data extraction")
	cmd.Stdout = nil
	err = cmd.Run()
	if err != nil {
		logger.Error().Err(err).Msgf("transcoder: Error starting ffmpeg")
	}

	return err
}
