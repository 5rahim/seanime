package transcoder

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/mediastream/videofile"
	"os"
	"os/exec"
	"path/filepath"
)

func Extract(path string, sha string, mediaInfo *videofile.MediaInfo, settings *Settings, logger *zerolog.Logger) (err error) {

	defer printExecTime(logger, "Data extraction of %s", path)()

	attachmentPath := filepath.Join(settings.MetadataDir, sha, "/att")
	subsPath := filepath.Join(settings.MetadataDir, sha, "/sub")
	_ = os.MkdirAll(attachmentPath, 0o755)
	_ = os.MkdirAll(subsPath, 0o755)

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
