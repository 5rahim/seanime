package anime_test

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"seanime/internal/library/anime"
	"seanime/internal/util"
	"strings"
	"testing"
)

func TestLocalFile_GetNormalizedPath(t *testing.T) {

	tests := []struct {
		filePath       string
		libraryPath    string
		expectedResult string
	}{
		{
			filePath:       "E:\\Anime\\Bungou Stray Dogs 5th Season\\Bungou Stray Dogs\\[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			libraryPath:    "E:/ANIME",
			expectedResult: "e:/anime/bungou stray dogs 5th season/bungou stray dogs/[subsplease] bungou stray dogs - 61 (1080p) [f609b947].mkv",
		},
		{
			filePath:       "E:\\Anime\\Shakugan No Shana\\Shakugan No Shana I\\Opening\\OP01.mkv",
			libraryPath:    "E:/ANIME",
			expectedResult: "e:/anime/shakugan no shana/shakugan no shana i/opening/op01.mkv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			lf := anime.NewLocalFile(tt.filePath, tt.libraryPath)

			if assert.NotNil(t, lf) {

				if assert.Equal(t, tt.expectedResult, lf.GetNormalizedPath()) {
					spew.Dump(lf.GetNormalizedPath())
				}
			}

		})
	}

}

func TestLocalFile_IsInDir(t *testing.T) {

	tests := []struct {
		filePath       string
		libraryPath    string
		dir            string
		expectedResult bool
	}{
		{
			filePath:       "E:\\Anime\\Bungou Stray Dogs 5th Season\\Bungou Stray Dogs\\[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			libraryPath:    "E:/ANIME",
			dir:            "E:/ANIME/Bungou Stray Dogs 5th Season",
			expectedResult: true,
		},
		{
			filePath:       "E:\\Anime\\Shakugan No Shana\\Shakugan No Shana I\\Opening\\OP01.mkv",
			libraryPath:    "E:/ANIME",
			dir:            "E:/ANIME/Shakugan No Shana",
			expectedResult: true,
		},
		{
			filePath:       "E:\\Anime\\Shakugan No Shana\\Shakugan No Shana I\\Opening\\OP01.mkv",
			libraryPath:    "E:/ANIME",
			dir:            "E:/ANIME/Shakugan No Shana I",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			lf := anime.NewLocalFile(tt.filePath, tt.libraryPath)

			if assert.NotNil(t, lf) {

				if assert.Equal(t, tt.expectedResult, lf.IsInDir(tt.dir)) {
					spew.Dump(lf.IsInDir(tt.dir))
				}
			}

		})
	}

}

func TestLocalFile_IsAtRootOf(t *testing.T) {

	tests := []struct {
		filePath       string
		libraryPath    string
		dir            string
		expectedResult bool
	}{
		{
			filePath:       "E:\\Anime\\Bungou Stray Dogs 5th Season\\Bungou Stray Dogs\\[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			libraryPath:    "E:/ANIME",
			dir:            "E:/ANIME/Bungou Stray Dogs 5th Season",
			expectedResult: false,
		},
		{
			filePath:       "E:\\Anime\\Shakugan No Shana\\Shakugan No Shana I\\Opening\\OP01.mkv",
			libraryPath:    "E:/ANIME",
			dir:            "E:/ANIME/Shakugan No Shana",
			expectedResult: false,
		},
		{
			filePath:       "E:\\Anime\\Shakugan No Shana\\Shakugan No Shana I\\Opening\\OP01.mkv",
			libraryPath:    "E:/ANIME",
			dir:            "E:/ANIME/Shakugan No Shana/Shakugan No Shana I/Opening",
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			lf := anime.NewLocalFile(tt.filePath, tt.libraryPath)

			if assert.NotNil(t, lf) {

				if !assert.Equal(t, tt.expectedResult, lf.IsAtRootOf(tt.dir)) {
					t.Log(filepath.Dir(lf.GetNormalizedPath()))
					t.Log(strings.TrimSuffix(util.NormalizePath(tt.dir), "/"))
				}
			}

		})

	}

}

func TestLocalFile_Equals(t *testing.T) {

	tests := []struct {
		filePath1      string
		filePath2      string
		libraryPath    string
		expectedResult bool
	}{
		{
			filePath1:      "E:\\Anime\\Bungou Stray Dogs 5th Season\\Bungou Stray Dogs\\[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			filePath2:      "E:/ANIME/Bungou Stray Dogs 5th Season/Bungou Stray Dogs/[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			libraryPath:    "E:/Anime",
			expectedResult: true,
		},
		{
			filePath1:      "E:\\Anime\\Bungou Stray Dogs 5th Season\\Bungou Stray Dogs\\[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			filePath2:      "E:\\Anime\\Bungou Stray Dogs 5th Season\\Bungou Stray Dogs\\[SubsPlease] Bungou Stray Dogs - 62 (1080p) [F609B947].mkv",
			libraryPath:    "E:/ANIME",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.filePath1, func(t *testing.T) {
			lf1 := anime.NewLocalFile(tt.filePath1, tt.libraryPath)
			lf2 := anime.NewLocalFile(tt.filePath2, tt.libraryPath)

			if assert.NotNil(t, lf1) && assert.NotNil(t, lf2) {
				assert.Equal(t, tt.expectedResult, lf1.Equals(lf2))
			}

		})

	}

}

func TestLocalFile_GetTitleVariations(t *testing.T) {

	tests := []struct {
		filePath       string
		libraryPath    string
		expectedTitles []string
	}{
		{
			filePath:    "E:\\Anime\\Bungou Stray Dogs 5th Season\\Bungou Stray Dogs\\[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			libraryPath: "E:/ANIME",
			expectedTitles: []string{
				"Bungou Stray Dogs 5th Season",
				"Bungou Stray Dogs Season 5",
				"Bungou Stray Dogs S5",
			},
		},
		{
			filePath:    "E:\\Anime\\Shakugan No Shana\\Shakugan No Shana I\\Opening\\OP01.mkv",
			libraryPath: "E:/ANIME",
			expectedTitles: []string{
				"Shakugan No Shana I",
			},
		},
		{
			filePath:    "E:\\ANIME\\Neon Genesis Evangelion Death & Rebirth\\[Anime Time] Neon Genesis Evangelion - Rebirth.mkv",
			libraryPath: "E:/ANIME",
			expectedTitles: []string{
				"Neon Genesis Evangelion - Rebirth",
				"Neon Genesis Evangelion Death & Rebirth",
			},
		},
		{
			filePath:    "E:\\ANIME\\Omoi, Omoware, Furi, Furare\\[GJM] Love Me, Love Me Not (BD 1080p) [841C23CD].mkv",
			libraryPath: "E:/ANIME",
			expectedTitles: []string{
				"Love Me, Love Me Not",
				"Omoi, Omoware, Furi, Furare",
			},
		},
		{
			filePath:    "E:\\ANIME\\Violet Evergarden Gaiden Eien to Jidou Shuki Ningyou\\Violet.Evergarden.Gaiden.2019.1080..Dual.Audio.BDRip.10.bits.DD.x265-EMBER.mkv",
			libraryPath: "E:/ANIME",
			expectedTitles: []string{
				"Violet Evergarden Gaiden Eien to Jidou Shuki Ningyou",
				"Violet Evergarden Gaiden 2019",
			},
		},
		{
			filePath:    "E:\\ANIME\\Violet Evergarden S01+Movies+OVA 1080p Dual Audio BDRip 10 bits DD x265-EMBER\\01. Season 1 + OVA\\S01E01-'I Love You' and Auto Memory Dolls [F03E1F7A].mkv",
			libraryPath: "E:/ANIME",
			expectedTitles: []string{
				"Violet Evergarden",
				"Violet Evergarden S1",
				"Violet Evergarden Season 1",
				"Violet Evergarden 1st Season",
			},
		},
		{
			filePath:    "E:\\ANIME\\Golden Kamuy 4th Season\\[Judas] Golden Kamuy (Season 4) [1080p][HEVC x265 10bit][Multi-Subs]\\[Judas] Golden Kamuy - S04E01.mkv",
			libraryPath: "E:/ANIME",
			expectedTitles: []string{
				"Golden Kamuy S4",
				"Golden Kamuy Season 4",
				"Golden Kamuy 4th Season",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			lf := anime.NewLocalFile(tt.filePath, tt.libraryPath)

			if assert.NotNil(t, lf) {
				tv := lo.Map(lf.GetTitleVariations(), func(item *string, _ int) string { return *item })

				if assert.ElementsMatch(t, tt.expectedTitles, tv) {
					spew.Dump(lf.GetTitleVariations())
				}
			}

		})
	}

}

func TestLocalFile_GetParsedTitle(t *testing.T) {

	tests := []struct {
		filePath            string
		libraryPath         string
		expectedParsedTitle string
	}{
		{
			filePath:            "E:\\Anime\\Bungou Stray Dogs 5th Season\\Bungou Stray Dogs\\[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			libraryPath:         "E:/ANIME",
			expectedParsedTitle: "Bungou Stray Dogs",
		},
		{
			filePath:            "E:\\Anime\\Shakugan No Shana\\Shakugan No Shana I\\Opening\\OP01.mkv",
			libraryPath:         "E:/ANIME",
			expectedParsedTitle: "Shakugan No Shana I",
		},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			lf := anime.NewLocalFile(tt.filePath, tt.libraryPath)

			if assert.NotNil(t, lf) {

				if assert.Equal(t, tt.expectedParsedTitle, lf.GetParsedTitle()) {
					spew.Dump(lf.GetParsedTitle())
				}
			}

		})
	}

}

func TestLocalFile_GetFolderTitle(t *testing.T) {

	tests := []struct {
		filePath            string
		libraryPath         string
		expectedFolderTitle string
	}{
		{
			filePath:            "E:\\Anime\\Bungou Stray Dogs 5th Season\\S05E11 - Episode Title.mkv",
			libraryPath:         "E:/ANIME",
			expectedFolderTitle: "Bungou Stray Dogs",
		},
		{
			filePath:            "E:\\Anime\\Shakugan No Shana\\Shakugan No Shana I\\Opening\\OP01.mkv",
			libraryPath:         "E:/ANIME",
			expectedFolderTitle: "Shakugan No Shana I",
		},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			lf := anime.NewLocalFile(tt.filePath, tt.libraryPath)

			if assert.NotNil(t, lf) {

				if assert.Equal(t, tt.expectedFolderTitle, lf.GetFolderTitle()) {
					spew.Dump(lf.GetFolderTitle())
				}
			}

		})
	}

}
