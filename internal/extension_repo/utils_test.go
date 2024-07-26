package extension_repo

import "testing"

func TestExtensionID(t *testing.T) {

	tests := []struct {
		id       string
		expected bool
	}{
		{"my-extension", true},
		{"my-extension-", false},
		{"-my-extension", false},
		{"my-extension-1", true},
		{"my.extension", false},
		{"my_extension", false},
	}

	for _, test := range tests {
		if isValidExtensionIDString(test.id) != test.expected {
			t.Errorf("isValidExtensionID(%v) != %v", test.id, test.expected)
		}
	}

}
