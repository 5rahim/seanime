package debrid_client

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func PrintPathStructure(path string, indent string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	for _, entry := range entries {
		fmt.Println(indent + entry.Name())

		if entry.IsDir() {
			newIndent := indent + "  "
			newPath := filepath.Join(path, entry.Name())
			if err := PrintPathStructure(newPath, newIndent); err != nil {
				return err
			}
		} else {
		}
	}
	return nil
}

func TestCreateTempDir(t *testing.T) {

	files := []string{
		"/12345/Anime/Ep1.mkv",
		"/12345/Anime/Ep2.mkv",
	}

	root := t.TempDir()
	for _, file := range files {
		path := filepath.Join(root, file)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		if err := os.WriteFile(path, []byte("dummy content"), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", path, err)
		}
	}
	defer os.RemoveAll(root)

	err := PrintPathStructure(root, "")
	require.NoError(t, err)

}

func TestMoveContentsTo(t *testing.T) {
	tests := []struct {
		name      string
		files     []string
		dest      string
		expected  string
		expectErr bool
	}{
		{
			name: "Case 1: Move folder with files",
			files: []string{
				"/Anime/Ep1.mkv",
				"/Anime/Ep2.mkv",
			},
			dest:      "./dest",
			expected:  "./dest/Anime",
			expectErr: false,
		},
		{
			name: "Case 2: Move folder with single hash directory",
			files: []string{
				"/12345/Anime/Ep1.mkv",
				"/12345/Anime/Ep2.mkv",
			},
			dest:      "./dest",
			expected:  "./dest/Anime",
			expectErr: false,
		},
		{
			name: "Case 3: Move single file",
			files: []string{
				"/12345/Anime/Ep1.mkv",
			},
			dest:      "./dest",
			expected:  "./dest/Ep1.mkv",
			expectErr: false,
		},
		{
			name:      "Case 5: Source directory does not exist",
			files:     []string{},
			dest:      "./dest",
			expected:  "",
			expectErr: true,
		},
		{
			name: "Case 6: Move single file with hash directory",
			files: []string{
				"/12345/Anime/Ep1.mkv",
			},
			dest:     "./dest",
			expected: "./dest/Ep1.mkv",
		},
		{
			name: "Case 7",
			files: []string{
				"Ep1.mkv",
			},
			dest:     "./dest",
			expected: "./dest/Ep1.mkv",
		},
		{
			name: "Case 8",
			files: []string{
				"Ep1.mkv",
				"Ep2.mkv",
			},
			dest:     "./dest",
			expected: "./dest/Ep2.mkv",
		},
		{
			name: "Case 9",
			files: []string{
				"/12345/Anime/Anime 1/Ep1.mkv",
				"/12345/Anime/Anime 1/Ep2.mkv",
				"/12345/Anime/Anime 2/Ep1.mkv",
				"/12345/Anime/Anime 2/Ep2.mkv",
				"/12345/Anime 2/Anime 3/Ep1.mkv",
				"/12345/Anime 2/Anime 3/Ep2.mkv",
			},
			dest:      "./dest",
			expected:  "./dest/12345",
			expectErr: false,
		},
		{
			name: "Case 10",
			files: []string{
				"/Users/r/Downloads/b6aa416f662a2df83c6f5f79da95004ced59b8ef/Tsue to Tsurugi no Wistoria S01 1080p WEBRip DD+ x265-EMBER/[EMBER] Tsue to Tsurugi no Wistoria - 01.mkv",
				"/Users/r/Downloads/b6aa416f662a2df83c6f5f79da95004ced59b8ef/Tsue to Tsurugi no Wistoria S01 1080p WEBRip DD+ x265-EMBER/[EMBER] Tsue to Tsurugi no Wistoria - 02.mkv",
			},
			dest:      "./dest",
			expected:  "./dest/Tsue to Tsurugi no Wistoria S01 1080p WEBRip DD+ x265-EMBER",
			expectErr: false,
		},
		{
			name: "Case 11",
			files: []string{
				"/Users/rahim/Downloads/80431b4f9a12f4e06616062d3d3973b9ef99b5e6/[SubsPlease] Bocchi the Rock! - 01 (1080p) [E04F4EFB]/[SubsPlease] Bocchi the Rock! - 01 (1080p) [E04F4EFB].mkv",
			},
			dest:      "./dest",
			expected:  "./dest/[SubsPlease] Bocchi the Rock! - 01 (1080p) [E04F4EFB].mkv",
			expectErr: false,
		},
		{
			name: "Case 12",
			files: []string{
				"/tmp/.tmp-123456/[EMBER] Tsue to Tsurugi no Wistoria - 01.mkv",
			},
			dest:      "./dest",
			expected:  "./dest/[EMBER] Tsue to Tsurugi no Wistoria - 01.mkv",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the source directory structure
			root := t.TempDir()
			for _, file := range tt.files {
				path := filepath.Join(root, file)
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}
				if err := os.WriteFile(path, []byte("dummy content"), 0644); err != nil {
					t.Fatalf("failed to create file %s: %v", path, err)
				}
			}

			PrintPathStructure(root, "")
			println("-----------------------------")

			// Create the destination directory
			dest := t.TempDir()

			// Move the contents
			err := moveContentsTo(root, dest)

			if (err != nil) != tt.expectErr {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectErr {
				expected := filepath.Join(dest, filepath.Base(tt.expected))
				if _, err := os.Stat(expected); os.IsNotExist(err) {
					t.Errorf("expected directory or file does not exist: %s", expected)
				}

				PrintPathStructure(dest, "")
			}
		})
	}
}
