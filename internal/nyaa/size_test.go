package nyaa

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

		bytes, err := stringToBytes(test.size)
		if err != nil {
			t.Errorf("Error converting size to bytes: %s", err)
		}
		if bytes != test.bytes {
			t.Errorf("Expected %d bytes, got %d", test.bytes, bytes)
		}

	}

}
