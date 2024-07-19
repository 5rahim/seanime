package filler

import (
	"github.com/davecgh/go-spew/spew"
	"seanime/internal/util"
	"testing"
)

func TestAnimeFillerList_Search(t *testing.T) {

	af := NewAnimeFillerList(util.NewLogger())

	opts := SearchOptions{
		Titles: []string{"Hunter x Hunter (2011)"},
	}

	ret, err := af.Search(opts)
	if err != nil {
		t.Error(err)
	}

	spew.Dump(ret)
}

func TestAnimeFillerList_FindFillerEpisodes(t *testing.T) {

	af := NewAnimeFillerList(util.NewLogger())

	ret, err := af.FindFillerEpisodes("/shows/one-piece")
	if err != nil {
		t.Error(err)
	}

	spew.Dump(ret)
}
