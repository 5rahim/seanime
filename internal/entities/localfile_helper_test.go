package entities_test

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/stretchr/testify/assert"
	"testing"
)

var dirPath = "E:\\Anime"

func TestTitleVariations(t *testing.T) {

	tests := []struct {
		filePath string
		titles   []string
	}{
		{
			filePath: "E:\\Anime\\Bungou Stray Dogs 5th Season\\Bungou Stray Dogs\\[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			titles: []string{
				"Bungou Stray Dogs 5th Season",
				"Bungou Stray Dogs Season 5",
				"Bungou Stray Dogs S5",
			},
		},
		{
			filePath: "E:\\Anime\\Shakugan No Shana\\Shakugan No Shana I\\Opening\\OP01.mkv",
			titles: []string{
				"Shakugan No Shana I",
			},
		},
		{
			filePath: "E:\\ANIME\\Neon Genesis Evangelion Death & Rebirth\\[Anime Time] Neon Genesis Evangelion - Rebirth.mkv",
			titles: []string{
				"Neon Genesis Evangelion - Rebirth",
				"Neon Genesis Evangelion Death & Rebirth",
			},
		},
		{
			filePath: "E:\\ANIME\\Omoi, Omoware, Furi, Furare\\[GJM] Love Me, Love Me Not (BD 1080p) [841C23CD].mkv",
			titles: []string{
				"Love Me, Love Me Not",
				"Omoi, Omoware, Furi, Furare",
			},
		},
		{
			filePath: "E:\\ANIME\\Violet Evergarden Gaiden Eien to Jidou Shuki Ningyou\\Violet.Evergarden.Gaiden.2019.1080..Dual.Audio.BDRip.10.bits.DD.x265-EMBER.mkv",
			titles: []string{
				"Violet Evergarden Gaiden Eien to Jidou Shuki Ningyou",
				"Violet Evergarden Gaiden 2019",
			},
		},
		{
			filePath: "E:\\ANIME\\Violet Evergarden S01+Movies+OVA 1080p Dual Audio BDRip 10 bits DD x265-EMBER\\01. Season 1 + OVA\\S01E01-'I Love You' and Auto Memory Dolls [F03E1F7A].mkv",
			titles: []string{
				"Violet Evergarden",
				"Violet Evergarden S1",
				"Violet Evergarden Season 1",
				"Violet Evergarden 1st Season",
			},
		},
		{
			filePath: "E:\\ANIME\\Golden Kamuy 4th Season\\[Judas] Golden Kamuy (Season 4) [1080p][HEVC x265 10bit][Multi-Subs]\\[Judas] Golden Kamuy - S04E01.mkv",
			titles: []string{
				"Golden Kamuy S4",
				"Golden Kamuy Season 4",
				"Golden Kamuy 4th Season",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			lf := entities.NewLocalFile(tt.filePath, dirPath)

			if assert.NotNil(t, lf) {
				tv := lo.Map(lf.GetTitleVariations(), func(item *string, _ int) string { return *item })

				if assert.ElementsMatch(t, tv, tt.titles) {
					t.Log(spew.Sdump(lf.GetTitleVariations()))
				}
			}

		})
	}

}
