package seanime_parser

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"testing"
)

func testEnclosedDetection(t *testing.T, tkns []*token) {
	for idx, tkn := range tkns {
		if tkn.getValue() == "ENCLOSED" {
			assert.Truef(t, tkn.isEnclosed(), "expected token to be enclosed at index %d", idx)
		}
		if tkn.getValue() == "ABC" {
			assert.Falsef(t, tkn.isEnclosed(), "expected token not to be enclosed at index %d", idx)
		}
	}
}

func TestEnclosedDetection(t *testing.T) {

	ret := tokenize("[ENCLOSED] ABC ABC ABC (ENCLOSED) [ENCLOSED]")
	test := tokens{}
	test.setTokens(ret)
	testEnclosedDetection(t, ret)

	ret = tokenize("[ENCLOSED (ENCLOSED] ABC ABC [ENCLOSED]")
	testEnclosedDetection(t, ret)

	ret = tokenize("[ENCLOSED (ENCLOSED) ENCLOSED ENCLOSED [ENCLOSED]")
	testEnclosedDetection(t, ret)

	ret = tokenize("[ENCLOSED] (ENCLOSED) ABC ABC [ABC ABC")
	testEnclosedDetection(t, ret)

}

func TestTokenizeCategories(t *testing.T) {
	input := "[enclosed] abc-abc.mkv"
	tkns := tokenize(input)

	categories := []tokenCategory{
		tokenCatOpeningBracket, // [
		tokenCatUnknown,        // enclosed
		tokenCatClosingBracket, // ]
		tokenCatDelimiter,      // " "
		tokenCatUnknown,        // abc
		tokenCatSeparator,      // -
		tokenCatUnknown,        // abc
		tokenCatDelimiter,      // .
		tokenCatUnknown,        // mkv
	}

	assert.Equal(t, len(tkns), len(categories))

	for idx, tkn := range tkns {
		assert.Equal(t, tkn.Category, categories[idx])
		if tkn.Category != categories[idx] {
			spew.Dump(tkn)
		}
	}

}

func TestTokenizeKinds(t *testing.T) {
	input := "01 ABC 01v3 2024 1080p"
	tkns := tokenize(input)

	kinds := []tokenKind{
		tokenKindNumber,
		tokenKindCharacter,
		tokenKindWord,
		tokenKindCharacter,
		tokenKindNumberLike,
		tokenKindCharacter,
		tokenKindYear,
		tokenKindCharacter,
		tokenKindPossibleVideoRes,
	}

	assert.Equal(t, len(tkns), len(kinds))

	for idx, tkn := range tkns {
		assert.Equal(t, tkn.Kind, kinds[idx])
		if tkn.Kind != kinds[idx] {
			spew.Dump(tkn)
		}
	}

}

func TestIsMatchingClosingBracket(t *testing.T) {
	ret := isMatchingClosingBracket("[", "]")
	assert.True(t, ret)

	ret = isMatchingClosingBracket("[", ")")
	assert.False(t, ret)
}
