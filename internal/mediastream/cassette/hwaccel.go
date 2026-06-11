package cassette

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"seanime/internal/mediastream/videofile"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
)

// HwAccelOptions are configuration knobs for hardware acceleration
type HwAccelOptions struct {
	Kind           string
	Preset         string
	CustomSettings string // JSON-encoded HwAccelProfile for "custom" kind.
}

// BuildHwAccelProfile returns a profile for the requested hardware backend
func BuildHwAccelProfile(opts HwAccelOptions, ffmpegPath string, logger *zerolog.Logger) HwAccelProfile {
	name := opts.Kind
	if name == "" || name == "cpu" || name == "none" {
		name = "disabled"
	}

	// Handle custom JSON profile.
	var custom HwAccelProfile
	if name == "custom" {
		if opts.CustomSettings == "" {
			logger.Warn().Msg("cassette: custom hwaccel selected but no settings provided, falling back to CPU")
			name = "disabled"
		} else if err := json.Unmarshal([]byte(opts.CustomSettings), &custom); err != nil {
			logger.Error().Err(err).Msg("cassette: failed to parse custom hwaccel settings, falling back to CPU")
			name = "disabled"
		} else {
			custom.Name = "custom"
		}
	}

	// probe for the best encoder
	if name == "auto" {
		name = probeHardwareEncoder(ffmpegPath, logger)
	}

	logger.Debug().Str("backend", name).Msg("cassette: hardware acceleration resolved")

	defaultDevice := "/dev/dri/renderD128"
	if runtime.GOOS == "windows" {
		defaultDevice = "auto"
	}

	preset := opts.Preset
	if preset == "" {
		preset = "fast"
	}

	switch name {
	case "disabled":
		return cpuProfile(preset)
	case "vaapi":
		return vaApiProfile(defaultDevice)
	case "qsv", "intel":
		return qsvProfile(defaultDevice, preset, false)
	case "qsv-low-power", "qsv-lp", "intel-low-power", "intel-lp":
		return qsvProfile(defaultDevice, preset, true)
	case "nvidia":
		return nvidiaProfile(preset)
	case "videotoolbox":
		return videotoolboxProfile()
	case "custom":
		return custom
	default:
		logger.Warn().Str("name", name).Msg("cassette: unknown hwaccel, falling back to CPU")
		return cpuProfile(preset)
	}
}

// hardware probing

// probeHardwareEncoder tests encoders and returns the best backend
func probeHardwareEncoder(ffmpegPath string, logger *zerolog.Logger) string {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}

	type candidate struct {
		name    string
		encoder string
	}

	candidates := []candidate{
		{"nvidia", "h264_nvenc"},
		{"qsv", "h264_qsv"},
		{"vaapi", "h264_vaapi"},
	}
	if runtime.GOOS == "darwin" {
		candidates = append(candidates, candidate{"videotoolbox", "h264_videotoolbox"})
	}

	for _, c := range candidates {
		if testEncoder(ffmpegPath, c.encoder) {
			logger.Info().
				Str("encoder", c.encoder).
				Str("backend", c.name).
				Msg("cassette: hardware encoder probe succeeded")
			return c.name
		}
		logger.Trace().Str("encoder", c.encoder).Msg("cassette: hardware encoder probe failed")
	}

	logger.Info().Msg("cassette: no hardware encoder available, using CPU")
	return "disabled"
}

// TestEncoder attempts a minimal encode to verify if it works
func testEncoder(ffmpegPath, encoder string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Generate 1 frame of black video and encode it with the candidate
	// encoder. If this succeeds, the encoder is functional.
	cmd := exec.CommandContext(ctx, ffmpegPath,
		"-f", "lavfi",
		"-i", "color=black:s=64x64:d=0.04",
		"-c:v", encoder,
		"-frames:v", "1",
		"-f", "null", "-",
	)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// profile constructors

func cpuProfile(preset string) HwAccelProfile {
	return HwAccelProfile{
		Name:        "disabled",
		DecodeFlags: []string{},
		EncodeFlags: []string{
			"-c:v", "libx264",
			"-preset", preset,
			"-profile:v", "high", // ?
			"-tune", "animation", // ?
			// "-tune", "fastdecode,zerolatency", // ?
			"-sc_threshold", "0",
			"-pix_fmt", "yuv420p",
		},
		ScaleFilter:   "scale=%d:%d",
		NoScaleFilter: "format=yuv420p",
		ForcedIDR:     true,
	}
}

func vaApiProfile(device string) HwAccelProfile {
	return HwAccelProfile{
		Name: "vaapi",
		DecodeFlags: []string{
			"-hwaccel", "vaapi",
			"-hwaccel_device", GetEnvOr("SEANIME_TRANSCODER_VAAPI_RENDERER", device),
			"-hwaccel_output_format", "vaapi",
		},
		EncodeFlags: []string{
			"-c:v", "h264_vaapi",
			"-profile:v", "high", // ?
		},
		ScaleFilter:   "format=nv12|vaapi,hwupload,scale_vaapi=%d:%d:format=nv12",
		NoScaleFilter: "format=nv12|vaapi,hwupload",
		ForcedIDR:     true,
	}
}

func qsvProfile(device, preset string, lowPower bool) HwAccelProfile {
	name := "qsv"
	encodeFlags := []string{
		"-c:v", "h264_qsv",
		"-preset", preset,
		"-profile:v", "high", // ?
		"-async_depth", "1", // ? reduce latency
		"-look_ahead", "0", // ?
		"-bf", "3", // ?
	}
	if lowPower {
		name = "qsv-low-power"
		encodeFlags = []string{
			"-c:v", "h264_qsv",
			"-low_power", "1",
			"-preset", preset,
			"-profile:v", "high", // ?
			"-async_depth", "1", // ? reduce latency
			"-look_ahead", "0", // ?
			"-bf", "0", // low-power QSV is more broadly compatible without B-frames
		}
	}

	return HwAccelProfile{
		Name: name,
		DecodeFlags: []string{
			"-hwaccel", "qsv",
			"-qsv_device", GetEnvOr("SEANIME_TRANSCODER_QSV_RENDERER", device),
			"-hwaccel_output_format", "qsv",
		},
		EncodeFlags:   encodeFlags,
		ScaleFilter:   "format=nv12|qsv,hwupload,scale_qsv=%d:%d:format=nv12",
		NoScaleFilter: "format=nv12|qsv,hwupload",
		ForcedIDR:     true,
	}
}

func nvidiaProfile(preset string) HwAccelProfile {
	// map to nvenc presets
	switch preset {
	case "ultrafast":
		preset = "p1"
	case "superfast", "veryfast":
		preset = "p2"
	case "faster", "fast":
		preset = "p3"
	case "medium":
		preset = "p4"
	case "slow", "slower":
		preset = "p6"
	case "veryslow", "placebo":
		preset = "p7"
	}
	return HwAccelProfile{
		Name: "nvidia",
		DecodeFlags: []string{
			"-hwaccel", "cuda",
			"-hwaccel_output_format", "cuda",
		},
		EncodeFlags: []string{
			"-c:v", "h264_nvenc",
			"-preset", preset,
			"-profile:v", "high", // ?
			"-rc:v", "vbr", // ?
			"-bf", "0", // ?
			"-spatial-aq", "1", // ?
			"-temporal-aq", "1", // ?
			"-rc-lookahead", "0", // ?
			"-delay", "0",
			"-no-scenecut", "1",
		},
		ScaleFilter:   "format=nv12|cuda,hwupload,scale_cuda=%d:%d:format=nv12",
		NoScaleFilter: "format=nv12|cuda,hwupload",
		ForcedIDR:     true,
	}
}

func videotoolboxProfile() HwAccelProfile {
	return HwAccelProfile{
		Name: "videotoolbox",
		DecodeFlags: []string{
			"-hwaccel", "videotoolbox",
		},
		EncodeFlags: []string{
			"-c:v", "h264_videotoolbox",
			// "-realtime", "true",
			// "-prio_speed", "true",
			"-profile:v", "main",
		},
		ScaleFilter:   "scale=%d:%d",
		NoScaleFilter: "format=yuv420p",
		ForcedIDR:     true,
	}
}

// runtime fallback and adjustments

// BuildVideoFilter generates the scale filter string
func BuildVideoFilter(hw *HwAccelProfile, video *videofile.Video, width, height int32) string {
	noScale := false
	if video.Width == uint32(width) && video.Height == uint32(height) {
		noScale = true
	}

	lower := strings.ToLower(video.PixFmt)
	is10Bit := strings.Contains(lower, "10le") || strings.Contains(lower, "12le") || strings.Contains(lower, "p010")

	// use the scale filter to convert pixel formats if video is 10bit
	// even if we are not resizing the video
	// h264 encoders (nvenc, qsv, vaapi) typically only accept 8-bit formats (like nv12)
	if is10Bit && hw.Name != "custom" && hw.Name != "disabled" && hw.Name != "videotoolbox" {
		noScale = false
	}

	if hw.Name == "custom" {
		if noScale && hw.NoScaleFilter != "" {
			return hw.NoScaleFilter
		}
		return fmt.Sprintf(hw.ScaleFilter, width, height)
	}

	var filter string
	if noScale && hw.NoScaleFilter != "" {
		filter = hw.NoScaleFilter
	} else {
		filter = fmt.Sprintf(hw.ScaleFilter, width, height)
	}

	// Enable p010 hwupload if the source is 10-bit or 12-bit
	if is10Bit && strings.HasPrefix(filter, "format=nv12|") {
		filter = strings.Replace(filter, "format=nv12|", "format=p010|", 1)
	}

	// Software and Videotoolbox use scale= directly
	if !noScale && (hw.Name == "disabled" || hw.Name == "videotoolbox") {
		return filter
	}

	return filter
}

// FallbackToCPU returns a cpu profile
func FallbackToCPU(preset string) HwAccelProfile {
	return cpuProfile(preset)
}

// DetectHwAccelFailure checks for hardware acceleration failures
func DetectHwAccelFailure(stderr string) bool {
	lower := strings.ToLower(stderr)
	failureSignals := []string{
		"hwaccel", "vaapi", "cuvid", "vdpau", "qsv",
		"cuda", "nvenc", "videotoolbox",
		"no capable devices found",
		"device creation failed",
		"initialization failed",
	}
	if !strings.Contains(lower, "failed") && !strings.Contains(lower, "error") {
		return false
	}
	for _, sig := range failureSignals {
		if strings.Contains(lower, sig) {
			return true
		}
	}
	return false
}

// FormatHwAccelSummary returns a summary of the active profile
func FormatHwAccelSummary(p HwAccelProfile) string {
	if p.Name == "disabled" {
		return "CPU (software encoding)"
	}
	encoder := "unknown"
	for i, f := range p.EncodeFlags {
		if f == "-c:v" && i+1 < len(p.EncodeFlags) {
			encoder = p.EncodeFlags[i+1]
			break
		}
	}
	return fmt.Sprintf("%s (%s)", strings.ToUpper(p.Name), encoder)
}
