package seanime_parser

func (t *token) getValue() string {
	if t == nil {
		return ""
	}
	return t.Value
}

func (t *token) isIdentifiedMetadata() bool {
	if t == nil {
		return false
	}
	return t.MetadataCategory != metadataUnknown
}

func (t *token) isMetadataCategory(kind metadataCategory) bool {
	if t == nil {
		return false
	}
	return t.MetadataCategory == kind
}

func (t *token) getMetadataCategory() (metadataCategory, bool) {
	if t == nil {
		return 0, false
	}
	return t.MetadataCategory, t.MetadataCategory != metadataUnknown
}

func (t *token) getNormalizedValue() string {
	if t == nil {
		return ""
	}
	return normalize(t.Value)
}

func (t *token) getCategory() tokenCategory {
	if t == nil {
		return ""
	}
	return t.Category
}

func (t *token) getIdentifiedKeywordCategory() (keywordCategory, bool) {
	if t == nil {
		return keywordCatNone, false
	}
	return t.IdentifiedKeywordCategory, t.IdentifiedKeywordCategory != keywordCatNone
}

func (t *token) getKind() tokenKind {
	if t == nil {
		return ""
	}
	return t.Kind
}

func (t *token) isCategory(cat tokenCategory) bool {
	if t == nil {
		return false
	}
	return t.Category == cat
}

func (t *token) isKeyword() bool {
	if t == nil {
		return false
	}
	return t.IdentifiedKeywordCategory != keywordCatNone
}

func (t *token) isStandaloneKeyword() bool {
	if t == nil {
		return false
	}
	if t.IdentifiedKeywordCategory == keywordCatNone {
		return false
	}
	return t.IdentifiedKeywordKind == keywordKindStandalone
}

func (t *token) isKeywordCategory(c keywordCategory) bool {
	if t == nil {
		return false
	}
	return t.IdentifiedKeywordCategory == c
}

func (t *token) isKind(kind tokenKind) bool {
	if t == nil {
		return false
	}
	return t.Kind == kind
}

func (t *token) isOpeningBracket() bool {
	if t == nil {
		return false
	}
	return t.Category == tokenCatOpeningBracket
}

func (t *token) isClosingBracket() bool {
	if t == nil {
		return false
	}
	return t.Category == tokenCatClosingBracket
}

func (t *token) isUnknown() bool {
	if t == nil {
		return false
	}
	return t.Category == tokenCatUnknown
}

func (t *token) isDelimiter() bool {
	if t == nil {
		return false
	}
	return t.Category == tokenCatDelimiter
}

func (t *token) isSeparator() bool {
	if t == nil {
		return false
	}
	return t.Category == tokenCatSeparator
}

func (t *token) isWordKind() bool {
	if t == nil {
		return false
	}
	return t.Kind == tokenKindWord
}

func (t *token) isCRC32() bool {
	if t == nil {
		return false
	}
	return t.Kind == tokenKindCrc32
}

func (t *token) isDotDelimiter() bool {
	if t == nil {
		return false
	}
	return t.Category == tokenCatDelimiter && t.Value == "."
}

func (t *token) isPlusSeparator() bool {
	if t == nil {
		return false
	}
	return t.Category == tokenCatSeparator && t.Value == "+"
}

func (t *token) isDashSeparator() bool {
	if t == nil {
		return false
	}
	return t.Category == tokenCatSeparator && (t.Value == "-" ||
		t.Value == "–" ||
		t.Value == "—" ||
		t.Value == "―")
}

func (t *token) isEnclosed() bool {
	if t == nil {
		return false
	}
	return t.Enclosed
}

// isNumberKind checks if the token comprises digits only.
func (t *token) isNumberKind() bool {
	if t == nil {
		return false
	}
	return t.Kind == tokenKindNumber
}

// isNumberLikeKind checks if the token comprises digits with version identifiers
func (t *token) isNumberLikeKind() bool {
	if t == nil {
		return false
	}
	return t.Kind == tokenKindNumberLike
}

func (t *token) isYear() bool {
	if t == nil {
		return false
	}
	return t.Kind == tokenKindYear
}

func (t *token) isFileExt() bool {
	if t == nil {
		return false
	}
	return t.IdentifiedKeywordCategory == keywordCatFileExtension
}

// isNumberLikeKind checks if the token comprises only digits or digits with version identifiers
func (t *token) isNumberOrLikeKind() bool {
	if t == nil {
		return false
	}
	return t.isNumberKind() || t.isNumberLikeKind()
}

func (t *token) isOrdinalNumber() bool {
	if t == nil {
		return false
	}
	return t.Kind == tokenKindOrdinalNumber
}

func (t *token) isMostlyLatinString() bool {
	if len(t.getValue()) <= 0 {
		return false
	}
	latinLength := 0
	nonLatinLength := 0
	for _, r := range t.getValue() {
		if isLatinRune(r) {
			latinLength++
		} else {
			nonLatinLength++
		}
	}
	return latinLength > nonLatinLength
}

// isFileInfoMetadata checks if the token is classified as file information metadata.
// It returns true if the token DOES NOT belong to one of the following identified keyword categories:
//
// - keywordCatNone
// - keywordCatYear
// - keywordCatReleaseVersion
// - keywordCatReleaseGroup
// - keywordCatSeasonPrefix
// - keywordCatPartPrefix
// - keywordCatVolumePrefix
// - keywordCatEpisodePrefix
func (t *token) isFileInfoMetadata() bool {
	if t == nil {
		return false
	}

	switch t.IdentifiedKeywordCategory {
	case keywordCatNone:
		return false
	case keywordCatYear:
		return false
	case keywordCatReleaseVersion:
		return false
	case keywordCatReleaseGroup:
		return false
	case keywordCatSeasonPrefix:
		return false
	case keywordCatPartPrefix:
		return false
	case keywordCatVolumePrefix:
		return false
	case keywordCatEpisodePrefix:
		return false
	default:
		return true
	}
}

func (t *token) isAnimeInfoMetadata() bool {
	if t == nil {
		return false
	}
	return !t.isFileInfoMetadata()
}
