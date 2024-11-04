package autodownloader

import (
	"github.com/5rahim/habari"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/database/models"
	"seanime/internal/library/anime"
	"testing"
)

func TestComparison(t *testing.T) {
	ad := AutoDownloader{
		metadataProvider: metadata.GetMockProvider(t),
		settings: &models.AutoDownloaderSettings{
			EnableSeasonCheck: true,
		},
	}
	name1 := "[Oshi no Ko] 2nd Season"
	name2 := "Oshi no Ko Season 2"
	aniListEntry := &anilist.AnimeListEntry{
		Media: &anilist.BaseAnime{
			Title: &anilist.BaseAnime_Title{
				Romaji:  &name1,
				English: &name2,
			},
			Episodes: lo.ToPtr(13),
			Format:   lo.ToPtr(anilist.MediaFormatTv),
		},
	}

	rule := &anime.AutoDownloaderRule{
		MediaId:             166531,
		ReleaseGroups:       []string{"SubsPlease", "Erai-raws"},
		Resolutions:         []string{"1080p"},
		TitleComparisonType: "likely",
		EpisodeType:         "recent",
		EpisodeNumbers:      []int{3}, // ignored
		Destination:         "/data/seanime/library/[Oshi no Ko] 2nd Season",
		ComparisonTitle:     "[Oshi no Ko] 2nd Season",
	}

	tests := []struct {
		torrentName                  string
		succeedTitleComparison       bool
		succeedSeasonAndEpisodeMatch bool
		enableSeasonCheck            bool
	}{
		{
			torrentName:                  "[Erai-raws] Oshi no Ko 2nd Season - 03 [720p][Multiple Subtitle] [ENG][FRE]",
			succeedTitleComparison:       true,
			succeedSeasonAndEpisodeMatch: true,
			enableSeasonCheck:            true,
		},
		{
			torrentName:                  "[SubsPlease] Oshi no Ko - 14 (1080p)",
			succeedTitleComparison:       true,
			succeedSeasonAndEpisodeMatch: true,
			enableSeasonCheck:            true,
		},
		{
			torrentName:                  "[Erai-raws] Oshi no Ko 3rd Season - 03 [720p][Multiple Subtitle] [ENG][FRE]",
			succeedTitleComparison:       true,
			succeedSeasonAndEpisodeMatch: false,
			enableSeasonCheck:            true,
		},
		{
			torrentName:                  "[Erai-raws] Oshi no Ko 2nd Season - 03 [720p][Multiple Subtitle] [ENG][FRE]",
			succeedTitleComparison:       true,
			succeedSeasonAndEpisodeMatch: true,
			enableSeasonCheck:            false,
		},
		{
			torrentName:                  "[SubsPlease] Oshi no Ko - 14 (1080p)",
			succeedTitleComparison:       true,
			succeedSeasonAndEpisodeMatch: true,
			enableSeasonCheck:            false,
		},
		{
			torrentName:                  "[Erai-raws] Oshi no Ko 3rd Season - 03 [720p][Multiple Subtitle] [ENG][FRE]",
			succeedTitleComparison:       true,
			succeedSeasonAndEpisodeMatch: true,
			enableSeasonCheck:            false,
		},
	}

	lfw := anime.NewLocalFileWrapper([]*anime.LocalFile{
		{
			Path: "/data/seanime/library/[Oshi no Ko] 2nd Season/[SubsPlease] Oshi no Ko - 12 (1080p).mkv",
			Name: "Oshi no Ko - 12 (1080p).mkv",
			ParsedData: &anime.LocalFileParsedData{
				Original:     "Oshi no Ko - 12 (1080p).mkv",
				Title:        "Oshi no Ko",
				ReleaseGroup: "SubsPlease",
			},
			ParsedFolderData: []*anime.LocalFileParsedData{
				{
					Original: "[Oshi no Ko] 2nd Season",
					Title:    "[Oshi no Ko]",
				},
			},
			Metadata: &anime.LocalFileMetadata{
				Episode:      1,
				AniDBEpisode: "1",
				Type:         "main",
			},
			MediaId: 166531,
		},
	})

	for _, tt := range tests {
		t.Run(tt.torrentName, func(t *testing.T) {

			ad.settings.EnableSeasonCheck = tt.enableSeasonCheck

			p := habari.Parse(tt.torrentName)
			if tt.succeedTitleComparison {
				require.True(t, ad.isTitleMatch(p, tt.torrentName, rule, aniListEntry))
			} else {
				require.False(t, ad.isTitleMatch(p, tt.torrentName, rule, aniListEntry))
			}
			lfwe, ok := lfw.GetLocalEntryById(166531)
			require.True(t, ok)
			_, ok = ad.isSeasonAndEpisodeMatch(p, rule, aniListEntry, lfwe, []*models.AutoDownloaderItem{})
			if tt.succeedSeasonAndEpisodeMatch {
				require.True(t, ok)
			} else {
				require.False(t, ok)
			}
		})
	}

}

func TestComparison2(t *testing.T) {
	ad := AutoDownloader{
		metadataProvider: metadata.GetMockProvider(t),
		settings: &models.AutoDownloaderSettings{
			EnableSeasonCheck: true,
		},
	}
	//name1 := "DANDADAN"
	//name2 := "Dandadan"
	//aniListEntry := &anilist.AnimeListEntry{
	//	Media: &anilist.BaseAnime{
	//		Title: &anilist.BaseAnime_Title{
	//			Romaji:  &name1,
	//			English: &name2,
	//		},
	//		Episodes: lo.ToPtr(12),
	//		Format:   lo.ToPtr(anilist.MediaFormatTv),
	//	},
	//}

	rule := &anime.AutoDownloaderRule{
		MediaId:             166531,
		ReleaseGroups:       []string{},
		Resolutions:         []string{"1080p"},
		TitleComparisonType: "likely",
		EpisodeType:         "recent",
		EpisodeNumbers:      []int{},
		Destination:         "/data/seanime/library/Dandadan",
		ComparisonTitle:     "Dandadan",
	}

	tests := []struct {
		torrentName                 string
		succeedAdditionalTermsMatch bool
		ruleAdditionalTerms         []string
	}{
		{
			torrentName:                 "[Anime Time] Dandadan - 04 [Dual Audio][1080p][HEVC 10bit x265][AAC][Multi Sub] [Weekly]",
			ruleAdditionalTerms:         []string{},
			succeedAdditionalTermsMatch: true,
		},
		{
			torrentName: "[Anime Time] Dandadan - 04 [Dual Audio][1080p][HEVC 10bit x265][AAC][Multi Sub] [Weekly]",
			ruleAdditionalTerms: []string{
				"H265,H.265, H 265,x265",
				"10bit,10-bit,10 bit",
			},
			succeedAdditionalTermsMatch: true,
		},
		{
			torrentName: "[Raze] Dandadan - 04 x265 10bit 1080p 143.8561fps.mkv",
			ruleAdditionalTerms: []string{
				"H265,H.265, H 265,x265",
				"10bit,10-bit,10 bit",
			},
			succeedAdditionalTermsMatch: true,
		},
		{
			torrentName: "[Sokudo] DAN DA DAN | Dandadan - S01E03 [1080p EAC-3 AV1][Dual Audio] (weekly)",
			ruleAdditionalTerms: []string{
				"H265,H.265, H 265,x265",
				"10bit,10-bit,10 bit",
			},
			succeedAdditionalTermsMatch: false,
		},
		{
			torrentName: "[Raze] Dandadan - 04 x265 10bit 1080p 143.8561fps.mkv",
			ruleAdditionalTerms: []string{
				"H265,H.265, H 265,x265",
				"10bit,10-bit,10 bit",
				"AAC",
			},
			succeedAdditionalTermsMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.torrentName, func(t *testing.T) {

			rule.AdditionalTerms = tt.ruleAdditionalTerms

			ok := ad.isAdditionalTermsMatch(tt.torrentName, rule)
			if tt.succeedAdditionalTermsMatch {
				require.True(t, ok)
			} else {
				require.False(t, ok)
			}
		})
	}

}
