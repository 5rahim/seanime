package mappings

import (
	"github.com/goccy/go-json"
	"net/http"
	"seanime/internal/util"
)

type (

	// ReducedAnimeListResponse is the response from the reduced anime list API.
	// It does not contain AniList ids for example.
	ReducedAnimeListResponse struct {
		items          []*ReducedAnimeListItem
		itemsByAnidbID map[int]*ReducedAnimeListItem
		Count          int
	}
	ReducedAnimeListItem struct {
		TheTvdbID interface{} `json:"thetvdb_id,omitempty"`
		AnidbID   int         `json:"anidb_id,omitempty"`
	}
)

func GetReducedAnimeLists() (resp *ReducedAnimeListResponse, err error) {
	client := http.Client{}

	req, err := http.NewRequest("GET", "https://raw.githubusercontent.com/Fribb/anime-lists/master/anime-lists-reduced.json", nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var items []*ReducedAnimeListItem
	if err := json.NewDecoder(res.Body).Decode(&items); err != nil {
		return nil, err
	}

	itemsByAnidbID := make(map[int]*ReducedAnimeListItem)
	for _, item := range items {
		if item.AnidbID == 0 {
			continue
		}
		itemsByAnidbID[item.AnidbID] = item
	}

	return &ReducedAnimeListResponse{
		items:          items,
		itemsByAnidbID: itemsByAnidbID,
		Count:          len(items),
	}, nil
}

func (i *ReducedAnimeListResponse) GetItems() []*ReducedAnimeListItem {
	return i.items
}

// FindTvdbIDFromAnidbID will return the TVDB ID for the given AniDB ID.
// If the AniDB ID is not found, the second return value will be false, and the first return value will be 0.
func (i *ReducedAnimeListResponse) FindTvdbIDFromAnidbID(anidbID int) (tvdbID int, ok bool) {
	defer util.HandlePanicInModuleThen("api/mappings/FindTvdbIDFromAnidbID", func() {
		ok = false
	})

	if i == nil {
		return 0, false
	}

	item, ok := i.itemsByAnidbID[anidbID]
	if !ok {
		return 0, false
	}

	return item.GetTvdbID()
}

func (i *ReducedAnimeListItem) GetTvdbID() (int, bool) {
	if i == nil {
		return 0, false
	}

	switch v := i.TheTvdbID.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}
