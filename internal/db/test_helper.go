package db

import (
	"github.com/goccy/go-json"
	"io"
	"log"
	"os"
	"path/filepath"
)

func GetTestDatabaseInfo() *struct {
	DataDir string `json:"dataDir"`
	Name    string `json:"name"`
} {

	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// Open the JSON file
	file, err := os.Open(filepath.Join(path, "../../test/db.json"))
	if err != nil {
		println("Error opening file:", err.Error())
		log.Fatal(err)
		return nil
	}
	defer file.Close()

	jsonData, err := io.ReadAll(file)
	if err != nil {
		println("Error reading file:", err.Error())
		log.Fatal(err)
		return nil
	}

	var data *struct {
		DataDir string `json:"dataDir"`
		Name    string `json:"name"`
	}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		println("Error unmarshaling JSON:", err.Error())
		log.Fatal(err)
		return nil
	}

	return data

}
