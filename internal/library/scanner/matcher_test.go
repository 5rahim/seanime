package scanner

import (
	"context"
	"github.com/stretchr/testify/assert"
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
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
	allMedia := animeCollection.GetAllAnime()

	dir := "E:/Anime"

	tests := []struct {
		name            string
		paths           []string
		expectedMediaId int
	}{
		{
			name: "should match with media id 21202",
			paths: []string{
				"E:/Anime/Kono Subarashii Sekai ni Shukufuku wo!/Kono Subarashii Sekai ni Shukufuku wo! (01-10) [1080p] (Batch)/[HorribleSubs] Kono Subarashii Sekai ni Shukufuku wo! - 01 [1080p].mkv",
				"E:/Anime/Kono Subarashii Sekai ni Shukufuku wo!/Kono Subarashii Sekai ni Shukufuku wo! (01-10) [1080p] (Batch)/[HorribleSubs] Kono Subarashii Sekai ni Shukufuku wo! - 02 [1080p].mkv",
				"E:/Anime/Kono Subarashii Sekai ni Shukufuku wo!/Kono Subarashii Sekai ni Shukufuku wo! (01-10) [1080p] (Batch)/[HorribleSubs] Kono Subarashii Sekai ni Shukufuku wo! - 03 [1080p].mkv",
			},
			expectedMediaId: 21202, // Kono Subarashii Sekai ni Shukufuku wo!
		},
		{
			name: "should match with media id 21699",
			paths: []string{
				"E:/Anime/Kono Subarashii Sekai ni Shukufuku wo! 2/KonoSuba.God's.Blessing.On.This.Wonderful.World.S02.1080p.BluRay.10-Bit.Dual-Audio.FLAC2.0.x265-YURASUKA/KonoSuba.God's.Blessing.On.This.Wonderful.World.S02E01.1080p.BluRay.10-Bit.Dual-Audio.FLAC2.0.x265-YURASUKA.mkv",
				"E:/Anime/Kono Subarashii Sekai ni Shukufuku wo! 2/KonoSuba.God's.Blessing.On.This.Wonderful.World.S02.1080p.BluRay.10-Bit.Dual-Audio.FLAC2.0.x265-YURASUKA/KonoSuba.God's.Blessing.On.This.Wonderful.World.S02E02.1080p.BluRay.10-Bit.Dual-Audio.FLAC2.0.x265-YURASUKA.mkv",
				"E:/Anime/Kono Subarashii Sekai ni Shukufuku wo! 2/KonoSuba.God's.Blessing.On.This.Wonderful.World.S02.1080p.BluRay.10-Bit.Dual-Audio.FLAC2.0.x265-YURASUKA/KonoSuba.God's.Blessing.On.This.Wonderful.World.S02E03.1080p.BluRay.10-Bit.Dual-Audio.FLAC2.0.x265-YURASUKA.mkv",
			},
			expectedMediaId: 21699, // Kono Subarashii Sekai ni Shukufuku wo! 2
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
