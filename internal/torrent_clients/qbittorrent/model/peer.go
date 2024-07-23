package qbittorrent_model

type Peer struct {
	Client           string  `json:"client"`
	Connection       string  `json:"connection"`
	Country          string  `json:"country"`
	CountryCode      string  `json:"country_code"`
	DLSpeed          int     `json:"dlSpeed"`
	Downloaded       int     `json:"downloaded"`
	Files            string  `json:"files"`
	Flags            string  `json:"flags"`
	FlagsDescription string  `json:"flags_desc"`
	IP               string  `json:"ip"`
	Port             int     `json:"port"`
	Progress         float64 `json:"progress"`
	Relevance        int     `json:"relevance"`
	ULSpeed          int     `json:"up_speed"`
	Uploaded         int     `json:"uploaded"`
}
