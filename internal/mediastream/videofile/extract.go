package videofile

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"path/filepath"
	"seanime/internal/util"
)

func GetFileSubsCacheDir(outDir string, hash string) string {
	return filepath.Join(outDir, "videofiles", hash, "/subs")
}

func GetFileAttCacheDir(outDir string, hash string) string {
	return filepath.Join(outDir, "videofiles", hash, "/att")
}

func ExtractAttachment(ffmpegPath string, path string, hash string, mediaInfo *MediaInfo, cacheDir string, logger *zerolog.Logger) (err error) {
	logger.Debug().Str("hash", hash).Msgf("videofile: Starting media attachment extraction")

	attachmentPath := GetFileAttCacheDir(cacheDir, hash)
	subsPath := GetFileSubsCacheDir(cacheDir, hash)
	_ = os.MkdirAll(attachmentPath, 0755)
	_ = os.MkdirAll(subsPath, 0755)

	subsDir, err := os.ReadDir(subsPath)
	if err == nil {
		if len(subsDir) == len(mediaInfo.Subtitles) {
			logger.Debug().Str("hash", hash).Msgf("videofile: Attachments already extracted")
			return
		}
	}

	for _, sub := range mediaInfo.Subtitles {
		if sub.Extension == nil || *sub.Extension == "" {
			logger.Error().Msgf("videofile: Subtitle format is not supported")
			return fmt.Errorf("videofile: Unsupported subtitle format")
		}
	}

	cmd := util.NewCmdCtx(
		context.Background(),
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
	//cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		logger.Error().Err(err).Msgf("videofile: Error starting FFmepg")
	}

	return err
}
