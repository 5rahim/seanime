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
		Category: "anime-eng",
		SortBy:   "seeders",
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
	entry, found := collec.GetListEntryFromMediaId(145064) // jjk2
	//entry, found := collec.GetListEntryFromMediaId(163205) // mononogatari 2
	//entry, found := collec.GetListEntryFromMediaId(146065) // mushoku tensei season 2
	//entry, found := collec.GetListEntryFromMediaId(140439) // mob psycho 3
	//entry, found := collec.GetListEntryFromMediaId(119661) // rezero season 2 part 2
	//entry, found := collec.GetListEntryFromMediaId(131681) // attack on titan season 4 part 2
	//entry, found := collec.GetListEntryFromMediaId(154116) // undead unluck
	assert.True(t, found)
	assert.NotNil(t, entry.Media)

	queries, ok := BuildSearchQuery(&BuildSearchQueryOptions{
		Media:          entry.Media,
		Batch:          lo.ToPtr(false),
		EpisodeNumber:  lo.ToPtr(16),
		AbsoluteOffset: lo.ToPtr(24),
		Resolution:     lo.ToPtr(""),
		//Title:          lo.ToPtr("Re zero"),
	})
	assert.True(t, ok)

	res, err := SearchMultiple(SearchMultipleOptions{
		Provider: "nyaa",
		Query:    queries,
		Category: "anime-eng",
		SortBy:   "seeders",
		Filter:   "",
	})
	assert.NoError(t, err, "error searching nyaa")

	t.Log("=====================================")
	for _, torrent := range res {
		t.Log(spew.Sdump(torrent.Name))
	}

}
