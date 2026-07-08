package videofile

import (
	"os"
	"path/filepath"
	"seanime/internal/util"
	"testing"

	"gopkg.in/vansante/go-ffprobe.v2"
)

func TestStreamToMimeCodec(t *testing.T) {
	tests := []struct {
		name     string
		stream   ffprobe.Stream
		expected string // empty means nil expected
	}{
		{
			name:     "h264 high level 4.0",
			stream:   ffprobe.Stream{CodecName: "h264", Profile: "High", Level: 40},
			expected: "avc1.640028",
		},
		{
			name:     "hevc main level 4.0",
			stream:   ffprobe.Stream{CodecName: "hevc", Profile: "Main", Level: 120},
			expected: "hvc1.1.6.L120.B0",
		},
		{
			name:     "hevc main 10 level 5.1",
			stream:   ffprobe.Stream{CodecName: "hevc", Profile: "Main 10", Level: 153},
			expected: "hvc1.2.4.L153.B0",
		},
		{
			name:     "hevc unknown level",
			stream:   ffprobe.Stream{CodecName: "hevc", Profile: "Main", Level: 0},
			expected: "hvc1.1.6.L120.B0",
		},
		{
			name:     "av1 main level 4.0 8bit",
			stream:   ffprobe.Stream{CodecName: "av1", Profile: "Main", Level: 8, BitsPerRawSample: "8"},
			expected: "av01.0.08M.08",
		},
		{
			name:     "av1 main level 5.1 10bit",
			stream:   ffprobe.Stream{CodecName: "av1", Profile: "Main", Level: 13, BitsPerRawSample: "10"},
			expected: "av01.0.13M.10",
		},
		{
			name:     "vp9 profile 0",
			stream:   ffprobe.Stream{CodecName: "vp9", Profile: "Profile 0", Level: -99},
			expected: "vp09.00.10.08",
		},
		{
			name:     "vp9 profile 2 10bit",
			stream:   ffprobe.Stream{CodecName: "vp9", Profile: "Profile 2", Level: 40, BitsPerRawSample: "10"},
			expected: "vp09.02.40.10",
		},
		{
			name:     "vp8",
			stream:   ffprobe.Stream{CodecName: "vp8"},
			expected: "vp8",
		},
		{
			name:     "aac lc",
			stream:   ffprobe.Stream{CodecName: "aac", Profile: "LC"},
			expected: "mp4a.40.2",
		},
		{
			name:     "opus",
			stream:   ffprobe.Stream{CodecName: "opus"},
			expected: "opus",
		},
		{
			name:     "flac",
			stream:   ffprobe.Stream{CodecName: "flac"},
			expected: "fLaC",
		},
		{
			name:     "ac3",
			stream:   ffprobe.Stream{CodecName: "ac3"},
			expected: "ac-3",
		},
		{
			name:     "eac3",
			stream:   ffprobe.Stream{CodecName: "eac3"},
			expected: "ec-3",
		},
		{
			name:     "mp3",
			stream:   ffprobe.Stream{CodecName: "mp3"},
			expected: "mp4a.40.34",
		},
		{
			name:     "vorbis",
			stream:   ffprobe.Stream{CodecName: "vorbis"},
			expected: "vorbis",
		},
		{
			name:     "unknown codec",
			stream:   ffprobe.Stream{CodecName: "dts"},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := streamToMimeCodec(&tt.stream)
			if tt.expected == "" {
				if got != nil {
					t.Fatalf("expected nil, got %q", *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected %q, got nil", tt.expected)
			}
			if *got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, *got)
			}
		})
	}
}

func TestFfprobeGetInfo_1(t *testing.T) {
	t.Skip()

	testFilePath := ""

	mi, err := FfprobeGetInfo("", testFilePath, "1")
	if err != nil {
		t.Fatalf("Error getting media info: %v", err)
	}

	util.Spew(mi)
}

func TestExtractAttachment(t *testing.T) {
	t.Skip()

	testFilePath := ""

	testDir := t.TempDir()

	mi, err := FfprobeGetInfo("", testFilePath, "1")
	if err != nil {
		t.Fatalf("Error getting media info: %v", err)
	}

	util.Spew(mi)

	err = ExtractAttachment("", testFilePath, "1", mi, testDir, util.NewLogger())
	if err != nil {
		t.Fatalf("Error extracting attachment: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(testDir, "videofiles", "1", "att"))
	if err != nil {
		t.Fatalf("Error reading directory: %v", err)
	}

	for _, entry := range entries {
		info, _ := entry.Info()
		t.Logf("Entry: %s, Size: %d\n", entry.Name(), info.Size())
	}
}
