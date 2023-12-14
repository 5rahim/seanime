package seanime_parser

import (
	"strconv"
	"strings"
)

// parseSeason looks for season/volume/part prefixes and numbers.
// It will also check for episode numbers.
// e.g. S01E01, Season 1, S1, 1st Season, 1st Volume, 1st Part, 1st Season - 03, 1st Season - 03v2, 1st Season - 03' etc...
func (p *parser) parseSeason() {

	for _, tkn := range *p.tokenManager.tokens {

		if !tkn.isUnknown() { // Don't bother if token is already known
			continue // Skip to next token
		}

		// Parse S01E01 by checking if the token follows the pattern
		if strings.HasPrefix(tkn.getNormalizedValue(), "S") && len(tkn.getValue()) > 3 {
			// Extract season and episode
			if season, sep, episode, ok := extractSeasonAndEpisode(tkn.getValue()); ok {
				seasonPrefixTkn := newToken("S")
				seasonPrefixTkn.setIdentifiedKeywordCategory(keywordCatSeasonPrefix)
				seasonPrefixTkn.setKind(tokenKindCharacter)

				seasonTkn := newToken(season)
				seasonTkn.setMetadataCategory(metadataSeason)
				seasonTkn.setKind(tokenKindNumber)

				sepTkn := newToken(sep)
				sepTkn.setIdentifiedKeywordCategory(keywordCatEpisodePrefix)
				sepTkn.setKind(tokenKindCharacter)

				episodeTkn := newToken(episode)
				episodeTkn.setMetadataCategory(metadataEpisodeNumber)
				episodeTkn.setKind(tokenKindNumberLike)
				if isNumber(episode) {
					episodeTkn.setKind(tokenKindNumber)
					p.tokenManager.tokens.checkNumberWithDecimal(episodeTkn) // Check if number is decimal
				}

				p.tokenManager.tokens.overwriteAndInsertManyAt(p.tokenManager.tokens.getIndexOf(tkn), []*token{seasonPrefixTkn, seasonTkn, sepTkn, episodeTkn})

				episodeIsZeroPadded := isNumberZeroPadded(episodeTkn.getValue())

				// Check range
				// e.g. S1-3, S01-03, S1 ~ 3, S01 ~ 03
				if nextNumTkn, found, kind := checkNumberRangeAfterToken(p, seasonTkn, episodeIsZeroPadded); found {
					// e.g. S1-3, S01-03
					if kind == 0 {
						nextNumTkn.setMetadataCategory(metadataEpisodeNumber)
						p.tokenManager.tokens.checkNumberWithDecimal(nextNumTkn) // Check if number is decimal
						continue                                                 // Skip to next token
					}
				}

				continue // Skip to next token
			}
		}

		// Combined or separated seasons/volumes/parts
		if strings.HasPrefix(tkn.getNormalizedValue(), "S") ||
			strings.HasPrefix(tkn.getNormalizedValue(), "P") ||
			strings.HasPrefix(tkn.getNormalizedValue(), "V") {

			keywords, found := p.tokenManager.keywordManager.findKeywordsBy(func(kw *keyword) bool {
				return (kw.isSeasonPrefix() || kw.isVolumePrefix() || kw.isPartPrefix()) && // Season/Part/Volume prefix
					strings.HasPrefix(tkn.getNormalizedValue(), kw.value) // Token starts with season prefix
			})

			if !found {
				continue // Skip to next token
			}

		keywordLoop:
			for _, keyword := range keywords {

				// e.g. S01
				if keyword.isCombinedWithNumber() {

					metadataCat := getMetadataCategoryFromKeywordPrefix(keyword.category)

					// Check if token is after file metadata
					if p.tokenManager.tokens.isTokenAfterFileMetadata(tkn) {
						break keywordLoop // Skip to next token
					}

					// Check if prefix is followed by a number or number-like (e.g. 01, 01v2)
					remaining := strings.TrimPrefix(tkn.getNormalizedValue(), keyword.value)

					if len(remaining) > 0 && isNumberOrLike(remaining) {

						// e.g. S
						seasonPrefixTkn := newToken(keyword.value)
						seasonPrefixTkn.setIdentifiedKeywordCategory(keyword.category)
						seasonPrefixTkn.setKind(tokenKindWord)

						// e.g. 01, 1, 3
						seasonTkn := newToken(tkn.getValue()[len(keyword.value):])
						seasonTkn.setMetadataCategory(metadataCat)
						seasonTkn.setKind(tokenKindNumberLike)
						if isNumber(remaining) {
							seasonTkn.setKind(tokenKindNumber)
							p.tokenManager.tokens.checkNumberWithDecimal(seasonTkn) // Check if number is decimal
						}

						firstSeasonIsZeroPadded := isNumberZeroPadded(remaining)

						// Overwrite and insert tokens
						p.tokenManager.tokens.overwriteAndInsertManyAt(p.tokenManager.tokens.getIndexOf(tkn), []*token{seasonPrefixTkn, seasonTkn})

						// Check if the number identified is a decimal number
						// e.g. "S1" "." "5" -> "S1.5"
						if isNumber(remaining) { // e.g. S1.5, don't bother if S1v2
							p.tokenManager.tokens.checkNumberWithDecimal(seasonTkn) // Check if number is decimal
						}

						// Check range
						// e.g. S1-3, S01-03, S1 ~ 3, S01 ~ 03
						if nextNumTkn, found, kind := checkNumberRangeAfterToken(p, seasonTkn, firstSeasonIsZeroPadded); found {
							// e.g. S1-3, S01-03
							if kind == 0 {
								nextNumTkn.setMetadataCategory(metadataCat)
								p.tokenManager.tokens.checkNumberWithDecimal(nextNumTkn) // Check if number is decimal
								break keywordLoop                                        // Skip to next token
							}
							// e.g. S01 - 03
							if kind == 1 {
								nextNumTkn.setMetadataCategory(metadataEpisodeNumber)
								p.tokenManager.tokens.checkNumberWithDecimal(nextNumTkn) // Check if number is decimal
								break keywordLoop                                        // Skip to next token
							}
						}

					}

				}

				// e.g. Season 01
				if keyword.isSeparatedWithNumber() {

					metadataCat := getMetadataCategoryFromKeywordPrefix(keyword.category)

					// Get next token, by skipping delimiters
					// Check if next token is a number or number-like
					if nextTkn, found, _ := p.tokenManager.tokens.getTokenAfterSD(tkn); found &&
						(nextTkn.isNumberOrLikeKind() && nextTkn.isUnknown()) {

						tkn.setIdentifiedKeywordCategory(keyword.category)
						nextTkn.setMetadataCategory(metadataCat)

						// Check range
						firstSeasonIsZeroPadded := isNumberZeroPadded(nextTkn.getValue())
						// e.g. Season 1-3, Season 01-03, Season 1 ~ 3, Season 01 ~ 03
						if nextNumTkn, found, kind := checkNumberRangeAfterToken(p, nextTkn, firstSeasonIsZeroPadded); found {
							// e.g. Season 1-3, Season 01-03
							if kind == 0 {
								nextNumTkn.setMetadataCategory(metadataCat)
								break keywordLoop // Skip to next token
							}
							// e.g. Season 01 - 03
							if kind == 1 {
								nextNumTkn.setMetadataCategory(metadataEpisodeNumber)
								break keywordLoop // Skip to next token
							}
						}

						break keywordLoop // Skip to next token

					}

				}

				// e.g. 1st Season, first season
				if keyword.isOrdinalSuffix() {

					// Get previous token, by skipping delimiters
					// Check if next token is an ordinal number
					if nextTkn, found, _ := p.tokenManager.tokens.getTokenBeforeSD(tkn); found &&
						(nextTkn.isOrdinalNumber() && nextTkn.isUnknown()) {

						tkn.setIdentifiedKeywordCategory(keyword.category)

						if num, ok := getNumberFromOrdinal(nextTkn.getValue()); ok {
							nextTkn.setValue(strconv.Itoa(num))
							nextTkn.setMetadataCategory(metadataSeason)
							nextTkn.setKind(tokenKindNumber)

							// Check range ONLY for episode
							// e.g. 1st Season - 03
							if nextNumTkn, found, kind := checkNumberRangeAfterToken(p, tkn, false); found {
								// e.g. S01 - 03
								if kind == 1 {
									nextNumTkn.setMetadataCategory(metadataEpisodeNumber)
									break keywordLoop // Skip to next token
								}
							}

							break keywordLoop // Skip to next token
						}

					}

				}

			}

		}

		// Parse 01x01
		if strings.Contains(tkn.getNormalizedValue(), "X") && len(tkn.getValue()) > 3 {
			// Extract season and episode
			if season, sep, episode, ok := extractSeasonAndEpisode(tkn.getValue()); ok {
				if len(season) > 2 {
					continue // Skip to next token
				}

				seasonTkn := newToken(season)
				seasonTkn.setMetadataCategory(metadataSeason)
				seasonTkn.setKind(tokenKindNumber)

				sepTkn := newToken(sep)
				sepTkn.setIdentifiedKeywordCategory(keywordCatEpisodePrefix)
				sepTkn.setKind(tokenKindCharacter)

				episodeTkn := newToken(episode)
				episodeTkn.setMetadataCategory(metadataEpisodeNumber)
				if isNumber(episode) {
					episodeTkn.setKind(tokenKindNumber)
				} else {
					episodeTkn.setKind(tokenKindNumberLike)
				}

				p.tokenManager.tokens.overwriteAndInsertManyAt(p.tokenManager.tokens.getIndexOf(tkn), []*token{seasonTkn, sepTkn, episodeTkn})
				continue // Skip to next token
			}
		}

	}

}

func checkNumberRangeAfterToken(p *parser, tkn *token, prevNumberIsPadded bool) (*token, bool, int) {

	var nextNumTkn *token
	found := false
	var kind int // 0 = season, 1 = episode

	// Check range
	// e.g. S1-3, S01-03, S1 ~ 3, S01 ~ 03
	for {
		if rangeTkns, ok, dlSkipped := p.tokenManager.tokens.getCategorySequenceAfter(p.tokenManager.tokens.getIndexOf(tkn), []tokenCategory{
			tokenCatSeparator, // -
			tokenCatUnknown,   // 05
		}, true); ok {

			nextNumTkn = rangeTkns[1]

			// Check episode
			if rangeTkns[1].isNumberOrLikeKind() && rangeTkns[0].isDashSeparator() {

				// e.g. S1 - 03, S01- 03
				if dlSkipped > 0 {
					if intVal, err := strconv.Atoi(nextNumTkn.getValue()); err == nil {
						// e.g. if < 10 -> 01, 02, 03. if > 10 -> 11, 12, 13
						if (intVal < 10 && isNumberZeroPadded(nextNumTkn.getValue())) || (intVal >= 10) {
							nextNumTkn.setMetadataCategory(metadataEpisodeNumber)
							kind = 1
							found = true
							break
						}
					} else { // /!\ might need to do some additional checks on the number
						nextNumTkn.setMetadataCategory(metadataEpisodeNumber)
						kind = 1
						found = true
						break
					}

					// e.g. S1-03 (dlSkipped = 0) Where 03 might be an episode. This is not very likely
				} else if !prevNumberIsPadded && isNumberZeroPadded(nextNumTkn.getValue()) {
					nextNumTkn.setMetadataCategory(metadataSeason)
					kind = 1
					found = true
					break
				}
			}

			// Avoid this case: S1-2v2
			if !nextNumTkn.isNumberKind() {
				found = false
				break
			}

			intVal, err := strconv.Atoi(nextNumTkn.getValue())
			if err != nil {
				found = false
				break
			}

			// Avoid this case: S01 - 3
			if intVal < 10 && (prevNumberIsPadded && !isNumberZeroPadded(nextNumTkn.getValue())) {
				found = false
				break
			}

			// e.g. S1-3
			nextNumTkn.setMetadataCategory(metadataSeason)
			kind = 0
			found = true
			break

		}
		break
	}

	return nextNumTkn, found, kind

}
