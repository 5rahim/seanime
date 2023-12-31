package scanner

import (
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/util"
	"io"
	"os"
	"path/filepath"
	"testing"
)

//----------------------------------------------------------------------------------------------------------------------

func TestScanner_Scan(t *testing.T) {

	anilistClient, _, data := anilist.MockAnilistAccount()
	wsEventManager := events.NewMockWSEventManager(util.NewLogger())
	dir := "E:/Anime"

	tests := []struct {
		name  string
		paths []string
	}{
		{
			name: "Scan",
			paths: []string{
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 20v2 (1080p) [30072859].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 21v2 (1080p) [4B1616A5].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 22v2 (1080p) [58BF43B4].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 23v2 (1080p) [D94B4894].mkv",
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			existingLfs := make([]*entities.LocalFile, 0)
			for _, path := range tt.paths {
				lf := entities.NewLocalFile(path, dir)
				existingLfs = append(existingLfs, lf)
			}

			// +---------------------+
			// |        Scan         |
			// +---------------------+

			scanner := &Scanner{
				DirPath:            dir,
				Username:           data.Username,
				Enhanced:           false,
				AnilistClient:      anilistClient,
				Logger:             util.NewLogger(),
				WSEventManager:     wsEventManager,
				ExistingLocalFiles: existingLfs,
				SkipLockedFiles:    false,
				SkipIgnoredFiles:   false,
			}
			lfs, err := scanner.Scan()
			if err != nil {
				t.Fatal("expected result, got error:", err.Error())
			}

			for _, lf := range lfs {
				t.Log(lf.Name)
			}

		})

	}

}

func getMockedAllMedia(t *testing.T) []*anilist.BaseMedia {

	path, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// Open the JSON file
	file, err := os.Open(filepath.Join(path, "./scanner_test_mock_data.json"))
	if err != nil {
		t.Fatal("Error opening file:", err.Error())
	}
	defer file.Close()

	jsonData, err := io.ReadAll(file)
	if err != nil {
		t.Fatal("Error reading file:", err.Error())
	}

	var data struct {
		AllMedia []*anilist.BaseMedia `json:"allMedia"`
	}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		t.Fatal("Error unmarshaling JSON:", err.Error())
	}

	return data.AllMedia
}
