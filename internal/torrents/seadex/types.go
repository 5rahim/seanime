package seadex

type (
	RecordsResponse struct {
		Items []*RecordItem `json:"items"`
	}

	RecordItem struct {
		AlID           int    `json:"alID"`
		CollectionID   string `json:"collectionId"`
		CollectionName string `json:"collectionName"`
		Comparison     string `json:"comparison"`
		Created        string `json:"created"`
		Expand         struct {
			Trs []*Tr `json:"trs"`
		} `json:"expand"`
		Trs             []string `json:"trs"`
		Updated         string   `json:"updated"`
		ID              string   `json:"id"`
		Incomplete      bool     `json:"incomplete"`
		Notes           string   `json:"notes"`
		TheoreticalBest string   `json:"theoreticalBest"`
	}

	Tr struct {
		Created        string    `json:"created"`
		CollectionID   string    `json:"collectionId"`
		CollectionName string    `json:"collectionName"`
		DualAudio      bool      `json:"dualAudio"`
		Files          []*TrFile `json:"files"`
		ID             string    `json:"id"`
		InfoHash       string    `json:"infoHash"`
		IsBest         bool      `json:"isBest"`
		ReleaseGroup   string    `json:"releaseGroup"`
		Tracker        string    `json:"tracker"`
		URL            string    `json:"url"`
	}
	TrFile struct {
		Length int    `json:"length"`
		Name   string `json:"name"`
	}
)
