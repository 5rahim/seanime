package scanner

import (
	"bytes"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/sourcegraph/conc/pool"
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

	var data []*entities.LocalFile
	if err := json.Unmarshal(jsonData, &data); err != nil {
		t.Error("Error unmarshaling JSON:", err)
		return
	}

	if err != nil {
		t.Error("expected success, got error")
	}

	titles := entities.GetUniqueAnimeTitlesFromLocalFiles(data)

	for _, title := range titles {
		fmt.Println(title)
	}

}

func TestLocalFile_GetTitleVariations(t *testing.T) {

	localFiles, ok := entities.MockGetLocalFiles()
	if !ok {
		t.Fatalf("expected local files")
	}

	p := pool.NewWithResults[[]*string]()
	for _, lf := range localFiles {
		lf := lf
		p.Go(func() []*string {
			return lf.GetTitleVariations()
		})
	}
	res := p.Wait()

	for _, r := range res {
		t.Log(formatArr(r...))
	}

}

func formatArr(arr ...*string) string {
	bf := bytes.NewBuffer([]byte{})
	for _, el := range arr {
		bf.WriteString(*el)
		bf.WriteString(" --- ")
	}
	return bf.String()
}
