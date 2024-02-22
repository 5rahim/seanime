package animetosho

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestSearch(t *testing.T) {
	results, err := Search("Akuyaku Reijou Level 99: Watashi wa Ura Boss desu ga Maou de wa Arimasen\n")
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(results)

	if len(results) == 0 {
		t.Fatal("no results")
	}

	for i, result := range results {
		if result.Title == "" {
			t.Errorf("results[%d].Title = %#v; want non-empty string", i, result.Title)
		}
		if result.URL == "" {
			t.Errorf("results[%d].URL = %#v; want non-empty string", i, result.URL)
		}
		if result.MagnetURL == "" {
			t.Errorf("results[%d].MagnetURL = %#v; want non-empty string", i, result.MagnetURL)
		}
		if result.TorrentURL == "" {
			t.Errorf("results[%d].TorrentURL = %#v; want non-empty string", i, result.TorrentURL)
		}
	}
}

func TestSearch2(t *testing.T) {
	err := Search2("metallic rouge")
	if err != nil {
		t.Fatal(err)
	}
}
