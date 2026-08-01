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
			name:     "filename with comma",
			input:    "https://example.com/dld/file?token=abc%2F123%3D&filename=Title, Subtitle S01E03 [1080p].zip",
			expected: "https://example.com/dld/file?token=abc%2F123%3D&filename=Title%2C%20Subtitle%20S01E03%20%5B1080p%5D.zip",
		},
		{
			name:     "filename with reserved and unicode characters",
			input:    "https://example.com/dld/file?token=abc%2F123%3D&filename=100% [Group] A&B #1 + 日本?.zip",
			expected: "https://example.com/dld/file?token=abc%2F123%3D&filename=100%25%20%5BGroup%5D%20A%26B%20%231%20%2B%20%E6%97%A5%E6%9C%AC%3F.zip",
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
