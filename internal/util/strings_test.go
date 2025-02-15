package util

import "testing"

func TestSizeInBytes(t *testing.T) {

	tests := []struct {
		size  string
		bytes int64
	}{
		{"1.5 gb", 1610612736},
		{"1.5 GB", 1610612736},
		{"1.5 GiB", 1610612736},
		{"385.5 mib", 404226048},
	}

	for _, test := range tests {

		bytes, err := StringSizeToBytes(test.size)
		if err != nil {
			t.Errorf("Error converting size to bytes: %s", err)
		}
		if bytes != test.bytes {
			t.Errorf("Expected %d bytes, got %d", test.bytes, bytes)
		}

	}

}

func TestIsBase64Encoded(t *testing.T) {

	tests := []struct {
		str      string
		isBase64 bool
	}{
		{"SGVsbG8gV29ybGQ=", true},    // "Hello World"
		{"", false},                   // Empty string
		{"SGVsbG8gV29ybGQ", false},    // Invalid padding
		{"SGVsbG8gV29ybGQ==", false},  // Invalid padding
		{"SGVsbG8=V29ybGQ=", false},   // Padding in middle
		{"SGVsbG8gV29ybGQ!!", false},  // Invalid characters
		{"=SGVsbG8gV29ybGQ=", false},  // Padding at start
		{"SGVsbG8gV29ybGQ===", false}, // Too much padding
		{"A", false},                  // Single character
		{"AA==", true},                // Valid minimal string
		{"YWJjZA==", true},            // "abcd"
	}

	for _, test := range tests {
		if IsBase64(test.str) != test.isBase64 {
			t.Errorf("Expected %t for %s, got %t", test.isBase64, test.str, IsBase64(test.str))
		}
	}
}
