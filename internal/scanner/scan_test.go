package scanner

import (
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime/internal/anilist"
	"io"
	"os"
	"path/filepath"
	"testing"
)

//----------------------------------------------------------------------------------------------------------------------

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
