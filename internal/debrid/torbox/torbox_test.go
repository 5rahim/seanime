package torbox

import "testing"

func TestNormalizeDownloadURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "raw filename",
			input:    "https://example.com/dld/file?token=abc%2F123%3D&filename=The Ramparts of Ice S01E11.mkv",
			expected: "https://example.com/dld/file?token=abc%2F123%3D&filename=The%20Ramparts%20of%20Ice%20S01E11.mkv",
		},
		{
			name:     "encoded filename",
			input:    "https://example.com/dld/file?token=abc+123&filename=The%20Ramparts%20of%20Ice.mkv",
			expected: "https://example.com/dld/file?token=abc+123&filename=The%20Ramparts%20of%20Ice.mkv",
		},
		{
			name:     "raw path",
			input:    "https://example.com/dld/The Ramparts of Ice.mkv?token=abc%2F123%3D",
			expected: "https://example.com/dld/The%20Ramparts%20of%20Ice.mkv?token=abc%2F123%3D",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := normalizeDownloadUrl(tt.input)
			if err != nil {
				t.Fatalf("expected valid URL: %v", err)
			}
			if actual != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}
