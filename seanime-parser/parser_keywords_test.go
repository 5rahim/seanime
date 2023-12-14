package seanime_parser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKeywordGroups(t *testing.T) {

	tests := []struct {
		name             string
		input            string
		expectedTknValue string
		keywordCat       keywordCategory
	}{
		{"1", "BLU RAY 1080P", "BLU RAY", keywordCatSource},
		{"2", "BLU-RAY 1080P", "BLU-RAY", keywordCatSource},
		{"3", "TV RIP 1080P", "TV RIP", keywordCatSource},
		{"4", "10 bits 1080P", "10 bits", keywordCatVideoTerm},
		{"4", "10-bit 1080P", "10-bit", keywordCatVideoTerm},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(tt.input)
			tm := p.tokenManager

			tkn, _ := tm.tokens.getAtSafe(0)
			found := p.identifyKeyword(tkn, "normal")

			assert.True(t, found)
			assert.Equal(t, tt.expectedTknValue, tkn.getValue())
			assert.True(t, tkn.isKeywordCategory(tt.keywordCat))

			t.Log(tm.tokens.sPrint())
		})
	}

}

func TestStandaloneKeywords(t *testing.T) {

	tests := []struct {
		name               string
		input              string
		expectedKeywordCat keywordCategory
	}{
		{"1", "BLURAY 1080P", keywordCatSource},
		{"2", "60FPS 1080P", keywordCatVideoTerm},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(tt.input)
			tm := p.tokenManager

			tkn, _ := tm.tokens.getAtSafe(0)
			found := p.identifyKeyword(tkn, "normal")

			assert.True(t, found)

			tknKeywordCat, found := tkn.getIdentifiedKeywordCategory()

			assert.Equal(t, tt.expectedKeywordCat, tknKeywordCat)

			t.Log(tm.tokens.sPrint())
		})
	}

}
