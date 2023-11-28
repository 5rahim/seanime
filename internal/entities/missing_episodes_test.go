package entities

import (
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMissingEpisodes(t *testing.T) {

	anilistCollection := anilist.MockGetCollection()
	lfs, ok := MockGetSelectedLocalFiles()
	assert.True(t, ok)

	missingEps := NewMissingEpisodes(&NewMissingEpisodesOptions{
		AnilistCollection: anilistCollection,
		LocalFiles:        lfs,
		AnizipCache:       anizip.NewCache(),
	})

	assert.Equal(t, 5, len(missingEps.Episodes))

}
