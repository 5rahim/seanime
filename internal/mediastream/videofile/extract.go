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

func ExtractAttachment(ffmpegPath string, path string, hash string, mediaInfo *MediaInfo, cacheDir string, logger *zerolog.Logger) (err error) {
	logger.Trace().Str("path", path).Msgf("transcoder: Starting media attachment extraction")

	attachmentPath := GetFileAttCacheDir(cacheDir, hash)
	subsPath := GetFileSubsCacheDir(cacheDir, hash)
	_ = os.MkdirAll(attachmentPath, 0755)
	_ = os.MkdirAll(subsPath, 0755)

	subsDir, err := os.ReadDir(subsPath)
	if err == nil {
		if len(subsDir) == len(mediaInfo.Subtitles) {
			logger.Trace().Str("path", path).Msgf("transcoder: Attachments already extracted")
			return
		}
	}

	cmd := exec.Command(
		ffmpegPath,
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
	cmd.Stdout = nil
	err = cmd.Run()
	if err != nil {
		logger.Error().Err(err).Msgf("transcoder: Error starting ffmpeg")
	}

	return err
}
