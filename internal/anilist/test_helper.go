package anilist

import (
	"github.com/goccy/go-json"
	"io"
	"log"
	"os"
	"path/filepath"
)

func MockGetAnilistClient() *Client {

	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// Open the JSON file
	file, err := os.Open(filepath.Join(path, "../../test/sample/jwt.json"))
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

	var data *struct{ JWT string }
	if err := json.Unmarshal(jsonData, &data); err != nil {
		println("Error unmarshaling JSON:", err.Error())
		return nil
	}

	anilistClient := NewAuthedClient(data.JWT)

	return anilistClient
}

func MockGetAllMedia() *[]*BaseMedia {

	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// Open the JSON file
	file, err := os.Open(filepath.Join(path, "../../test/sample/media.json"))
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

	var data []*BaseMedia
	if err := json.Unmarshal(jsonData, &data); err != nil {
		println("Error unmarshaling JSON:", err.Error())
		return nil
	}

	return &data
}

func MockGetCollection() *AnimeCollection {

	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// Open the JSON file
	file, err := os.Open(filepath.Join(path, "../../test/sample/anilist_collection.json"))
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

	var data AnimeCollection
	if err := json.Unmarshal(jsonData, &data); err != nil {
		println("Error unmarshaling JSON:", err.Error())
		return nil
	}

	return &data
}

func MockGetCollectionEntry(mId int) (*AnimeCollection_MediaListCollection_Lists_Entries, bool) {

	collection := MockGetCollection()
	if collection == nil {
		return nil, false
	}

	entries, ok := collection.GetListEntryFromMediaId(mId)

	return entries, ok
}
