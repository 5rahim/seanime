package torrent

import "seanime/internal/torrents/nyaa"

func NewNsfwSearch(query string) (ret *SearchData, err error) {
	ret = &SearchData{
		Torrents: make([]*AnimeTorrent, 0),
	}

	if query == "" {
		return ret, nil
	}

	// +---------------------+
	// |       Query         |
	// +---------------------+

	res, err := nyaa.Search(nyaa.BuildURLOptions{
		Provider: "sukebei",
		Query:    query,
		Category: "art-anime",
		SortBy:   "seeders",
		Filter:   "",
	})
	if err != nil {
		return nil, err
	}

	for _, torrent := range res {
		ret.Torrents = append(ret.Torrents, NewAnimeTorrentFromNyaa(torrent))
	}

	return
}
