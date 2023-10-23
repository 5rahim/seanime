package scanner

import (
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"io"
	"os"
)

func MockGetTestLocalFiles() ([]*LocalFile, bool) {

	// Open the JSON file
	file, err := os.Open("../../test/sample/localfiles.json")
	if err != nil {
		println("Error opening file:", err)
		return nil, false
	}
	defer file.Close()

	jsonData, err := io.ReadAll(file)
	if err != nil {
		println("Error reading file:", err)
		return nil, false
	}

	var data []*LocalFile
	if err := json.Unmarshal(jsonData, &data); err != nil {
		println("Error unmarshaling JSON:", err)
		return nil, false
	}

	return data, true

}

type JWT struct {
	JWT string `json:"jwt"`
}

func MockGetAnilistClient() *anilist.Client {

	// Open the JSON file
	file, err := os.Open("../../test/sample/jwt.json")
	if err != nil {
		println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	jsonData, err := io.ReadAll(file)
	if err != nil {
		println("Error reading file:", err)
		return nil
	}

	var data *JWT
	if err := json.Unmarshal(jsonData, &data); err != nil {
		println("Error unmarshaling JSON:", err)
		return nil
	}

	anilistClient := anilist.NewAuthedClient(data.JWT)

	return anilistClient
}
