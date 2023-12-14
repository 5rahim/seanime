package seanime_parser

import (
	"slices"
	"strconv"
	"strings"
)

func (p *parser) parseEpisode() {

	// Check alt episode number or range //TODO
	// e.g. 01 (12)
	if found := p.parseKnownEpisodeAltNumber(); found {
		return // Stop if an episode number is found (even if the alt number is not found)
	}

	// Search by alt episode number
	// We check if any unknown number token is followed by "({number})", e.g. {tkn} (14)
	if found := p.parseEpisodeBySearchingForAltNumber(); found {
		return // Stop if an episode number with alt number is found
	}

	// Check combined or separated keywords other than season prefixes
	// e.g. Ep1, ED1, ED 1, OVAs 1-3, OVAs 1 ~ 3, OVA1, OVA 1v2
	if found := p.parseKeywordsWithEpisodes(); found {
		return // Stop if an episode number (e.g. Ep1) is found, not an OVA, ED, OP, ...
	}

	// e.g. 01 of 24
	if found := p.parseEpisodeByRangeSeparator("OF"); found {
		return // Stop if an episode number is found
	}

	// Check last number before the first opening bracket (if there is one at the beginning [subgroup], then, before the second opening bracket)
	// e.g. Title - 01
	if found := p.parseEpisodeBySearching(false); found {
		return // Stop if an episode number is found
	}

	// e.g. [12]
	if found := p.parseEpisodeByEnclosedNumber(); found {
		return // Stop if an episode number is found
	}

	// e.g Title 01
	if found := p.parseEpisodeBySearching(true); found {
		return // Stop if an episode number is found
	}

}

// ---------------------------------------------------------------------------------------------------------------------
// Searching for alt episode number
// ---------------------------------------------------------------------------------------------------------------------

func (p *parser) parseEpisodeBySearchingForAltNumber() bool {
	for _, numTkn := range *p.tokenManager.tokens {

		if !numTkn.isNumberOrLikeKind() || !numTkn.isUnknown() {
			continue // Check next token
		}

		// Check if token is followed by "({number})", e.g. {tkn} (14)
		nextTkns, found, _ := p.tokenManager.tokens.getCategorySequenceAfter(p.tokenManager.tokens.getIndexOf(numTkn), []tokenCategory{
			tokenCatOpeningBracket, // (
			tokenCatUnknown,        // 12
			tokenCatClosingBracket, // )
		}, true)

		if !found {
			continue // Check next token
		}

		// Verify that the sequence is correct
		if nextTkns[0].getValue() != "(" || !nextTkns[1].isNumberKind() || nextTkns[2].getValue() != ")" {
			continue
		}

		// Update tokens
		numTkn.setMetadataCategory(metadataEpisodeNumber)
		nextTkns[1].setMetadataCategory(metadataEpisodeNumberAlt)
		return true
	}
	return false
}

// parseKnownEpisodeAltNumber parses the alt episode number if an episode number already exists.
func (p *parser) parseKnownEpisodeAltNumber() (foundEpisode bool) {
	found, tkns := p.tokenManager.tokens.findWithMetadataCategory(metadataEpisodeNumber)
	if !found {
		return false
	}

	// We found an episode number anyway
	foundEpisode = true

	last := tkns[len(tkns)-1]

	// Check if token is followed by "({number})", e.g. {tkn} (14)
	nextTkns, found, _ := p.tokenManager.tokens.getCategorySequenceAfter(p.tokenManager.tokens.getIndexOf(last), []tokenCategory{
		tokenCatOpeningBracket, // (
		tokenCatUnknown,        // 12
		tokenCatClosingBracket, // )
	}, true)
	if !found {

		// Check range after found episode
		{
			if len(tkns) != 1 {
				return
			}
			// Check range, e.g. 01-03
			// making sure that the range has no delimiters
			rangeTkns, foundRange := p.tokenManager.tokens.checkNumberRangeAfter(last, false)
			if !foundRange {
				return
			}
			rangeTkns[1].setMetadataCategory(metadataEpisodeNumber)
		}

		return
	}

	if nextTkns[0].getValue() != "(" || !nextTkns[1].isNumberKind() || nextTkns[2].getValue() != ")" {
		return
	}

	// Update token
	nextTkns[1].setMetadataCategory(metadataEpisodeNumberAlt)
	return
}

// ---------------------------------------------------------------------------------------------------------------------
// Searching by patterns
// ---------------------------------------------------------------------------------------------------------------------

// parseEpisodeBySearching parses the episode number by searching for different patterns.
func (p *parser) parseEpisodeBySearching(aggressive bool) bool {

	// Check "- 01 [...]" or "- 01 480p"
	for {
		var openingBracketTkn *token

		for _, tkn := range *p.tokenManager.tokens {
			// Find the first opening bracket or metadata keyword, whatever comes first
			if tkn.isOpeningBracket() || tkn.isKeyword() {
				if p.tokenManager.tokens.getIndexOf(tkn) == 0 { // Skip first token if it's an opening bracket
					continue
				}
				openingBracketTkn = tkn
				break
			}
		}
		if openingBracketTkn == nil {
			break
		}

		// Get previous token
		numTkn, found, _ := p.tokenManager.tokens.getTokenBeforeSD(openingBracketTkn)
		if !found {
			break
		}
		// Check if previous token is an unknown number or number-like
		if !numTkn.isNumberOrLikeKind() || !numTkn.isUnknown() {
			break
		}

		// Check we find a range before
		// e.g. "01-{numTkn} [...]"
		// making sure that the range has no delimiters
		if rangeTkns, found := p.tokenManager.tokens.checkEpisodeRangeBefore(numTkn); found {
			// Make sure that there is no number range before the range
			// e.g. Avoid this "009-1-02 [...]", where "1-02" is considered as a range
			if _, found := p.tokenManager.tokens.checkNumberRangeBefore(rangeTkns[1], false); !found {
				rangeTkns[1].setMetadataCategory(metadataEpisodeNumber)
				numTkn.setMetadataCategory(metadataEpisodeNumber)
				return true // Found episode number, end
			}
		}

		// If we are not searching aggressively, check if there is a dash separator before the number
		// e.g. "- 01"
		if !aggressive && !p.tokenManager.tokens.foundDashSeparatorBefore(numTkn) {
			break
		}

		// When searching aggressively
		// Check that the number might really be an episode number
		// e.g. if {lastNumTkn} < 10, lastNumTkn should be zero padded to avoid false positives like "Title 2"
		if aggressive {
			if numTkn.isNumberKind() {
				intVal, err := strconv.Atoi(numTkn.getValue())
				if err != nil {
					break
				}
				if intVal < 10 && !isNumberZeroPadded(numTkn.getValue()) {
					break
				}
				// should be isolated
				// avoids Evangelion 1.11 You Can (Not) Redo -> 11 being identified as episode number
				if !p.tokenManager.tokens.isIsolated(numTkn) {
					break
				}
			}
			// in the case of "Title 2v2", we can safely identify "2v2" as an episode number
		}

		numTkn.setMetadataCategory(metadataEpisodeNumber)
		return true // Found episode number, end
	}

	// Check for first occurrence of unknown number preceded and followed by a dash separator
	// e.g. "- 01 -"
	for _, numTkn := range *p.tokenManager.tokens {
		if !numTkn.isUnknown() || !numTkn.isNumberOrLikeKind() {
			continue // Check next token
		}
		// Check dash separator before
		if !p.tokenManager.tokens.foundDashSeparatorBefore(numTkn) {
			continue // Check next token
		}
		// Check dash separator after
		if !p.tokenManager.tokens.foundDashSeparatorAfter(numTkn) {
			continue // Check next token
		}
		// Check that it is not a range
		// e.g. "01-03"
		if _, found := p.tokenManager.tokens.checkNumberRangeBefore(numTkn, false); found {
			continue // Check next token
		}
		if _, found := p.tokenManager.tokens.checkNumberRangeAfter(numTkn, false); found {
			continue // Check next token
		}

		numTkn.setMetadataCategory(metadataEpisodeNumber)
		return true // Found episode number, end
	}

	// Check for last unknown number
	for {
		var lastNumTkn *token
		var count int

		// Get the last unknown number token
		for _, tkn := range *p.tokenManager.tokens {
			if tkn.isYear() {
				continue
			}
			if tkn.isNumberOrLikeKind() && tkn.isUnknown() {
				lastNumTkn = tkn
				count++
			}
		}

		if lastNumTkn == nil {
			break
		}

		// Check we find a range before
		// e.g. "1 - {lastNumTkn} [...]"
		if rangeTkns, found := p.tokenManager.tokens.checkNumberRangeBefore(lastNumTkn, true); found {
			if isNumberZeroPadded(lastNumTkn.getValue()) && !isNumberZeroPadded(rangeTkns[1].getValue()) {

			} else {
				rangeTkns[1].setMetadataCategory(metadataEpisodeNumber)
			}
			lastNumTkn.setMetadataCategory(metadataEpisodeNumber)
			return true // Found episode number, end
		}

		// If we are not searching aggressively, check if there is a dash separator before the number
		// e.g. "- 01"
		if !aggressive && !p.tokenManager.tokens.foundDashSeparatorBefore(lastNumTkn) {
			break
		}

		// When searching aggressively
		// Check that the number might really be an episode number
		// e.g. if {lastNumTkn} < 10, lastNumTkn should be zero padded to avoid false positives like "Title 2"
		if aggressive {
			// If we have more than one number, we can't be sure that it's an episode number
			if count > 1 {
				break
			}
			if p.tokenManager.tokens.foundDashSeparatorAfter(lastNumTkn) && !p.tokenManager.tokens.isFirstToken(lastNumTkn) {
				break
			}
			if lastNumTkn.isNumberKind() {
				intVal, err := strconv.Atoi(lastNumTkn.getValue())
				if err != nil {
					break
				}
				if intVal < 10 && !isNumberZeroPadded(lastNumTkn.getValue()) {
					break
				}
				// should be isolated
				if !p.tokenManager.tokens.isIsolated(lastNumTkn) {
					break
				}
			} else if lastNumTkn.isNumberLikeKind() {
				if !isReasonableEpisodeNumber(lastNumTkn.getValue()) {
					break
				}
			}
			// in the case of "Title 2v2", we can safely identify "2v2" as an episode number
		}

		lastNumTkn.setMetadataCategory(metadataEpisodeNumber)
		return true // Found episode number, end
	}

	return false
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// ---------------------------------------------------------------------------------------------------------------------
// Keywords
// ---------------------------------------------------------------------------------------------------------------------

// parseKeywordsWithEpisodes parses keywords that are combined or separated with a number.
// It does not handle season/volume/part prefixes (keywordCatSeasonPrefix, ...) as those are handled by parseSeason.
//
// It handles the following cases:
// keywordKindCombinedWithNumber
// keywordKindSeparatedWithNumber
//
// e.g. Ep1, ED1, ED 1, OVAs 1-3, OVAs 1 ~ 3, OVA1, OVA 1v2
//
// "foundEpisode" is set to true if an actual episode number is found. (not an OVA, ED, OP, ...)
func (p *parser) parseKeywordsWithEpisodes() (foundEpisode bool) {

	for _, tkn := range *p.tokenManager.tokens {

		if tkn.isKeyword() || !tkn.isUnknown() { // Don't bother if token is already a keyword
			continue // Skip to next token
		}

		keywords, found := p.tokenManager.keywordManager.findKeywordsBy(func(kw *keyword) bool {
			// Get keywords that are combined or separated with a number AND not a season prefix or episode prefix
			// e.g. ED1, ED 1
			return (kw.isCombinedWithNumber() || kw.isSeparatedWithNumber()) &&
				!kw.isSeasonPrefix() && !kw.isVolumePrefix() && !kw.isPartPrefix() && // Skip all these because they are handled in parseSeason()
				strings.HasPrefix(tkn.getNormalizedValue(), kw.value) // Token value starts with keyword value
		})

		if !found {
			continue // Skip to next token
		}

	keywordLoop:
		for _, keyword := range keywords {

			// e.g. ED1
			if keyword.isCombinedWithNumber() {

				// Check if prefix is followed by a number or number-like (e.g. 01, 01v2)
				remaining := strings.TrimPrefix(tkn.getNormalizedValue(), keyword.value)

				if len(remaining) > 0 && isNumberOrLike(remaining) {

					// e.g. ED
					prefixTkn := newToken(keyword.value)
					prefixTkn.setIdentifiedKeywordCategory(keyword.category)
					prefixTkn.setKind(tokenKindWord)

					// e.g. 01, 1, 3
					numberTkn := newToken(tkn.getValue()[len(keyword.value):])
					numberTkn.setMetadataCategory(metadataOtherEpisodeNumber)
					if keyword.isEpisodePrefix() {
						numberTkn.setMetadataCategory(metadataEpisodeNumber)
					}
					numberTkn.setKind(tokenKindNumberLike)
					if isNumber(remaining) {
						numberTkn.setKind(tokenKindNumber)
					}

					//firstNumberIsZeroPadded := isNumberZeroPadded(remaining)

					// Overwrite token and insert new tokens
					// "ED1" -> "ED", "1"
					p.tokenManager.tokens.overwriteAndInsertManyAt(p.tokenManager.tokens.getIndexOf(tkn), []*token{prefixTkn, numberTkn})
					foundEpisode = true

					if isNumber(remaining) { // e.g. ED1.5, don't bother if ED1v2
						p.tokenManager.tokens.checkNumberWithDecimal(numberTkn) // Check if number is decimal
					}

					// Check range
					// e.g. ED1-3, ED01-03, ED1 ~ 3, ED01 ~ 03
					if rangeTkns, found := p.tokenManager.tokens.checkNumberRangeAfter(numberTkn, false); found {
						// e.g. ED1-3, ED01-03
						rangeTkns[1].setMetadataCategory(metadataOtherEpisodeNumber)
						if keyword.isEpisodePrefix() {
							rangeTkns[1].setMetadataCategory(metadataEpisodeNumber)
						}
						break keywordLoop // Skip to next token
					}
					//if nextNumTkn, found, kind := checkNumberRangeAfterToken(p, numberTkn, firstNumberIsZeroPadded); found {
					//	// e.g. ED1-3, ED01-03
					//	if kind == 0 {
					//		nextNumTkn.setMetadataCategory(metadataOtherEpisodeNumber)
					//		p.tokenManager.tokens.checkNumberWithDecimal(nextNumTkn) // Check if number is decimal
					//		break keywordLoop                                        // Skip to next token
					//	}
					//}

					break keywordLoop // Skip to next token

				}

			}

			// e.g. ED 1
			if keyword.isSeparatedWithNumber() {

				// Get next token, by skipping delimiters
				// Check if next token is a number or number-like
				if nextTkn, found, _ := p.tokenManager.tokens.getTokenAfterSD(tkn); found &&
					(nextTkn.isNumberOrLikeKind() && nextTkn.isUnknown()) {

					tkn.setIdentifiedKeywordCategory(keyword.category)
					nextTkn.setMetadataCategory(metadataOtherEpisodeNumber)
					if keyword.isEpisodePrefix() {
						foundEpisode = true
						nextTkn.setMetadataCategory(metadataEpisodeNumber)
					}

					// Check range
					firstSeasonIsZeroPadded := isNumberZeroPadded(nextTkn.getValue())
					// e.g. ED 1-3, ED 01-03, ED 1 ~ 3, ED 01 ~ 03
					if nextNumTkn, found, kind := checkNumberRangeAfterToken(p, nextTkn, firstSeasonIsZeroPadded); found {
						// e.g. ED 1-3, ED 01-03
						if kind == 0 {
							nextNumTkn.setMetadataCategory(metadataOtherEpisodeNumber)
							if keyword.isEpisodePrefix() {
								nextNumTkn.setMetadataCategory(metadataEpisodeNumber)
							}
							break keywordLoop // Skip to next token
						}
					}

					break keywordLoop // Skip to next token

				}

			}

		}
	}

	return

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// ---------------------------------------------------------------------------------------------------------------------
// Enclosed episode number
// ---------------------------------------------------------------------------------------------------------------------

func (p *parser) parseEpisodeByEnclosedNumber() bool {

	//if !p.tokenManager.tokens.allUnknownTokensAreEnclosed() {
	//	return false
	//}

	for _, numTkn := range *p.tokenManager.tokens {
		if !numTkn.isUnknown() || !numTkn.isNumberOrLikeKind() || !numTkn.isEnclosed() {
			continue // Check next token
		}

		// e.g. [
		prevTkn, found := p.tokenManager.tokens.getTokenBefore(numTkn)
		if !found || !prevTkn.isOpeningBracket() {
			continue
		}
		// e.g. ]
		nextTkn, found := p.tokenManager.tokens.getTokenAfter(numTkn)
		if !found || !nextTkn.isClosingBracket() {
			continue
		}

		// e.g. We found [12] or (12)
		numTkn.setMetadataCategory(metadataEpisodeNumber)
		return true // Found

	}

	return false
}

// ---------------------------------------------------------------------------------------------------------------------
// Searching by range separator
// ---------------------------------------------------------------------------------------------------------------------

func (p *parser) parseEpisodeByRangeSeparator(value string) bool {

	for _, numTkn := range *p.tokenManager.tokens {
		if !numTkn.isUnknown() || !numTkn.isNumberOrLikeKind() {
			continue // Check next token
		}

		// e.g. of
		ofTkn, found, _ := p.tokenManager.tokens.getTokenAfterSD(numTkn)
		if !found || ofTkn.getNormalizedValue() != value {
			continue
		}
		// e.g. [
		secondNumTkn, found, _ := p.tokenManager.tokens.getTokenAfterSD(ofTkn)
		if !found || !secondNumTkn.isNumberOrLikeKind() {
			continue
		}

		// e.g. We found "01 of 24"
		numTkn.setMetadataCategory(metadataEpisodeNumber)
		ofTkn.setCategory(tokenCatKnown) // Set category to known to avoid incorrect episode title parsing
		secondNumTkn.setMetadataCategory(metadataOtherEpisodeNumber)
		return true

	}
	return false

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// e.g.
//   - 01 -> true
//   - 1 -> true
//   - 1v2 -> true
//   - 1.1 -> true
//   - 1.11 -> false
func isReasonableEpisodeNumber(s string) bool {
	if slices.Contains([]string{"3.0", "1.11"}, s) {
		return false
	}
	if isDigitsOnly(s) {
		return true
	}
	if isNumber(s) {
		return true
	}
	index := strings.IndexByte(s, '.')
	if index == -1 {
		return true
	}
	decimalIntVal, err := strconv.Atoi(s[index+1:])
	if err == nil {
		return false
	}
	if decimalIntVal > 5 {
		return false
	}
	return true
}
