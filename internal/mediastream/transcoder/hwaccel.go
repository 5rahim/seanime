package transcoder

import (
	"runtime"

	"github.com/goccy/go-json"
)

type (
	HwAccelOptions struct {
		Kind           string
		Preset         string
		CustomSettings string
	}
)

func GetHardwareAccelSettings(opts HwAccelOptions) HwAccelSettings {
	name := opts.Kind
	if name == "" || name == "auto" || name == "cpu" || name == "none" {
		name = "disabled"
	}
	streamLogger.Debug().Msgf("transcoder: Hardware acceleration: %s", name)

	var customHwAccelSettings HwAccelSettings
	if opts.CustomSettings != "" && name == "custom" {
		err := json.Unmarshal([]byte(opts.CustomSettings), &customHwAccelSettings)
		if err != nil {
			streamLogger.Error().Err(err).Msg("transcoder: Failed to parse custom hardware acceleration settings, falling back to CPU")
			name = "disabled"
		}
		customHwAccelSettings.Name = "custom"
	} else if opts.CustomSettings == "" && name == "custom" {
		name = "disabled"
	}

	defaultOSDevice := "/dev/dri/renderD128"
	switch runtime.GOOS {
	case "windows":
		defaultOSDevice = "auto"
	}

	// superfast or ultrafast would produce heavy files, so opt for "fast" by default.
	// vaapi does not have any presets so this flag is unused for vaapi hwaccel.
	preset := opts.Preset

	switch name {
	case "disabled":
		return HwAccelSettings{
			Name:        "disabled",
			DecodeFlags: []string{},
			EncodeFlags: []string{
				"-c:v", "libx264",
				"-preset", preset,
				// sc_threshold is a scene detection mechanism used to create a keyframe when the scene changes
				// this is on by default and inserts keyframes where we don't want to (it also breaks force_key_frames)
				// we disable it to prevents whole scenes from being removed due to the -f segment failing to find the corresponding keyframe
				"-sc_threshold", "0",
				// force 8bits output (by default it keeps the same as the source but 10bits is not playable on some devices)
				"-pix_fmt", "yuv420p",
			},
			// we could put :force_original_aspect_ratio=decrease:force_divisible_by=2 here but we already calculate a correct width and
			// aspect ratio in our code so there is no need.
			ScaleFilter:   "scale=%d:%d",
			WithForcedIdr: true,
		}
	case "vaapi":
		return HwAccelSettings{
			Name: name,
			DecodeFlags: []string{
				"-hwaccel", "vaapi",
				"-hwaccel_device", GetEnvOr("SEANIME_TRANSCODER_VAAPI_RENDERER", defaultOSDevice),
				"-hwaccel_output_format", "vaapi",
			},
			EncodeFlags: []string{
				// h264_vaapi does not have any preset or scenecut flags.
				"-c:v", "h264_vaapi",
			},
			// if the hardware decoder could not work and fallback to soft decode, we need to instruct ffmpeg to
			// upload back frames to gpu space (after converting them)
			// see https://trac.ffmpeg.org/wiki/Hardware/VAAPI#Encoding for more info
			// we also need to force the format to be nv12 since 10bits is not supported via hwaccel.
			// this filter is equivalent to this pseudocode:
			// if (vaapi) {
			//   hwupload, passthrough, keep vaapi as is
			//   convert whatever to nv12 on GPU
			// } else {
			//   convert whatever to nv12 on CPU
			//   hwupload to vaapi(nv12)
			//   convert whatever to nv12 on GPU // scale_vaapi doesn't support passthrough option, so it has to make a copy
			// }
			// See https://www.reddit.com/r/ffmpeg/comments/1bqn60w/hardware_accelerated_decoding_without_hwdownload/ for more info
			ScaleFilter:   "format=nv12|vaapi,hwupload,scale_vaapi=%d:%d:format=nv12",
			WithForcedIdr: true,
		}
	case "qsv", "intel":
		return HwAccelSettings{
			Name: name,
			DecodeFlags: []string{
				"-hwaccel", "qsv",
				"-qsv_device", GetEnvOr("SEANIME_TRANSCODER_QSV_RENDERER", defaultOSDevice),
				"-hwaccel_output_format", "qsv",
			},
			EncodeFlags: []string{
				"-c:v", "h264_qsv",
				"-preset", preset,
			},
			// see note on ScaleFilter of the vaapi HwAccel, this is the same filter but adapted to qsv
			ScaleFilter:   "format=nv12|qsv,hwupload,scale_qsv=%d:%d:format=nv12",
			WithForcedIdr: true,
		}
	case "nvidia":
		return HwAccelSettings{
			Name: "nvidia",
			DecodeFlags: []string{
				"-hwaccel", "cuda",
				// this flag prevents data to go from gpu space to cpu space
				// it forces the whole dec/enc to be on the gpu. We want that.
				"-hwaccel_output_format", "cuda",
			},
			EncodeFlags: []string{
				"-c:v", "h264_nvenc",
				"-preset", preset,
				// the exivalent of -sc_threshold on nvidia.
				"-no-scenecut", "1",
			},
			// see note on ScaleFilter of the vaapi HwAccel, this is the same filter but adapted to cuda
			ScaleFilter:   "format=nv12|cuda,hwupload,scale_cuda=%d:%d:format=nv12",
			WithForcedIdr: true,
		}
	case "videotoolbox":
		return HwAccelSettings{
			Name: "videotoolbox",
			DecodeFlags: []string{
				"-hwaccel", "videotoolbox",
			},
			EncodeFlags: []string{
				"-c:v", "h264_videotoolbox",
				"-profile:v", "main",
			},
			ScaleFilter:   "scale=%d:%d",
			WithForcedIdr: true,
		}
	case "custom":
		return customHwAccelSettings
	default:
		streamLogger.Fatal().Msgf("No hardware accelerator named: %s", name)
		panic("unreachable")
	}
}
