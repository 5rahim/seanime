package scanner

import (
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/comparison"
	"github.com/seanime-app/seanime/internal/result"
	"strings"
)

type (
	MediaContainerOptions struct {
		allMedia []*anilist.BaseMedia
	}

	MediaContainer struct {
		NormalizedMedia []*NormalizedMedia
		engTitles       []*string
		romTitles       []*string
		synonyms        []*string
		allMedia        []*anilist.BaseMedia
	}

	NormalizedMedia struct {
		*anilist.BasicMedia
	}

	NormalizedMediaCache struct {
		*result.Cache[int, *NormalizedMedia]
	}
)

func NewNormalizedMedia(m *anilist.BasicMedia) *NormalizedMedia {
	return &NormalizedMedia{
		BasicMedia: m,
	}
}

func NewNormalizedMediaCache() *NormalizedMediaCache {
	return &NormalizedMediaCache{result.NewCache[int, *NormalizedMedia]()}
}

// NewMediaContainer will create a list of all English titles, Romaji titles, and synonyms from all anilist.BaseMedia (used by Matcher).
//
// The list will include all anilist.BaseMedia and their relations (prequels, sequels, spin-offs, etc...) as NormalizedMedia.
//
// It also provides helper functions to get a NormalizedMedia from a title or synonym (used by FileHydrator).
func NewMediaContainer(opts *MediaContainerOptions) *MediaContainer {
	mc := new(MediaContainer)

	mc.NormalizedMedia = make([]*NormalizedMedia, 0)
	for _, m := range opts.allMedia {
		mc.NormalizedMedia = append(mc.NormalizedMedia, NewNormalizedMedia(m.ToBasicMedia()))
		if m.Relations != nil && m.Relations.Edges != nil && len(m.Relations.Edges) > 0 {
			for _, edgeM := range m.Relations.Edges {
				if edgeM.Node == nil || edgeM.Node.Format == nil || edgeM.RelationType == nil {
					continue
				}
				if *edgeM.Node.Format != anilist.MediaFormatMovie &&
					*edgeM.Node.Format != anilist.MediaFormatOva &&
					*edgeM.Node.Format != anilist.MediaFormatSpecial &&
					*edgeM.Node.Format != anilist.MediaFormatTv {
					continue
				}
				if *edgeM.RelationType != anilist.MediaRelationPrequel &&
					*edgeM.RelationType != anilist.MediaRelationSequel &&
					*edgeM.RelationType != anilist.MediaRelationSpinOff &&
					*edgeM.RelationType != anilist.MediaRelationAlternative &&
					*edgeM.RelationType != anilist.MediaRelationParent {
					continue
				}
				mc.NormalizedMedia = append(mc.NormalizedMedia, NewNormalizedMedia(edgeM.GetNode()))
			}
		}
	}
	mc.NormalizedMedia = lo.UniqBy(mc.NormalizedMedia, func(m *NormalizedMedia) int {
		return m.ID
	})

	engTitles := lop.Map(mc.NormalizedMedia, func(m *NormalizedMedia, index int) *string {
		if m.Title.English != nil {
			return m.Title.English
		}
		return new(string)
	})
	romTitles := lop.Map(mc.NormalizedMedia, func(m *NormalizedMedia, index int) *string {
		if m.Title.Romaji != nil {
			return m.Title.Romaji
		}
		return new(string)
	})
	_synonymsArr := lop.Map(mc.NormalizedMedia, func(m *NormalizedMedia, index int) []*string {
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

func (mc *MediaContainer) GetMediaFromTitleOrSynonym(title *string) (*NormalizedMedia, bool) {
	if title == nil {
		return nil, false
	}
	t := strings.ToLower(*title)
	res, found := lo.Find(mc.NormalizedMedia, func(m *NormalizedMedia) bool {
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

func (mc *MediaContainer) GetMediaFromId(id int) (*NormalizedMedia, bool) {
	res, found := lo.Find(mc.NormalizedMedia, func(m *NormalizedMedia) bool {
		if m.ID == id {
			return true
		}
		return false
	})
	return res, found
}
