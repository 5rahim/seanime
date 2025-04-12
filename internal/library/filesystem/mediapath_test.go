package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"seanime/internal/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeparateFilePathS(t *testing.T) {
	tests := []struct {
		path              string
		potentialPrefixes []string
		expected          *SeparatedFilePath
	}{
		{
			path:              "/path/to/file.mkv",
			potentialPrefixes: []string{"/path/to", "/path"},
			expected:          &SeparatedFilePath{Filename: "file.mkv", Dirnames: []string{}},
		},
		{
			path:              "/path/TO/to/file.mkv",
			potentialPrefixes: []string{"/path"},
			expected:          &SeparatedFilePath{Filename: "file.mkv", Dirnames: []string{"TO", "to"}},
		},
		{
			path:              "/path/to/file2.mkv",
			potentialPrefixes: []string{},
			expected:          &SeparatedFilePath{Filename: "file2.mkv", Dirnames: []string{"path", "to"}},
		},
		{
			path:              "/mnt/Anime/Bungou Stray Dogs/Bungou Stray Dogs 5th Season/[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			potentialPrefixes: []string{"/mnt/Anime", "/mnt/Anime/", "/mnt", "/var/"},
			expected:          &SeparatedFilePath{Filename: "[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv", Dirnames: []string{"Bungou Stray Dogs", "Bungou Stray Dogs 5th Season"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			fmt.Println("Here")
			res := SeparateFilePathS(tt.path, tt.potentialPrefixes)
			assert.Equal(t, tt.expected.Filename, res.Filename)
			assert.Equal(t, tt.expected.Dirnames, res.Dirnames)
		})
	}
}

// Test with symlinks
func TestGetVideoFilePathsFromDir_WithSymlinks(t *testing.T) {
	tmpDir := t.TempDir()

	libDir := filepath.Join(tmpDir, "library")
	externalLibDir := t.TempDir()
	os.Mkdir(libDir, 0755)
	// Create files in the external directory
	createFile(t, filepath.Join(externalLibDir, "external_video1.mkv"))
	createFile(t, filepath.Join(externalLibDir, "external_video2.mp4"))

	// Create directories and files
	dir1 := filepath.Join(libDir, "Anime1")
	os.Mkdir(dir1, 0755)
	createFile(t, filepath.Join(dir1, "Anime1_1.mkv"))
	createFile(t, filepath.Join(dir1, "Anime1_2.mp4"))

	dir2 := filepath.Join(libDir, "Anime2")
	os.Mkdir(dir2, 0755)
	createFile(t, filepath.Join(dir2, "Anime2_1.mkv"))

	// Create a symlink to the external directory
	symlinkPath := filepath.Join(libDir, "symlink_to_external")
	if err := os.Symlink(externalLibDir, symlinkPath); err != nil {
		t.Fatalf("Failed to create symlink: %s", err)
	}
	// Create a recursive symlink to the library directory
	symlinkToLibPath := filepath.Join(externalLibDir, "symlink_to_library")
	if err := os.Symlink(libDir, symlinkToLibPath); err != nil {
		t.Fatalf("Failed to create symlink: %s", err)
	}

	// Expected files
	expectedPaths := []string{
		filepath.Join(dir1, "Anime1_1.mkv"),
		filepath.Join(dir1, "Anime1_2.mp4"),
		filepath.Join(dir2, "Anime2_1.mkv"),
		filepath.Join(externalLibDir, "external_video1.mkv"),
		filepath.Join(externalLibDir, "external_video2.mp4"),
	}

	filePaths, err := GetMediaFilePathsFromDirS(libDir)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	util.Spew(filePaths)

	// Check results
	for _, expected := range expectedPaths {
		found := false
		for _, path := range filePaths {
			// if path == expected {
			// 	found = true
			// 	break
			// }
			// Compare the paths using stdlib
			info1, err := os.Stat(path)
			if err != nil {
				t.Fatalf("Failed to get file info for %s: %s", path, err)
			}
			info2, err := os.Stat(expected)
			if err != nil {
				t.Fatalf("Failed to get file info for %s: %s", expected, err)
			}
			if os.SameFile(info1, info2) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected file path %s not found in result", expected)
		}
	}
}

func createFile(t *testing.T, path string) {
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create file: %s", err)
	}
	defer file.Close()
}
