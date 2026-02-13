package scanner

import (
	"context"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/database/db"
	"seanime/internal/extension"
	"seanime/internal/library/anime"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"seanime/internal/util/limiter"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestMatcher1(t *testing.T) {

	anilistClient := anilist.TestGetMockAnilistClient()
	animeCollection, err := anilistClient.AnimeCollectionWithRelations(context.Background(), nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	allMedia := animeCollection.GetAllAnime()

	dir := "E:/Anime"

	tests := []struct {
		name            string
		paths           []string
		expectedMediaId int
	}{
		{
			// These local files are from "86 - Eighty Six Part 2" but should be matched with "86 - Eighty Six Part 1"
			// because there is no indication for the part. However, the FileHydrator will fix this issue.
			name: "should match with media id 116589",
			paths: []string{
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 20v2 (1080p) [30072859].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 21v2 (1080p) [4B1616A5].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 22v2 (1080p) [58BF43B4].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 23v2 (1080p) [D94B4894].mkv",
			},
			expectedMediaId: 116589, // 86 - Eighty Six Part 1
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			scanLogger, err := NewConsoleScanLogger()
			if err != nil {
				t.Fatal("expected result, got error:", err.Error())
			}

			// +---------------------+
			// |   Local Files       |
			// +---------------------+

			var lfs []*anime.LocalFile
			for _, path := range tt.paths {
				lf := anime.NewLocalFile(path, dir)
				lfs = append(lfs, lf)
			}

			// +---------------------+
			// |   MediaContainer    |
			// +---------------------+

			mc := NewMediaContainer(&MediaContainerOptions{
				AllMedia:   NormalizedMediaFromAnilistComplete(allMedia),
				ScanLogger: scanLogger,
			})

			// +---------------------+
			// |      Matcher        |
			// +---------------------+

			matcher := &Matcher{
				LocalFiles:        lfs,
				MediaContainer:    mc,
				Logger:            util.NewLogger(),
				ScanLogger:        scanLogger,
				ScanSummaryLogger: nil,
				Debug:             true,
			}

			err = matcher.MatchLocalFilesWithMedia()

			if assert.NoError(t, err, "Error while matching local files") {
				for _, lf := range lfs {
					if lf.MediaId != tt.expectedMediaId {
						t.Fatalf("expected media id %d, got %d", tt.expectedMediaId, lf.MediaId)
					}
					t.Logf("local file: %s,\nmedia id: %d\n", lf.Name, lf.MediaId)
				}
			}
		})
	}

}

func TestMatcher2(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := anilist.NewAnilistClient(test_utils.ConfigData.Provider.AnilistJwt, "")
	animeCollection, err := anilistClient.AnimeCollectionWithRelations(context.Background(), &test_utils.ConfigData.Provider.AnilistUsername)
	if err != nil {
		t.Fatal(err.Error())
	}

	dir := "E:/Anime"

	tests := []struct {
		name            string
		paths           []string
		expectedMediaId int
		otherMediaIds   []int
	}{
		{
			name: "Kono Subarashii Sekai ni Shukufuku wo! - 21202",
			paths: []string{
				"E:/Anime/Kono Subarashii Sekai ni Shukufuku wo!/Kono Subarashii Sekai ni Shukufuku wo! (01-10) [1080p] (Batch)/[HorribleSubs] Kono Subarashii Sekai ni Shukufuku wo! - 01 [1080p].mkv",
				"E:/Anime/Kono Subarashii Sekai ni Shukufuku wo!/Kono Subarashii Sekai ni Shukufuku wo! (01-10) [1080p] (Batch)/[HorribleSubs] Kono Subarashii Sekai ni Shukufuku wo! - 02 [1080p].mkv",
				"E:/Anime/Kono Subarashii Sekai ni Shukufuku wo!/Kono Subarashii Sekai ni Shukufuku wo! (01-10) [1080p] (Batch)/[HorribleSubs] Kono Subarashii Sekai ni Shukufuku wo! - 03 [1080p].mkv",
			},
			expectedMediaId: 21202, //
		},
		{
			name: "Kono Subarashii Sekai ni Shukufuku wo! 2 - 21699",
			paths: []string{
				"E:/Anime/Kono Subarashii Sekai ni Shukufuku wo! 2/KonoSuba.God's.Blessing.On.This.Wonderful.World.S02.1080p.BluRay.10-Bit.Dual-Audio.FLAC2.0.x265-YURASUKA/KonoSuba.God's.Blessing.On.This.Wonderful.World.S02E01.1080p.BluRay.10-Bit.Dual-Audio.FLAC2.0.x265-YURASUKA.mkv",
				"E:/Anime/Kono Subarashii Sekai ni Shukufuku wo! 2/KonoSuba.God's.Blessing.On.This.Wonderful.World.S02.1080p.BluRay.10-Bit.Dual-Audio.FLAC2.0.x265-YURASUKA/KonoSuba.God's.Blessing.On.This.Wonderful.World.S02E02.1080p.BluRay.10-Bit.Dual-Audio.FLAC2.0.x265-YURASUKA.mkv",
				"E:/Anime/Kono Subarashii Sekai ni Shukufuku wo! 2/KonoSuba.God's.Blessing.On.This.Wonderful.World.S02.1080p.BluRay.10-Bit.Dual-Audio.FLAC2.0.x265-YURASUKA/KonoSuba.God's.Blessing.On.This.Wonderful.World.S02E03.1080p.BluRay.10-Bit.Dual-Audio.FLAC2.0.x265-YURASUKA.mkv",
			},
			expectedMediaId: 21699,
		},
		{
			name: "KnY 145139",
			paths: []string{
				"E:/Anime/Kimetsu no Yaiba Katanakaji no Sato-hen/Demon Slayer S03 1080p Dual Audio BDRip 10 bits DD x265-EMBER/S03E07-Awful Villain [703A5C5B].mkv",
			},
			expectedMediaId: 145139,
		},
		{
			name: "MT 108465",
			paths: []string{
				"E:/Anime/Mushoku Tensei Isekai Ittara Honki Dasu/Mushoku Tensei S01+SP 1080p Dual Audio BDRip 10 bits DDP x265-EMBER/Mushoku Tensei S01P01 1080p Dual Audio BDRip 10 bits DD x265-EMBER/S01E01-Jobless Reincarnation V2 [911C3607].mkv",
			},
			expectedMediaId: 108465,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			// Add media to collection if it doesn't exist
			allMedia := animeCollection.GetAllAnime()
			hasExpectedMediaId := false
			for _, media := range allMedia {
				if media.ID == tt.expectedMediaId {
					hasExpectedMediaId = true
					break
				}
			}
			if !hasExpectedMediaId {
				anilist.TestAddAnimeCollectionWithRelationsEntry(animeCollection, tt.expectedMediaId, anilist.TestModifyAnimeCollectionEntryInput{Status: lo.ToPtr(anilist.MediaListStatusCurrent)}, anilistClient)
				allMedia = animeCollection.GetAllAnime()
			}

			for _, otherMediaId := range tt.otherMediaIds {
				hasOtherMediaId := false
				for _, media := range allMedia {
					if media.ID == otherMediaId {
						hasOtherMediaId = true
						break
					}
				}
				if !hasOtherMediaId {
					anilist.TestAddAnimeCollectionWithRelationsEntry(animeCollection, otherMediaId, anilist.TestModifyAnimeCollectionEntryInput{Status: lo.ToPtr(anilist.MediaListStatusCurrent)}, anilistClient)
					allMedia = animeCollection.GetAllAnime()
				}
			}

			scanLogger, err := NewConsoleScanLogger()
			if err != nil {
				t.Fatal("expected result, got error:", err.Error())
			}

			// +---------------------+
			// |   Local Files       |
			// +---------------------+

			var lfs []*anime.LocalFile
			for _, path := range tt.paths {
				lf := anime.NewLocalFile(path, dir)
				lfs = append(lfs, lf)
			}

			// +---------------------+
			// |   MediaContainer    |
			// +---------------------+

			mc := NewMediaContainer(&MediaContainerOptions{
				AllMedia:   NormalizedMediaFromAnilistComplete(allMedia),
				ScanLogger: scanLogger,
			})

			// +---------------------+
			// |      Matcher        |
			// +---------------------+

			matcher := &Matcher{
				LocalFiles:        lfs,
				MediaContainer:    mc,
				Logger:            util.NewLogger(),
				ScanLogger:        scanLogger,
				ScanSummaryLogger: nil,
				Debug:             true,
			}

			err = matcher.MatchLocalFilesWithMedia()

			if assert.NoError(t, err, "Error while matching local files") {
				for _, lf := range lfs {
					if lf.MediaId != tt.expectedMediaId {
						t.Fatalf("expected media id %d, got %d", tt.expectedMediaId, lf.MediaId)
					}
					t.Logf("local file: %s,\nmedia id: %d\n", lf.Name, lf.MediaId)
				}
			}
		})
	}

}

func TestMatcher3(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := anilist.NewAnilistClient(test_utils.ConfigData.Provider.AnilistJwt, "")
	animeCollection, err := anilistClient.AnimeCollectionWithRelations(context.Background(), &test_utils.ConfigData.Provider.AnilistUsername)
	if err != nil {
		t.Fatal(err.Error())
	}

	dir := "E:/Anime"

	tests := []struct {
		name            string
		paths           []string
		expectedMediaId int
		// Optional ids of other media that should be in the collection to test conflict resolution
		otherMediaIds []int
	}{
		{
			name: "Frieren - Simple title matching - 154587",
			paths: []string{
				"E:/Anime/Frieren/Frieren - 01.mkv",
			},
			expectedMediaId: 154587,
		},
		{
			name: "Jujutsu Kaisen Season 2 - Ordinal season format - 145064",
			paths: []string{
				"E:/Anime/Jujutsu Kaisen Season 2/[SubsPlease] Jujutsu Kaisen 2nd Season - 01 (1080p) [12345678].mkv",
			},
			expectedMediaId: 145064,
			otherMediaIds:   []int{113415},
		},
		{
			name: "Dungeon Meshi - 153518",
			paths: []string{
				"E:/Anime/Dungeon Meshi/Dungeon Meshi - 01.mkv",
			},
			expectedMediaId: 153518,
		},
		{
			name: "Violet Evergarden - 21827",
			paths: []string{
				"E:/Anime/Violet Evergarden/[SubsPlease] Violet Evergarden - 01 (1080p) [A1B2C3D4].mkv",
				"E:/Anime/Violet Evergarden/[SubsPlease] Violet Evergarden - 02 (1080p) [E5F6G7H8].mkv",
			},
			expectedMediaId: 21827,
		},
		{
			name: "Flying Witch - 21284",
			paths: []string{
				"E:/Anime/Flying Witch/[Erai-raws] Flying Witch - 01 [1080p][HEVC][Multiple Subtitle].mkv",
			},
			expectedMediaId: 21284,
		},
		{
			name: "Durarara - 6746",
			paths: []string{
				"E:/Anime/Durarara/Durarara.S01E01.1080p.BluRay.x264-GROUP.mkv",
				"E:/Anime/Durarara/Durarara.S01E02.1080p.BluRay.x264-GROUP.mkv",
			},
			expectedMediaId: 6746,
		},
		{
			name: "HIGH CARD - 135778",
			paths: []string{
				"E:/Anime/HIGH CARD (01-12) [1080p] [Dual-Audio]/[ASW] HIGH CARD - 01 [1080p HEVC x265 10Bit][AAC].mkv",
			},
			expectedMediaId: 135778,
		},
		{
			name: "Baccano - 2251",
			paths: []string{
				"E:/Anime/Baccano!/[Judas] Baccano! - S01E01.mkv",
				"E:/Anime/Baccano!/[Judas] Baccano! - S01E05.mkv",
			},
			expectedMediaId: 2251,
		},
		{
			name: "Kimi ni Todoke - 6045",
			paths: []string{
				"E:/Anime/Kimi ni Todoke/Kimi.ni.Todoke.S01.1080p.BluRay.10-Bit.Dual-Audio.FLAC.x265-YURASUKA/Kimi.ni.Todoke.S01E01.mkv",
			},
			expectedMediaId: 6045,
		},
		{
			name: "Zom 100 - 159831",
			paths: []string{
				"E:/Anime/Zom 100/Zom.100.Bucket.List.of.the.Dead.S01.1080p.BluRay.Remux.Dual.Audio.x265-EMBER/S01E01-Zom 100 [12345678].mkv",
			},
			expectedMediaId: 159831,
		},
		{
			name: "Kimi ni Todoke 2ND SEASON - 9656",
			paths: []string{
				"E:/Anime/Kimi ni Todoke 2ND SEASON/[SubsPlease] Kimi ni Todoke 2nd Season - 01 (1080p).mkv",
			},
			expectedMediaId: 9656,
			otherMediaIds:   []int{6045},
		},
		{
			name: "Durarara!!x2 Shou - 20652",
			paths: []string{
				"E:/Anime/Durarara x2 Shou/[HorribleSubs] Durarara!! x2 Shou - 01 [1080p].mkv",
			},
			expectedMediaId: 20652,
			otherMediaIds:   []int{6746},
		},
		{
			name: "HIGH CARD Season 2 - 163151",
			paths: []string{
				"E:/Anime/HIGH CARD Season 2/[SubsPlease] HIGH CARD Season 2 - 01 (1080p) [ABCD1234].mkv",
			},
			expectedMediaId: 163151,
			otherMediaIds:   []int{135778},
		},
		{
			name: "86 EIGHTY-SIX Part 2 - 131586",
			paths: []string{
				"E:/Anime/86 Eighty-Six Part 2/[SubsPlease] 86 Eighty-Six Part 2 - 01 (1080p).mkv",
			},
			expectedMediaId: 131586,
			otherMediaIds:   []int{116589},
		},
		{
			name: "Evangelion 1.0 - 2759",
			paths: []string{
				"E:/Anime/Evangelion Rebuild/Evangelion.1.0.You.Are.Not.Alone.2007.1080p.BluRay.x264-GROUP.mkv",
			},
			expectedMediaId: 2759,
		},
		{
			name: "Evangelion 2.0 - 3784",
			paths: []string{
				"E:/Anime/Evangelion Rebuild/Evangelion.2.22.You.Can.Not.Advance.2009.1080p.BluRay.x265-GROUP.mkv",
			},
			expectedMediaId: 3784,
			otherMediaIds:   []int{2759, 3786}, // Include Eva 1.0 and Eva 3.0+1.0 for conflict testing
		},
		{
			// One Piece Film Gold
			name: "One Piece Film Gold - 21335",
			paths: []string{
				"E:/Anime/One Piece Movies/One.Piece.Film.Gold.2016.1080p.BluRay.x264-GROUP.mkv",
			},
			expectedMediaId: 21335,
		},
		{
			name: "Violet Evergarden - 21827",
			paths: []string{
				"E:/Anime/Violet Evergarden/Season 01/Violet Evergarden - S01E01 - Episode Title.mkv",
			},
			expectedMediaId: 21827,
		},
		{
			name: "Flying Witch (2016) - 21284",
			paths: []string{
				"E:/Anime/Flying Witch (2016)/Season 01/Flying Witch (2016) - S01E01 - Stone Seeker.mkv",
			},
			expectedMediaId: 21284,
		},
		{
			name: "Baccano! with punctuation - 2251",
			paths: []string{
				"E:/Anime/Baccano!/Baccano! - 01 [BD 1080p] [5.1 Dual Audio].mkv",
			},
			expectedMediaId: 2251,
		},
		{
			name: "86 - Eighty Six with dashes - 116589",
			paths: []string{
				"E:/Anime/86 - Eighty Six/86 - Eighty Six - 01 - Undertaker.mkv",
			},
			expectedMediaId: 116589,
		},
		{
			name: "Evangelion 3.0+1.0 - 3786",
			paths: []string{
				"E:/Anime/Evangelion 3.0+1.0/Evangelion.3.0+1.0.Thrice.Upon.a.Time.2021.1080p.AMZN.WEB-DL.DDP5.1.x264-GROUP.mkv",
			},
			expectedMediaId: 3786,
		},
		{
			name: "Insomniacs After School x265 - 143653",
			paths: []string{
				"E:/Anime/Kimi wa Houkago Insomnia/[ASW] Kimi wa Houkago Insomnia - 01 [1080p HEVC][AAC].mkv",
			},
			expectedMediaId: 143653,
		},
		{
			name: "Kimi wa Houkago Insomnia 10bit - 143653",
			paths: []string{
				"E:/Anime/Insomniacs After School/Insomniacs.After.School.S01E01.1080p.WEB-DL.10bit.x265-GROUP.mkv",
			},
			expectedMediaId: 143653,
		},
		{
			name: "One Piece Stampede WEB-DL - 105143",
			paths: []string{
				"E:/Anime/One Piece Movies/One.Piece.Stampede.2019.1080p.NF.WEB-DL.DDP5.1.H.264-GROUP.mkv",
			},
			expectedMediaId: 105143,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			// Add media to collection if it doesn't exist
			allMedia := animeCollection.GetAllAnime()

			// Helper to ensure media exists in collection
			hasMedia := false
			for _, media := range allMedia {
				if media.ID == tt.expectedMediaId {
					hasMedia = true
					break
				}
			}
			if !hasMedia {
				anilist.TestAddAnimeCollectionWithRelationsEntry(animeCollection, tt.expectedMediaId, anilist.TestModifyAnimeCollectionEntryInput{Status: lo.ToPtr(anilist.MediaListStatusCurrent)}, anilistClient)
				allMedia = animeCollection.GetAllAnime()
			}

			// Ensure other media exists
			for _, id := range tt.otherMediaIds {
				hasMedia := false
				for _, media := range allMedia {
					if media.ID == id {
						hasMedia = true
						break
					}
				}
				if !hasMedia {
					anilist.TestAddAnimeCollectionWithRelationsEntry(animeCollection, id, anilist.TestModifyAnimeCollectionEntryInput{Status: lo.ToPtr(anilist.MediaListStatusCurrent)}, anilistClient)
					allMedia = animeCollection.GetAllAnime()
				}
			}

			scanLogger, err := NewConsoleScanLogger()
			if err != nil {
				t.Fatal("expected result, got error:", err.Error())
			}

			// +---------------------+
			// |   Local Files       |
			// +---------------------+

			var lfs []*anime.LocalFile
			for _, path := range tt.paths {
				lf := anime.NewLocalFile(path, dir)
				lfs = append(lfs, lf)
			}

			// +---------------------+
			// |   MediaContainer    |
			// +---------------------+

			mc := NewMediaContainer(&MediaContainerOptions{
				AllMedia:   NormalizedMediaFromAnilistComplete(allMedia),
				ScanLogger: scanLogger,
			})

			// +---------------------+
			// |      Matcher        |
			// +---------------------+

			matcher := &Matcher{
				LocalFiles:        lfs,
				MediaContainer:    mc,
				Logger:            util.NewLogger(),
				ScanLogger:        scanLogger,
				ScanSummaryLogger: nil,
			}

			err = matcher.MatchLocalFilesWithMedia()

			if assert.NoError(t, err, "Error while matching local files") {
				for _, lf := range lfs {
					if lf.MediaId != tt.expectedMediaId {
						t.Errorf("FAILED: expected media id %d, got %d for file %s", tt.expectedMediaId, lf.MediaId, lf.Name)
					} else {
						t.Logf("SUCCESS: local file: %s -> media id: %d", lf.Name, lf.MediaId)
					}
				}
			}
		})
	}

}

func TestMatcher4(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := anilist.NewAnilistClient(test_utils.ConfigData.Provider.AnilistJwt, "")
	animeCollection, err := anilistClient.AnimeCollectionWithRelations(context.Background(), &test_utils.ConfigData.Provider.AnilistUsername)
	if err != nil {
		t.Fatal(err.Error())
	}

	dir := "E:/Anime"

	tests := []struct {
		name            string
		paths           []string
		expectedMediaId int
		otherMediaIds   []int
	}{
		// Abbreviated titles
		{
			name: "Bunny Girl Senpai abbreviated - 101291",
			paths: []string{
				"E:/Anime/Bunny Girl Senpai/[SubsPlease] Bunny Girl Senpai - 01 (1080p).mkv",
			},
			expectedMediaId: 101291,
		},
		{
			// Romaji title
			name: "Seishun Buta Yarou full title - 101291",
			paths: []string{
				"E:/Anime/Seishun Buta Yarou/Seishun.Buta.Yarou.wa.Bunny.Girl.Senpai.no.Yume.wo.Minai.S01E01.1080p.BluRay.x264.mkv",
			},
			expectedMediaId: 101291,
		},
		// Mushoku Tensei parts/seasons
		{
			name: "Mushoku Tensei S2 - 146065",
			paths: []string{
				"E:/Anime/Mushoku Tensei S2/[SubsPlease] Mushoku Tensei S2 - 01 (1080p) [EC64C8B1].mkv",
			},
			expectedMediaId: 146065,
			otherMediaIds:   []int{108465, 127720, 166873}, // Part 1, Cour 2, Season 2 Part 2
		},
		{
			// Season 2 Part 2 (Erai-raws)
			name: "Mushoku Tensei II Part 2 Erai-raws - 166873",
			paths: []string{
				"E:/Anime/Mushoku Tensei II Part 2/[Erai-raws] Mushoku Tensei II Part 2 - 06 [1080p][HEVC][Multiple Subtitle][7509990E].mkv",
			},
			expectedMediaId: 166873, // Season 2 Part 2
			otherMediaIds:   []int{108465, 146065},
		},
		{
			// Jobless Reincarnation (English)
			name: "Jobless Reincarnation S2 - 146065",
			paths: []string{
				"E:/Anime/Jobless Reincarnation/Mushoku.Tensei.Jobless.Reincarnation.S02E01.1080p.CR.WEB-DL.x264.mkv",
			},
			expectedMediaId: 146065,
			otherMediaIds:   []int{108465},
		},
		// Bungo Stray Dogs seasons
		{
			name: "Bungou Stray Dogs S1 - 21311",
			paths: []string{
				"E:/Anime/Bungou Stray Dogs/[Judas] Bungo Stray Dogs - S01E01.mkv",
			},
			expectedMediaId: 21311,
			otherMediaIds:   []int{21679}, // S2
		},
		{
			name: "Bungou Stray Dogs S2 - 21679",
			paths: []string{
				"E:/Anime/Bungou Stray Dogs 2nd Season/Bungou.Stray.Dogs.S02E01.1080p.BluRay.x264-GROUP.mkv",
			},
			expectedMediaId: 21679,
			otherMediaIds:   []int{21311}, // S1
		},
		{
			name: "BSD 5th Season abbreviated - 163263",
			paths: []string{
				"E:/Anime/BSD S5/[SubsPlease] Bungou Stray Dogs 5th Season - 01 (1080p).mkv",
			},
			expectedMediaId: 163263,
			otherMediaIds:   []int{21311, 21679}, // S1, S2
		},
		// Golden Kamuy
		{
			name: "Golden Kamuy S3 - 110355",
			paths: []string{
				"E:/Anime/Golden Kamuy 3rd Season/Golden.Kamuy.S03E01.1080p.WEB-DL.x264.mkv",
			},
			expectedMediaId: 110355,
			otherMediaIds:   []int{102977}, // S2
		},
		// Blue Lock
		{
			name: "Blue Lock S1 - 137822",
			paths: []string{
				"E:/Anime/Blue Lock/[SubsPlease] Blue Lock - 01 (1080p).mkv",
			},
			expectedMediaId: 137822,
			otherMediaIds:   []int{163146}, // S2
		},
		{
			name: "Blue Lock 2nd Season - 163146",
			paths: []string{
				"E:/Anime/Blue Lock 2nd Season/[SubsPlease] Blue Lock 2nd Season - 01 (1080p) [HASH].mkv",
			},
			expectedMediaId: 163146,
			otherMediaIds:   []int{137822}, // S1
		},
		{
			name: "Violet Evergarden Gaiden - 109190",
			paths: []string{
				"E:/Anime/Violet Evergarden Gaiden/Violet.Evergarden.Eternity.and.the.Auto.Memory.Doll.2019.1080p.BluRay.x264.mkv",
			},
			expectedMediaId: 109190,
			otherMediaIds:   []int{21827}, // Main series
		},
		{
			name: "Zom 100 short name - 159831",
			paths: []string{
				"E:/Anime/Zom 100/[ASW] Zom 100 - 01 [1080p HEVC].mkv",
			},
			expectedMediaId: 159831,
		},
		{
			name: "Insomniacs main series not special - 143653",
			paths: []string{
				"E:/Anime/Kimi wa Houkago Insomnia/[Erai-raws] Kimi wa Houkago Insomnia - 01 [1080p].mkv",
			},
			expectedMediaId: 143653,
			otherMediaIds:   []int{160205}, // Special Animation PV
		},
		{
			name: "Kekkai Sensen - 20727",
			paths: []string{
				"E:/Anime/[Anime Time] Kekkai Sensen (Blood Blockade Battlefront) S01+02+OVA+Extra [Dual Audio][BD][1080p][HEVC 10bit x265][AAC][Eng Sub]/Blood Blockade Battlefront/NC/Blood Blockade Battlefront - NCED.mkv",
			},
			expectedMediaId: 20727,
		},
		{
			name: "BnH 8 - 182896",
			paths: []string{
				"E:/Anime/Boku no Hero Academia FINAL SEASON/My.Hero.Academia.S08E07.From.Aizawa.1080p.NF.WEB-DL.AAC2.0.H.264-VARYG.mkv",
			},
			expectedMediaId: 182896,
		},
		{
			name: "Frieren NCED",
			paths: []string{
				"E:/Anime/Sousou no Frieren/Frieren Beyond Journey's End (BD Remux 1080p AVC FLAC AAC) [Dual Audio] [PMR]/Extras/NCED 02 (BD Remux 1080p AVC FLAC) [PMR].mkv",
			},
			expectedMediaId: 154587,
		},
		// Extras
		{
			name: "NCED in Extras folder, should match base series (Violet Evergarden)",
			paths: []string{
				"E:/Anime/Violet Evergarden/Violet Evergarden (BD 1080p)/Extras/NCED 01 (BD 1080p).mkv",
			},
			expectedMediaId: 21827, // Violet Evergarden (base series)
		},
		{
			name: "NCOP in Extras folder, should match base series (Flying Witch)",
			paths: []string{
				"E:/Anime/Flying Witch/Flying Witch BD/Extras/NCOP (BD 1080p).mkv",
			},
			expectedMediaId: 21284, // Flying Witch
		},
		{
			name: "Ending file in Extras, should match base series (Durarara)",
			paths: []string{
				"E:/Anime/Durarara/Durarara BD/Extras/Ending 01.mkv",
			},
			expectedMediaId: 6746, // Durarara!!
		},
		// Parts
		{
			name: "86 Eighty Six Part 2 explicit",
			paths: []string{
				"E:/Anime/86 Eighty Six Part 2/[SubsPlease] 86 - Eighty Six - 20 (1080p).mkv",
			},
			expectedMediaId: 131586, // 86: Eighty Six Part 2
		},
		// Shortened name
		{
			name: "Bunny Girl Senpai",
			paths: []string{
				"E:/Anime/Bunny Girl/Bunny.Girl.Senpai.E01.mkv",
			},
			expectedMediaId: 101291, // Seishun Buta Yarou wa Bunny Girl Senpai no Yume wo Minai
		},
		// Year
		{
			name: "Year in filename - Flying Witch (2016)",
			paths: []string{
				"E:/Anime/Flying Witch (2016)/Flying.Witch.2016.E01.1080p.BluRay.mkv",
			},
			expectedMediaId: 21284, // Flying Witch
		},
		{
			name: "Year in path - Evangelion rebuild (2007)",
			paths: []string{
				"E:/Anime/Evangelion Rebuild/Evangelion.1.0.You.Are.Not.Alone.2007.1080p.BluRay.mkv",
			},
			expectedMediaId: 2759, // Evangelion Shin Movie: Jo
		},
		// Movies and specials
		{
			name: "Movie - Violet Evergarden Gaiden",
			paths: []string{
				"E:/Anime/Violet Evergarden Gaiden/Violet.Evergarden.Gaiden.2019.1080p.BluRay.mkv",
			},
			expectedMediaId: 109190, // Violet Evergarden Gaiden: Eien to Jidou Shuki Ningyou
		},
		{
			name: "Movie - Bunny Girl Senpai Movie (Dreaming Girl)",
			paths: []string{
				"E:/Anime/Seishun Buta Yarou/Seishun.Buta.Yarou.wa.Yumemiru.Shoujo.no.Yume.wo.Minai.2019.1080p.BluRay.mkv",
			},
			expectedMediaId: 104157, // Seishun Buta Yarou wa Yumemiru Shoujo no Yume wo Minai
		},
		{
			name: "Danmachi IV Part 2",
			paths: []string{
				"E:/Anime/Re Zero/Danmachi S04P02 1080p Dual Audio BDRip 10 bits DD x265-EMBER/S04E12-Amphisbaena A Song of Despair [080E734C].mkv",
			},
			expectedMediaId: 155211,
			otherMediaIds:   []int{},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			allMedia := animeCollection.GetAllAnime()

			hasMedia := false
			for _, media := range allMedia {
				if media.ID == tt.expectedMediaId {
					hasMedia = true
					break
				}
			}
			if !hasMedia {
				anilist.TestAddAnimeCollectionWithRelationsEntry(animeCollection, tt.expectedMediaId, anilist.TestModifyAnimeCollectionEntryInput{Status: lo.ToPtr(anilist.MediaListStatusCurrent)}, anilistClient)
				allMedia = animeCollection.GetAllAnime()
			}

			for _, id := range tt.otherMediaIds {
				hasMedia := false
				for _, media := range allMedia {
					if media.ID == id {
						hasMedia = true
						break
					}
				}
				if !hasMedia {
					anilist.TestAddAnimeCollectionWithRelationsEntry(animeCollection, id, anilist.TestModifyAnimeCollectionEntryInput{Status: lo.ToPtr(anilist.MediaListStatusCurrent)}, anilistClient)
					allMedia = animeCollection.GetAllAnime()
				}
			}

			scanLogger, err := NewConsoleScanLogger()
			if err != nil {
				t.Fatal("expected result, got error:", err.Error())
			}

			var lfs []*anime.LocalFile
			for _, path := range tt.paths {
				lf := anime.NewLocalFile(path, dir)
				lfs = append(lfs, lf)
			}

			mc := NewMediaContainer(&MediaContainerOptions{
				AllMedia:   NormalizedMediaFromAnilistComplete(allMedia),
				ScanLogger: scanLogger,
			})

			matcher := &Matcher{
				LocalFiles:        lfs,
				MediaContainer:    mc,
				Logger:            util.NewLogger(),
				ScanLogger:        scanLogger,
				ScanSummaryLogger: nil,
				Debug:             true,
			}

			err = matcher.MatchLocalFilesWithMedia()

			if assert.NoError(t, err, "Error while matching local files") {
				for _, lf := range lfs {
					if lf.MediaId != tt.expectedMediaId {
						t.Errorf("FAILED: expected media id %d, got %d for file %s", tt.expectedMediaId, lf.MediaId, lf.Name)
					} else {
						t.Logf("SUCCESS: local file: %s -> media id: %d", lf.Name, lf.MediaId)
					}
				}
			}
		})
	}

}

func TestMatcherComplexCases(t *testing.T) {
	t.Skip()
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := anilist.NewAnilistClient(test_utils.ConfigData.Provider.AnilistJwt, "")
	animeCollection, err := anilistClient.AnimeCollectionWithRelations(context.Background(), &test_utils.ConfigData.Provider.AnilistUsername)
	if err != nil {
		t.Fatal(err.Error())
	}

	dir := "E:/Anime"

	tests := []struct {
		name            string
		paths           []string
		expectedMediaId int
		otherMediaIds   []int
	}{
		{
			name: "DanMachi S5 - Ryuu release with full romaji title",
			paths: []string{
				"E:/Anime/DanMachi/[Ryuu] Dungeon ni Deai wo Motomeru no wa Machigatteiru Darou ka - S05E11 (WEB 1080p HEVC x265 10-bit EAC-3) [Dual-Audio].mkv",
			},
			expectedMediaId: 170732,
			otherMediaIds:   []int{20920, 101167, 116006},
		},
		{
			name: "DanMachi S5 - SubsPlus English title format",
			paths: []string{
				"E:/Anime/Is It Wrong to Try to Pick Up Girls in a Dungeon/[SubsPlus+] Is It Wrong to Try to Pick Up Girls in a Dungeon - S05E14 (CR WEB-DL 1080p AVC EAC3).mkv",
			},
			expectedMediaId: 170732,
			otherMediaIds:   []int{20920, 101167},
		},
		{
			name: "DanMachi S4 - EMBER release with abbreviated title",
			paths: []string{
				"E:/Anime/DanMachi/[EMBER] Dungeon ni Deai wo Motomeru no wa Machigatteiru Darou ka (DanMachi) S05E13 [1080p] [HEVC WEBRip DDP].mkv",
			},
			expectedMediaId: 170732,
			otherMediaIds:   []int{20920, 142861},
		},
		{
			name: "DanMachi S4 BD batch with complex naming",
			paths: []string{
				"E:/Anime/DanMachi IV/[Anipakku] DanMachi (S04) [BD 1080p x265 OPUS][PT-BR]/[Anipakku] DanMachi - S04E01.mkv",
			},
			expectedMediaId: 142861,
			otherMediaIds:   []int{20920, 170732},
		},
		{
			name: "DanMachi S1 no season indicator",
			paths: []string{
				"E:/Anime/DanMachi/[HorribleSubs] DanMachi - 01 [1080p].mkv",
			},
			expectedMediaId: 20920,
			otherMediaIds:   []int{101167, 116006, 170732},
		},
		{
			name: "Oregairu S2 - Senjou BD release with Zoku",
			paths: []string{
				"E:/Anime/Oregairu/[Senjou] Yahari Ore no Seishun Lovecome wa Machigatte Iru. Zoku (My Teen Romantic Comedy SNAFU TOO!) (Oregairu) [BDRip 1920x1080 x264 FLAC]/[Senjou] Oregairu Zoku - 01.mkv",
			},
			expectedMediaId: 23847,
			otherMediaIds:   []int{14813, 108489},
		},
		{
			name: "Oregairu complete batch - EMBER release",
			paths: []string{
				"E:/Anime/Oregairu/[EMBER] Oregairu-My Teen Romantic Comedy SNAFU (2013-2020) (Season 1+2+3+OVA) [BDRip] [1080p Dual Audio HEVC 10 bits]/Season 1/[EMBER] Oregairu S01E01.mkv",
			},
			expectedMediaId: 14813,
			otherMediaIds:   []int{23847, 108489},
		},
		{
			name: "Oregairu S3 Climax - EMBER release with Kan",
			paths: []string{
				"E:/Anime/Oregairu/[EMBER] Oregairu-My Teen Romantic Comedy SNAFU Climax! (2020) (Season 3) [BDRip] [1080p Dual Audio HEVC 10 bits DDP]/[EMBER] Oregairu Kan - 01.mkv",
			},
			expectedMediaId: 108489,
			otherMediaIds:   []int{14813, 23847},
		},
		{
			name: "Oregairu S3 - MTBB BD release",
			paths: []string{
				"E:/Anime/Oregairu/[MTBB] My Teen Romantic Comedy SNAFU Climax (BD 1080p)/[MTBB] My Teen Romantic Comedy SNAFU Climax - 01.mkv",
			},
			expectedMediaId: 108489,
			otherMediaIds:   []int{14813, 23847},
		},
		{
			name: "Oregairu S2 - MTBB BD release with TOO",
			paths: []string{
				"E:/Anime/Oregairu/[MTBB] My Teen Romantic Comedy SNAFU TOO! (BD 1080p)/[MTBB] My Teen Romantic Comedy SNAFU TOO - 01.mkv",
			},
			expectedMediaId: 23847,
			otherMediaIds:   []int{14813, 108489},
		},
		{
			name: "Oregairu S3 - Judas release with explicit Season 3",
			paths: []string{
				"E:/Anime/Oregairu/[Judas] Yahari Ore no Seishun Love Come wa Machigatteiru (Season 3 \"Kan\") [1080p][HEVC x265 10bit][Multi-Subs]/[Judas] Oregairu Kan - 01.mkv",
			},
			expectedMediaId: 108489,
			otherMediaIds:   []int{14813, 23847},
		},
		{
			name: "AoT Final Season - Yameii S04E30 with Final Chapters",
			paths: []string{
				"E:/Anime/Attack on Titan/[Yameii] Attack on Titan - S04E30 [English Dub] [AS WEB-DL 1080p] [19ED00C6].mkv",
			},
			expectedMediaId: 162314,
			otherMediaIds:   []int{16498, 110277, 131681},
		},
		{
			name: "AoT Final Season Part 4 - Rapta release",
			paths: []string{
				"E:/Anime/Attack on Titan/Attack.on.Titan.Shingeki.no.Kyojin.The.Final.Season.Part.4.The.Final.Chapters.Special.2.1080p.CR.WEBRip.10bits.x265-Rapta.mkv",
			},
			expectedMediaId: 162314,
			otherMediaIds:   []int{110277, 131681},
		},
		{
			name: "AoT Final Season - EMBER Season 4 Part 04 batch",
			paths: []string{
				"E:/Anime/Attack on Titan/[EMBER] Shingeki no Kyojin (2023) (Season 4 Part 04) [1080p] [HEVC WEBRip DDP]/[EMBER] Shingeki no Kyojin S04P04E01.mkv",
			},
			expectedMediaId: 162314,
			otherMediaIds:   []int{110277, 131681},
		},
		{
			name: "AoT Final Season - Anime Time Part 3 Part 2",
			paths: []string{
				"E:/Anime/Attack on Titan/[Anime Time] Shingeki no Kyojin - The Final Season Part 3 - Part 2 (Season 04 - 32-36) [1080p][HEVC 10bit x265][AAC][Multi Sub]/[Anime Time] Shingeki no Kyojin S04E35.mkv",
			},
			expectedMediaId: 162314,
			otherMediaIds:   []int{110277, 131681, 163263},
		},
		{
			name: "AoT Season 1 - no indicator in simple filename",
			paths: []string{
				"E:/Anime/Shingeki no Kyojin/[HorribleSubs] Shingeki no Kyojin - 01 [1080p].mkv",
			},
			expectedMediaId: 16498,
			otherMediaIds:   []int{20958, 110277},
		},
		{
			name: "Fate/Zero S01 - Prof BD release",
			paths: []string{
				"E:/Anime/Fate Zero/Fate Zero (2011) S01 [1080p x265 HEVC 10bit BluRay Dual Audio AAC] [Prof]/[Prof] Fate Zero - S01E01.mkv",
			},
			expectedMediaId: 10087,
			otherMediaIds:   []int{11741},
		},
		{
			name: "Fate/Zero S02 - Prof BD release",
			paths: []string{
				"E:/Anime/Fate Zero/Fate Zero (2012) S02 [1080p x265 HEVC 10bit BluRay Dual Audio AAC] [Prof]/[Prof] Fate Zero - S02E01.mkv",
			},
			expectedMediaId: 11741,
			otherMediaIds:   []int{10087},
		},
		{
			name: "Fate/Zero - DragsterPS multi-audio S02",
			paths: []string{
				"E:/Anime/Fate Zero/[DragsterPS] Fate Zero S02 [1080p] [Multi-Audio] [Multi-Subs]/[DragsterPS] Fate Zero S02E01.mkv",
			},
			expectedMediaId: 11741,
			otherMediaIds:   []int{10087},
		},
		{
			name: "Fate/Zero - UTW classic release no season",
			paths: []string{
				"E:/Anime/Fate Zero/[UTW] Fate Zero - 01-25 + Specials [BD][h264-720p_AC3]/[UTW] Fate Zero - 01 [BD][h264-720p][AC3].mkv",
			},
			expectedMediaId: 10087,
			otherMediaIds:   []int{11741},
		},
		{
			name: "Fate/Zero - Coalgirls classic release",
			paths: []string{
				"E:/Anime/Fate Zero/[Coalgirls]_Fate_Zero_(1920x1080_Blu-ray_FLAC)/[Coalgirls]_Fate_Zero_-_01_(1920x1080_Blu-ray_FLAC).mkv",
			},
			expectedMediaId: 10087,
			otherMediaIds:   []int{11741},
		},
		{
			name: "KonoSuba S03 - ToonsHub with full English title",
			paths: []string{
				"E:/Anime/KonoSuba/[ToonsHub] KONOSUBA -Gods blessing on this wonderful world S03E12 1080p CR WEB-DL AAC2.0 H.264.mkv",
			},
			expectedMediaId: 136804,
			otherMediaIds:   []int{21202, 21699},
		},
		{
			name: "KonoSuba S03 - fig BD remux with season indicator",
			paths: []string{
				"E:/Anime/KonoSuba/[fig] Konosuba S03 (BluRay Remux 1080p AVC FLAC AAC) [Dual-Audio]/[fig] Konosuba S03E01.mkv",
			},
			expectedMediaId: 136804,
			otherMediaIds:   []int{21202, 21699},
		},
		{
			name: "KonoSuba S02 - fig BD with Season 2 in folder",
			paths: []string{
				"E:/Anime/KonoSuba/[fig] Konosuba S02 (BluRay Remux 1080p AVC FLAC) [Dual-Audio] (Season 2 + OVA)/[fig] Konosuba S02E01.mkv",
			},
			expectedMediaId: 21699,
			otherMediaIds:   []int{21202, 136804},
		},
		{
			name: "KonoSuba S01 - fig BD with Season 1 in folder",
			paths: []string{
				"E:/Anime/KonoSuba/[fig] Konosuba S01 (BluRay Remux 1080p AVC FLAC) [Dual-Audio] (Season 1 + OVA)/[fig] Konosuba S01E01.mkv",
			},
			expectedMediaId: 21202,
			otherMediaIds:   []int{21699, 136804},
		},
		{
			name: "KonoSuba - EMBER S3 batch with Japanese title",
			paths: []string{
				"E:/Anime/KonoSuba/[EMBER] Kono Subarashii Sekai ni Shukufuku wo! (2024) (Season 3) [1080p] [Dual Audio HEVC WEBRip DDP]/[EMBER] Kono Subarashii Sekai ni Shukufuku wo! - S03E01.mkv",
			},
			expectedMediaId: 136804,
			otherMediaIds:   []int{21202, 21699},
		},
		{
			name: "KonoSuba no season - should match S1",
			paths: []string{
				"E:/Anime/KonoSuba/[HorribleSubs] Kono Subarashii Sekai ni Shukufuku wo! - 01 [1080p].mkv",
			},
			expectedMediaId: 21202,
			otherMediaIds:   []int{21699, 136804},
		},
		{
			name: "SAO Alicization WoU Part 2 - Lulu BD release",
			paths: []string{
				"E:/Anime/Sword Art Online/[Lulu] Sword Art Online Alicization - War of Underworld Part 2 (BD 1080p HEVC FLAC)[Dual-Audio]/[Lulu] SAO Alicization WoU Part 2 - 01.mkv",
			},
			expectedMediaId: 114308,
			otherMediaIds:   []int{100182, 108759},
		},
		{
			name: "SAO Alicization WoU Part 1 - Lulu BD release",
			paths: []string{
				"E:/Anime/Sword Art Online/[Lulu] Sword Art Online Alicization - War of Underworld Part 1 (BD 1080p Hi10 FLAC)[Dual-Audio]/[Lulu] SAO Alicization WoU - 01.mkv",
			},
			expectedMediaId: 108759,
			otherMediaIds:   []int{100182, 114308},
		},
		{
			name: "SAO complete - Tenrai-Sensei mega batch",
			paths: []string{
				"E:/Anime/Sword Art Online/[Tenrai-Sensei] Sword Art Online S1+S2+S3+S4+The Movie+OVAs [BD][1080p][HEVC 10bit x265][Dual Audio]/S1/[Tenrai-Sensei] SAO - S01E01.mkv",
			},
			expectedMediaId: 11757,
			otherMediaIds:   []int{21881, 100182},
		},
		{
			name: "SAO Alicization - HorribleRips release",
			paths: []string{
				"E:/Anime/Sword Art Online/[HorribleRips] Sword Art Online Alicization [1080p]/[HorribleRips] Sword Art Online Alicization - 01 [1080p].mkv",
			},
			expectedMediaId: 100182,
			otherMediaIds:   []int{11757, 108759},
		},
		{
			name: "SAO Alicization WoU - HorribleRips release",
			paths: []string{
				"E:/Anime/Sword Art Online/[HorribleRips] Sword Art Online Alicization - War of Underworld [1080p]/[HorribleRips] SAO Alicization WoU - 01 [1080p].mkv",
			},
			expectedMediaId: 108759,
			otherMediaIds:   []int{100182, 114308},
		},
		{
			name: "Mob Psycho S3 - SubsPlease release",
			paths: []string{
				"E:/Anime/Mob Psycho 100/[SubsPlease] Mob Psycho 100 S3 - 12 (1080p) [E5058D7B].mkv",
			},
			expectedMediaId: 140439,
			otherMediaIds:   []int{21507, 101338},
		},
		{
			name: "Mob Psycho III - EMBER batch release",
			paths: []string{
				"E:/Anime/Mob Psycho 100/[EMBER] Mob Psycho 100 (2022) (Season 3) [1080p] [Dual Audio HEVC WEBRip]/[EMBER] Mob Psycho 100 III - 01.mkv",
			},
			expectedMediaId: 140439,
			otherMediaIds:   []int{21507, 101338},
		},
		{
			name: "Mob Psycho III - Judas batch",
			paths: []string{
				"E:/Anime/Mob Psycho 100/[Judas] Mob Psycho 100 (Season 3) [1080p][HEVC x265 10bit][Multi-Subs]/[Judas] Mob Psycho 100 III - 01.mkv",
			},
			expectedMediaId: 140439,
			otherMediaIds:   []int{21507, 101338},
		},
		{
			name: "Mob Psycho III - scoot WEB release",
			paths: []string{
				"E:/Anime/Mob Psycho 100/[scoot] Mob Psycho 100 III [WEB 1080p HEVC]/[scoot] Mob Psycho 100 III - 01.mkv",
			},
			expectedMediaId: 140439,
			otherMediaIds:   []int{21507, 101338},
		},
		{
			name: "Mob Psycho 100 base - no season indicator",
			paths: []string{
				"E:/Anime/Mob Psycho 100/[HorribleSubs] Mob Psycho 100 - 01 [1080p].mkv",
			},
			expectedMediaId: 21507,
			otherMediaIds:   []int{101338, 140439},
		},
		{
			name: "Index III - Gremlin BD release",
			paths: []string{
				"E:/Anime/Toaru Majutsu no Index/[Gremlin] Toaru Majutsu no Index III [BD][1080p]/[Gremlin] Toaru Majutsu no Index III - 01.mkv",
			},
			expectedMediaId: 36432,
			otherMediaIds:   []int{4654, 8937},
		},
		{
			name: "Index II - Erai-raws batch",
			paths: []string{
				"E:/Anime/Toaru Majutsu no Index/[Erai-raws] Toaru Majutsu no Index II - 01 ~ 24 [1080p][Multiple Subtitle]/[Erai-raws] Toaru Majutsu no Index II - 01.mkv",
			},
			expectedMediaId: 8937,
			otherMediaIds:   []int{4654, 36432},
		},
		{
			name: "Index base - Erai-raws batch no numeral",
			paths: []string{
				"E:/Anime/Toaru Majutsu no Index/[Erai-raws] Toaru Majutsu no Index - 01 ~ 24 [1080p][Multiple Subtitle]/[Erai-raws] Toaru Majutsu no Index - 01.mkv",
			},
			expectedMediaId: 4654,
			otherMediaIds:   []int{8937, 36432},
		},
		{
			name: "Index III - HorribleSubs weekly",
			paths: []string{
				"E:/Anime/Toaru Majutsu no Index III/[HorribleSubs] Toaru Majutsu no Index III - 26 [1080p].mkv",
			},
			expectedMediaId: 36432,
			otherMediaIds:   []int{4654, 8937},
		},
		{
			name: "Index III - English title alternative",
			paths: []string{
				"E:/Anime/A Certain Magical Index/[Cleo] A Certain Magical Index III [Dual Audio 10bit 720p][HEVC-x265]/[Cleo] A Certain Magical Index III - 01.mkv",
			},
			expectedMediaId: 36432,
			otherMediaIds:   []int{4654, 8937},
		},
		{
			name: "Re:Zero S3 - EMBER batch",
			paths: []string{
				"E:/Anime/Re Zero/[EMBER] Re Zero kara Hajimeru Isekai Seikatsu (2025) (Season 3) [1080p] [Dual Audio HEVC WEBRip DDP]/[EMBER] Re Zero S03E01.mkv",
			},
			expectedMediaId: 163134,
			otherMediaIds:   []int{21355, 108632},
		},
		{
			name: "Re:Zero S3 - Judas batch with colon",
			paths: []string{
				"E:/Anime/Re Zero/[Judas] Re Zero kara Hajimeru Isekai Seikatsu (Re Zero - Starting Life in Another World) (Season 03) [1080p][HEVC x265 10bit][Dual-Audio][Multi-Subs]/[Judas] ReZero S03E01.mkv",
			},
			expectedMediaId: 163134,
			otherMediaIds:   []int{21355, 108632},
		},
		{
			name: "Re:Zero S3 - FLE release with various aliases",
			paths: []string{
				"E:/Anime/Re Zero/[FLE] Re ZERO Starting Life in Another World - S03 (WEB 1080p H.264 E-AC-3) [Dual Audio]/[FLE] Re ZERO - S03E01.mkv",
			},
			expectedMediaId: 163134,
			otherMediaIds:   []int{21355, 108632},
		},
		{
			name: "Re:Zero S2 - Golumpa dub release",
			paths: []string{
				"E:/Anime/Re Zero/[Golumpa] Re ZERO -Starting Life in Another World- S02 [English Dub] [CR WEB-DL 720p]/[Golumpa] ReZero S02E01.mkv",
			},
			expectedMediaId: 108632,
			otherMediaIds:   []int{21355, 163134},
		},
		{
			name: "Re:Zero no season - should match base",
			paths: []string{
				"E:/Anime/Re Zero/[HorribleSubs] Re Zero kara Hajimeru Isekai Seikatsu - 01 [1080p].mkv",
			},
			expectedMediaId: 21355,
			otherMediaIds:   []int{108632, 163134},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure all required media is in the collection
			allMedia := animeCollection.GetAllAnime()

			// Add expected media
			hasMedia := false
			for _, media := range allMedia {
				if media.ID == tt.expectedMediaId {
					hasMedia = true
					break
				}
			}
			if !hasMedia {
				anilist.TestAddAnimeCollectionWithRelationsEntry(animeCollection, tt.expectedMediaId, anilist.TestModifyAnimeCollectionEntryInput{Status: lo.ToPtr(anilist.MediaListStatusCurrent)}, anilistClient)
				allMedia = animeCollection.GetAllAnime()
			}

			// Add other competing media
			for _, id := range tt.otherMediaIds {
				hasOther := false
				for _, media := range allMedia {
					if media.ID == id {
						hasOther = true
						break
					}
				}
				if !hasOther {
					anilist.TestAddAnimeCollectionWithRelationsEntry(animeCollection, id, anilist.TestModifyAnimeCollectionEntryInput{Status: lo.ToPtr(anilist.MediaListStatusCurrent)}, anilistClient)
					allMedia = animeCollection.GetAllAnime()
				}
			}

			scanLogger, err := NewConsoleScanLogger()
			if err != nil {
				t.Fatal("expected result, got error:", err.Error())
			}

			// Create local files
			var lfs []*anime.LocalFile
			for _, path := range tt.paths {
				lf := anime.NewLocalFile(path, dir)
				lfs = append(lfs, lf)
			}

			// Create MediaContainer
			mc := NewMediaContainer(&MediaContainerOptions{
				AllMedia:   NormalizedMediaFromAnilistComplete(allMedia),
				ScanLogger: scanLogger,
			})

			// Run matcher
			matcher := &Matcher{
				LocalFiles:        lfs,
				MediaContainer:    mc,
				Logger:            util.NewLogger(),
				ScanLogger:        scanLogger,
				ScanSummaryLogger: nil,
				Debug:             true,
			}

			err = matcher.MatchLocalFilesWithMedia()

			if assert.NoError(t, err, "Error while matching local files") {
				for _, lf := range lfs {
					if lf.MediaId != tt.expectedMediaId {
						t.Errorf("FAILED [%s]: expected media id %d, got %d for file %s",
							tt.name, tt.expectedMediaId, lf.MediaId, lf.Name)
					} else {
						t.Logf("SUCCESS: %s -> media id: %d", lf.Name, lf.MediaId)
					}
				}
			}
		})
	}
}

// TestMatcherWithOfflineDB tests matching using the anime-offline-database.
// MediaFetcher is initialized with DisableAnimeCollection=true and Enhanced=true.
func TestMatcherWithOfflineDB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := anilist.TestGetMockAnilistClient()
	logger := util.NewLogger()

	database, err := db.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, logger)
	if err != nil {
		t.Fatal(err)
	}
	anilistClientRef := util.NewRef(anilistClient)
	extensionBankRef := util.NewRef(extension.NewUnifiedBank())
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClientRef, extensionBankRef, logger, database)
	anilistPlatform.SetUsername(test_utils.ConfigData.Provider.AnilistUsername)
	metadataProvider := metadata_provider.GetFakeProvider(t, database)
	completeAnimeCache := anilist.NewCompleteAnimeCache()
	anilistRateLimiter := limiter.NewAnilistLimiter()

	scanLogger, err := NewConsoleScanLogger()
	if err != nil {
		t.Fatal("expected result, got error:", err.Error())
	}

	dir := "E:/Anime"

	t.Log("Initializing MediaFetcher with anime-offline-database...")

	mf, err := NewMediaFetcher(t.Context(), &MediaFetcherOptions{
		Enhanced:                   true,
		EnhanceWithOfflineDatabase: true, // Use offline database
		PlatformRef:                util.NewRef(anilistPlatform),
		LocalFiles:                 []*anime.LocalFile{}, // Empty, we don't need local files for fetching
		CompleteAnimeCache:         completeAnimeCache,
		MetadataProviderRef:        util.NewRef(metadataProvider),
		Logger:                     logger,
		AnilistRateLimiter:         anilistRateLimiter,
		ScanLogger:                 scanLogger,
		DisableAnimeCollection:     true, // Only use offline database
	})
	if err != nil {
		t.Fatal("Failed to create MediaFetcher:", err.Error())
	}

	t.Logf("MediaFetcher initialized with %d media entries", len(mf.AllMedia))

	mc := NewMediaContainer(&MediaContainerOptions{
		AllMedia:   mf.AllMedia,
		ScanLogger: scanLogger,
	})

	tests := []struct {
		name            string
		paths           []string
		expectedMediaId int
	}{
		{
			name: "Death Note - 1535",
			paths: []string{
				"E:/Anime/Death Note/[SubsPlease] Death Note - 01 (1080p).mkv",
				"E:/Anime/Death Note/[SubsPlease] Death Note - 02 (1080p).mkv",
			},
			expectedMediaId: 1535,
		},
		{
			name: "Fullmetal Alchemist Brotherhood - 5114",
			paths: []string{
				"E:/Anime/Fullmetal Alchemist Brotherhood/[HorribleSubs] Fullmetal Alchemist Brotherhood - 01 [1080p].mkv",
			},
			expectedMediaId: 5114,
		},
		{
			name: "Attack on Titan S1 - 16498",
			paths: []string{
				"E:/Anime/Attack on Titan/Shingeki.no.Kyojin.S01E01.1080p.BluRay.x264.mkv",
			},
			expectedMediaId: 16498,
		},
		{
			name: "Demon Slayer S1 - 101922",
			paths: []string{
				"E:/Anime/Kimetsu no Yaiba/[SubsPlease] Kimetsu no Yaiba - 01 (1080p).mkv",
			},
			expectedMediaId: 101922,
		},
		{
			name: "Jujutsu Kaisen S1 - 113415",
			paths: []string{
				"E:/Anime/Jujutsu Kaisen/[SubsPlease] Jujutsu Kaisen - 01 (1080p).mkv",
			},
			expectedMediaId: 113415,
		},
		{
			name: "Spy x Family S1 - 140960",
			paths: []string{
				"E:/Anime/Spy x Family/[SubsPlease] Spy x Family - 01 (1080p).mkv",
			},
			expectedMediaId: 140960,
		},
		{
			name: "One Punch Man S1 - 21087",
			paths: []string{
				"E:/Anime/One Punch Man/[HorribleSubs] One Punch Man - 01 [1080p].mkv",
			},
			expectedMediaId: 21087,
		},
		{
			name: "My Hero Academia S1 - 21459",
			paths: []string{
				"E:/Anime/Boku no Hero Academia/[SubsPlease] Boku no Hero Academia - 01 (1080p).mkv",
			},
			expectedMediaId: 21459,
		},
		{
			name: "Spirited Away - 199",
			paths: []string{
				"E:/Anime/Spirited Away/Spirited.Away.2001.1080p.BluRay.x264.mkv",
			},
			expectedMediaId: 199,
		},
		{
			name: "Your Name - 21519",
			paths: []string{
				"E:/Anime/Your Name/Kimi.no.Na.wa.2016.1080p.BluRay.x264.mkv",
			},
			expectedMediaId: 21519,
		},
		{
			name: "Steins Gate - 9253",
			paths: []string{
				"E:/Anime/Steins Gate/Steins.Gate.S01E01.1080p.BluRay.x264.mkv",
			},
			expectedMediaId: 9253,
		},
		{
			name: "Re Zero S1 - 21355",
			paths: []string{
				"E:/Anime/Re Zero/[SubsPlease] Re Zero kara Hajimeru Isekai Seikatsu - 01 (1080p).mkv",
			},
			expectedMediaId: 21355,
		},
		{
			name: "Mob Psycho 100 S1 - 21507",
			paths: []string{
				"E:/Anime/Mob Psycho 100/[HorribleSubs] Mob Psycho 100 - 01 [1080p].mkv",
			},
			expectedMediaId: 21507,
		},
		{
			name: "Chainsaw Man - 127230",
			paths: []string{
				"E:/Anime/Chainsaw Man/[SubsPlease] Chainsaw Man - 01 (1080p).mkv",
			},
			expectedMediaId: 127230,
		},
		{
			name: "KonoSuba S1 - 21202",
			paths: []string{
				"E:/Anime/KonoSuba/[HorribleSubs] Kono Subarashii Sekai ni Shukufuku wo! - 01 [1080p].mkv",
			},
			expectedMediaId: 21202,
		},
		{
			name: "FMAB alternate name - 5114",
			paths: []string{
				"E:/Anime/FMAB/FMAB.S01E01.1080p.BluRay.x264.mkv",
			},
			expectedMediaId: 5114,
		},
		{
			name: "Kekkai Sensen - 20727",
			paths: []string{
				"E:/Anime/[Anime Time] Kekkai Sensen (Blood Blockade Battlefront) S01+02+OVA+Extra [Dual Audio][BD][1080p][HEVC 10bit x265][AAC][Eng Sub]/Blood Blockade Battlefront/NC/Blood Blockade Battlefront - NCED.mkv",
			},
			expectedMediaId: 20727,
		},
		{
			name: "ACCA 13-ku Kansatsu-ka - 21823",
			paths: []string{
				"E:/Anime/ACCA 13-ku Kansatsu-ka/[Judas] ACCA 13-ku Kansatsu-ka (Season 1) [BD 1080p][HEVC x265 10bit][Dual-Audio][Eng-Subs]/Extras/[Judas] ACCA 13-ku Kansatsu-ka - Ending.mkv",
			},
			expectedMediaId: 21823,
		},
		{
			name: "Akebi-chan no Sailor Fuku - 131548",
			paths: []string{
				"E:/Anime/Akebi-chan no Sailor Fuku/[Anime Time] Akebi-chan no Sailor Fuku - 01 [1080p][HEVC 10bit x265][AAC][Multi Sub].mkv",
			},
			expectedMediaId: 131548,
		},
		{
			name: "Pluto - 99088",
			paths: []string{
				"E:/Anime/PLUTO/Pluto S01 1080p Dual Audio WEBRip DD+ x265-EMBER/S01E01-Episode 1 [59596368].mkv",
			},
			expectedMediaId: 99088,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create local files for this test case
			var lfs []*anime.LocalFile
			for _, path := range tt.paths {
				lf := anime.NewLocalFile(path, dir)
				lfs = append(lfs, lf)
			}

			matcher := &Matcher{
				LocalFiles:        lfs,
				MediaContainer:    mc,
				Logger:            logger,
				ScanLogger:        scanLogger,
				ScanSummaryLogger: nil,
				Debug:             true,
			}

			err := matcher.MatchLocalFilesWithMedia()
			if err != nil {
				t.Fatal("Error while matching:", err.Error())
			}

			for _, lf := range lfs {
				if lf.MediaId == tt.expectedMediaId {
					t.Logf("SUCCESS: %s -> media id: %d", lf.Name, lf.MediaId)
				} else if lf.MediaId == 0 {
					t.Errorf("UNMATCHED: %s (expected %d)", lf.Name, tt.expectedMediaId)
				} else {
					t.Errorf("WRONG MATCH: %s -> got %d, expected %d", lf.Name, lf.MediaId, tt.expectedMediaId)
				}
			}
		})
	}
}

func TestIgnoredSynonyms(t *testing.T) {
	tests := []struct {
		name                     string
		candidates               []*anime.NormalizedMedia
		singularizedSynonym      string
		singularizedSynonymForId int
	}{
		{
			name: "Common synonym 'MS' should be kept for shortest title",
			candidates: []*anime.NormalizedMedia{
				{
					ID: 33,
					Title: &anime.NormalizedMediaTitle{
						English: lo.ToPtr("MultiSet Iterator"),
					},
					Synonyms: []*string{lo.ToPtr("MS")},
				},
				{
					ID: 44,
					Title: &anime.NormalizedMediaTitle{
						English: lo.ToPtr("MultiSet Iterator II"),
					},
					Synonyms: []*string{lo.ToPtr("MultiSet 2"), lo.ToPtr("MS")},
				},
				{
					ID: 55,
					Title: &anime.NormalizedMediaTitle{
						English: lo.ToPtr("MultiSet Iterator III"),
					},
					Synonyms: []*string{lo.ToPtr("MultiSet 3"), lo.ToPtr("MS")},
				},
			},
			singularizedSynonym:      "MS",
			singularizedSynonymForId: 33,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := &Matcher{
				MediaContainer: &MediaContainer{
					NormalizedTitlesCache: map[int][]*NormalizedTitle{
						33: {
							{Original: "MultiSet Iterator", Normalized: "MultiSet Iterator", IsMain: true},
							{Original: "MS", Normalized: "MS", IsMain: false},
						},
						44: {
							{Original: "MultiSet Iterator II", Normalized: "MultiSet Iterator II", IsMain: true},
							{Original: "MultiSet 2", Normalized: "MultiSet", IsMain: false},
							{Original: "MS", Normalized: "MS", IsMain: false},
						},
						55: {
							{Original: "MultiSet Iterator III", Normalized: "MultiSet Iterator III", IsMain: true},
							{Original: "MultiSet 3", Normalized: "MultiSet", IsMain: false},
							{Original: "MS", Normalized: "MS", IsMain: false},
						},
					},
				},
			}

			is := matcher.getIgnoredSynonyms(tt.candidates)

			for _, candidate := range tt.candidates {
				ignored, hasEntry := is[candidate.ID]

				if candidate.ID == tt.singularizedSynonymForId {
					// candidate should keep the synonym
					// if it has an entry, it shouldn't contain the singularized synonym
					if hasEntry {
						_, contains := ignored[tt.singularizedSynonym]
						assert.Falsef(t, contains, "Synonym '%s' should not be ignored for media %d (it has the shortest title)", tt.singularizedSynonym, candidate.ID)
					}
				} else {
					// This candidate should have the synonym ignored
					if assert.Truef(t, hasEntry, "Media %d should have ignored synonyms entry", candidate.ID) {
						_, contains := ignored[tt.singularizedSynonym]
						assert.Truef(t, contains, "Synonym '%s' should be ignored for media %d", tt.singularizedSynonym, candidate.ID)
					}
				}
			}
		})
	}
}
