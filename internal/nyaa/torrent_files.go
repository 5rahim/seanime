package nyaa

import "github.com/gocolly/colly"

func TorrentFiles(viewURL string) ([]string, error) {
	var folders []string
	var files []string

	c := colly.NewCollector()

	c.OnHTML(".folder", func(e *colly.HTMLElement) {
		folders = append(folders, e.Text)
	})

	c.OnHTML(".torrent-file-list", func(e *colly.HTMLElement) {
		files = append(files, e.ChildText("li"))
	})

	var e error
	c.OnError(func(r *colly.Response, err error) {
		e = err
	})
	if e != nil {
		return nil, e
	}

	c.Visit(viewURL)

	if len(folders) == 0 {
		return files, nil
	}

	return folders, nil
}
