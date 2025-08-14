package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidVideoExtension(t *testing.T) {
	tests := []struct {
		ext      string
		expected bool
	}{
		{ext: ".mp4", expected: true},
		{ext: ".avi", expected: true},
		{ext: ".mkv", expected: true},
		{ext: ".mov", expected: true},
		{ext: ".unknown", expected: false},
		{ext: ".MP4", expected: true},
		{ext: ".AVI", expected: true},
		{ext: "", expected: false},
	}

	for _, test := range tests {
		t.Run(test.ext, func(t *testing.T) {
			result := IsValidVideoExtension(test.ext)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestSubdirectory(t *testing.T) {
	tests := []struct {
		parent   string
		child    string
		expected bool
	}{
		{parent: "C:\\parent", child: "C:\\parent\\child", expected: true},
		{parent: "C:\\parent", child: "C:\\parent\\child.txt", expected: true},
		{parent: "C:\\parent", child: "C:/PARENT/child.txt", expected: true},
		{parent: "C:\\parent", child: "C:\\parent\\..\\child", expected: false},
		{parent: "C:\\parent", child: "C:\\parent", expected: false},
	}

	for _, test := range tests {
		t.Run(test.child, func(t *testing.T) {
			result := IsSubdirectory(test.parent, test.child)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestIsFileUnderDir(t *testing.T) {
	tests := []struct {
		parent   string
		child    string
		expected bool
	}{
		{parent: "C:\\parent", child: "C:\\parent\\child", expected: true},
		{parent: "C:\\parent", child: "C:\\parent\\child.txt", expected: true},
		{parent: "C:\\parent", child: "C:/PARENT/child.txt", expected: true},
		{parent: "C:\\parent", child: "C:\\parent\\..\\child", expected: false},
		{parent: "C:\\parent", child: "C:\\parent", expected: false},
	}

	for _, test := range tests {
		t.Run(test.child, func(t *testing.T) {
			result := IsFileUnderDir(test.parent, test.child)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestSameDir(t *testing.T) {
	tests := []struct {
		dir1     string
		dir2     string
		expected bool
	}{
		{dir1: "C:\\dir", dir2: "C:\\dir", expected: true},
		{dir1: "C:\\dir", dir2: "C:\\DIR", expected: true},
		{dir1: "C:\\dir1", dir2: "C:\\dir2", expected: false},
	}

	for _, test := range tests {
		t.Run(test.dir2, func(t *testing.T) {
			result := IsSameDir(test.dir1, test.dir2)
			require.Equal(t, test.expected, result)
		})
	}
}
