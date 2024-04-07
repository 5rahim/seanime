package torrent

import "github.com/seanime-app/seanime/internal/torrents/nyaa"

func NewNsfwSearch(query string, cache *nyaa.SearchCache) (ret *SearchData, err error) {
	ret = &SearchData{
		Torrents: make([]*AnimeTorrent, 0),
	}

	if query == "" {
		return ret, nil
	}

	// +---------------------+
	// |       Query         |
	// +---------------------+

	res, err := nyaa.Search(nyaa.SearchOptions{
		Provider: "sukebei",
		Query:    query,
		Category: "art-anime",
		SortBy:   "seeders",
		Filter:   "",
		Cache:    cache,
	})
	if err != nil {
		return nil, err
	}

	for _, torrent := range res {
		ret.Torrents = append(ret.Torrents, NewAnimeTorrentFromNyaa(torrent))
	}

	return
}
