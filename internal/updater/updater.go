package updater

type (
	Updater struct {
	}

	LatestReleaseResponse struct {
		Release Release `json:"release"`
	}
	Release struct {
		Url         string         `json:"url"`
		HtmlUrl     string         `json:"html_url"`
		NodeId      string         `json:"node_id"`
		TagName     string         `json:"tag_name"`
		Name        string         `json:"name"`
		Body        string         `json:"body"`
		PublishedAt string         `json:"published_at"`
		Released    bool           `json:"released"`
		Assets      []ReleaseAsset `json:"assets"`
	}
	ReleaseAsset struct {
		Url                string `json:"url"`
		Id                 int    `json:"id"`
		NodeId             string `json:"node_id"`
		Name               string `json:"name"`
		ContentType        string `json:"content_type"`
		Uploaded           bool   `json:"uploaded"`
		Size               int    `json:"size"`
		BrowserDownloadUrl string `json:"browser_download_url"`
	}
)
