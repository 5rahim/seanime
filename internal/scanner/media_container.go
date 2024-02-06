package scanner

import (
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/comparison"
	"github.com/seanime-app/seanime/internal/entities"
	"strings"
)

type (
	MediaContainerOptions struct {
		allMedia   []*anilist.BaseMedia
		ScanLogger *ScanLogger
	}

	MediaContainer struct {
		NormalizedMedia []*entities.NormalizedMedia
		ScanLogger      *ScanLogger
		engTitles       []*string
		romTitles       []*string
		synonyms        []*string
		allMedia        []*anilist.BaseMedia
	}
)

// NewMediaContainer will create a list of all English titles, Romaji titles, and synonyms from all anilist.BaseMedia (used by Matcher).
//
// The list will include all anilist.BaseMedia and their relations (prequels, sequels, spin-offs, etc...) as NormalizedMedia.
//
// It also provides helper functions to get a NormalizedMedia from a title or synonym (used by FileHydrator).
func NewMediaContainer(opts *MediaContainerOptions) *MediaContainer {
	mc := new(MediaContainer)
	mc.ScanLogger = opts.ScanLogger

	mc.NormalizedMedia = make([]*entities.NormalizedMedia, 0)

	normalizedMediaMap := make(map[int]*entities.NormalizedMedia)

	for _, m := range opts.allMedia {
		normalizedMediaMap[m.ID] = entities.NewNormalizedMedia(m.ToBasicMedia())
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
				// DEVNOTE: Edges fetched from the AniList AnimeCollection query do not contain NextAiringEpisode
				// Make sure we don't overwrite the NextAiringEpisode from the original media
				if _, found := normalizedMediaMap[edgeM.Node.ID]; !found {
					normalizedMediaMap[edgeM.Node.ID] = entities.NewNormalizedMedia(edgeM.Node)
				}
			}
		}
	}
	for _, m := range normalizedMediaMap {
		mc.NormalizedMedia = append(mc.NormalizedMedia, m)
	}

	engTitles := lop.Map(mc.NormalizedMedia, func(m *entities.NormalizedMedia, index int) *string {
		if m.Title.English != nil {
			return m.Title.English
		}
		return new(string)
	})
	romTitles := lop.Map(mc.NormalizedMedia, func(m *entities.NormalizedMedia, index int) *string {
		if m.Title.Romaji != nil {
			return m.Title.Romaji
		}
		return new(string)
	})
	_synonymsArr := lop.Map(mc.NormalizedMedia, func(m *entities.NormalizedMedia, index int) []*string {
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

	mc.ScanLogger.LogMediaContainer(zerolog.InfoLevel).
		Any("inputCount", len(opts.allMedia)).
		Any("mediaCount", len(mc.NormalizedMedia)).
		Any("titles", len(mc.engTitles)+len(mc.romTitles)+len(mc.synonyms)).
		Msg("Created media container")

	return mc
}

func (mc *MediaContainer) GetMediaFromTitleOrSynonym(title *string) (*entities.NormalizedMedia, bool) {
	if title == nil {
		return nil, false
	}
	t := strings.ToLower(*title)
	res, found := lo.Find(mc.NormalizedMedia, func(m *entities.NormalizedMedia) bool {
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

func (mc *MediaContainer) GetMediaFromId(id int) (*entities.NormalizedMedia, bool) {
	res, found := lo.Find(mc.NormalizedMedia, func(m *entities.NormalizedMedia) bool {
		if m.ID == id {
			return true
		}
		return false
	})
	return res, found
}
