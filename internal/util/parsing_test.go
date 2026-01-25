package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToBytes(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"1GB", 1073741824, false},
		{"1 GB", 1073741824, false},
		{"1.5 GB", 1610612736, false},
		{"1 GiB", 1073741824, false},
		{"500MB", 524288000, false},
		{"500 MiB", 524288000, false},
		{"100KB", 102400, false},
		{"1024 B", 1024, false},
		{"", 0, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			val, err := StringToBytes(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, val)
			}
		})
	}
}
