package anilist

import (
	"github.com/seanime-app/seanime/internal/limiter"
	"github.com/seanime-app/seanime/internal/util"
	"testing"
)

func TestAddMediaToPlanning(t *testing.T) {

	_, data, _ := MockAnilistAccount()

	if data == nil {
		t.Skip("no mock data")
	}

	err := data.AddMediaToPlanning(
		[]int{131586},
		limiter.NewAnilistLimiter(),
		util.NewLogger(),
	)
	if err != nil {
		t.Fatal(err)
	}

}
