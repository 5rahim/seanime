package mpv

import (
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

var testFilePath = "E:\\ANIME\\[SubsPlease] Bocchi the Rock! (01-12) (1080p) [Batch]\\[SubsPlease] Bocchi the Rock! - 01v2 (1080p) [ABDDAE16].mkv"

func TestMpv_OpenAndPlay(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	m := New(util.NewLogger(), "", "")

	err := m.OpenAndPlay(testFilePath)
	if err != nil {
		t.Fatal(err)
	}

	sub := m.Subscribe("test")

	go func() {
		time.Sleep(2 * time.Second)
		m.CloseAll()
	}()

	select {
	case v, _ := <-sub.Closed():
		t.Logf("mpv exited, %+v", v)
		break
	}

	t.Log("Done")

}

func TestMpv_OpenAndPlayPath(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	m := New(util.NewLogger(), "", test_utils.ConfigData.Provider.MpvPath)

	err := m.OpenAndPlay(testFilePath)
	if err != nil {
		t.Fatal(err)
	}

	sub := m.Subscribe("test")

	select {
	case v, _ := <-sub.Closed():
		t.Logf("mpv exited, %+v", v)
		break
	}

	t.Log("Done")

}

func TestMpv_Playback(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	m := New(util.NewLogger(), "", "")

	err := m.OpenAndPlay(testFilePath)
	if err != nil {
		t.Fatal(err)
	}

	sub := m.Subscribe("test")

loop:
	for {
		select {
		case v, _ := <-sub.Closed():
			t.Logf("mpv exited, %+v", v)
			break loop
		default:
			spew.Dump(m.GetPlaybackStatus())
			time.Sleep(2 * time.Second)
		}
	}

	t.Log("Done")

}

func TestMpv_Multiple(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	m := New(util.NewLogger(), "", "")

	err := m.OpenAndPlay(testFilePath)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	err = m.OpenAndPlay(testFilePath)
	if !assert.NoError(t, err) {
		t.Log("error opening mpv instance twice")
	}

	sub := m.Subscribe("test")

	go func() {
		time.Sleep(2 * time.Second)
		m.CloseAll()
	}()

	select {
	case v, _ := <-sub.Closed():
		t.Logf("mpv exited, %+v", v)
		break
	}

	t.Log("Done")

}

// Test parseArgs function
func TestParseArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
		hasError bool
	}{
		{
			name:     "simple arguments",
			input:    "--fullscreen --volume=50",
			expected: []string{"--fullscreen", "--volume=50"},
			hasError: false,
		},
		{
			name:     "double quoted argument",
			input:    "--title=\"My Movie Name\"",
			expected: []string{"--title=My Movie Name"},
			hasError: false,
		},
		{
			name:     "single quoted argument",
			input:    "--title='My Movie Name'",
			expected: []string{"--title=My Movie Name"},
			hasError: false,
		},
		{
			name:     "space separated quoted argument",
			input:    "--title \"My Movie Name\"",
			expected: []string{"--title", "My Movie Name"},
			hasError: false,
		},
		{
			name:     "single space separated quoted argument",
			input:    "--title 'My Movie Name'",
			expected: []string{"--title", "My Movie Name"},
			hasError: false,
		},
		{
			name:     "mixed arguments",
			input:    "--fullscreen --title \"My Movie\" --volume=50",
			expected: []string{"--fullscreen", "--title", "My Movie", "--volume=50"},
			hasError: false,
		},
		{
			name:     "path with spaces",
			input:    "--subtitle-file \"C:\\Program Files\\subtitles\\movie.srt\"",
			expected: []string{"--subtitle-file", "C:\\Program Files\\subtitles\\movie.srt"},
			hasError: false,
		},
		{
			name:     "escaped quotes",
			input:    "--title \"Movie with \\\"quotes\\\" in title\"",
			expected: []string{"--title", "Movie with \"quotes\" in title"},
			hasError: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
			hasError: false,
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: []string{},
			hasError: false,
		},
		{
			name:     "tabs and spaces",
			input:    "--fullscreen\t\t--volume=50   --loop",
			expected: []string{"--fullscreen", "--volume=50", "--loop"},
			hasError: false,
		},
		{
			name:     "unclosed double quote",
			input:    "--title \"My Movie",
			expected: nil,
			hasError: true,
		},
		{
			name:     "unclosed single quote",
			input:    "--title 'My Movie",
			expected: nil,
			hasError: true,
		},
		{
			name:     "nested quotes",
			input:    "--title \"Movie 'with' nested quotes\"",
			expected: []string{"--title", "Movie 'with' nested quotes"},
			hasError: false,
		},
		{
			name:     "complex mixed case",
			input:    "--fullscreen --title=\"Complex Movie\" --volume 75 --subtitle-file 'path/with spaces/sub.srt' --loop",
			expected: []string{"--fullscreen", "--title=Complex Movie", "--volume", "75", "--subtitle-file", "path/with spaces/sub.srt", "--loop"},
			hasError: false,
		},
		{
			name:     "empty quoted string",
			input:    "--title \"\"",
			expected: []string{"--title", ""},
			hasError: false,
		},
		{
			name:     "multiple spaces between args",
			input:    "--fullscreen     --volume=50",
			expected: []string{"--fullscreen", "--volume=50"},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseArgs(tt.input)

			if tt.hasError {
				assert.Error(t, err, "Expected error for input: %q", tt.input)
				assert.Nil(t, result, "Expected nil result when error occurs")
			} else {
				assert.NoError(t, err, "Unexpected error for input: %q", tt.input)
				assert.Equal(t, tt.expected, result, "Mismatch for input: %q", tt.input)
			}
		})
	}
}
