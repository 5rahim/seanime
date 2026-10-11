package videofile

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
)

func TestExtractAttachmentSpacedFont(t *testing.T) {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg is not installed")
	}

	for _, withSubtitles := range []bool{false, true} {
		name := "without subtitles"
		if withSubtitles {
			name = "with subtitles"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			fontName := "Font With Spaces.ttf"
			fontData := []byte("sample font attachment")
			fontPath := filepath.Join(dir, fontName)
			if err := os.WriteFile(fontPath, fontData, 0644); err != nil {
				t.Fatal(err)
			}

			args := []string{
				"-hide_banner", "-loglevel", "error", "-y",
				"-f", "lavfi", "-i", "color=c=black:s=16x16:r=1:d=1",
			}
			mediaInfo := &MediaInfo{Fonts: []string{fontName}}
			if withSubtitles {
				subtitlePath := filepath.Join(dir, "input.srt")
				if err := os.WriteFile(subtitlePath, []byte("1\n00:00:00,000 --> 00:00:01,000\nHello\n\n"), 0644); err != nil {
					t.Fatal(err)
				}
				args = append(args, "-f", "srt", "-i", subtitlePath, "-map", "0:v:0", "-map", "1:s:0")
				ext := "ass"
				mediaInfo.Subtitles = []Subtitle{{Index: 0, Extension: &ext}}
			}
			args = append(args, "-t", "1", "-c:v", "ffv1")
			if withSubtitles {
				args = append(args, "-c:s", "ass")
			}
			args = append(args,
				"-attach", fontPath,
				"-metadata:s:t:0", "mimetype=application/x-truetype-font",
				filepath.Join(dir, "input.mkv"),
			)
			cmd := exec.Command(ffmpegPath, args...)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("create fixture: %v: %s", err, output)
			}

			cacheDir := filepath.Join(dir, "cache")
			logger := zerolog.Nop()
			extract := func() {
				t.Helper()
				if err := ExtractAttachment(ffmpegPath, filepath.Join(dir, "input.mkv"), "test", mediaInfo, cacheDir, &logger); err != nil {
					t.Fatalf("extract attachments: %v", err)
				}
			}
			fontOutput := filepath.Join(GetFileAttCacheDir(cacheDir, "test"), fontName)
			checkFont := func() {
				t.Helper()
				got, err := os.ReadFile(fontOutput)
				if err != nil {
					t.Fatalf("read extracted font: %v", err)
				}
				if !bytes.Equal(got, fontData) {
					t.Fatalf("extracted font differs from attachment")
				}
			}

			extract()
			checkFont()
			if withSubtitles {
				if _, err := os.Stat(filepath.Join(GetFileSubsCacheDir(cacheDir, "test"), "0.ass")); err != nil {
					t.Fatalf("subtitle not extracted: %v", err)
				}
			}

			// A missing font must invalidate the cache, even when subtitles exist.
			if err := os.Remove(fontOutput); err != nil {
				t.Fatal(err)
			}
			extract()
			checkFont()
		})
	}
}
