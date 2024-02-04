package anilist

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestClientCustomDo(t *testing.T) {

	// Get Anilist client
	anilistClientWrapper := MockAnilistClientWrapper()

	id := 1

	res, err := anilistClientWrapper.Client.BaseMediaByID(context.Background(), &id)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Success!")
	spew.Dump(res)

}
