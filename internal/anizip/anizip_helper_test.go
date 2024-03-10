package anizip

import (
	"testing"
)

func TestOffsetEpisode(t *testing.T) {

	cases := []struct {
		input    string
		expected string
	}{
		{"S1", "S2"},
		{"OP1", "OP2"},
		{"1", "2"},
		{"OP", "OP"},
	}

	for _, c := range cases {
		actual := OffsetEpisode(c.input, 1)
		if actual != c.expected {
			t.Errorf("OffsetEpisode(%s, 1) == %s, expected %s", c.input, actual, c.expected)
		}
	}

}
