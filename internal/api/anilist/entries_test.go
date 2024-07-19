package anilist

import (
	"github.com/stretchr/testify/assert"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"seanime/internal/util/limiter"
	"testing"
)

func TestAddMediaToPlanning(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist(), test_utils.AnilistMutation())

	anilistClient := TestGetMockAnilistClient()

	err := anilistClient.AddMediaToPlanning(
		[]int{131586},
		limiter.NewAnilistLimiter(),
		util.NewLogger(),
	)
	assert.NoError(t, err)
}
