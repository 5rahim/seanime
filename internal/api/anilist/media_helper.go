package anilist

import (
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/util/comparison"
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
func (m *BaseMedia) GetRomajiTitleSafe() string {
	if m.GetTitle().GetRomaji() != nil {
		return *m.GetTitle().GetRomaji()
	}
	if m.GetTitle().GetEnglish() != nil {
		return *m.GetTitle().GetEnglish()
	}
	return "N/A"
}

func (m *BaseMedia) GetPreferredTitle() string {
	if m.GetTitle().GetUserPreferred() != nil {
		return *m.GetTitle().GetUserPreferred()
	}
	return m.GetTitleSafe()
}

func (m *BaseMedia) GetCoverImageSafe() string {
	if m.GetCoverImage().GetExtraLarge() != nil {
		return *m.GetCoverImage().GetExtraLarge()
	}
	if m.GetCoverImage().GetLarge() != nil {
		return *m.GetCoverImage().GetLarge()
	}
	if m.GetBannerImage() != nil {
		return *m.GetBannerImage()
	}
	return ""
}

func (m *BaseMedia) GetBannerImageSafe() string {
	if m.GetBannerImage() != nil {
		return *m.GetBannerImage()
	}
	return m.GetCoverImageSafe()
}

func (m *BaseMedia) IsMovieOrSingleEpisode() bool {
	if m == nil {
		return false
	}
	if m.GetTotalEpisodeCount() == 1 {
		return true
	}
	return false
}

func (m *BaseMedia) IsMovie() bool {
	if m == nil {
		return false
	}
	if m.Format == nil {
		return false
	}

	return *m.Format == MediaFormatMovie
}

func (m *BaseMedia) IsFinished() bool {
	if m == nil {
		return false
	}
	if m.Status == nil {
		return false
	}

	return *m.Status == MediaStatusFinished
}

func (m *BaseMedia) GetAllTitles() []*string {
	titles := make([]*string, 0)
	if m.HasRomajiTitle() {
		titles = append(titles, m.Title.Romaji)
	}
	if m.HasEnglishTitle() {
		titles = append(titles, m.Title.English)
	}
	if m.HasSynonyms() && len(m.Synonyms) > 1 {
		titles = append(titles, lo.Filter(m.Synonyms, func(s *string, i int) bool { return comparison.ValueContainsSeason(*s) })...)
	}
	return titles
}

func (m *BaseMedia) GetAllTitlesDeref() []string {
	titles := make([]string, 0)
	if m.HasRomajiTitle() {
		titles = append(titles, *m.Title.Romaji)
	}
	if m.HasEnglishTitle() {
		titles = append(titles, *m.Title.English)
	}
	if m.HasSynonyms() && len(m.Synonyms) > 1 {
		syn := lo.Filter(m.Synonyms, func(s *string, i int) bool { return comparison.ValueContainsSeason(*s) })
		for _, s := range syn {
			titles = append(titles, *s)
		}
	}
	return titles
}

// GetCurrentEpisodeCount returns the current episode number for that media and -1 if it doesn't have one.
// i.e. -1 is returned if the media has no episodes AND the next airing episode is not set.
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

// GetPossibleSeasonNumber returns the possible season number for that media and -1 if it doesn't have one.
// It looks at the synonyms and returns the highest season number found.
func (m *BaseMedia) GetPossibleSeasonNumber() int {
	if m == nil || m.Synonyms == nil || len(m.Synonyms) == 0 {
		return -1
	}
	titles := lo.Filter(m.Synonyms, func(s *string, i int) bool { return comparison.ValueContainsSeason(*s) })
	if m.HasEnglishTitle() {
		titles = append(titles, m.Title.English)
	}
	if m.HasRomajiTitle() {
		titles = append(titles, m.Title.Romaji)
	}
	seasons := lo.Map(titles, func(s *string, i int) int { return comparison.ExtractSeasonNumber(*s) })
	return lo.Max(seasons)
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *CompleteMedia) GetTitleSafe() string {
	if m.GetTitle().GetEnglish() != nil {
		return *m.GetTitle().GetEnglish()
	}
	if m.GetTitle().GetRomaji() != nil {
		return *m.GetTitle().GetRomaji()
	}
	return "N/A"
}
func (m *CompleteMedia) GetRomajiTitleSafe() string {
	if m.GetTitle().GetRomaji() != nil {
		return *m.GetTitle().GetRomaji()
	}
	if m.GetTitle().GetEnglish() != nil {
		return *m.GetTitle().GetEnglish()
	}
	return "N/A"
}

func (m *CompleteMedia) GetPreferredTitle() string {
	if m.GetTitle().GetUserPreferred() != nil {
		return *m.GetTitle().GetUserPreferred()
	}
	return m.GetTitleSafe()
}

func (m *CompleteMedia) GetCoverImageSafe() string {
	if m.GetCoverImage().GetExtraLarge() != nil {
		return *m.GetCoverImage().GetExtraLarge()
	}
	if m.GetCoverImage().GetLarge() != nil {
		return *m.GetCoverImage().GetLarge()
	}
	if m.GetBannerImage() != nil {
		return *m.GetBannerImage()
	}
	return ""
}

func (m *CompleteMedia) GetBannerImageSafe() string {
	if m.GetBannerImage() != nil {
		return *m.GetBannerImage()
	}
	return m.GetCoverImageSafe()
}

func (m *CompleteMedia) IsMovieOrSingleEpisode() bool {
	if m == nil {
		return false
	}
	if m.GetTotalEpisodeCount() == 1 {
		return true
	}
	return false
}

func (m *CompleteMedia) IsMovie() bool {
	if m == nil {
		return false
	}
	if m.Format == nil {
		return false
	}

	return *m.Format == MediaFormatMovie
}

func (m *CompleteMedia) IsFinished() bool {
	if m == nil {
		return false
	}
	if m.Status == nil {
		return false
	}

	return *m.Status == MediaStatusFinished
}

func (m *CompleteMedia) GetAllTitles() []*string {
	titles := make([]*string, 0)
	if m.HasRomajiTitle() {
		titles = append(titles, m.Title.Romaji)
	}
	if m.HasEnglishTitle() {
		titles = append(titles, m.Title.English)
	}
	if m.HasSynonyms() && len(m.Synonyms) > 1 {
		titles = append(titles, lo.Filter(m.Synonyms, func(s *string, i int) bool { return comparison.ValueContainsSeason(*s) })...)
	}
	return titles
}

func (m *CompleteMedia) GetAllTitlesDeref() []string {
	titles := make([]string, 0)
	if m.HasRomajiTitle() {
		titles = append(titles, *m.Title.Romaji)
	}
	if m.HasEnglishTitle() {
		titles = append(titles, *m.Title.English)
	}
	if m.HasSynonyms() && len(m.Synonyms) > 1 {
		syn := lo.Filter(m.Synonyms, func(s *string, i int) bool { return comparison.ValueContainsSeason(*s) })
		for _, s := range syn {
			titles = append(titles, *s)
		}
	}
	return titles
}

// GetCurrentEpisodeCount returns the current episode number for that media and -1 if it doesn't have one.
// i.e. -1 is returned if the media has no episodes AND the next airing episode is not set.
func (m *CompleteMedia) GetCurrentEpisodeCount() int {
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
func (m *CompleteMedia) GetTotalEpisodeCount() int {
	ceil := -1
	if m.Episodes != nil {
		ceil = *m.Episodes
	}
	return ceil
}

// GetPossibleSeasonNumber returns the possible season number for that media and -1 if it doesn't have one.
// It looks at the synonyms and returns the highest season number found.
func (m *CompleteMedia) GetPossibleSeasonNumber() int {
	if m == nil || m.Synonyms == nil || len(m.Synonyms) == 0 {
		return -1
	}
	titles := lo.Filter(m.Synonyms, func(s *string, i int) bool { return comparison.ValueContainsSeason(*s) })
	if m.HasEnglishTitle() {
		titles = append(titles, m.Title.English)
	}
	if m.HasRomajiTitle() {
		titles = append(titles, m.Title.Romaji)
	}
	seasons := lo.Map(titles, func(s *string, i int) int { return comparison.ExtractSeasonNumber(*s) })
	return lo.Max(seasons)
}

func (m *CompleteMedia) HasEnglishTitle() bool {
	return m.Title.English != nil
}

func (m *CompleteMedia) HasRomajiTitle() bool {
	return m.Title.Romaji != nil
}

func (m *CompleteMedia) HasSynonyms() bool {
	return m.Synonyms != nil
}

//----------------------------------------------------------------------------------------------------------------------

var EdgeNarrowFormats = []MediaFormat{MediaFormatTv, MediaFormatTvShort}
var EdgeBroaderFormats = []MediaFormat{MediaFormatTv, MediaFormatTvShort, MediaFormatOna, MediaFormatOva, MediaFormatMovie, MediaFormatSpecial}

func (m *CompleteMedia) FindEdge(relation string, formats []MediaFormat) (*BaseMedia, bool) {
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

func (e *CompleteMedia_Relations_Edges) IsBroadRelationFormat() bool {
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
func (e *CompleteMedia_Relations_Edges) IsNarrowRelationFormat() bool {
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

func (m *CompleteMedia) ToBaseMedia() *BaseMedia {
	if m == nil {
		return nil
	}

	var trailer *BaseMedia_Trailer
	if m.GetTrailer() != nil {
		trailer = &BaseMedia_Trailer{
			ID:        m.GetTrailer().GetID(),
			Site:      m.GetTrailer().GetSite(),
			Thumbnail: m.GetTrailer().GetThumbnail(),
		}
	}

	var nextAiringEpisode *BaseMedia_NextAiringEpisode
	if m.GetNextAiringEpisode() != nil {
		nextAiringEpisode = &BaseMedia_NextAiringEpisode{
			AiringAt:        m.GetNextAiringEpisode().GetAiringAt(),
			TimeUntilAiring: m.GetNextAiringEpisode().GetTimeUntilAiring(),
			Episode:         m.GetNextAiringEpisode().GetEpisode(),
		}
	}

	var startDate *BaseMedia_StartDate
	if m.GetStartDate() != nil {
		startDate = &BaseMedia_StartDate{
			Year:  m.GetStartDate().GetYear(),
			Month: m.GetStartDate().GetMonth(),
			Day:   m.GetStartDate().GetDay(),
		}
	}

	var endDate *BaseMedia_EndDate
	if m.GetEndDate() != nil {
		endDate = &BaseMedia_EndDate{
			Year:  m.GetEndDate().GetYear(),
			Month: m.GetEndDate().GetMonth(),
			Day:   m.GetEndDate().GetDay(),
		}
	}

	return &BaseMedia{
		ID:              m.GetID(),
		IDMal:           m.GetIDMal(),
		SiteURL:         m.GetSiteURL(),
		Format:          m.GetFormat(),
		Episodes:        m.GetEpisodes(),
		Status:          m.GetStatus(),
		Synonyms:        m.GetSynonyms(),
		BannerImage:     m.GetBannerImage(),
		Season:          m.GetSeason(),
		Type:            m.GetType(),
		IsAdult:         m.GetIsAdult(),
		CountryOfOrigin: m.GetCountryOfOrigin(),
		Genres:          m.GetGenres(),
		Duration:        m.GetDuration(),
		Description:     m.GetDescription(),
		MeanScore:       m.GetMeanScore(),
		Trailer:         trailer,
		Title: &BaseMedia_Title{
			UserPreferred: m.GetTitle().GetUserPreferred(),
			Romaji:        m.GetTitle().GetRomaji(),
			English:       m.GetTitle().GetEnglish(),
			Native:        m.GetTitle().GetNative(),
		},
		CoverImage: &BaseMedia_CoverImage{
			ExtraLarge: m.GetCoverImage().GetExtraLarge(),
			Large:      m.GetCoverImage().GetLarge(),
			Medium:     m.GetCoverImage().GetMedium(),
			Color:      m.GetCoverImage().GetColor(),
		},
		StartDate:         startDate,
		EndDate:           endDate,
		NextAiringEpisode: nextAiringEpisode,
	}
}
