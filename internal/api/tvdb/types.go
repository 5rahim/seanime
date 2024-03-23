package tvdb

var (
	ApiUrl  = "https://api4.thetvdb.com/v4"
	URL     = "https://thetvdb.com"
	ApiKeys = []string{"f5744a13-9203-4d02-b951-fbd7352c1657", "8f406bec-6ddb-45e7-8f4b-e1861e10f1bb", "5476e702-85aa-45fd-a8da-e74df3840baf", "51020266-18f7-4382-81fc-75a4014fa59f"}
)

type Episode struct {
	ID      int64  `json:"id"`
	Image   string `json:"image"`
	Number  int    `json:"number"`
	AiredAt string `json:"airedAt"`
	//Title       string `json:"title"` // Not used - since we need to fetch the translations
	//Description string `json:"description"` // Not used - since we need to fetch the translations
}

type Chapter struct {
}
