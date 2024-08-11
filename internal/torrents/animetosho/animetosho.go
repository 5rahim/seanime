package animetosho

type (
	Torrent struct {
		Id                   int         `json:"id"`
		Title                string      `json:"title"`
		Link                 string      `json:"link"`
		Timestamp            int         `json:"timestamp"`
		Status               string      `json:"status"`
		ToshoId              int         `json:"tosho_id,omitempty"`
		NyaaId               int         `json:"nyaa_id,omitempty"`
		NyaaSubdom           interface{} `json:"nyaa_subdom,omitempty"`
		AniDexId             int         `json:"anidex_id,omitempty"`
		TorrentUrl           string      `json:"torrent_url"`
		InfoHash             string      `json:"info_hash"`
		InfoHashV2           string      `json:"info_hash_v2,omitempty"`
		MagnetUri            string      `json:"magnet_uri"`
		Seeders              int         `json:"seeders"`
		Leechers             int         `json:"leechers"`
		TorrentDownloadCount int         `json:"torrent_download_count"`
		TrackerUpdated       interface{} `json:"tracker_updated,omitempty"`
		NzbUrl               string      `json:"nzb_url,omitempty"`
		TotalSize            int64       `json:"total_size"`
		NumFiles             int         `json:"num_files"`
		AniDbAid             int         `json:"anidb_aid"`
		AniDbEid             int         `json:"anidb_eid"`
		AniDbFid             int         `json:"anidb_fid"`
		ArticleUrl           string      `json:"article_url"`
		ArticleTitle         string      `json:"article_title"`
		WebsiteUrl           string      `json:"website_url"`
	}
)
