package entities

import (
	"github.com/seanime-app/seanime-server/internal/anify"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMissingEpisodes(t *testing.T) {

	anilistCollection := anilist.MockGetCollection()
	lfs, ok := MockGetSelectedLocalFiles()
	assert.True(t, ok)

	missingEps := NewMissingEpisodes(&NewMissingEpisodesOptions{
		AnilistCollection:          anilistCollection,
		LocalFiles:                 lfs,
		AnizipCache:                anizip.NewCache(),
		AnifyEpisodeImageContainer: anify.NewEpisodeImageContainer(),
	})

	assert.Equal(t, 5, len(missingEps.Episodes))

}
