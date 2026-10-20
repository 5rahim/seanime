package mkvparser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertVTTToASS(t *testing.T) {
	content := `WEBVTT

00:00:00.000 --> 00:00:05.000 line:0% align:end size:15%
<ruby><rt>kaa.lt</rt></ruby>

00:01:05.566 --> 00:01:06.942
'Cause it'll make her heart race!

00:01:08.318 --> 00:01:09.403
Yeah?

00:02:40.619 --> 00:02:41.787 line:0%
What are these?

00:02:42.371 --> 00:02:44.873 line:0%
- They're love's secret weapon.
- Huh?

00:02:44.956 --> 00:02:48.168 line:0%
The plan is to use these boxes
to set up a dramatic surprise,

00:02:48.251 --> 00:02:50.754
to reveal the White Day gift
you forgot to give Hina.
`
	for name, from := range map[string]int{"explicit": SubtitleTypeWEBVTT, "detected": SubtitleTypeUnknown} {
		t.Run(name, func(t *testing.T) {
			out, err := ConvertToASS(content, from)
			require.NoError(t, err)
			require.Equal(t, 7, strings.Count(out, "Dialogue:"))
			require.Contains(t, out, "kaa.lt")
			require.Contains(t, out, "What are these?")
			require.Contains(t, out, "- They're love's secret weapon.\\N- Huh?")
		})
	}
	require.Equal(t, SubtitleTypeWEBVTT, DetectSubtitleType(content))
}
