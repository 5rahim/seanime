package entities

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	"github.com/stretchr/testify/assert"
	"testing"
)

const mediaEntryMediaid = 98314

func TestNewMediaEntryDownloadInfo(t *testing.T) {

	mediaLfs, _ := MockGetSelectedLocalFilesByMediaId(mediaEntryMediaid)

	anizipData, err := anizip.FetchAniZipMedia("anilist", mediaEntryMediaid)
	if err != nil {
		t.Fatalf("could not get anizip data for %d", mediaEntryMediaid)
	}

	anilistEntry, ok := anilist.MockGetCollectionEntry(mediaEntryMediaid)
	if !ok {
		t.Fatalf("could not get anilist entry for %d", mediaEntryMediaid)
	}

	info, err := NewMediaEntryDownloadInfo(&NewMediaEntryDownloadInfoOptions{
		localFiles:   mediaLfs,
		anizipMedia:  anizipData,
		anilistEntry: anilistEntry,
		media:        anilistEntry.Media,
	})

	if assert.NoError(t, err) {
		assert.NotNil(t, info)

		t.Log(spew.Sdump(info))
	}

}
