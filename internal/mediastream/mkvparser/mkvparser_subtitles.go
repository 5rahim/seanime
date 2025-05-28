package mkvparser

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/5rahim/go-astisub"
)

const (
	SubtitleTypeASS = iota
	SubtitleTypeSRT
	SubtitleTypeSTL
	SubtitleTypeTTML
	SubtitleTypeWEBVTT
	SubtitleTypeUnknown
)

func isProbablySrt(content string) bool {
	separatorCounts := strings.Count(content, "-->")
	return separatorCounts > 5
}

func DetectSubtitleType(content string) int {
	if strings.HasPrefix(strings.TrimSpace(content), "[Script Info]") {
		return SubtitleTypeASS
	} else if isProbablySrt(content) {
		return SubtitleTypeSRT
	} else if strings.Contains(content, "<tt ") || strings.Contains(content, "<tt>") {
		return SubtitleTypeTTML
	} else if strings.HasPrefix(strings.TrimSpace(content), "WEBVTT") {
		return SubtitleTypeWEBVTT
	} else if strings.Contains(content, "{\\") || strings.Contains(content, "\\N") {
		return SubtitleTypeSTL
	}
	return SubtitleTypeUnknown
}

func ConvertToASS(content string, from int) (string, error) {
	var o *astisub.Subtitles
	var err error

	reader := bytes.NewReader([]byte(content))

read:
	switch from {
	case SubtitleTypeSRT:
		o, err = astisub.ReadFromSRT(reader)
	case SubtitleTypeSTL:
		o, err = astisub.ReadFromSTL(reader, astisub.STLOptions{IgnoreTimecodeStartOfProgramme: true})
	case SubtitleTypeTTML:
		o, err = astisub.ReadFromTTML(reader)
	case SubtitleTypeWEBVTT:
		o, err = astisub.ReadFromWebVTT(reader)
	case SubtitleTypeUnknown:
		detectedType := DetectSubtitleType(content)
		if detectedType == SubtitleTypeUnknown {
			return "", fmt.Errorf("failed to detect subtitle format from content")
		}
		from = detectedType
		goto read
	default:
		return "", fmt.Errorf("unsupported subtitle format: %d", from)
	}

	if err != nil {
		return "", fmt.Errorf("failed to read subtitles: %w", err)
	}

	if o == nil {
		return "", fmt.Errorf("failed to read subtitles: %w", err)
	}

	o.Metadata = &astisub.Metadata{
		SSAScriptType:            "v4.00+",
		SSAWrapStyle:             "0",
		SSAPlayResX:              &[]int{640}[0],
		SSAPlayResY:              &[]int{360}[0],
		SSAScaledBorderAndShadow: true,
	}

	//Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
	//Style: Default, Roboto Medium,24,&H00FFFFFF,&H000000FF,&H00000000,&H00000000,0,0,0,0,100,100,0,0,1,1.3,0,2,20,20,23,0
	o.Styles["Default"] = &astisub.Style{
		ID: "Default",
		InlineStyle: &astisub.StyleAttributes{
			SSAFontName: "Roboto Medium",
			SSAFontSize: &[]float64{24}[0],
			SSAPrimaryColour: &astisub.Color{
				Red:   255,
				Green: 255,
				Blue:  255,
				Alpha: 0,
			},
			SSASecondaryColour: &astisub.Color{
				Red:   255,
				Green: 0,
				Blue:  0,
				Alpha: 0,
			},
			SSAOutlineColour: &astisub.Color{
				Red:   0,
				Green: 0,
				Blue:  0,
				Alpha: 0,
			},
			SSABackColour: &astisub.Color{
				Red:   0,
				Green: 0,
				Blue:  0,
				Alpha: 0,
			},
			SSABold:           &[]bool{false}[0],
			SSAItalic:         &[]bool{false}[0],
			SSAUnderline:      &[]bool{false}[0],
			SSAStrikeout:      &[]bool{false}[0],
			SSAScaleX:         &[]float64{100}[0],
			SSAScaleY:         &[]float64{100}[0],
			SSASpacing:        &[]float64{0}[0],
			SSAAngle:          &[]float64{0}[0],
			SSABorderStyle:    &[]int{1}[0],
			SSAOutline:        &[]float64{1.3}[0],
			SSAShadow:         &[]float64{0}[0],
			SSAAlignment:      &[]int{2}[0],
			SSAMarginLeft:     &[]int{20}[0],
			SSAMarginRight:    &[]int{20}[0],
			SSAMarginVertical: &[]int{23}[0],
			SSAEncoding:       &[]int{0}[0],
		},
	}

	for _, item := range o.Items {
		item.Style = &astisub.Style{
			ID: "Default",
		}
	}

	w := &bytes.Buffer{}
	err = o.WriteToSSA(w)
	if err != nil {
		return "", fmt.Errorf("failed to write subtitles: %w", err)
	}

	return w.String(), nil
}
