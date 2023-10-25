package scanner

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime-server/internal/util"
	"io"
	"os"
	"testing"
)

func TestGetLocalFilesFromDir(t *testing.T) {
	logger := util.NewLogger()

	os.Setenv("SEA_LOCAL_DIR", "E:/Anime")

	localFiles, err := GetLocalFilesFromDir(os.Getenv("SEA_LOCAL_DIR"), logger)
	if err != nil {
		t.Error("expected localfiles, got error")
	}

	fmt.Printf("localFiles: %v", localFiles)

}

func TestGetUniqueAnimeTitles(t *testing.T) {

	// Open the JSON file
	file, err := os.Open("../../test/sample/localfiles.json")
	if err != nil {
		t.Error("Error opening file:", err)
		return
	}
	defer file.Close()

	jsonData, err := io.ReadAll(file)
	if err != nil {
		t.Error("Error reading file:", err)
		return
	}

	var data []*LocalFile
	if err := json.Unmarshal(jsonData, &data); err != nil {
		t.Error("Error unmarshaling JSON:", err)
		return
	}

	if err != nil {
		t.Error("expected success, got error")
	}

	titles := GetUniqueAnimeTitles(data)

	for _, title := range titles {
		fmt.Println(title)
	}

}
