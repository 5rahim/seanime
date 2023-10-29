package nyaa

import "testing"

func TestSearch(t *testing.T) {

	res, err := Search(SearchOptions{
		Provider: "nyaa",
		Query:    "one piece",
		Category: "anime",
		SortBy:   "downloads",
		Filter:   "",
	})

	if err != nil {
		t.Fatal(err)
	}

	for _, torrent := range res {
		t.Log(torrent)
	}
}
