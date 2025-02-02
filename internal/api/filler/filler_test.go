package filler

import (
	"seanime/internal/util"
	"testing"

	"github.com/davecgh/go-spew/spew"
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
