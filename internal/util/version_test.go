package util

import "testing"

func TestCompareVersion(t *testing.T) {
	testCases := []struct {
		name           string
		prevVersion    string
		currVersion    string
		expectedOutput int
		expectedBool   bool
	}{
		{
			name:           "Current version is newer by major version",
			prevVersion:    "1.0.0",
			currVersion:    "2.0.0",
			expectedOutput: 3,
			expectedBool:   true,
		},
		{
			name:           "Current version is older by major version",
			prevVersion:    "3.0.0",
			currVersion:    "2.0.0",
			expectedOutput: -3,
			expectedBool:   false,
		},
		{
			name:           "Current version is newer by minor version",
			prevVersion:    "1.1.0",
			currVersion:    "1.2.0",
			expectedOutput: 2,
			expectedBool:   true,
		},
		{
			name:           "Current version is older by minor version",
			prevVersion:    "1.3.0",
			currVersion:    "1.2.0",
			expectedOutput: -2,
			expectedBool:   false,
		},
		{
			name:           "Current version is newer by patch version",
			prevVersion:    "1.1.1",
			currVersion:    "1.1.2",
			expectedOutput: 1,
			expectedBool:   true,
		},
		{
			name:           "Current version is older by patch version",
			prevVersion:    "1.1.3",
			currVersion:    "1.1.2",
			expectedOutput: -1,
			expectedBool:   false,
		},
		{
			name:           "Versions are equal",
			prevVersion:    "1.1.1",
			currVersion:    "1.1.1",
			expectedOutput: 0,
			expectedBool:   false,
		},
		{
			name:           "Invalid version format",
			prevVersion:    "1.1",
			currVersion:    "1.1.1",
			expectedOutput: 0,
			expectedBool:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, boolOutput := CompareVersion(tc.prevVersion, tc.currVersion)
			if output != tc.expectedOutput || boolOutput != tc.expectedBool {
				t.Errorf("Expected output to be %d and bool output to be %v, got %d and %v", tc.expectedOutput, tc.expectedBool, output, boolOutput)
			}
		})
	}
}
