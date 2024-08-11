package manga

import (
	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
)

// GetChapter returns a chapter from the container
func (cc *ChapterContainer) GetChapter(id string) (ret *hibikemanga.ChapterDetails, found bool) {
	for _, c := range cc.Chapters {
		if c.ID == id {
			return c, true
		}
	}
	return nil, false
}
