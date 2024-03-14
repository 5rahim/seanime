package nyaa

import (
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/comparison"
	"github.com/seanime-app/seanime/seanime-parser"
)

type (
	Torrent struct {
		Category    string `json:"category"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Date        string `json:"date"`
		Size        string `json:"size"`
		Seeders     string `json:"seeders"`
		Leechers    string `json:"leechers"`
		Downloads   string `json:"downloads"`
		IsTrusted   string `json:"isTrusted"`
		IsRemake    string `json:"isRemake"`
		Comments    string `json:"comments"`
		Link        string `json:"link"`
		GUID        string `json:"guid"`
		CategoryID  string `json:"categoryID"`
		InfoHash    string `json:"infoHash"`
	}

	Comment struct {
		User string `json:"user"`
		Date string `json:"date"`
		Text string `json:"text"`
	}

	DetailedTorrent struct {
		Torrent
		Resolution string `json:"resolution"`
		IsBatch    bool   `json:"isBatch"` // /!\ will be true if the torrent is a movie
	}
)

func (t *Torrent) toDetailedTorrent() *DetailedTorrent {
	elements := seanime_parser.Parse(t.Name)

	isBatch := false

	if len(elements.EpisodeNumber) > 1 || comparison.ValueContainsBatchKeywords(t.Name) {
		isBatch = true
	}

	return &DetailedTorrent{
		Torrent:    *t,
		Resolution: elements.VideoResolution,
		IsBatch:    isBatch,
	}
}

func (t *Torrent) GetSizeInBytes() int64 {
	bytes, _ := util.StringSizeToBytes(t.Size)
	return bytes
}
