package videocore

import (
	"seanime/internal/mkvparser"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInSight_CleanSubtitle(t *testing.T) {
	is := &InSight{}

	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "Hello World"},
		{"{Hello} World", "World"},
		{"{\\an8}Hello", "Hello"},
		{"<i>Hello</i>", "Hello"},
		{"  Hello  ", "Hello"},
	}

	for _, tt := range tests {
		result := is.CleanSubtitle(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestInSight_FindMatches(t *testing.T) {
	is := &InSight{}

	cache := []insightSearchEntry{
		{MalID: 1, Tokens: []string{"miyuki", "shirogane"}},
		{MalID: 2, Tokens: []string{"kaguya", "shinomiya"}},
		{MalID: 3, Tokens: []string{"chika", "fujiwara"}},
	}

	tests := []struct {
		text     string
		expected []int
	}{
		{"Miyuki aaaaaa", []int{1}},
		{"Kaguya aaaa aa aa a miyuki", []int{1, 2}},
		{"Chika aaaa a aaa aaa", []int{3}},
		{"aaaaa,Ishigami aaaa", []int{}},
		{"aaaaaa, Shirogane aaaaaa aaa a", []int{1}},
	}

	for _, tt := range tests {
		result := is.FindMatches(tt.text, cache)
		assert.ElementsMatch(t, tt.expected, result)
	}
}

func TestInSight_CreateSegments(t *testing.T) {
	is := &InSight{}
	matches := []int{1, 2}
	startTime := 10.0
	duration := 5.0

	segments := is.CreateSegments(matches, startTime, duration)

	assert.Len(t, segments, 2)
	assert.Equal(t, 1, segments[0].CharacterId)
	assert.Equal(t, 10.0, segments[0].StartTime)
	assert.Equal(t, 18.0, segments[0].EndTime)

	assert.Equal(t, 2, segments[1].CharacterId)
	assert.Equal(t, 10.0, segments[1].StartTime)
	assert.Equal(t, 18.0, segments[1].EndTime)

	// test minimum duration
	startTime = 10.0
	duration = 1.0
	segments = is.CreateSegments(matches, startTime, duration)
	// end time should be start + min duration (6) = 16
	// since 10+1+3 = 14 which is < 16
	assert.Equal(t, 16.0, segments[0].EndTime)
}

func TestInSight_Analyze(t *testing.T) {
	is := NewInSight(nil, nil)
	is.characters = []*InSightCharacter{
		{MalID: 1, Name: "Miyuki Shirogane"},
	}
	is.searchCache = []insightSearchEntry{
		{MalID: 1, Tokens: []string{"miyuki", "shirogane"}},
	}

	events := []*mkvparser.SubtitleEvent{
		{StartTime: 0, Duration: 2, Text: "Miyuki!"},
		{StartTime: 5, Duration: 2, Text: "Hello"},
	}

	is.Analyze(events)

	assert.NotNil(t, is.inSightData)
	assert.Len(t, is.inSightData.Suggestions, 1)
	assert.Equal(t, 1, is.inSightData.Suggestions[0].CharacterId)
}

func TestInSight_ParseSubtitleContent(t *testing.T) {
	is := &InSight{}

	srtContent := `1
00:00:01,000 --> 00:00:04,000
Hello World

2
00:00:05,000 --> 00:00:08,000
Second Line`

	events, err := is.ParseSubtitleContent(srtContent, "srt")
	assert.NoError(t, err)
	assert.Len(t, events, 2)
	assert.Equal(t, 1.0, events[0].StartTime)
	assert.Equal(t, 3.0, events[0].Duration)
	assert.Equal(t, "Hello World", events[0].Text)

	assert.Equal(t, 5.0, events[1].StartTime)
	assert.Equal(t, 3.0, events[1].Duration)
	assert.Equal(t, "Second Line", events[1].Text)
}
