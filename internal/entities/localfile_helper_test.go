package entities

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

var dirPath = "E:\\Anime"
var testFilePath = "E:\\Anime\\Bungou Stray Dogs 5th Season\\[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv"

func TestLocalFile_GetTitleVariations(t *testing.T) {

	lf := NewLocalFile(testFilePath, dirPath)

	if assert.NotNil(t, lf) {
		tv := lo.Map(lf.GetTitleVariations(), func(item *string, _ int) string { return *item })

		if assert.ElementsMatch(t, tv, []string{
			"Bungou Stray Dogs 5th Season", "Bungou Stray Dogs Season 5", "Bungou Stray Dogs S5",
		}) {
			t.Log(spew.Sdump(lf.GetTitleVariations()))
		}
	}

}

var testFilePath2 = "E:\\Anime\\Shakugan No Shana\\Shakugan No Shana I\\Opening\\OP01.mkv"

func TestLocalFile_GetTitleVariations2(t *testing.T) {

	lf := NewLocalFile(testFilePath2, dirPath)

	if assert.NotNil(t, lf) {
		tv := lo.Map(lf.GetTitleVariations(), func(item *string, _ int) string { return *item })

		assert.Contains(t, tv, "Shakugan No Shana")
	}

}
