package util

import "testing"

func TestToHumanReadableSize(t *testing.T) {

	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0.0 B"},
		{1, "1.0 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, test := range tests {
		actual := ToHumanReadableSize(int(test.size))
		if actual != test.expected {
			t.Errorf("ToHumanReadableSize(%d): expected %s, actual %s", test.size, test.expected, actual)
		}
	}

}

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
