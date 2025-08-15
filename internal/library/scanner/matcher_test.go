package scanner

import (
	"context"
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

// Add more media to this file if needed
// scanner_test_mock_data.json

func TestMatcher_MatchLocalFileWithMedia(t *testing.T) {

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
				AllMedia:   allMedia,
				ScanLogger: scanLogger,
			})

			// +---------------------+
			// |      Matcher        |
			// +---------------------+

			matcher := &Matcher{
				LocalFiles:         lfs,
				MediaContainer:     mc,
				CompleteAnimeCache: nil,
				Logger:             util.NewLogger(),
				ScanLogger:         scanLogger,
				ScanSummaryLogger:  nil,
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

func TestMatcher_MatchLocalFileWithMedia2(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := anilist.NewAnilistClient(test_utils.ConfigData.Provider.AnilistJwt)
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
			name: "Demon Slayer: Kimetsu no Yaiba Entertainment District Arc - 142329",
			paths: []string{
				"E:/Anime/Kimetsu no Yaiba Yuukaku-hen/[Salieri] Demon Slayer - Kimetsu No Yaiba - S3 - Entertainment District - BD (1080P) (HDR) [Dual-Audio]/[Salieri] Demon Slayer S3 - Kimetsu No Yaiba- Entertainment District - 03 (1080P) (HDR) [Dual-Audio].mkv",
			},
			expectedMediaId: 142329,
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
				AllMedia:   allMedia,
				ScanLogger: scanLogger,
			})

			// +---------------------+
			// |      Matcher        |
			// +---------------------+

			matcher := &Matcher{
				LocalFiles:         lfs,
				MediaContainer:     mc,
				CompleteAnimeCache: nil,
				Logger:             util.NewLogger(),
				ScanLogger:         scanLogger,
				ScanSummaryLogger:  nil,
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
