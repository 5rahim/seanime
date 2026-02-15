package anime

import (
	"context"
	"seanime/internal/api/anilist"
	"seanime/internal/util/comparison"
	"seanime/internal/util/limiter"
	"seanime/internal/util/result"

	"github.com/samber/lo"
)

type NormalizedMedia struct {
	ID          int
	IdMal       *int
	Title       *NormalizedMediaTitle
	Synonyms    []*string
	Format      *anilist.MediaFormat
	Status      *anilist.MediaStatus
	Season      *anilist.MediaSeason
	Year        *int
	StartDate   *NormalizedMediaDate
	Episodes    *int
	BannerImage *string
	CoverImage  *NormalizedMediaCoverImage
	//Relations         *anilist.CompleteAnimeById_Media_CompleteAnime_Relations
	NextAiringEpisode *NormalizedMediaNextAiringEpisode
	// Whether it was fetched from AniList
	fetched bool
}

type NormalizedMediaTitle struct {
	Romaji        *string
	English       *string
	Native        *string
	UserPreferred *string
}

type NormalizedMediaDate struct {
	Year  *int
	Month *int
	Day   *int
}

type NormalizedMediaCoverImage struct {
	ExtraLarge *string
	Large      *string
	Medium     *string
	Color      *string
}

type NormalizedMediaNextAiringEpisode struct {
	AiringAt        int
	TimeUntilAiring int
	Episode         int
}

type NormalizedMediaCache struct {
	*result.Cache[int, *NormalizedMedia]
}

func NewNormalizedMedia(m *anilist.BaseAnime) *NormalizedMedia {
	var startDate *NormalizedMediaDate
	if m.GetStartDate() != nil {
		startDate = &NormalizedMediaDate{
			Year:  m.GetStartDate().GetYear(),
			Month: m.GetStartDate().GetMonth(),
			Day:   m.GetStartDate().GetDay(),
		}
	}

	var title *NormalizedMediaTitle
	if m.GetTitle() != nil {
		title = &NormalizedMediaTitle{
			Romaji:        m.GetTitle().GetRomaji(),
			English:       m.GetTitle().GetEnglish(),
			Native:        m.GetTitle().GetNative(),
			UserPreferred: m.GetTitle().GetUserPreferred(),
		}
	}

	var coverImage *NormalizedMediaCoverImage
	if m.GetCoverImage() != nil {
		coverImage = &NormalizedMediaCoverImage{
			ExtraLarge: m.GetCoverImage().GetExtraLarge(),
			Large:      m.GetCoverImage().GetLarge(),
			Medium:     m.GetCoverImage().GetMedium(),
			Color:      m.GetCoverImage().GetColor(),
		}
	}

	var nextAiringEpisode *NormalizedMediaNextAiringEpisode
	if m.GetNextAiringEpisode() != nil {
		nextAiringEpisode = &NormalizedMediaNextAiringEpisode{
			AiringAt:        m.GetNextAiringEpisode().GetAiringAt(),
			TimeUntilAiring: m.GetNextAiringEpisode().GetTimeUntilAiring(),
			Episode:         m.GetNextAiringEpisode().GetEpisode(),
		}
	}

	return &NormalizedMedia{
		ID:                m.GetID(),
		IdMal:             m.GetIDMal(),
		Title:             title,
		Synonyms:          m.GetSynonyms(),
		Format:            m.GetFormat(),
		Status:            m.GetStatus(),
		Season:            m.GetSeason(),
		Year:              m.GetSeasonYear(),
		StartDate:         startDate,
		Episodes:          m.GetEpisodes(),
		BannerImage:       m.GetBannerImage(),
		CoverImage:        coverImage,
		NextAiringEpisode: nextAiringEpisode,
		fetched:           true,
	}
}

// NewNormalizedMediaFromOfflineDB creates a NormalizedMedia from the anime-offline-database.
// The media is marked as not fetched (fetched=false) since it lacks some AniList-specific data.
func NewNormalizedMediaFromOfflineDB(
	id int,
	idMal *int,
	title *NormalizedMediaTitle,
	synonyms []*string,
	format *anilist.MediaFormat,
	status *anilist.MediaStatus,
	season *anilist.MediaSeason,
	year *int,
	startDate *NormalizedMediaDate,
	episodes *int,
	coverImage *NormalizedMediaCoverImage,
) *NormalizedMedia {
	return &NormalizedMedia{
		ID:         id,
		IdMal:      idMal,
		Title:      title,
		Synonyms:   synonyms,
		Format:     format,
		Status:     status,
		Season:     season,
		Year:       year,
		StartDate:  startDate,
		Episodes:   episodes,
		CoverImage: coverImage,
		fetched:    false,
	}
}

func FetchNormalizedMedia(anilistClient anilist.AnilistClient, l *limiter.Limiter, cache *anilist.CompleteAnimeCache, m *NormalizedMedia) error {
	if anilistClient == nil || m == nil {
		return nil
	}

	if m.fetched {
		return nil
	}

	if cache != nil {
		if complete, found := cache.Get(m.ID); found {
			*m = *NewNormalizedMedia(complete.ToBaseAnime())
		}
	}

	l.Wait()
	complete, err := anilistClient.CompleteAnimeByID(context.Background(), &m.ID)
	if err != nil {
		return err
	}

	if cache != nil {
		cache.Set(m.ID, complete.GetMedia())
	}
	*m = *NewNormalizedMedia(complete.GetMedia().ToBaseAnime())
	m.fetched = true
	return nil
}

func NewNormalizedMediaCache() *NormalizedMediaCache {
	return &NormalizedMediaCache{result.NewCache[int, *NormalizedMedia]()}
}

// Helper methods

func (m *NormalizedMedia) GetTitleSafe() string {
	if m.Title == nil {
		return ""
	}
	if m.Title.UserPreferred != nil {
		return *m.Title.UserPreferred
	}
	if m.Title.English != nil {
		return *m.Title.English
	}
	if m.Title.Romaji != nil {
		return *m.Title.Romaji
	}
	if m.Title.Native != nil {
		return *m.Title.Native
	}
	return ""
}

func (m *NormalizedMedia) HasRomajiTitle() bool {
	return m.Title != nil && m.Title.Romaji != nil
}

func (m *NormalizedMedia) HasEnglishTitle() bool {
	return m.Title != nil && m.Title.English != nil
}

func (m *NormalizedMedia) HasSynonyms() bool {
	return len(m.Synonyms) > 0
}

func (m *NormalizedMedia) GetAllTitles() []*string {
	titles := make([]*string, 0)
	if m.Title == nil {
		return titles
	}
	if m.Title.Romaji != nil {
		titles = append(titles, m.Title.Romaji)
	}
	if m.Title.English != nil {
		titles = append(titles, m.Title.English)
	}
	if m.Title.Native != nil {
		titles = append(titles, m.Title.Native)
	}
	if m.Title.UserPreferred != nil {
		titles = append(titles, m.Title.UserPreferred)
	}
	titles = append(titles, m.Synonyms...)
	return titles
}

// GetPossibleSeasonNumber returns the possible season number for that media and -1 if it doesn't have one.
// It looks at the synonyms and returns the highest season number found.
func (m *NormalizedMedia) GetPossibleSeasonNumber() int {
	if m == nil || len(m.Synonyms) == 0 {
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

func (m *NormalizedMedia) FetchMediaTree(
	rel anilist.FetchMediaTreeRelation,
	anilistClient anilist.AnilistClient,
	rl *limiter.Limiter,
	tree *anilist.CompleteAnimeRelationTree,
	cache *anilist.CompleteAnimeCache,
) error {
	if m == nil {
		return nil
	}

	rl.Wait()
	res, err := anilistClient.CompleteAnimeByID(context.Background(), &m.ID)
	if err != nil {
		return err
	}
	return res.GetMedia().FetchMediaTree(rel, anilistClient, rl, tree, cache)
}

// GetCurrentEpisodeCount returns the current episode number for that media and -1 if it doesn't have one.
// i.e. -1 is returned if the media has no episodes AND the next airing episode is not set.
func (m *NormalizedMedia) GetCurrentEpisodeCount() int {
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
func (m *NormalizedMedia) GetTotalEpisodeCount() int {
	ceil := -1
	if m.Episodes != nil {
		ceil = *m.Episodes
	}
	return ceil
}
