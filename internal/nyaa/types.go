package nyaa

type Torrent struct {
	Category    string
	Name        string
	Description string
	Date        string
	Size        string
	Seeders     string
	Leechers    string
	Downloads   string
	IsTrusted   string
	IsRemake    string
	Comments    string
	Link        string
	GUID        string
	CategoryID  string
	InfoHash    string
}

type Comment struct {
	User string
	Date string
	Text string
}
