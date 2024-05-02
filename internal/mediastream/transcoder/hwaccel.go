package transcoder

import (
	"log"
	"os"
)

func DetectHardwareAccel() HwAccelT {
	name := GetEnvOr("SEANIME_TRANSCODER_HWACCEL", "disabled")
	log.Printf("Using hardware acceleration: %s", name)

	switch name {
	case "disabled":
		return HwAccelT{
			Name:        "disabled",
			DecodeFlags: []string{},
			EncodeFlags: []string{
				"-c:v", "libx264",
				// superfast or ultrafast would produce a huge fige, so we prefer veryfast or faster.
				"-preset", "fast",
				// sc_threshold is a scene detection mechanism used to create a keyframe when the scene changes
				// this is on by default and inserts keyframes where we don't want to (it also breaks force_key_frames)
				// we disable it to prevents whole scenes from being removed due to the -f segment failing to find the corresponding keyframe
				"-sc_threshold", "0",
				"-pix_fmt", "yuv420p",
			},
			// we could put :force_original_aspect_ratio=decrease:force_divisible_by=2 here but we already calculate a correct width and
			// aspect ratio in our code so there is no need.
			ScaleFilter: "scale=%d:%d",
		}
	case "nvidia":
		return HwAccelT{
			Name: "nvidia",
			DecodeFlags: []string{
				"-hwaccel", "cuda",
				// this flag prevents data to go from gpu space to cpu space
				// it forces the whole dec/enc to be on the gpu. We want that.
				"-hwaccel_output_format", "cuda",
			},
			EncodeFlags: []string{
				"-c:v", "h264_nvenc",
				"-preset", "fast",
				// the exivalent of -sc_threshold on nvidia.
				"-no-scenecut", "1",
				"-pix_fmt", "yuv420p",
			},
			// if the decode goes into system memory, we need to prepend the filters with "hwupload_cuda".
			// since we use hwaccel_output_format, decoded data stays in gpu memory so we must not specify it (it errors)
			ScaleFilter: "scale_cuda=%d:%d",
		}
	case "vaapi":
		return HwAccelT{
			Name: name,
			DecodeFlags: []string{
				"-hwaccel", "vaapi",
				"-hwaccel_device", GetEnvOr("GOTRANSCODER_VAAPI_RENDERER", "/dev/dri/renderD128"),
				"-hwaccel_output_format", "vaapi",
			},
			EncodeFlags: []string{
				// h264_vaapi does not have any preset or scenecut flags.
				"-c:v", "h264_vaapi",
				// if the hardware decoder could not work and fallback to soft decode, we need to instruct ffmpeg to
				// upload back frames to gpu space (after converting them)
				// see https://trac.ffmpeg.org/wiki/Hardware/VAAPI#Encoding for more info
				// "-vf", "format=nv12|vaapi,hwupload",
				"-pix_fmt", "yuv420p",
			},
			ScaleFilter: "scale_vaapi=%d:%d",
		}
	case "qsv", "intel":
		return HwAccelT{
			Name: name,
			DecodeFlags: []string{
				"-hwaccel", "qsv",
				// "-qsv_device", GetEnvOr("GOTRANSCODER_QSV_RENDERER", "/dev/dri/renderD128"),
				"-hwaccel_output_format", "qsv",
			},
			EncodeFlags: []string{
				"-c:v", "h264_qsv",
				"-preset", "fast",
				"-pix_fmt", "yuv420p",
			},
			ScaleFilter: "scale_qsv=%d:%d",
		}
	default:
		log.Printf("No hardware accelerator named: %s", name)
		os.Exit(2)
		panic("unreachable")
	}
}
