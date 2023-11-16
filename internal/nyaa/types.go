package nyaa

type Torrent struct {
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

type Comment struct {
	User string `json:"user"`
	Date string `json:"date"`
	Text string `json:"text"`
}
