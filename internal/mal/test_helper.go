package mal

import (
	"github.com/goccy/go-json"
	"io"
	"log"
	"os"
	"path/filepath"
)

func MockJWTs() *struct {
	JWT       string `json:"jwt"`
	Username  string `json:"username"`
	JWT2      string `json:"jwt2"`
	Username2 string `json:"username2"`
	MALJwt    string `json:"mal_jwt"`
} {

	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// Open the JSON file
	file, err := os.Open(filepath.Join(path, "../../test/jwt.json"))
	if err != nil {
		println("Error opening file:", err.Error())
		return nil
	}
	defer file.Close()

	jsonData, err := io.ReadAll(file)
	if err != nil {
		println("Error reading file:", err.Error())
		return nil
	}

	var data *struct {
		JWT       string `json:"jwt"`
		Username  string `json:"username"`
		JWT2      string `json:"jwt2"`
		Username2 string `json:"username2"`
		MALJwt    string `json:"mal_jwt"`
	}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		println("Error unmarshaling JSON:", err.Error())
		return nil
	}

	return data

}
