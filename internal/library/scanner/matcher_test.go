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
				anilist.TestAddAnimeCollectionWithRelationsEntry(animeCollection, tt.expectedMediaId, anilist.TestModifyAnimeCollectionEntryInput{Status: new(anilist.MediaListStatusCurrent)}, anilistClient)
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
					anilist.TestAddAnimeCollectionWithRelationsEntry(animeCollection, otherMediaId, anilist.TestModifyAnimeCollectionEntryInput{Status: new(anilist.MediaListStatusCurrent)}, anilistClient)
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
			name: "Frieren - 154587",
			paths: []string{
				"E:/Anime/Frieren/Frieren - 01.mkv",
			},
			expectedMediaId: 154587,
		},
		{
			name: "Mob Psycho 100 S3 - 140439",
			paths: []string{
				"E:/Anime/Mob Psycho 100/[HorribleSubs] Mob Psycho 100 III - 01 [1080p].mkv",
			},
			expectedMediaId: 140439,
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
				"E:/Anime/Danmachi S04P02 1080p Dual Audio BDRip 10 bits DD x265-EMBER/S04E12-Amphisbaena A Song of Despair [080E734C].mkv",
			},
			expectedMediaId: 155211,
			otherMediaIds:   []int{},
		},
		{
			name: "Gintama last movie",
			paths: []string{
				"E:/Anime/Gintama Movie Collection/Gintama The Movie - Capitolo Finale - Tuttofare per sempre (2013).1080p.h265.[EAC3, EAC3].[ITA, JPN].{imdb-tt2374144}.mkv",
			},
			expectedMediaId: 15335,
			otherMediaIds:   []int{},
		},
		{
			name: "Gotoubun no Hanayome＊",
			paths: []string{
				"E:/Anime/Go-toubun no Hanayome/[Shimai] Gotoubun no Hanayome＊ - 01.mkv",
			},
			expectedMediaId: 177191,
			otherMediaIds:   []int{},
		},
		{
			name: "One Punch Man OVA",
			paths: []string{
				"E:/Anime/One Punch Man OVA/[Judas] One Punch Man (Seasons 1-2 + OVAs + Specials) [BD 1080p][HEVC x265 10bit][Dual-Audio][Multi-Subs]/[Judas] One-Punch Man S1/Extras/[Judas] One Punch Man - S01OVA01.mkv",
			},
			expectedMediaId: 21416,
			otherMediaIds:   []int{},
		},
		{
			name: "Revue Starlight Movie",
			paths: []string{
				"E:/Anime/Shoujo Kageki Revue Starlight/[neoHEVC] Revue Starlight [Season 1 + Specials + Movie] [BD 1080p x265 HEVC AAC] [Dual Audio]/Specials/Revue Starlight - S00E05 - (S1M1 - Movie).mkv",
			},
			expectedMediaId: 113024,
			otherMediaIds:   []int{},
		},
		{
			name: "ReZero 3",
			paths: []string{
				"E:/Anime/Re Zero kara Hajimeru Isekai Seikatsu 3rd Season (Batch + OVAs)/ReZero S03 1080p Dual Audio WEBRip DD+ x265-EMBER/S03E01-Theatrical Malice [4AB8AF98].mkv",
			},
			expectedMediaId: 163134,
			otherMediaIds:   []int{},
		},
		{
			name: "ReZero 3_2",
			paths: []string{
				"E:/Anime/Re Zero kara Hajimeru Isekai Seikatsu 3rd Season (Batch + OVAs)/ReZero S03 1080p Dual Audio WEBRip DD+ x265-EMBER/ReZero - S03E01-Theatrical Malice [4AB8AF98].mkv",
			},
			expectedMediaId: 163134,
			otherMediaIds:   []int{},
		},
		// generic root title shouldn't cause mismatches
		{
			name: "Bakemonogatari",
			paths: []string{
				"E:/Anime/Monogatari Series/Bakemonogatari/[Coalgirls]_Bakemonogatari_01_(1920x1080_Blu-ray_FLAC)_[E32CBBD3].mkv",
				"E:/Anime/Monogatari Series/Bakemonogatari/[Commie] Bakemonogatari - 02 [BD 1080p AAC] [1B3D8A29].mkv",
				"E:/Anime/Monogatari Series/Bakemonogatari/Bakemonogatari - 03 (BDRip 1920x1080 x264 FLAC).mkv",
				"E:/Anime/Monogatari Series/Bakemonogatari - 03 (BDRip 1920x1080 x264 FLAC).mkv",
			},
			expectedMediaId: 5081,
			otherMediaIds:   []int{11597, 15689, 17074},
		},
		{
			name: "Nisemonogatari",
			paths: []string{
				"E:/Anime/Monogatari Series/Nisemonogatari/[MTBB] Nisemonogatari - 01 (1080p BD) [5C0A0065].mkv",
				"E:/Anime/Monogatari Series/Nisemonogatari/[Commie] Nisemonogatari - 02 [DB005E82].mkv",
				"E:/Anime/Monogatari Series/Nisemonogatari/[Coalgirls]_Nisemonogatari_03_(1920x1080_Blu-ray_FLAC).mkv",
			},
			expectedMediaId: 11597,
			otherMediaIds:   []int{5081, 15689, 17074},
		},
		{
			name: "Nekomonogatari Kuro",
			paths: []string{
				"E:/Anime/Monogatari Series/Nekomonogatari (Kuro)/[Commie] Nekomonogatari (Kuro) - 01 [B085F6CB].mkv",
				"E:/Anime/Monogatari Series/Nekomonogatari (Kuro)/[Coalgirls]_Nekomonogatari_(Kuro)_02_(1920x1080_Blu-ray_FLAC)_[A38F72D0].mkv",
				"E:/Anime/Monogatari Series/Nekomonogatari (Kuro)/[MTBB] Nekomonogatari: Kuro - 03 (1080p BD) [5C0A0065].mkv",
			},
			expectedMediaId: 15689,
			otherMediaIds:   []int{5081, 11597, 17074},
		},
		{
			name: "Monogatari Series: Second Season",
			paths: []string{
				"E:/Anime/Monogatari Series/Monogatari Series Second Season/[Edo] Monogatari Series: Second Season - 01 (1080p, AAC) [54F919E3].mkv",
				"E:/Anime/Monogatari Series/Monogatari Series Second Season/[Coalgirls]_Monogatari_Series_Second_Season_02_(1920x1080_Blu-ray_FLAC).mkv",
				"E:/Anime/Monogatari Series/Monogatari Series Second Season/[Commie] Monogatari Series Second Season - 03 [1080p][33BBD87B].mkv",
			},
			expectedMediaId: 17074,
			otherMediaIds:   []int{5081, 11597, 15689},
		},
		{
			name: "Hanamonogatari",
			paths: []string{
				"E:/Anime/Monogatari Series/Hanamonogatari/[Commie] Hanamonogatari - 01 [1080p][33BBD87B].mkv",
				"E:/Anime/Monogatari Series/Hanamonogatari/[MTBB] Hanamonogatari - 02 (1080p BD) [69C45B90].mkv",
				"E:/Anime/Monogatari Series/Hanamonogatari/[Coalgirls]_Hanamonogatari_03_(1920x1080_Blu-Ray_FLAC)_[C1D38A94].mkv",
			},
			expectedMediaId: 20593,
			otherMediaIds:   []int{17074},
		},
		{
			name: "Tsukimonogatari",
			paths: []string{
				"E:/Anime/Monogatari Series/Tsukimonogatari/[Commie] Tsukimonogatari - 01 [1080p][33BBD87B].mkv",
				"E:/Anime/Monogatari Series/Tsukimonogatari/[MTBB] Tsukimonogatari - 02 (1080p BD) [69C45B90].mkv",
				"E:/Anime/Monogatari Series/Tsukimonogatari/[Coalgirls]_Tsukimonogatari_03_(1920x1080_Blu-Ray_FLAC)_[C1D38A94].mkv",
			},
			expectedMediaId: 20918,
			otherMediaIds:   []int{17074, 21262},
		},
		{
			name: "Owarimonogatari",
			paths: []string{
				"E:/Anime/Monogatari Series/Owarimonogatari/[Commie] Owarimonogatari - 01 [1080p][F3E836D4].mkv",
				"E:/Anime/Monogatari Series/Owarimonogatari/[MTBB] Owarimonogatari - 02 (1080p BD).mkv",
				"E:/Anime/Monogatari Series/Owarimonogatari/[Coalgirls]_Owarimonogatari_03_(1920x1080_Blu-Ray_FLAC).mkv",
			},
			expectedMediaId: 21262,
			otherMediaIds:   []int{21746, 17074, 100815},
		},
		{
			name: "Koyomimonogatari",
			paths: []string{
				"E:/Anime/Monogatari Series/Koyomimonogatari/[Commie] Koyomimonogatari - 01 [1080p].mkv",
				"E:/Anime/Monogatari Series/Koyomimonogatari/[MTBB] Koyomimonogatari - 02 (1080p BD).mkv",
				"E:/Anime/Monogatari Series/Koyomimonogatari/[Edo] Koyomimonogatari - 03 (1080p, AAC).mkv",
			},
			expectedMediaId: 21520,
			otherMediaIds:   []int{17074, 21262},
		},
		{
			name: "Kizumonogatari I: Tekketsu-hen",
			paths: []string{
				"E:/Anime/Monogatari Series/Kizumonogatari/[Commie] Kizumonogatari I - Tekketsu-hen [1080p].mkv",
				"E:/Anime/Monogatari Series/Kizumonogatari/Kizumonogatari Part 1 Tekketsu [1080p].mkv",
				"E:/Anime/Monogatari Series/Kizumonogatari/[Coalgirls] Kizumonogatari I - Tekketsu-hen (1920x1080 Blu-Ray FLAC).mkv",
			},
			expectedMediaId: 9260,
			otherMediaIds:   []int{21399, 21400, 5081},
		},
		{
			name: "Kizumonogatari II: Nekketsu-hen",
			paths: []string{
				"E:/Anime/Monogatari Series/Kizumonogatari/[Commie] Kizumonogatari II - Nekketsu-hen [1080p].mkv",
				//"E:/Anime/Monogatari Series/Kizumonogatari/Kizumonogatari Part 2 Nekketsu [1080p].mkv",
				"E:/Anime/Monogatari Series/Kizumonogatari/[Coalgirls] Kizumonogatari II - Nekketsu-hen (1920x1080 Blu-Ray FLAC).mkv",
			},
			expectedMediaId: 21399,
			otherMediaIds:   []int{9260, 21400, 5081},
		},
		{
			name: "Kizumonogatari III: Reiketsu-hen",
			paths: []string{
				"E:/Anime/Monogatari Series/Kizumonogatari/[Commie] Kizumonogatari III - Reiketsu-hen [1080p].mkv",
				//"E:/Anime/Monogatari Series/Kizumonogatari/Kizumonogatari Part 3 Reiketsu [1080p].mkv",
				"E:/Anime/Monogatari Series/Kizumonogatari/[Coalgirls] Kizumonogatari III - Reiketsu-hen (1920x1080 Blu-Ray FLAC).mkv",
			},
			expectedMediaId: 21400,
			otherMediaIds:   []int{9260, 21399, 5081},
		},
		{
			name: "Owarimonogatari 2nd Season",
			paths: []string{
				"E:/Anime/Monogatari Series/Owarimonogatari 2nd Season/[Commie] Owarimonogatari Second Season - 01 [1080p].mkv",
				"E:/Anime/Monogatari Series/Owarimonogatari 2nd Season/[MTBB] Owarimonogatari 2nd Season - 02 (1080p BD).mkv",
				"E:/Anime/Monogatari Series/Owarimonogatari 2nd Season/[SomeSubs] Owarimonogatari S2 - 03 [1080p BD].mkv",
			},
			expectedMediaId: 21745,
			otherMediaIds:   []int{21262, 17074, 100815},
		},
		{
			name: "Zoku Owarimonogatari",
			paths: []string{
				"E:/Anime/Monogatari Series/Zoku Owarimonogatari/[Erai-raws] Zoku Owarimonogatari - 01 [1080p][Multiple Subtitle].mkv",
				"E:/Anime/Monogatari Series/Zoku Owarimonogatari/[MTBB] Zoku Owarimonogatari - 02 (1080p BD).mkv",
				"E:/Anime/Monogatari Series/Zoku Owarimonogatari/[Commie] Zoku Owarimonogatari - 03 [1080p].mkv",
			},
			expectedMediaId: 100815,
			otherMediaIds:   []int{21746, 21262, 17074},
		},
		{
			name: "Monogatari Series: Off & Monster Season",
			paths: []string{
				"E:/Anime/Monogatari Series/Monogatari Series Off & Monster Season/[Erai-raws] Monogatari Series - Off & Monster Season - 01 [1080p][Multiple Subtitle][64667E46].mkv",
				"E:/Anime/Monogatari Series/Monogatari Series Off & Monster Season/[SubsPlease] Monogatari Series - Off and Monster Season - 02 (1080p) [2B23C6D3].mkv",
				"E:/Anime/Monogatari Series/Monogatari Series Off & Monster Season/[Kawaiika-Raws] Monogatari Series Off & Monster Season - 03 [WEB 1080p x265 E-AC-3][Dual-Audio].mkv",
			},
			expectedMediaId: 173533,
			otherMediaIds:   []int{17074, 5081, 100815},
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
				anilist.TestAddAnimeCollectionWithRelationsEntry(animeCollection, tt.expectedMediaId, anilist.TestModifyAnimeCollectionEntryInput{Status: new(anilist.MediaListStatusCurrent)}, anilistClient)
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
					anilist.TestAddAnimeCollectionWithRelationsEntry(animeCollection, id, anilist.TestModifyAnimeCollectionEntryInput{Status: new(anilist.MediaListStatusCurrent)}, anilistClient)
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
						t.Errorf("FAILED: expected media id %d, got %d for file %s", tt.expectedMediaId, lf.MediaId, lf.Name)
					} else {
						t.Logf("SUCCESS: local file: %s -> media id: %d", lf.Name, lf.MediaId)
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
			name: "Mob Psycho 100 S3 - 140439",
			paths: []string{
				"E:/Anime/Mob Psycho 100/[HorribleSubs] Mob Psycho 100 III - 01 [1080p].mkv",
			},
			expectedMediaId: 140439,
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
						English: new("MultiSet Iterator"),
					},
					Synonyms: []*string{new("MS")},
				},
				{
					ID: 44,
					Title: &anime.NormalizedMediaTitle{
						English: new("MultiSet Iterator II"),
					},
					Synonyms: []*string{new("MultiSet 2"), new("MS")},
				},
				{
					ID: 55,
					Title: &anime.NormalizedMediaTitle{
						English: new("MultiSet Iterator III"),
					},
					Synonyms: []*string{new("MultiSet 3"), new("MS")},
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

func TestGetFileFormatType(t *testing.T) {
	dir := "E:/Anime"
	tests := []struct {
		name     string
		path     string
		expected fileFormatType
	}{
		// filename-level detection
		{
			name:     "OVA in filename: S01OVA01",
			path:     "E:/Anime/One Punch Man/[Judas] One Punch Man - S01OVA01.mkv",
			expected: fileFormatOVA,
		},
		{
			name:     "OVA in filename: standalone OVA keyword",
			path:     "E:/Anime/Title/Title OVA 01.mkv",
			expected: fileFormatOVA,
		},
		{
			name:     "OAD in filename",
			path:     "E:/Anime/Title/Title OAD 01.mkv",
			expected: fileFormatOVA,
		},
		{
			name:     "SP in filename",
			path:     "E:/Anime/Title/[Sub] Title SP01.mkv",
			expected: fileFormatSpecial,
		},
		{
			name:     "Movie in filename",
			path:     "E:/Anime/Title/Title Movie 1080p.mkv",
			expected: fileFormatMovie,
		},
		{
			name:     "NCED in filename",
			path:     "E:/Anime/Title/Title - NCED.mkv",
			expected: fileFormatNC,
		},
		{
			name:     "NCOP in filename",
			path:     "E:/Anime/Title/Title - NCOP 01.mkv",
			expected: fileFormatNC,
		},
		// regular episodes (no format detected)
		{
			name:     "Regular episode, no format keyword",
			path:     "E:/Anime/One Punch Man/One Punch Man - 01.mkv",
			expected: fileFormatUnknown,
		},
		{
			name:     "Regular episode with S01E01 format",
			path:     "E:/Anime/One Punch Man/One.Punch.Man.S01E01.1080p.mkv",
			expected: fileFormatUnknown,
		},
		// folder names should not trigger format detection
		{
			name:     "OVA folder name only, no OVA in filename (folder not used)",
			path:     "E:/Anime/One Punch Man OVA/One Punch Man - 01.mkv",
			expected: fileFormatUnknown,
		},
		{
			name:     "Batch folder: Seasons 1-2 + OVAs + Specials (regular ep inside)",
			path:     "E:/Anime/[Judas] One Punch Man (Seasons 1-2 + OVAs + Specials) [BD 1080p]/[Judas] One-Punch Man S1/One Punch Man - 01.mkv",
			expected: fileFormatUnknown,
		},
		{
			name:     "Batch folder: Series (+OVA) with regular episode",
			path:     "E:/Anime/One Punch Man Series (+OVA)/One Punch Man/One Punch Man - 01.mkv",
			expected: fileFormatUnknown,
		},
		{
			name:     "Batch folder: Complete + OVA collection with regular episode",
			path:     "E:/Anime/Title Complete (TV + OVA + Specials)/Season 1/Title - 01.mkv",
			expected: fileFormatUnknown,
		},
		// filename still wins regardless of folder
		{
			name:     "OVA in filename wins regardless of folder context",
			path:     "E:/Anime/One Punch Man Series (+OVA)/Extras/[Judas] One Punch Man - S01OVA01.mkv",
			expected: fileFormatOVA,
		},
		{
			name:     "OVA folder with OVA in filename",
			path:     "E:/Anime/One Punch Man OVA/One Punch Man OVA - 01.mkv",
			expected: fileFormatOVA,
		},
		{
			name:     "SP in filename regardless of folder",
			path:     "E:/Anime/Title Specials/Title - SP01.mkv",
			expected: fileFormatSpecial,
		},
		// extras folder
		{
			name:     "Extras folder with NCED (NC takes priority)",
			path:     "E:/Anime/Title/Extras/NCED 01.mkv",
			expected: fileFormatNC,
		},
		{
			name:     "Extras folder with non-NC file (fallback to Special)",
			path:     "E:/Anime/Title/Extras/Extra Episode 01.mkv",
			expected: fileFormatSpecial,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lf := anime.NewLocalFile(tt.path, dir)
			result := getFileFormatType(lf)
			if result != tt.expected {
				formatNames := map[fileFormatType]string{
					fileFormatUnknown: "Unknown",
					fileFormatOVA:     "OVA",
					fileFormatSpecial: "Special",
					fileFormatMovie:   "Movie",
					fileFormatNC:      "NC",
				}
				t.Errorf("expected %s, got %s for path: %s", formatNames[tt.expected], formatNames[result], tt.path)
			}
		})
	}
}

func TestMatcher_applyMatchingRule(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := anilist.NewAnilistClient(test_utils.ConfigData.Provider.AnilistJwt, "")
	animeCollection, err := anilistClient.AnimeCollectionWithRelations(context.Background(), &test_utils.ConfigData.Provider.AnilistUsername)
	if err != nil {
		t.Fatal(err.Error())
	}

	dir := "E:/Anime"

	tests := []struct {
		name             string
		paths            []string
		rules            []*MatchingRule
		expectedResults  map[string]int
		expectedMediaIds []int
	}{
		{
			name: "One rule",
			paths: []string{
				"E:/Anime/Some Folder/Mob Psycho - S01E05 - Test5.mkv",
				"E:/Anime/Some Folder/Mob Psycho - S01E06 - Test6.mkv",
				"E:/Anime/Some Folder/TestEpisode - S01E07 - Test7.mkv",
			},
			rules: []*MatchingRule{
				{Pattern: ".*Some Folder.*", MediaID: 21507},
			},
			expectedResults: map[string]int{
				"Mob Psycho - S01E05 - Test5.mkv":  21507,
				"Mob Psycho - S01E06 - Test6.mkv":  21507,
				"TestEpisode - S01E07 - Test7.mkv": 21507,
			},
			expectedMediaIds: []int{21507},
		},

		{
			name: "Multiple rules",
			paths: []string{
				"E:/Anime/One Piece/One Piece - E100.mkv",
				"E:/Anime/Mob Psycho/Mob Psycho - E01.mkv",
				"E:/Anime/Naruto/Naruto Shippuden - E05.mkv",
			},
			rules: []*MatchingRule{
				{Pattern: ".*One Piece.*", MediaID: 21},
				{Pattern: ".*Mob Psycho.*", MediaID: 21507},
				{Pattern: ".*Naruto.*", MediaID: 20},
			},
			expectedResults: map[string]int{
				"One Piece - E100.mkv":       21,
				"Mob Psycho - E01.mkv":       21507,
				"Naruto Shippuden - E05.mkv": 20,
			},
			expectedMediaIds: []int{21, 21507, 20},
		},

		{
			name: "Case insensitive rule",
			paths: []string{
				"E:/Anime/some folder/MOB PSYCHO - E01.mkv",
				"E:/Anime/Some Folder/mob psycho 100 - E02.mkv",
				"E:/Anime/SOME FOLDER/Mob Psycho - E03.mkv",
			},
			rules: []*MatchingRule{
				{Pattern: "(?i).*Some Folder.*", MediaID: 21507},
			},
			expectedResults: map[string]int{
				"MOB PSYCHO - E01.mkv":     21507,
				"mob psycho 100 - E02.mkv": 21507,
				"Mob Psycho - E03.mkv":     21507,
			},
			expectedMediaIds: []int{21507},
		},

		{
			name: "Rule for some files only",
			paths: []string{
				"E:/Anime/Test/One Piece - E01.mkv",
				"E:/Anime/Test/Mob Psycho - S02E05.mkv",
			},
			rules: []*MatchingRule{
				{Pattern: ".*Mob Psycho.*", MediaID: 21507},
			},
			expectedResults: map[string]int{
				"One Piece - E01.mkv":     21,
				"Mob Psycho - S02E05.mkv": 21507,
			},
			expectedMediaIds: []int{21507, 21}, // медиа всё равно должны быть в коллекции
		},
		{
			name: "Special characters in filename",
			paths: []string{
				"E:/Anime/Test Folder/Attack on Titan [1080p] - E01.mkv",
				"E:/Anime/Test Folder/Attack on Titan (2013) - E02.mkv",
				"E:/Anime/Test Folder/Attack on Titan [Final Season] - E03.mkv",
				"E:/Anime/Test Folder/Attack on Titan Final_Season_Part_2 - [04] [1080p].mkv",
				"E:/Anime/Test Folder/Attack.on.Titan.Final.Season.Part.2.05.mkv",
			},
			rules: []*MatchingRule{
				{Pattern: ".*Attack on Titan \\[1080p\\].*", MediaID: 16498},
				{Pattern: ".*Attack on Titan \\[2013\\].*", MediaID: 16498},
				{Pattern: ".*Attack on Titan \\[Final Season\\].*", MediaID: 110277},
				{Pattern: ".*Final_Season_Part_2.*", MediaID: 131681},
				{Pattern: ".*Attack\\.on\\.Titan\\.Final\\.Season\\.Part\\.2.*", MediaID: 131681},
			},
			expectedResults: map[string]int{
				"Attack on Titan [1080p] - E01.mkv":                      16498,
				"Attack on Titan (2013) - E02.mkv":                       16498,
				"Attack on Titan [Final Season] - E03.mkv":               110277,
				"Attack on Titan Final_Season_Part_2 - [04] [1080p].mkv": 131681,
				"Attack.on.Titan.Final.Season.Part.2.05.mkv":             131681,
			},
			expectedMediaIds: []int{16498, 110277, 131681},
		},
		{
			name: "Rules with Unicode characters",
			paths: []string{
				"E:/Anime/(アニメ) さらい屋五葉 第01話 「形ばかりの」(CX 1440x1080 x264-aac).mp4",
				"E:/Anime/鬼滅の刃/鬼滅の刃 - E02.mkv",
			},
			rules: []*MatchingRule{
				{Pattern: ".*さらい屋五葉.*", MediaID: 7588},
				{Pattern: ".*鬼滅の刃.*", MediaID: 101922},
			},
			expectedResults: map[string]int{
				"(アニメ) さらい屋五葉 第01話 「形ばかりの」(CX 1440x1080 x264-aac).mp4": 7588,
				"鬼滅の刃 - E02.mkv": 101922,
			},
			expectedMediaIds: []int{101922, 7588},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			// Add medias to collection if it doesn't exist
			allMedia := animeCollection.GetAllAnime()
			expectedIDs := make([]int, len(tt.expectedMediaIds))
			copy(expectedIDs, tt.expectedMediaIds)

			for _, media := range allMedia {
				for i, expectedID := range expectedIDs {
					if media.ID == expectedID {
						last := len(expectedIDs) - 1
						expectedIDs[i] = expectedIDs[last]
						expectedIDs = expectedIDs[:last]
						break
					}
				}
			}

			for _, missingID := range expectedIDs {
				anilist.TestAddAnimeCollectionWithRelationsEntry(
					animeCollection,
					missingID,
					anilist.TestModifyAnimeCollectionEntryInput{
						Status: new(anilist.MediaListStatusCurrent),
					},
					anilistClient,
				)
			}

			allMedia = animeCollection.GetAllAnime()

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

			config := &Config{
				Matching: MatchingConfig{
					Rules: tt.rules,
				},
			}
			matcher := &Matcher{
				LocalFiles:        lfs,
				MediaContainer:    mc,
				Logger:            util.NewLogger(),
				ScanLogger:        scanLogger,
				ScanSummaryLogger: nil,
				Debug:             true,
				Config:            config,
			}

			err = matcher.MatchLocalFilesWithMedia()

			for _, lf := range lfs {
				expectedID, exists := tt.expectedResults[lf.Name]
				if !exists {
					t.Errorf("Unexpected file in results: %s", lf.Name)
					continue
				}
				assert.Equal(t, expectedID, lf.MediaId,
					"File %q: expected media ID %d, got %d", lf.Name, expectedID, lf.MediaId)

				if expectedID != 0 {
					t.Logf("✓ %q → MediaID: %d", lf.Name, lf.MediaId)
				} else {
					t.Logf("✓ %q → unmatched (as expected)", lf.Name)
				}
			}
		})
	}

}
