package videofile

import (
	"os"
	"path/filepath"
	"seanime/internal/util"
	"testing"

	"gopkg.in/vansante/go-ffprobe.v2"
)

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

func TestStreamToMimeCodec(t *testing.T) {
	tests := []struct {
		name   string
		stream ffprobe.Stream
		want   string
	}{
		{
			name:   "h264 high level 4.0",
			stream: ffprobe.Stream{CodecName: "h264", Profile: "High", Level: 40},
			want:   "avc1.640028",
		},
		{
			name:   "h264 baseline level 3.1",
			stream: ffprobe.Stream{CodecName: "h264", Profile: "Baseline", Level: 31},
			want:   "avc1.42001F",
		},
		{
			name:   "h264 constrained baseline level 3.1",
			stream: ffprobe.Stream{CodecName: "h264", Profile: "Constrained Baseline", Level: 31},
			want:   "avc1.42E01F",
		},
		{
			name:   "h264 unknown profile",
			stream: ffprobe.Stream{CodecName: "h264", Profile: "High 10", Level: 40},
		},
		{
			name:   "h264 unknown level",
			stream: ffprobe.Stream{CodecName: "h264", Profile: "High", Level: 0},
		},
		{
			name:   "hevc main level 4.0",
			stream: ffprobe.Stream{CodecName: "hevc", Profile: "Main", Level: 120},
			want:   "hvc1.1.6.L120.B0",
		},
		{
			name:   "hevc main 10 level 5.1",
			stream: ffprobe.Stream{CodecName: "h265", Profile: "Main 10", Level: 153},
			want:   "hvc1.2.4.L153.B0",
		},
		{
			name:   "hevc unknown profile",
			stream: ffprobe.Stream{CodecName: "hevc", Profile: "Rext", Level: 120},
		},
		{
			name:   "hevc unknown level",
			stream: ffprobe.Stream{CodecName: "hevc", Profile: "Main", Level: 0},
		},
		{
			name:   "av1 main level 4.0 8 bit",
			stream: ffprobe.Stream{CodecName: "av1", Profile: "Main", Level: 8, BitsPerRawSample: "8"},
			want:   "av01.0.08M.08",
		},
		{
			name:   "av1 pix fmt bit depth",
			stream: ffprobe.Stream{CodecName: "av1", Profile: "Main", Level: 13, PixFmt: "yuv420p10le"},
			want:   "av01.0.13M.10",
		},
		{
			name:   "av1 unknown profile",
			stream: ffprobe.Stream{CodecName: "av1", Profile: "Unknown", Level: 8, BitsPerRawSample: "8"},
		},
		{
			name:   "av1 unknown level",
			stream: ffprobe.Stream{CodecName: "av1", Profile: "Main", Level: -99, BitsPerRawSample: "8"},
		},
		{
			name:   "av1 reserved level",
			stream: ffprobe.Stream{CodecName: "av1", Profile: "Main", Level: 24, BitsPerRawSample: "8"},
		},
		{
			name:   "av1 unknown bit depth",
			stream: ffprobe.Stream{CodecName: "av1", Profile: "Main", Level: 8},
		},
		{
			name:   "vp9 profile 0",
			stream: ffprobe.Stream{CodecName: "vp9", Profile: "Profile 0", Level: 40},
			want:   "vp09.00.40.08",
		},
		{
			name:   "vp9 profile 2 10 bit",
			stream: ffprobe.Stream{CodecName: "vp9", Profile: "Profile 2", Level: 40, BitsPerRawSample: "10"},
			want:   "vp09.02.40.10",
		},
		{
			name:   "vp9 profile 3 pix fmt bit depth",
			stream: ffprobe.Stream{CodecName: "vp9", Profile: "Profile 3", Level: 41, PixFmt: "yuv444p12le"},
			want:   "vp09.03.41.12",
		},
		{
			name:   "vp9 unknown profile",
			stream: ffprobe.Stream{CodecName: "vp9", Profile: "Unknown", Level: 40},
		},
		{
			name:   "vp9 unknown level",
			stream: ffprobe.Stream{CodecName: "vp9", Profile: "Profile 0", Level: -99},
		},
		{
			name:   "vp9 undefined level",
			stream: ffprobe.Stream{CodecName: "vp9", Profile: "Profile 0", Level: 0},
		},
		{
			name:   "vp9 unknown high bit depth",
			stream: ffprobe.Stream{CodecName: "vp9", Profile: "Profile 2", Level: 40},
		},
		{
			name:   "vp8",
			stream: ffprobe.Stream{CodecName: "vp8"},
			want:   "vp8",
		},
		{
			name:   "aac lc",
			stream: ffprobe.Stream{CodecName: "aac", Profile: "LC"},
			want:   "mp4a.40.2",
		},
		{
			name:   "aac he",
			stream: ffprobe.Stream{CodecName: "aac", Profile: "HE-AAC"},
			want:   "mp4a.40.5",
		},
		{
			name:   "aac he v2",
			stream: ffprobe.Stream{CodecName: "aac", Profile: "HE-AACv2"},
			want:   "mp4a.40.29",
		},
		{
			name:   "aac unknown profile",
			stream: ffprobe.Stream{CodecName: "aac", Profile: "Unknown"},
		},
		{
			name:   "mp3",
			stream: ffprobe.Stream{CodecName: "mp3"},
			want:   "mp3",
		},
		{
			name:   "opus",
			stream: ffprobe.Stream{CodecName: "opus"},
			want:   "opus",
		},
		{
			name:   "vorbis",
			stream: ffprobe.Stream{CodecName: "vorbis"},
			want:   "vorbis",
		},
		{
			name:   "ac3",
			stream: ffprobe.Stream{CodecName: "ac3"},
			want:   "ac-3",
		},
		{
			name:   "eac3",
			stream: ffprobe.Stream{CodecName: "eac3"},
			want:   "ec-3",
		},
		{
			name:   "flac",
			stream: ffprobe.Stream{CodecName: "flac"},
			want:   "fLaC",
		},
		{
			name:   "alac",
			stream: ffprobe.Stream{CodecName: "alac"},
			want:   "alac",
		},
		{
			name:   "unknown codec",
			stream: ffprobe.Stream{CodecName: "dts"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := streamToMimeCodec(&tt.stream)
			if tt.want == "" {
				if got != nil {
					t.Fatalf("got %q, want nil", *got)
				}
				return
			}

			if got == nil {
				t.Fatalf("got nil, want %q", tt.want)
			}

			if *got != tt.want {
				t.Errorf("got %q, want %q", *got, tt.want)
			}
		})
	}
}

func TestContainerMimeType(t *testing.T) {
	tests := map[string]string{
		"mkv":  "video/matroska",
		".MKV": "video/matroska",
		"mka":  "audio/matroska",
		"mk3d": "video/matroska-3d",
		"mp4":  "video/mp4",
		"m4v":  "video/mp4",
		"webm": "video/webm",
		"mov":  "video/quicktime",
		"avi":  "video/x-msvideo",
		"":     "",
	}

	for extension, want := range tests {
		t.Run(extension, func(t *testing.T) {
			if got := containerMimeType(extension); got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}

func TestMediaMimeCodec(t *testing.T) {
	avc := "avc1.640028"
	aac := "mp4a.40.2"

	tests := []struct {
		name      string
		extension string
		videos    []Video
		audios    []Audio
		want      string
	}{
		{
			name:      "video and audio",
			extension: "mkv",
			videos:    []Video{{MimeCodec: &avc}},
			audios:    []Audio{{MimeCodec: &aac}},
			want:      "video/matroska; codecs=\"avc1.640028, mp4a.40.2\"",
		},
		{
			name:      "video only",
			extension: "mp4",
			videos:    []Video{{MimeCodec: &avc}},
			want:      "video/mp4; codecs=\"avc1.640028\"",
		},
		{
			name:      "audio only",
			extension: "mka",
			audios:    []Audio{{MimeCodec: &aac}},
			want:      "audio/matroska; codecs=\"mp4a.40.2\"",
		},
		{
			name:      "default audio",
			extension: "mkv",
			videos:    []Video{{MimeCodec: &avc}},
			audios: []Audio{
				{MimeCodec: nil},
				{MimeCodec: &aac, IsDefault: true},
			},
			want: "video/matroska; codecs=\"avc1.640028, mp4a.40.2\"",
		},
		{
			name:      "unknown video codec",
			extension: "mkv",
			videos:    []Video{{MimeCodec: nil}},
			audios:    []Audio{{MimeCodec: &aac}},
		},
		{
			name:      "unknown audio codec",
			extension: "mkv",
			videos:    []Video{{MimeCodec: &avc}},
			audios:    []Audio{{MimeCodec: nil}},
		},
		{
			name:      "unknown default audio codec",
			extension: "mkv",
			videos:    []Video{{MimeCodec: &avc}},
			audios: []Audio{
				{MimeCodec: &aac},
				{MimeCodec: nil, IsDefault: true},
			},
		},
		{
			name:      "no streams",
			extension: "mkv",
		},
		{
			name:      "unknown container",
			extension: "not-a-real-container-867",
			videos:    []Video{{MimeCodec: &avc}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mediaMimeCodec(tt.extension, tt.videos, tt.audios)
			if tt.want == "" {
				if got != nil {
					t.Fatalf("got %q, want nil", *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("got nil, want %q", tt.want)
			}
			if *got != tt.want {
				t.Errorf("got %q, want %q", *got, tt.want)
			}
		})
	}
}
