package videofile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"seanime/internal/util"
	"seanime/internal/util/crashlog"
	"time"

	"github.com/rs/zerolog"
)

func GetFileSubsCacheDir(outDir string, hash string) string {
	return filepath.Join(outDir, "videofiles", hash, "subs")
}

func GetFileAttCacheDir(outDir string, hash string) string {
	return filepath.Join(outDir, "videofiles", hash, "att")
}

// ExtractAttachment extracts subtitles and font attachments from a media file
// using ffmpeg. It skips extraction if the output directory already contains
// the expected number of subtitle files.
//
// Improvements over the previous version:
//   - 120-second timeout prevents hangs on corrupt/huge files.
//   - Validates subtitle extensions before starting ffmpeg.
//   - Uses context cancellation for clean cleanup on timeout.
func ExtractAttachment(ffmpegPath string, path string, hash string, mediaInfo *MediaInfo, cacheDir string, logger *zerolog.Logger) (err error) {
	logger.Debug().Str("hash", hash).Msgf("videofile: Starting media attachment extraction")

	attachmentPath := GetFileAttCacheDir(cacheDir, hash)
	subsPath := GetFileSubsCacheDir(cacheDir, hash)
	_ = os.MkdirAll(attachmentPath, 0755)
	_ = os.MkdirAll(subsPath, 0755)

	// Check if subtitles are already extracted.
	subsDir, err := os.ReadDir(subsPath)
	if err == nil && len(subsDir) >= len(mediaInfo.Subtitles) {
		logger.Debug().Str("hash", hash).Msgf("videofile: Attachments already extracted")
		return nil
	}

	// Validate all subtitles have supported extensions before starting ffmpeg.
	for _, sub := range mediaInfo.Subtitles {
		if sub.Extension == nil || *sub.Extension == "" {
			logger.Error().Uint32("index", sub.Index).Msgf("videofile: Subtitle format is not supported, skipping")
			continue
		}
	}

	// Use a timeout context to prevent hangs on corrupt files.
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Instantiate a new crash logger
	crashLogger := crashlog.GlobalCrashLogger.InitArea("ffmpeg")
	defer crashLogger.Close()

	crashLogger.LogInfof("Extracting attachments from %s", path)

	// Build ffmpeg command: dump font attachments and extract subtitles.
	args := []string{"-y"}
	for i, font := range mediaInfo.Fonts {
		if font == "" {
			continue
		}
		font = filepath.Base(font)
		args = append(args,
			fmt.Sprintf("-dump_attachment:t:%d", i),
			filepath.Join(attachmentPath, font),
		)
	}
	args = append(args, "-i", path)

	extractedCount := 0
	for _, sub := range mediaInfo.Subtitles {
		if sub.Extension == nil || *sub.Extension == "" {
			continue
		}
		args = append(args,
			"-map", fmt.Sprintf("0:s:%d", sub.Index),
			"-c:s", "copy",
			fmt.Sprintf("%s/%d.%s", subsPath, sub.Index, *sub.Extension),
		)
		extractedCount++
	}

	if extractedCount == 0 {
		logger.Debug().Str("hash", hash).Msg("videofile: No extractable subtitles found")
		return nil
	}

	cmd := util.NewCmdCtx(ctx, ffmpegPath, args...)
	cmd.Dir = attachmentPath

	cmd.Stdout = crashLogger.Stdout()
	cmd.Stderr = crashLogger.Stdout()
	err = cmd.Run()
	if err != nil {
		if ctx.Err() != nil {
			logger.Error().Str("hash", hash).Msg("videofile: FFmpeg attachment extraction timed out")
		} else {
			logger.Error().Err(err).Msgf("videofile: Error running FFmpeg")
		}
		crashlog.GlobalCrashLogger.WriteAreaLogToFile(crashLogger)
	} else {
		logger.Debug().Str("hash", hash).Int("subtitles", extractedCount).Msg("videofile: Attachment extraction complete")
	}

	return err
}
