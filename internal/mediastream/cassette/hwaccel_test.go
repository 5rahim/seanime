package cassette

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestValidateCustomProfile(t *testing.T) {
	t.Run("accepts built-in profiles", func(t *testing.T) {
		profiles := []HwAccelProfile{
			cpuProfile("fast"),
			vaApiProfile("/dev/dri/renderD128"),
			qsvProfile("auto", "medium", false),
			qsvProfile("/dev/dri/renderD128", "medium", true),
			nvidiaProfile("slow"),
			videotoolboxProfile(),
		}
		for _, p := range profiles {
			assert.NoError(t, validateCustomProfile(&p), p.Name)
		}
	})

	tests := []struct {
		name    string
		profile HwAccelProfile
	}{
		{
			name: "rejects output file and muxer options from the report",
			profile: HwAccelProfile{
				EncodeFlags: []string{"-c:v", "libx264", "-f", "mpegts", "-y", "/tmp/WRITE_PROOF.ts"},
				ScaleFilter: "scale=%d:%d",
			},
		},
		{
			name: "rejects unknown option",
			profile: HwAccelProfile{
				EncodeFlags: []string{"-c:v", "libx264", "-x264-params", "a=b"},
				ScaleFilter: "scale=%d:%d",
			},
		},
		{
			name: "rejects dangling option without value",
			profile: HwAccelProfile{
				EncodeFlags: []string{"-c:v", "libx264", "-preset"},
				ScaleFilter: "scale=%d:%d",
			},
		},
		{
			name: "rejects option as value",
			profile: HwAccelProfile{
				EncodeFlags: []string{"-c:v", "-i"},
				ScaleFilter: "scale=%d:%d",
			},
		},
		{
			name: "rejects device outside dri",
			profile: HwAccelProfile{
				DecodeFlags: []string{"-hwaccel_device", "/etc/passwd"},
				ScaleFilter: "scale=%d:%d",
			},
		},
		{
			name: "rejects file reading filter in scale filter",
			profile: HwAccelProfile{
				ScaleFilter: "scale=%d:%d,movie=/etc/passwd",
			},
		},
		{
			name: "rejects filter chain separated by semicolon",
			profile: HwAccelProfile{
				ScaleFilter: "scale=%d:%d;movie=/etc/passwd",
			},
		},
		{
			name: "rejects unknown filter in no-scale filter",
			profile: HwAccelProfile{
				ScaleFilter:   "scale=%d:%d",
				NoScaleFilter: "amovie=/etc/passwd",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, validateCustomProfile(&tt.profile))
		})
	}
}

func TestBuildHwAccelProfileCustomValidation(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("uses valid custom profile", func(t *testing.T) {
		opts := HwAccelOptions{
			Kind:           "custom",
			CustomSettings: `{"name":"custom","decodeFlags":[],"encodeFlags":["-c:v","h264_vaapi","-profile:v","high"],"scaleFilter":"scale=%d:%d"}`,
		}
		p := BuildHwAccelProfile(opts, "ffmpeg", &logger)
		assert.Equal(t, "custom", p.Name)
	})

	t.Run("falls back to CPU for the report payload", func(t *testing.T) {
		opts := HwAccelOptions{
			Kind:           "custom",
			CustomSettings: `{"name":"custom","decodeFlags":[],"encodeFlags":["-c:v","libx264","-preset","ultrafast","-f","mpegts","-y","/tmp/WRITE_PROOF.ts"],"scaleFilter":"scale=%d:%d"}`,
		}
		p := BuildHwAccelProfile(opts, "ffmpeg", &logger)
		assert.Equal(t, "disabled", p.Name)
	})
}

func TestCustomOptions(t *testing.T) {
	logger := zerolog.Nop()
	p := BuildHwAccelProfile(HwAccelOptions{
		Kind:           "custom",
		CustomSettings: `{"encodeFlags":["-c:v","libx264","-crf","20.5","-tune","fastdecode,zerolatency"],"scaleFilter":"scale=%d:%d"}`,
	}, "ffmpeg", &logger)
	assert.Equal(t, "custom", p.Name)
	assert.Contains(t, p.EncodeFlags, "20.5")
	assert.Contains(t, p.EncodeFlags, "fastdecode,zerolatency")
}

func TestCustomFilterFiles(t *testing.T) {
	filters := []string{
		"scale=/w=/tmp/width:h=16",
		"scale=w=16:/h=/tmp/height",
		"scale='/w'=/tmp/width:h=16",
		`scale=\/w=/tmp/width:h=16`,
		"scale=w=16: '/h'=/tmp/height",
		`scale=w=16:\'/h\'=/tmp/height`,
		"format=/pix_fmts=/tmp/format",
	}
	for _, filter := range filters {
		t.Run(filter, func(t *testing.T) {
			assert.Error(t, validateCustomProfile(&HwAccelProfile{ScaleFilter: filter}))
			assert.Error(t, validateCustomProfile(&HwAccelProfile{ScaleFilter: "scale=%d:%d", NoScaleFilter: filter}))
		})
	}

	t.Run("allows division", func(t *testing.T) {
		assert.NoError(t, validateCustomFilter("scale=iw/2:ih/2"))
	})
}

func TestCustomFilterFormat(t *testing.T) {
	for _, filter := range []string{"scale=%d:%d", "scale=%[1]d:%[2]d", "scale=iw/2:ih/2"} {
		assert.NoError(t, validateCustomProfile(&HwAccelProfile{ScaleFilter: filter}))
	}
	// formatting a height of 47 with %c would produce a slash
	assert.Error(t, validateCustomProfile(&HwAccelProfile{ScaleFilter: "scale=%[2]cw=width.txt:h=16"}))
}
