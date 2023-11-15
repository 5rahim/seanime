package nyaa

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearch(t *testing.T) {

	res, err := Search(SearchOptions{
		Provider: "nyaa",
		Query:    "one piece",
		Category: "anime",
		SortBy:   "downloads",
		Filter:   "",
	})

	if err != nil {
		t.Fatal(err)
	}

	for _, torrent := range res {
		t.Log(torrent)
	}
}

func TestBuildSearchQuery(t *testing.T) {

	collec := anilist.MockGetCollection()
	assert.NotNil(t, collec)

	//entry, found := collec.GetListEntryFromMediaId(161645) //
	//entry, found := collec.GetListEntryFromMediaId(145064) // jjk2
	//entry, found := collec.GetListEntryFromMediaId(163205) // mononogatari 2
	entry, found := collec.GetListEntryFromMediaId(146065) // mushoku tensei season 2
	//entry, found := collec.GetListEntryFromMediaId(140439) // mob psycho 3
	//entry, found := collec.GetListEntryFromMediaId(119661) // rezero season 2 part 2
	//entry, found := collec.GetListEntryFromMediaId(131681) // attack on titan season 4 part 2
	//entry, found := collec.GetListEntryFromMediaId(154116) // undead unluck
	assert.True(t, found)
	assert.NotNil(t, entry.Media)

	ret, ok := BuildSearchQuery(&BuildSearchQueryOptions{
		Media:          entry.Media,
		Batch:          lo.ToPtr(true),
		EpisodeNumber:  lo.ToPtr(6),
		AbsoluteOffset: lo.ToPtr(0),
		Quality:        lo.ToPtr("1080"),
	})
	assert.True(t, ok)

	t.Log(spew.Sdump(ret))

}
