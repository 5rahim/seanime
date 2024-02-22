package animetosho

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestSearch2(t *testing.T) {
	torrents, err := Search("metallic rouge 05")
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(torrents)
}
