package entities

import (
	"github.com/goccy/go-json"
	lop "github.com/samber/lo/parallel"
	"io"
	"log"
	"os"
	"path/filepath"
)

func MockGetLocalFiles() ([]*LocalFile, bool) {

	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Open the JSON file
	file, err := os.Open(filepath.Join(path, "../../test/sample/localfiles.json"))
	if err != nil {
		println("Error opening file:", err.Error())
		return nil, false
	}
	defer file.Close()

	jsonData, err := io.ReadAll(file)
	if err != nil {
		println("Error reading file:", err.Error())
		return nil, false
	}

	var data []*LocalFile
	if err := json.Unmarshal(jsonData, &data); err != nil {
		println("Error unmarshaling JSON:", err.Error())
		return nil, false
	}

	return data, true

}

func MockGetLocalFilesByMediaId(mId int) ([]*LocalFile, bool) {
	lfs, ok := MockGetLocalFiles()
	if !ok {
		return nil, false
	}
	groupedByMediaId := lop.GroupBy(lfs, func(i *LocalFile) int {
		return i.MediaId
	})
	res, ok := groupedByMediaId[mId]

	return res, ok
}

func MockGetSelectedLocalFilesByMediaId(mId int) ([]*LocalFile, bool) {
	lfs, ok := MockGetLocalFiles()
	if !ok {
		return nil, false
	}
	groupedByMediaId := lop.GroupBy(lfs, func(i *LocalFile) int {
		return i.MediaId
	})
	res, ok := groupedByMediaId[mId]

	return res, ok
}

func MockGetSelectedLocalFiles() ([]*LocalFile, bool) {

	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// Open the JSON file
	file, err := os.Open(filepath.Join(path, "../../test/sample/localfiles_selected.json"))
	if err != nil {
		println("Error opening file:", err.Error())
		return nil, false
	}
	defer file.Close()

	jsonData, err := io.ReadAll(file)
	if err != nil {
		println("Error reading file:", err.Error())
		return nil, false
	}

	var data []*LocalFile
	if err := json.Unmarshal(jsonData, &data); err != nil {
		println("Error unmarshaling JSON:", err.Error())
		return nil, false
	}

	return data, true

}
