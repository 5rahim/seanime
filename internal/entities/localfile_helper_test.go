package entities

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

var DirPath = "E:\\Anime"
var LocalFilePath = "E:\\Anime\\Bungou Stray Dogs 5th Season\\[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv"

func TestLocalFile_GetTitleVariations(t *testing.T) {

	lf := NewLocalFile(LocalFilePath, DirPath)

	if assert.NotNil(t, lf) {
		tv := lo.Map(lf.GetTitleVariations(), func(item *string, _ int) string { return *item })

		if assert.ElementsMatch(t, tv, []string{
			"Bungou Stray Dogs 5th Season", "Bungou Stray Dogs Season 5", "Bungou Stray Dogs S5",
		}) {
			t.Log(spew.Sdump(lf.GetTitleVariations()))
		}
	}

}
