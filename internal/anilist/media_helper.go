package anilist

import (
	"github.com/samber/lo"
	"github.com/seanime-app/seanime-server/internal/comparison"
)

func (m *BaseMedia) GetTitleSafe() string {
	if m.GetTitle().GetEnglish() != nil {
		return *m.GetTitle().GetEnglish()
	}
	if m.GetTitle().GetRomaji() != nil {
		return *m.GetTitle().GetRomaji()
	}
	return "N/A"
}

func (m *BasicMedia) GetTitleSafe() string {
	if m.GetTitle().GetEnglish() != nil {
		return *m.GetTitle().GetEnglish()
	}
	if m.GetTitle().GetRomaji() != nil {
		return *m.GetTitle().GetRomaji()
	}
	return "N/A"
}

func (m *BaseMedia) GetPreferredTitle() string {
	if m.Title.UserPreferred != nil {
		return *m.GetTitle().GetUserPreferred()
	}
	return m.GetTitleSafe()
}

func (m *BaseMedia) GetAllTitles() []*string {
	titles := make([]*string, 0)
	if m.HasEnglishTitle() {
		titles = append(titles, m.Title.English)
	}
	if m.HasRomajiTitle() {
		titles = append(titles, m.Title.Romaji)
	}
	if m.HasSynonyms() && len(m.Synonyms) > 1 {
		titles = append(titles, lo.Filter(m.Synonyms, func(s *string, i int) bool { return comparison.ValueContainsSeason(*s) })...)
	}
	return titles
}

func (m *BaseMedia) GetCurrentEpisodeCount() int {
	ceil := -1
	if m.Episodes != nil {
		ceil = *m.Episodes
	}
	if m.NextAiringEpisode != nil {
		if m.NextAiringEpisode.Episode > 0 {
			ceil = m.NextAiringEpisode.Episode - 1
		}
	}
	return ceil
}

// GetTotalEpisodeCount returns the total episode number for that media and -1 if it doesn't have one
func (m *BaseMedia) GetTotalEpisodeCount() int {
	ceil := -1
	if m.Episodes != nil {
		ceil = *m.Episodes
	}
	return ceil
}

func (m *BaseMedia) HasEnglishTitle() bool {
	return m.Title.English != nil
}
func (m *BaseMedia) HasRomajiTitle() bool {
	return m.Title.Romaji != nil
}
func (m *BaseMedia) HasSynonyms() bool {
	return m.Synonyms != nil
}
func (m *BasicMedia) HasEnglishTitle() bool {
	return m.Title.English != nil
}
func (m *BasicMedia) HasRomajiTitle() bool {
	return m.Title.Romaji != nil
}
func (m *BasicMedia) HasSynonyms() bool {
	return m.Synonyms != nil
}

//----------------------------------------------------------------------------------------------------------------------

var EdgeNarrowFormats = []MediaFormat{MediaFormatTv, MediaFormatTvShort}
var EdgeBroaderFormats = []MediaFormat{MediaFormatTv, MediaFormatTvShort, MediaFormatOna, MediaFormatOva, MediaFormatMovie, MediaFormatSpecial}

func (m *BaseMedia) FindEdge(relation string, formats []MediaFormat) (*BasicMedia, bool) {
	if m.GetRelations() == nil {
		return nil, false
	}

	edges := m.GetRelations().GetEdges()

	for _, edge := range edges {

		if edge.GetRelationType().String() == relation {
			for _, fm := range formats {
				if fm.String() == edge.GetNode().GetFormat().String() {
					return edge.GetNode(), true
				}
			}
		}

	}
	return nil, false
}

func (e *BaseMedia_Relations_Edges) IsBroadRelationFormat() bool {
	if e.GetNode() == nil {
		return false
	}
	if e.GetNode().GetFormat() == nil {
		return false
	}
	for _, fm := range EdgeBroaderFormats {
		if fm.String() == e.GetNode().GetFormat().String() {
			return true
		}
	}
	return false
}
func (e *BaseMedia_Relations_Edges) IsNarrowRelationFormat() bool {
	if e.GetNode() == nil {
		return false
	}
	if e.GetNode().GetFormat() == nil {
		return false
	}
	for _, fm := range EdgeNarrowFormats {
		if fm.String() == e.GetNode().GetFormat().String() {
			return true
		}
	}
	return false
}

//----------------------------------------------------------------------------------------------------------------------

func (m *BaseMedia) ToBasicMedia() *BasicMedia {
	if m == nil {
		return nil
	}
	return &BasicMedia{
		ID:              m.ID,
		IDMal:           m.IDMal,
		Format:          m.Format,
		Episodes:        m.Episodes,
		Status:          m.Status,
		Synonyms:        m.Synonyms,
		BannerImage:     m.BannerImage,
		Season:          m.Season,
		Type:            m.Type,
		IsAdult:         m.IsAdult,
		CountryOfOrigin: m.CountryOfOrigin,
		Title: &BasicMedia_Title{
			UserPreferred: m.GetTitle().GetUserPreferred(),
			Romaji:        m.GetTitle().GetRomaji(),
			English:       m.GetTitle().GetEnglish(),
			Native:        m.GetTitle().GetNative(),
		},
		CoverImage: &BasicMedia_CoverImage{
			ExtraLarge: m.GetCoverImage().GetExtraLarge(),
			Large:      m.GetCoverImage().GetLarge(),
			Medium:     m.GetCoverImage().GetMedium(),
			Color:      m.GetCoverImage().GetColor(),
		},
		StartDate: &BasicMedia_StartDate{
			Year:  m.GetStartDate().GetYear(),
			Month: m.GetStartDate().GetMonth(),
			Day:   m.GetStartDate().GetDay(),
		},
		NextAiringEpisode: &BasicMedia_NextAiringEpisode{
			AiringAt:        m.GetNextAiringEpisode().GetAiringAt(),
			TimeUntilAiring: m.GetNextAiringEpisode().GetTimeUntilAiring(),
			Episode:         m.GetNextAiringEpisode().GetEpisode(),
		},
	}
}
