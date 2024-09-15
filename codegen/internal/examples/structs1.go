package codegen

import (
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"seanime/internal/api/anilist"
)

//type Struct1 struct {
//	Struct2
//}
//
//type Struct2 struct {
//	Text string `json:"text"`
//}

//type Struct3 []string

type Struct4 struct {
	Torrents    []hibiketorrent.AnimeTorrent `json:"torrents"`
	Destination string                       `json:"destination"`
	SmartSelect struct {
		Enabled               bool  `json:"enabled"`
		MissingEpisodeNumbers []int `json:"missingEpisodeNumbers"`
	} `json:"smartSelect"`
	Media *anilist.BaseAnime `json:"media"`
}
