package anilist

import (
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/limiter"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddMediaToPlanning(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist(), test_utils.AnilistMutation())

	anilistClientWrapper := TestGetMockAnilistClientWrapper()

	err := anilistClientWrapper.AddMediaToPlanning(
		[]int{131586},
		limiter.NewAnilistLimiter(),
		util.NewLogger(),
	)
	assert.NoError(t, err)
}
