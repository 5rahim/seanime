package scanner

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	"github.com/seanime-app/seanime-server/internal/limiter"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const mediaId = 131586 // 86 part 2

func TestMediaTreeAnalysis(t *testing.T) {

	anilistRateLimiter := limiter.NewAnilistLimiter()

	// get media
	allMedia := anilist.MockGetAllMedia()
	media, found := lo.Find(*allMedia, func(m *anilist.BaseMedia) bool {
		return m.ID == mediaId
	})
	assert.True(t, found)

	// create the tree
	tree := anilist.NewBaseMediaRelationTree()

	// fetch the tree
	err := media.FetchMediaTree(anilist.FetchMediaTreeAll, anilist.NewAuthedClient(""), anilistRateLimiter, tree, anilist.NewBaseMediaCache())
	assert.NoError(t, err)

	// get analysis
	mta, err := NewMediaTreeAnalysis(&MediaTreeAnalysisOptions{
		tree:        tree,
		anizipCache: anizip.NewCache(),
		rateLimiter: limiter.NewLimiter(time.Minute, 25),
	})
	assert.NoError(t, err)

	t.Log(spew.Sdump(mta.branches))

	relEp, _, ok := mta.getRelativeEpisodeNumber(23)
	assert.True(t, ok)

	assert.Equal(t, 12, relEp)

}
