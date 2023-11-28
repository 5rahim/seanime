package entities

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

// mushoku tensei
const mediaEntryMediaid = 146065

// Fetching "Mushoku Tensei: Jobless Reincarnation Season 2" download info
// Anilist: 13 episodes, Anizip: 12 episodes + "S1"
// Info should include "S1" as episode 0
func TestNewMediaEntryDownloadInfo(t *testing.T) {

	mediaLfs, _ := MockGetSelectedLocalFilesByMediaId(mediaEntryMediaid)
	sort.Slice(mediaLfs, func(i, j int) bool {
		return mediaLfs[i].GetEpisodeNumber() < mediaLfs[j].GetEpisodeNumber()
	})

	anizipData, err := anizip.FetchAniZipMedia("anilist", mediaEntryMediaid)
	if err != nil {
		t.Fatalf("could not get anizip data for %d", mediaEntryMediaid)
	}

	anilistEntry, ok := anilist.MockGetCollectionEntry(mediaEntryMediaid)
	if !ok {
		t.Fatalf("could not get anilist entry for %d", mediaEntryMediaid)
	}

	info, err := NewMediaEntryDownloadInfo(&NewMediaEntryDownloadInfoOptions{
		localFiles:  nil,
		anizipMedia: anizipData,
		progress:    lo.ToPtr(0),
		status:      lo.ToPtr(anilist.MediaListStatusCurrent),
		media:       anilistEntry.Media,
	})

	if assert.NoError(t, err) {
		assert.NotNil(t, info)

		_, found := lo.Find(info.EpisodesToDownload, func(i *MediaEntryDownloadEpisode) bool {
			return i.EpisodeNumber == 0
		})

		assert.True(t, found)

		t.Log(spew.Sdump(info))
	}

}
