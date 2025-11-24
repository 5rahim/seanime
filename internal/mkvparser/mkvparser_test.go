package mkvparser

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"seanime/internal/util"
	"strings"
	"testing"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testFile   = ""
	testMagnet = ""

	// Timeout for torrent operations
	torrentInfoTimeout        = 60 * time.Second
	metadataTestTimeout       = 90 * time.Second
	initialPiecesToPrioritize = 20
)

func TestMetadataParser_File(t *testing.T) {
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found, skipping test")
		return
	}

	file, err := os.Open(testFile)
	require.NoError(t, err)
	defer file.Close()

	logger := util.NewLogger()
	parser := NewMetadataParser(file, logger)

	ctx := context.Background()
	metadata := parser.GetMetadata(ctx)

	require.NoError(t, metadata.Error)
	assert.NotNil(t, metadata)
	assert.Greater(t, len(metadata.Tracks), 0, "Should have at least one track")
	assert.Greater(t, metadata.Duration, 0.0, "Duration should be greater than 0")

	assertTestResult(t, metadata)

}

func TestMetadataParser_ExtractSubtitles(t *testing.T) {
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found, skipping test")
		return
	}

	file, err := os.Open(testFile)
	require.NoError(t, err)
	defer file.Close()

	logger := util.NewLogger()
	parser := NewMetadataParser(file, logger)

	// First get metadata to know available tracks
	ctx := context.Background()
	metadata := parser.GetMetadata(ctx)
	require.NoError(t, metadata.Error)

	if len(metadata.SubtitleTracks) == 0 {
		t.Skip("No subtitle tracks found, skipping subtitle extraction test")
		return
	}

	t.Logf("Found %d subtitle tracks", len(metadata.SubtitleTracks))

	// Open a new reader for subtitle extraction
	newFile, err := os.Open(testFile)
	require.NoError(t, err)
	defer newFile.Close()

	// Extract subtitles from the beginning
	subtitleCh, errCh, startedCh := parser.ExtractSubtitles(ctx, newFile, 123000000, 1024*1024)

	<-startedCh

	subtitleCount := 0
	maxSubtitles := 10 // Only check first 10 subtitles

	for subtitleCount < maxSubtitles {
		select {
		case subtitle, ok := <-subtitleCh:
			if !ok {
				t.Log("Subtitle channel closed")
				goto done
			}
			if subtitle != nil {
				subtitleCount++
				t.Logf("Subtitle %d: Track=%d, StartTime=%.2f, Duration=%.2f, Text=%q",
					subtitleCount, subtitle.TrackNumber, subtitle.StartTime, subtitle.Duration,
					truncateString(subtitle.Text, 50))
				assert.Greater(t, subtitle.StartTime, -1.0, "Start time should be valid")
			}
		case err, ok := <-errCh:
			if !ok {
				t.Log("Error channel closed")
				goto done
			}
			if err != nil {
				t.Logf("Subtitle extraction completed with: %v", err)
				goto done
			}
		}
	}

done:
	t.Logf("Extracted %d subtitle events", subtitleCount)
	assert.Greater(t, subtitleCount, 0, "Should have extracted at least one subtitle")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func TestReadIsMkvOrWebm(t *testing.T) {
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found, skipping test")
		return
	}

	file, err := os.Open(testFile)
	require.NoError(t, err)
	defer file.Close()

	mimeType, isMkv := ReadIsMkvOrWebm(file)
	assert.True(t, isMkv, "Should detect as MKV/WebM")
	assert.NotEmpty(t, mimeType, "Should return mime type")
	t.Logf("Detected mime type: %s", mimeType)
}

// Helper functions for torrent tests

// getTestTorrentClient creates a new torrent client for testing.
func getTestTorrentClient(t *testing.T, tempDir string) *torrent.Client {
	t.Helper()
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = filepath.Join(tempDir, "torrent_data")
	err := os.MkdirAll(cfg.DataDir, 0755)
	if err != nil {
		t.Fatalf("failed to create torrent data directory: %v", err)
	}

	client, err := torrent.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create torrent client: %v", err)
	}
	return client
}

// hasExt checks if a file path has a specific extension (case-insensitive).
func hasExt(name, ext string) bool {
	if len(name) < len(ext) {
		return false
	}
	return strings.ToLower(name[len(name)-len(ext):]) == strings.ToLower(ext)
}

// hasVideoExt checks for common video file extensions.
func hasVideoExt(name string) bool {
	return hasExt(name, ".mkv") || hasExt(name, ".mp4") || hasExt(name, ".avi") || hasExt(name, ".mov") || hasExt(name, ".webm")
}

// getTestTorrentFile adds the torrent, waits for metadata, returns the first video file.
func getTestTorrentFile(t *testing.T, magnet string, tempDir string) (*torrent.Client, *torrent.Torrent, *torrent.File) {
	t.Helper()
	client := getTestTorrentClient(t, tempDir)

	tctx, cancel := context.WithTimeout(context.Background(), torrentInfoTimeout)
	defer cancel()

	tor, err := client.AddMagnet(magnet)
	if err != nil {
		client.Close()
		t.Fatalf("failed to add magnet: %v", err)
	}

	t.Log("Waiting for torrent info...")
	select {
	case <-tor.GotInfo():
		t.Log("Torrent info received.")
	case <-tctx.Done():
		tor.Drop()
		client.Close()
		t.Fatalf("timeout waiting for torrent metadata (%v)", torrentInfoTimeout)
	}

	// Find the first video file
	for _, f := range tor.Files() {
		path := f.DisplayPath()
		if hasVideoExt(path) {
			t.Logf("Found video file: %s (Size: %d bytes)", path, f.Length())
			return client, tor, f
		}
	}

	t.Logf("No video file found in torrent info: %s", tor.Info().Name)
	tor.Drop()
	client.Close()
	t.Fatalf("no video file found in torrent")
	return nil, nil, nil
}

func assertTestResult(t *testing.T, result *Metadata) {
	if result.Error != nil {
		if errors.Is(result.Error, context.DeadlineExceeded) || errors.Is(result.Error, context.Canceled) {
			t.Logf("Warning: GetMetadata context deadline exceeded or canceled: %v", result.Error)
		} else {
			t.Errorf("GetMetadata failed with unexpected error: %v", result.Error)
		}
	}

	// Check Duration
	assert.True(t, result.Duration > 0, "Expected Duration to be positive, got %.2f", result.Duration)
	t.Logf("Duration: %.2f seconds (%.2f minutes)", result.Duration, result.Duration/60.0)

	// Check TimecodeScale
	assert.True(t, result.TimecodeScale > 0, "Expected TimecodeScale to be positive, got %f", result.TimecodeScale)
	t.Logf("TimecodeScale: %f", result.TimecodeScale)

	// Check Muxing/Writing App
	if result.MuxingApp != "" {
		t.Logf("MuxingApp: %s", result.MuxingApp)
	}
	if result.WritingApp != "" {
		t.Logf("WritingApp: %s", result.WritingApp)
	}

	// Check Tracks
	assert.NotEmpty(t, result.Tracks, "Expected to find tracks")
	t.Logf("Found %d total tracks:", len(result.Tracks))
	foundVideo := false
	foundAudio := false
	for i, track := range result.Tracks {
		t.Logf("  Track %d: Type=%s, Codec=%s, Lang=%s, LangIETF=%s Name='%s', Default=%v, Forced=%v, Enabled=%v",
			i, track.Type, track.CodecID, track.Language, track.LanguageIETF, track.Name, track.Default, track.Forced, track.Enabled)

		if track.Type == TrackTypeSubtitle {
			t.Logf("   Subtitle Track: CodecPrivate=%s", track.CodecPrivate)
		}

		if track.Video != nil {
			foundVideo = true
			assert.True(t, track.Video.PixelWidth > 0, "Video track should have PixelWidth > 0")
			assert.True(t, track.Video.PixelHeight > 0, "Video track should have PixelHeight > 0")
			t.Logf("    Video: %dx%d", track.Video.PixelWidth, track.Video.PixelHeight)
		}
		if track.Audio != nil {
			foundAudio = true
			assert.True(t, track.Audio.Channels > 0, "Audio track should have Channels > 0")
			t.Logf("    Audio: %.0f Hz, %d channels", track.Audio.SamplingFrequency, track.Audio.Channels)
		}
	}

	assert.True(t, foundVideo, "Expected to find at least one video track")
	assert.True(t, foundAudio, "Expected to find at least one audio track")

	// Check chapters
	t.Logf("Found %d chapters", len(result.Chapters))
	for i, chapter := range result.Chapters {
		t.Logf("  Chapter %d: Start=%.2fs, Text='%s'", i, chapter.Start, chapter.Text)
	}

	// Check attachments
	t.Logf("Found %d attachments", len(result.Attachments))
	for i, attachment := range result.Attachments {
		t.Logf("  Attachment %d: Filename=%s, IsCompressed=%v", i, attachment.Filename, attachment.IsCompressed)
	}
}

// TestMetadataParser_Torrent performs an integration test.
// It downloads the header of a real torrent and parses its metadata.
func TestMetadataParser_Torrent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	client, tor, file := getTestTorrentFile(t, testMagnet, tempDir)

	t.Cleanup(func() {
		t.Log("Dropping torrent...")
		tor.Drop()
		t.Log("Closing torrent client...")
		client.Close()
		t.Log("Cleanup finished.")
	})

	logger := util.NewLogger()
	parser := NewMetadataParser(file.NewReader(), logger)

	ctx, cancel := context.WithTimeout(context.Background(), metadataTestTimeout)
	defer cancel()

	t.Log("Calling file.Download() to enable piece requests...")
	file.Download()

	// Prioritize initial pieces
	torInfo := tor.Info()
	if torInfo != nil && torInfo.NumPieces() > 0 {
		numPieces := torInfo.NumPieces()
		piecesToFetch := initialPiecesToPrioritize
		if numPieces < piecesToFetch {
			piecesToFetch = numPieces
		}
		t.Logf("Prioritizing first %d pieces (out of %d) for header parsing...", piecesToFetch, numPieces)
		for i := 0; i < piecesToFetch; i++ {
			p := tor.Piece(i)
			if p != nil {
				p.SetPriority(torrent.PiecePriorityNow)
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	t.Log("Calling GetMetadata...")
	startTime := time.Now()
	metadata := parser.GetMetadata(ctx)
	elapsed := time.Since(startTime)
	t.Logf("GetMetadata took %v", elapsed)

	assertTestResult(t, metadata)
}

func TestConvertSRTToASS(t *testing.T) {
	srt := `1
00:00:00,000 --> 00:00:03,000
Hello, world!

2
00:00:04,000 --> 00:00:06,000
This is a <--> test.
`
	out, err := ConvertToASS(srt, SubtitleTypeSRT)
	require.NoError(t, err)

	require.Equal(t, `[Script Info]
PlayResX: 640
PlayResY: 360
ScriptType: v4.00+
WrapStyle: 0
ScaledBorderAndShadow: yes

[V4+ Styles]
Format: Name, Alignment, Angle, BackColour, Bold, BorderStyle, Encoding, Fontname, Fontsize, Italic, MarginL, MarginR, MarginV, Outline, OutlineColour, PrimaryColour, ScaleX, ScaleY, SecondaryColour, Shadow, Spacing, Strikeout, Underline
Style: Default,2,0.000,&H00000000,0,1,0,Roboto Medium,24.000,0,20,20,23,1.300,&H00000000,&H00ffffff,100.000,100.000,&H000000ff,0.000,0.000,0,0

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
Dialogue: 0,00:00:00.00,00:00:03.00,Default,,0,0,0,,Hello, world!
Dialogue: 0,00:00:04.00,00:00:06.00,Default,,0,0,0,,This is a <--> test.
`, out)
}
