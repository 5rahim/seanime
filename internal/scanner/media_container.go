package scanner

import (
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/comparison"
	"strings"
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

// NewMediaContainer will create a list of all English titles, Romaji titles, and synonyms from all anilist.BaseMedia.
// It also provides helper functions to get an anilist.BaseMedia from a title or synonym.
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
	synonyms = lo.Filter(synonyms, func(s *string, i int) bool { return comparison.ValueContainsSeason(*s) })

	mc.engTitles = engTitles
	mc.romTitles = romTitles
	mc.synonyms = synonyms
	mc.allMedia = opts.allMedia

	return mc
}

func (mc *MediaContainer) GetMediaFromTitleOrSynonym(title *string) (*anilist.BaseMedia, bool) {
	if title == nil {
		return nil, false
	}
	t := strings.ToLower(*title)
	res, found := lo.Find(mc.allMedia, func(m *anilist.BaseMedia) bool {
		if m.HasEnglishTitle() && t == strings.ToLower(*m.Title.English) {
			return true
		}
		if m.HasRomajiTitle() && t == strings.ToLower(*m.Title.Romaji) {
			return true
		}
		if m.HasSynonyms() {
			for _, syn := range m.Synonyms {
				if t == strings.ToLower(*syn) {
					return true
				}
			}
		}
		return false
	})

	return res, found
}

func (mc *MediaContainer) GetMediaFromId(id int) (*anilist.BaseMedia, bool) {
	res, found := lo.Find(mc.allMedia, func(m *anilist.BaseMedia) bool {
		if m.ID == id {
			return true
		}
		return false
	})
	return res, found
}
