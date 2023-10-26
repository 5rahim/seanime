package scanner

import (
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/comparison"
)

type MediaContainerOptions struct {
	allMedia []*anilist.BaseMedia
}

type MediaContainer struct {
	engTitles []*string
	romTitles []*string
	synonyms  []*string
	allMedia  []*anilist.BaseMedia
}

// NewMediaContainer creates a new MediaContainer
// It will create a list of all English titles, Romaji titles, and synonyms for all anilist.BaseMedia
func NewMediaContainer(opts *MediaContainerOptions) *MediaContainer {
	mc := new(MediaContainer)

	engTitles := lop.Map(opts.allMedia, func(m *anilist.BaseMedia, index int) *string {
		if m.Title.English != nil {
			return m.Title.English
		}
		return new(string)
	})
	romTitles := lop.Map(opts.allMedia, func(m *anilist.BaseMedia, index int) *string {
		if m.Title.Romaji != nil {
			return m.Title.Romaji
		}
		return new(string)
	})
	_synonymsArr := lop.Map(opts.allMedia, func(m *anilist.BaseMedia, index int) []*string {
		if m.Synonyms != nil {
			return m.Synonyms
		}
		return make([]*string, 0)
	})
	synonyms := lo.Flatten(_synonymsArr)
	engTitles = lo.Filter(engTitles, func(s *string, i int) bool { return s != nil && len(*s) > 0 })
	romTitles = lo.Filter(romTitles, func(s *string, i int) bool { return s != nil && len(*s) > 0 })
	synonyms = lo.Filter(synonyms, func(s *string, i int) bool { return comparison.ValueContainsSeson(*s) })

	mc.engTitles = engTitles
	mc.romTitles = romTitles
	mc.synonyms = synonyms
	mc.allMedia = opts.allMedia

	return mc
}
