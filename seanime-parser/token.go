package seanime_parser

import (
	"github.com/google/uuid"
)

type tokenCategory = string

const (
	tokenCatUnknown        tokenCategory = "unknown"
	tokenCatDelimiter      tokenCategory = "delimiter"
	tokenCatSeparator      tokenCategory = "separator"
	tokenCatKnown          tokenCategory = "known"
	tokenCatParts          tokenCategory = "parts"
	tokenCatOpeningBracket tokenCategory = "openingBracket"
	tokenCatClosingBracket tokenCategory = "closingBracket"
)

type tokenKind = string

const (
	tokenKindUnknown          tokenKind = ""
	tokenKindCharacter        tokenKind = "character"
	tokenKindWord             tokenKind = "word"
	tokenKindNumber           tokenKind = "number"
	tokenKindNumberLike       tokenKind = "numberLike"
	tokenKindOrdinalNumber    tokenKind = "ordinalNumber"
	tokenKindCrc32            tokenKind = "crc32"
	tokenKindPossibleVideoRes tokenKind = "possibleVideoRes"
	tokenKindYear             tokenKind = "year"
)

type token struct {
	UUID                      string
	Value                     string
	Category                  tokenCategory
	Kind                      tokenKind
	IdentifiedKeywordCategory keywordCategory
	Enclosed                  bool
	Parts                     []*token
	MetadataCategory          metadataCategory
}

func newToken(value string) *token {
	return &token{
		UUID:                      uuid.NewString(),
		Value:                     value,
		Category:                  tokenCatUnknown,
		Kind:                      tokenKindUnknown,
		IdentifiedKeywordCategory: keywordCatNone,
		Enclosed:                  false,
		Parts:                     nil,
		MetadataCategory:          0,
	}
}

// setMetadataCategory will update the token's MetadataCategory
// and update its Category to tokenCatKnown if the metadataCategory is not metadataUnknown.
func (t *token) setMetadataCategory(mk metadataCategory) {
	if t == nil {
		return
	}

	t.MetadataCategory = mk
	t.Category = tokenCatKnown
	if mk == metadataUnknown {
		t.Category = tokenCatUnknown
	}
}

func (t *token) setValue(s string) {
	if t == nil {
		return
	}

	t.Value = s
}

// setCategory will update the token's Category to the specified tokenCategory value.
// If the token is nil, the function will return without making any changes.
//
// Example:
//
//	tkn := &token{}
//	tkn.setCategory(tokenCatKnown)
//	fmt.Println(tkn.Category)  // Output: tokenCatKnown
func (t *token) setCategory(c tokenCategory) {
	if t == nil {
		return
	}

	t.Category = c
}

// setIdentifiedKeywordCategory updates the `IdentifiedKeywordCategory` of a token
// with the provided keyword category.
//
// Example:
//
//	tkn := &token{}
//	tkn.setIdentifiedKeywordCategory(keywordCatSeasonPrefix)
//	fmt.Println(tkn.IdentifiedKeywordCategory)  // Output: keywordCatSeasonPrefix
func (t *token) setIdentifiedKeywordCategory(c keywordCategory) {
	if t == nil {
		return
	}

	t.IdentifiedKeywordCategory = c
}

func (t *token) setKind(k tokenKind) {
	if t == nil {
		return
	}

	t.Kind = k
}

func (t *token) setParts(p []*token) {
	if t == nil {
		return
	}

	t.Category = tokenCatParts
	t.Parts = p
}

func (t *token) setEnclosed(v bool) {
	if t == nil {
		return
	}

	t.Enclosed = v
}

func (t *token) getParts() (tokenParts []*token, found bool) {
	if t.Parts == nil {
		found = false
	}
	tokenParts = t.Parts
	found = true
	return
}
