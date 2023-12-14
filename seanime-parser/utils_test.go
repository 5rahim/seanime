package seanime_parser

import (
	"testing"
)

func TestIsNumberLike(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"01", true},
		{"01v2", true},
		{"01x10", true},
		{"05'", true},

		{"1234v", false},
		{"1234z", false},
		{"test123", false},
		{"1234Xv2", false},
		{"125X'12", false},
		{"12A34X", false},
		{"", false},
	}

	for _, tc := range tests {
		if got := isNumberLike(tc.input); got != tc.want {
			t.Errorf("isNumberLike(%q) = %v; want %v", tc.input, got, tc.want)
		}
	}
}

func TestIsOrdinalNumber(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"1st", true},
		{"2nd", true},
		{"3rd", true},
		{"4th", true},
		{"60th", true},

		{"1234v", false},
		{"1234z", false},
		{"test123", false},
		{"1234Xv2", false},
		{"125X'12", false},
		{"12A34X", false},
		{"", false},
	}

	for _, tc := range tests {
		if got := isOrdinalNumber(tc.input); got != tc.want {
			t.Errorf("isOrdinalNumber(%q) = %v; want %v", tc.input, got, tc.want)
		}
	}
}

func TestIsResolution(t *testing.T) {
	var tests = []struct {
		input string
		want  bool
	}{
		{"1920x1080", true},
		{"720p", true},
		{"1080P", true},
		{"1280x720", true},
		{"FalseResolution1000x", false},
		{"FalseResolutionp", false},
		{"FalseResolutionP", false},
		{"", false},
	}

	for _, test := range tests {
		if got := isResolution(test.input); got != test.want {
			t.Errorf("isResolution(%q) = %v", test.input, got)
		}
	}
}
