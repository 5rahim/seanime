package videofile

import "errors"

type Quality string

const (
	P240     Quality = "240p"
	P360     Quality = "360p"
	P480     Quality = "480p"
	P720     Quality = "720p"
	P1080    Quality = "1080p"
	P1440    Quality = "1440p"
	P4k      Quality = "4k"
	P8k      Quality = "8k"
	Original Quality = "original"
)

// Qualities Purposefully removing Original from this list (since it require special treatments anyway)
var Qualities = []Quality{P240, P360, P480, P720, P1080, P1440, P4k, P8k}

func QualityFromString(str string) (Quality, error) {
	if str == string(Original) {
		return Original, nil
	}

	qualities := Qualities
	for _, quality := range qualities {
		if string(quality) == str {
			return quality, nil
		}
	}
	return Original, errors.New("invalid quality string")
}

// AverageBitrate
// I'm not entirely sure about the values for bit rates. Double-checking would be nice.
func (v Quality) AverageBitrate() uint32 {
	switch v {
	case P240:
		return 400_000
	case P360:
		return 800_000
	case P480:
		return 1_200_000
	case P720:
		return 2_400_000
	case P1080:
		return 4_800_000
	case P1440:
		return 9_600_000
	case P4k:
		return 16_000_000
	case P8k:
		return 28_000_000
	case Original:
		panic("Original quality must be handled specially")
	}
	panic("Invalid quality value")
}

func (v Quality) MaxBitrate() uint32 {
	switch v {
	case P240:
		return 700_000
	case P360:
		return 1_400_000
	case P480:
		return 2_100_000
	case P720:
		return 4_000_000
	case P1080:
		return 8_000_000
	case P1440:
		return 12_000_000
	case P4k:
		return 28_000_000
	case P8k:
		return 40_000_000
	case Original:
		panic("Original quality must be handled specially")
	}
	panic("Invalid quality value")
}

func (v Quality) Height() uint32 {
	switch v {
	case P240:
		return 240
	case P360:
		return 360
	case P480:
		return 480
	case P720:
		return 720
	case P1080:
		return 1080
	case P1440:
		return 1440
	case P4k:
		return 2160
	case P8k:
		return 4320
	case Original:
		panic("Original quality must be handled specially")
	}
	panic("Invalid quality value")
}

func GetQualityFromHeight(height uint32) Quality {
	qualities := Qualities
	for _, quality := range qualities {
		if quality.Height() >= height {
			return quality
		}
	}
	return P240
}
