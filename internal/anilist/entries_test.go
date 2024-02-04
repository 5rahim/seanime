package anilist

import (
	"github.com/seanime-app/seanime/internal/limiter"
	"github.com/seanime-app/seanime/internal/util"
	"testing"
)

func TestAddMediaToPlanning(t *testing.T) {

	_, anilistClientWrapper, _ := MockAnilistClientWrappers()

	if anilistClientWrapper == nil {
		t.Skip("no mock data")
	}

	err := anilistClientWrapper.Client.AddMediaToPlanning(
		[]int{131586},
		limiter.NewAnilistLimiter(),
		util.NewLogger(),
	)
	if err != nil {
		t.Fatal(err)
	}

}
