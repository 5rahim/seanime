package entities

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/stretchr/testify/assert"
	"testing"
)

const mediaId = 135778

// TestNewMediaEntry tests /library/entry endpoint.
func TestNewMediaEntry(t *testing.T) {

	localFiles, ok := MockGetSelectedLocalFiles()
	if !ok {
		t.Fatal("failed to get local files")
	}
	anilistCollection := anilist.MockGetCollection()
	assert.NotNil(t, anilistCollection)

	entry, err := NewMediaEntry(&NewMediaEntryOptions{
		MediaId:           mediaId,
		LocalFiles:        localFiles,
		AnizipCache:       anizip.NewCache(),
		AnilistCollection: anilistCollection,
		AnilistClient:     anilist.MockGetAnilistClient(),
	})
	if assert.NoError(t, err) {
		assert.NotNil(t, entry)

		assert.Equal(t, mediaId, entry.MediaId)
		t.Logf(spew.Sdump(entry.Episodes))
	}

}
