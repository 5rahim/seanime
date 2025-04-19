package anilist

import (
	"github.com/samber/lo"
	"seanime/internal/util/comparison"
)

func (m *BaseAnime) GetTitleSafe() string {
	if m.GetTitle().GetEnglish() != nil {
		return *m.GetTitle().GetEnglish()
	}
	if m.GetTitle().GetRomaji() != nil {
		return *m.GetTitle().GetRomaji()
	}
	return ""
}

func (m *BaseAnime) GetEnglishTitleSafe() string {
	if m.GetTitle().GetEnglish() != nil {
		return *m.GetTitle().GetEnglish()
	}
	return ""
}

func (m *BaseAnime) GetRomajiTitleSafe() string {
	if m.GetTitle().GetRomaji() != nil {
		return *m.GetTitle().GetRomaji()
	}
	if m.GetTitle().GetEnglish() != nil {
		return *m.GetTitle().GetEnglish()
	}
	return ""
}

func (m *BaseAnime) GetPreferredTitle() string {
	if m.GetTitle().GetUserPreferred() != nil {
		return *m.GetTitle().GetUserPreferred()
	}
	return m.GetTitleSafe()
}

func (m *BaseAnime) GetCoverImageSafe() string {
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

func (m *BaseAnime) GetBannerImageSafe() string {
	if m.GetBannerImage() != nil {
		return *m.GetBannerImage()
	}
	return m.GetCoverImageSafe()
}

func (m *BaseAnime) IsMovieOrSingleEpisode() bool {
	if m == nil {
		return false
	}
	if m.GetTotalEpisodeCount() == 1 {
		return true
	}
	return false
}

func (m *BaseAnime) GetSynonymsDeref() []string {
	if m.Synonyms == nil {
		return nil
	}
	return lo.Map(m.Synonyms, func(s *string, i int) string { return *s })
}

func (m *BaseAnime) GetSynonymsContainingSeason() []string {
	if m.Synonyms == nil {
		return nil
	}
	return lo.Filter(lo.Map(m.Synonyms, func(s *string, i int) string { return *s }), func(s string, i int) bool { return comparison.ValueContainsSeason(s) })
}

func (m *BaseAnime) GetStartYearSafe() int {
	if m == nil || m.StartDate == nil || m.StartDate.Year == nil {
		return 0
	}
	return *m.StartDate.Year
}

func (m *BaseAnime) IsMovie() bool {
	if m == nil {
		return false
	}
	if m.Format == nil {
		return false
	}

	return *m.Format == MediaFormatMovie
}

func (m *BaseAnime) IsFinished() bool {
	if m == nil {
		return false
	}
	if m.Status == nil {
		return false
	}

	return *m.Status == MediaStatusFinished
}

func (m *BaseAnime) GetAllTitles() []*string {
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

func (m *BaseAnime) GetAllTitlesDeref() []string {
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

func (m *BaseAnime) GetMainTitles() []*string {
	titles := make([]*string, 0)
	if m.HasRomajiTitle() {
		titles = append(titles, m.Title.Romaji)
	}
	if m.HasEnglishTitle() {
		titles = append(titles, m.Title.English)
	}
	return titles
}

func (m *BaseAnime) GetMainTitlesDeref() []string {
	titles := make([]string, 0)
	if m.HasRomajiTitle() {
		titles = append(titles, *m.Title.Romaji)
	}
	if m.HasEnglishTitle() {
		titles = append(titles, *m.Title.English)
	}
	return titles
}

// GetCurrentEpisodeCount returns the current episode number for that media and -1 if it doesn't have one.
// i.e. -1 is returned if the media has no episodes AND the next airing episode is not set.
func (m *BaseAnime) GetCurrentEpisodeCount() int {
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
func (m *BaseAnime) GetTotalEpisodeCount() int {
	ceil := -1
	if m.Episodes != nil {
		ceil = *m.Episodes
	}
	return ceil
}

// GetPossibleSeasonNumber returns the possible season number for that media and -1 if it doesn't have one.
// It looks at the synonyms and returns the highest season number found.
func (m *BaseAnime) GetPossibleSeasonNumber() int {
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

func (m *BaseAnime) HasEnglishTitle() bool {
	return m.Title.English != nil
}

func (m *BaseAnime) HasRomajiTitle() bool {
	return m.Title.Romaji != nil
}

func (m *BaseAnime) HasSynonyms() bool {
	return m.Synonyms != nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *CompleteAnime) GetTitleSafe() string {
	if m.GetTitle().GetEnglish() != nil {
		return *m.GetTitle().GetEnglish()
	}
	if m.GetTitle().GetRomaji() != nil {
		return *m.GetTitle().GetRomaji()
	}
	return "N/A"
}
func (m *CompleteAnime) GetRomajiTitleSafe() string {
	if m.GetTitle().GetRomaji() != nil {
		return *m.GetTitle().GetRomaji()
	}
	if m.GetTitle().GetEnglish() != nil {
		return *m.GetTitle().GetEnglish()
	}
	return "N/A"
}

func (m *CompleteAnime) GetPreferredTitle() string {
	if m.GetTitle().GetUserPreferred() != nil {
		return *m.GetTitle().GetUserPreferred()
	}
	return m.GetTitleSafe()
}

func (m *CompleteAnime) GetCoverImageSafe() string {
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

func (m *CompleteAnime) GetBannerImageSafe() string {
	if m.GetBannerImage() != nil {
		return *m.GetBannerImage()
	}
	return m.GetCoverImageSafe()
}

func (m *CompleteAnime) IsMovieOrSingleEpisode() bool {
	if m == nil {
		return false
	}
	if m.GetTotalEpisodeCount() == 1 {
		return true
	}
	return false
}

func (m *CompleteAnime) IsMovie() bool {
	if m == nil {
		return false
	}
	if m.Format == nil {
		return false
	}

	return *m.Format == MediaFormatMovie
}

func (m *CompleteAnime) IsFinished() bool {
	if m == nil {
		return false
	}
	if m.Status == nil {
		return false
	}

	return *m.Status == MediaStatusFinished
}

func (m *CompleteAnime) GetAllTitles() []*string {
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

func (m *CompleteAnime) GetAllTitlesDeref() []string {
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
func (m *CompleteAnime) GetCurrentEpisodeCount() int {
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
func (m *CompleteAnime) GetTotalEpisodeCount() int {
	ceil := -1
	if m.Episodes != nil {
		ceil = *m.Episodes
	}
	return ceil
}

// GetPossibleSeasonNumber returns the possible season number for that media and -1 if it doesn't have one.
// It looks at the synonyms and returns the highest season number found.
func (m *CompleteAnime) GetPossibleSeasonNumber() int {
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

func (m *CompleteAnime) HasEnglishTitle() bool {
	return m.Title.English != nil
}

func (m *CompleteAnime) HasRomajiTitle() bool {
	return m.Title.Romaji != nil
}

func (m *CompleteAnime) HasSynonyms() bool {
	return m.Synonyms != nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var EdgeNarrowFormats = []MediaFormat{MediaFormatTv, MediaFormatTvShort}
var EdgeBroaderFormats = []MediaFormat{MediaFormatTv, MediaFormatTvShort, MediaFormatOna, MediaFormatOva, MediaFormatMovie, MediaFormatSpecial}

func (m *CompleteAnime) FindEdge(relation string, formats []MediaFormat) (*BaseAnime, bool) {
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

func (e *CompleteAnime_Relations_Edges) IsBroadRelationFormat() bool {
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
func (e *CompleteAnime_Relations_Edges) IsNarrowRelationFormat() bool {
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *CompleteAnime) ToBaseAnime() *BaseAnime {
	if m == nil {
		return nil
	}

	var trailer *BaseAnime_Trailer
	if m.GetTrailer() != nil {
		trailer = &BaseAnime_Trailer{
			ID:        m.GetTrailer().GetID(),
			Site:      m.GetTrailer().GetSite(),
			Thumbnail: m.GetTrailer().GetThumbnail(),
		}
	}

	var nextAiringEpisode *BaseAnime_NextAiringEpisode
	if m.GetNextAiringEpisode() != nil {
		nextAiringEpisode = &BaseAnime_NextAiringEpisode{
			AiringAt:        m.GetNextAiringEpisode().GetAiringAt(),
			TimeUntilAiring: m.GetNextAiringEpisode().GetTimeUntilAiring(),
			Episode:         m.GetNextAiringEpisode().GetEpisode(),
		}
	}

	var startDate *BaseAnime_StartDate
	if m.GetStartDate() != nil {
		startDate = &BaseAnime_StartDate{
			Year:  m.GetStartDate().GetYear(),
			Month: m.GetStartDate().GetMonth(),
			Day:   m.GetStartDate().GetDay(),
		}
	}

	var endDate *BaseAnime_EndDate
	if m.GetEndDate() != nil {
		endDate = &BaseAnime_EndDate{
			Year:  m.GetEndDate().GetYear(),
			Month: m.GetEndDate().GetMonth(),
			Day:   m.GetEndDate().GetDay(),
		}
	}

	return &BaseAnime{
		ID:              m.GetID(),
		IDMal:           m.GetIDMal(),
		SiteURL:         m.GetSiteURL(),
		Format:          m.GetFormat(),
		Episodes:        m.GetEpisodes(),
		Status:          m.GetStatus(),
		Synonyms:        m.GetSynonyms(),
		BannerImage:     m.GetBannerImage(),
		Season:          m.GetSeason(),
		SeasonYear:      m.GetSeasonYear(),
		Type:            m.GetType(),
		IsAdult:         m.GetIsAdult(),
		CountryOfOrigin: m.GetCountryOfOrigin(),
		Genres:          m.GetGenres(),
		Duration:        m.GetDuration(),
		Description:     m.GetDescription(),
		MeanScore:       m.GetMeanScore(),
		Trailer:         trailer,
		Title: &BaseAnime_Title{
			UserPreferred: m.GetTitle().GetUserPreferred(),
			Romaji:        m.GetTitle().GetRomaji(),
			English:       m.GetTitle().GetEnglish(),
			Native:        m.GetTitle().GetNative(),
		},
		CoverImage: &BaseAnime_CoverImage{
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *AnimeListEntry) GetProgressSafe() int {
	if m == nil {
		return 0
	}
	if m.Progress == nil {
		return 0
	}
	return *m.Progress
}

func (m *AnimeListEntry) GetScoreSafe() float64 {
	if m == nil {
		return 0
	}
	if m.Score == nil {
		return 0
	}
	return *m.Score
}

func (m *AnimeListEntry) GetRepeatSafe() int {
	if m == nil {
		return 0
	}
	if m.Repeat == nil {
		return 0
	}
	return *m.Repeat
}

func (m *AnimeListEntry) GetStatusSafe() MediaListStatus {
	if m == nil {
		return ""
	}
	if m.Status == nil {
		return ""
	}
	return *m.Status
}
