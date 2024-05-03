package transcoder

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/mediastream/videofile"
	"os"
	"os/exec"
	"path/filepath"
)

var extracted = NewCMap[string, <-chan struct{}]()

func Extract(path string, sha string, mediaInfo *videofile.MediaInfo, settings *Settings, logger *zerolog.Logger) (<-chan struct{}, error) {
	ret := make(chan struct{})
	existing, created := extracted.GetOrSet(sha, ret)
	if !created {
		return existing, nil
	}

	go func() {
		defer printExecTime(logger, "Starting extraction of %s", path)()

		attachmentPath := filepath.Join(settings.MetadataDir, sha, "/att")
		subsPath := filepath.Join(settings.MetadataDir, sha, "/sub")
		_ = os.MkdirAll(attachmentPath, 0o644)
		_ = os.MkdirAll(subsPath, 0o755)

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
		logger.Trace().Msgf("transcoder: Starting attachment extraction with: %s", cmd)
		cmd.Stdout = nil
		err := cmd.Run()
		if err != nil {
			extracted.Remove(sha)
			fmt.Println("Error starting ffmpeg extract:", err)
		}
		close(ret)
	}()

	return ret, nil
}
