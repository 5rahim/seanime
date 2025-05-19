package mkvparser

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"seanime/internal/util"
	httputil "seanime/internal/util/http"
	"strings"
	"testing"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/at-wat/ebml-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testMagnet  = util.Decode("bWFnbmV0Oj94dD11cm46YnRpaDpRRVI1TFlQSkFYWlFBVVlLSE5TTE80TzZNTlY2VUQ2QSZ0cj1odHRwJTNBJTJGJTJGbnlhYS50cmFja2VyLndmJTNBNzc3NyUyRmFubm91bmNlJnRyPXVkcCUzQSUyRiUyRnRyYWNrZXIuY29wcGVyc3VyZmVyLnRrJTNBNjk2OSUyRmFubm91bmNlJnRyPXVkcCUzQSUyRiUyRnRyYWNrZXIub3BlbnRyYWNrci5vcmclM0ExMzM3JTJGYW5ub3VuY2UmdHI9dWRwJTNBJTJGJTJGOS5yYXJiZy50byUzQTI3MTAlMkZhbm5vdW5jZSZ0cj11ZHAlM0ElMkYlMkY5LnJhcmJnLm1lJTNBMjcxMCUyRmFubm91bmNlJmRuPSU1QlN1YnNQbGVhc2UlNUQlMjBTb3Vzb3UlMjBubyUyMEZyaWVyZW4lMjAtJTIwMjglMjAlMjgxMDgwcCUyOSUyMCU1QjhCQkJDMjhDJTVELm1rdg==")
	testHttpUrl = ""
	testFile    = util.Decode("L1VzZXJzL3JhaGltL0RvY3VtZW50cy9jb2xsZWN0aW9uL0RhbmRhZGFuL1tTdWJzUGxlYXNlXSBEYW5kYWRhbiAtIDAzICgxMDgwcCkgWzU2OUJBQTlDXS5ta3Y=")
	// Timeout for torrent operations
	torrentInfoTimeout = 60 * time.Second
	// Timeout for metadata parsing test
	metadataTestTimeout = 90 * time.Second
	// Number of initial pieces to prioritize for header metadata
	initialPiecesToPrioritize = 20
)

// getTestTorrentClient creates a new torrent client for testing.
func getTestTorrentClient(t *testing.T, tempDir string) *torrent.Client {
	t.Helper()
	cfg := torrent.NewDefaultClientConfig()
	// Use a subdirectory within the temp dir for torrent data
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
		client.Close() // Close client on error
		t.Fatalf("failed to add magnet: %v", err)
	}

	t.Log("Waiting for torrent info...")
	select {
	case <-tor.GotInfo():
		t.Log("Torrent info received.")
		// continue
	case <-tctx.Done():
		tor.Drop()     // Attempt to drop torrent
		client.Close() // Close client
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
	tor.Drop()     // Drop torrent if no suitable file found
	client.Close() // Close client
	t.Fatalf("no video file found in torrent")
	return nil, nil, nil // Should not be reached
}

func assertTestResult(t *testing.T, result *Metadata) {
	// Check for parsing errors (ebml.ErrReadStopped is OK here)
	if result.Error != nil && !errors.Is(result.Error, ebml.ErrReadStopped) {
		// If the error is context timeout/canceled, it's less severe but still worth noting
		if errors.Is(result.Error, context.DeadlineExceeded) || errors.Is(result.Error, context.Canceled) {
			t.Logf("Warning: GetMetadata context deadline exceeded or canceled: %v", result.Error)
		} else {
			t.Errorf("GetMetadata failed with unexpected error: %v", result.Error)
		}
	} else if result.Error != nil {
		t.Logf("Note: GetMetadata stopped with expected error: %v", result.Error)
	}

	// Check Duration (should be positive for this known file)
	assert.True(t, result.Duration > 0, "Expected Duration to be positive, got %.2f", result.Duration)
	t.Logf("Duration: %.2f seconds", result.Duration)

	// Check TimecodeScale
	assert.True(t, result.TimecodeScale > 0, "Expected TimecodeScale to be positive, got %f", result.TimecodeScale)
	t.Logf("TimecodeScale: %f", result.TimecodeScale)

	// Check Muxing/Writing App (often present)
	if result.MuxingApp != "" {
		t.Logf("MuxingApp: %s", result.MuxingApp)
	}
	if result.WritingApp != "" {
		t.Logf("WritingApp: %s", result.WritingApp)
	}

	// Check Tracks (expecting video, audio, subs for this file)
	assert.NotEmpty(t, result.Tracks, "Expected to find tracks")
	t.Logf("Found %d total tracks:", len(result.Tracks))
	foundVideo := false
	foundAudio := false
	for i, track := range result.Tracks {
		t.Logf("  Track %d: Type=%s, Codec=%s, Lang=%s, Name='%s', Default=%v, Forced=%v, Enabled=%v",
			i, track.Type, track.CodecID, track.Language, track.Name, track.Default, track.Forced, track.Enabled)
		if track.Type == "video" {
			foundVideo = true
			assert.True(t, track.PixelWidth > 0, "Video track should have PixelWidth > 0")
			assert.True(t, track.PixelHeight > 0, "Video track should have PixelHeight > 0")
			t.Logf("    Video Details: %dx%d", track.PixelWidth, track.PixelHeight)
		}
		if track.Type == "audio" {
			foundAudio = true
			assert.True(t, track.SamplingFrequency > 0, "Audio track should have SamplingFrequency > 0")
			assert.True(t, track.Channels > 0, "Audio track should have Channels > 0")
			t.Logf("    Audio Details: Freq=%.1f, Channels=%d, BitDepth=%d", track.SamplingFrequency, track.Channels, track.BitDepth)
		}
	}
	assert.True(t, foundVideo, "Expected to find at least one video track")
	assert.True(t, foundAudio, "Expected to find at least one audio track")

	// Print the JSON representation of the result
	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal result to JSON: %v", err)
	}
	t.Logf("JSON Result: %s", string(jsonResult))
}

func testStreamSubtitles(t *testing.T, parser *MetadataParser, reader io.ReadCloser, ctx context.Context) {
	if _, ok := reader.(io.ReadSeeker); !ok {
		t.Fatalf("Reader does not support seeking, cannot test StreamSubtitles")
	}

	// Stream for 30 seconds
	streamCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	subtitleCh, errCh := parser.ExtractSubtitles(streamCtx, reader)

	var streamedSubtitles []*SubtitleEvent

	// Collect subtitles with a timeout
	collectDone := make(chan struct{})
	go func() {
		defer reader.Close()
		defer close(collectDone)
		for {
			select {
			case subtitle, ok := <-subtitleCh:
				if !ok {
					return // Channel closed
				}
				streamedSubtitles = append(streamedSubtitles, subtitle)
			case <-streamCtx.Done():
				return // Timeout
			}
		}
	}()

	// Wait for all subtitles or timeout
	select {
	case <-collectDone:
		// All subtitles collected
	case <-streamCtx.Done():
		t.Log("StreamSubtitles collection timed out (this is expected for large files)")
	}

	// Check for errors
	select {
	case err := <-errCh:
		if err != nil {
			t.Logf("StreamSubtitles returned an error: %v", err)
		}
	default:
		// No errors yet
	}

	t.Logf("Found %d streamed subtitles:", len(streamedSubtitles))
	for i, sub := range streamedSubtitles {
		if i < 5 { // Log first 5 subtitles
			t.Logf("  Streamed Subtitle %d: TrackNumber=%d, StartTime=%.2f, Text='%s'",
				i, sub.TrackNumber, sub.StartTime, sub.Text)
		}
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

	// Ensure client and torrent are closed/dropped eventually
	t.Cleanup(func() {
		t.Log("Dropping torrent...")
		tor.Drop()
		t.Log("Closing torrent client...")
		client.Close()
		t.Log("Cleanup finished.")
	})

	logger := util.NewLogger()
	parser := NewMetadataParser(file.NewReader(), logger)

	// Create context with timeout for the metadata parsing operation itself
	ctx, cancel := context.WithTimeout(context.Background(), metadataTestTimeout)
	defer cancel()

	t.Log("Calling file.Download() to enable piece requests...")
	file.Download() // Start download requests

	// Prioritize initial pieces to ensure metadata is fetched quickly
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
		// Give a moment for prioritization to take effect and requests to start
		time.Sleep(500 * time.Millisecond)
	} else {
		t.Log("Torrent info or pieces not available for prioritization.")
	}

	t.Log("Calling GetMetadata...")
	startTime := time.Now()
	metadata := parser.GetMetadata(ctx)
	elapsed := time.Since(startTime)
	t.Logf("GetMetadata took %v", elapsed)

	assertTestResult(t, metadata)

	testStreamSubtitles(t, parser, file.NewReader(), ctx)
}

// TestMetadataParser_HTTPStream tests parsing from an HTTP stream
func TestMetadataParser_HTTPStream(t *testing.T) {
	if testHttpUrl == "" {
		t.Skip("Skipping HTTP stream test")
	}

	logger := util.NewLogger()

	res, err := http.Get(testHttpUrl)
	if err != nil {
		t.Fatalf("HTTP GET request failed: %v", err)
	}
	defer res.Body.Close()

	rs := httputil.NewHttpReadSeeker(res)

	if res.StatusCode != http.StatusOK {
		t.Fatalf("HTTP GET request returned non-OK status: %s", res.Status)
	}

	parser := NewMetadataParser(rs, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // 30-second timeout for parsing
	defer cancel()

	metadata := parser.GetMetadata(ctx)

	assertTestResult(t, metadata)

	_, err = rs.Seek(0, io.SeekStart)
	require.NoError(t, err)

	testStreamSubtitles(t, parser, rs, ctx)
}

func TestMetadataParser_File(t *testing.T) {
	if testFile == "" {
		t.Skip("Skipping file test")
	}

	logger := util.NewLogger()

	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Could not open file: %v", err)
	}
	defer file.Close()

	parser := NewMetadataParser(file, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // 30-second timeout for parsing
	defer cancel()

	metadata := parser.GetMetadata(ctx)

	assertTestResult(t, metadata)

	_, err = file.Seek(0, io.SeekStart)
	require.NoError(t, err)

	testStreamSubtitles(t, parser, file, ctx)
}
