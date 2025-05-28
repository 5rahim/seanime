package mkvparser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
