package seanime_parser

import (
	"github.com/samber/lo"
	"slices"
)

func (p *parser) parseEpisodeTitle() {

	episodeShouldBeEnclosed := p.tokenManager.tokens.allUnknownTokensAreEnclosed()

	// Get all tokens after the last episode number token and before an opening bracket/file info metadata/EOF
	found, epTkns := p.tokenManager.tokens.findWithMetadataCategory(metadataEpisodeNumber)
	if !found {
		return // Next try
	}

	// Get the last episode number token
	lastEpTkn := epTkns[0]
	if len(epTkns) > 0 {
		lastEpTkn = epTkns[len(epTkns)-1]
	}

	// Get all unknown tokens between the last episode number token and an opening bracket/file info metadata/EOF
	// e.g. "... `01` -> "episode title" -| `[` ... ]
	tkns, found := p.tokenManager.tokens.walkAndCollecIf(
		p.tokenManager.tokens.getIndexOf(lastEpTkn)+1,
		func(tkn *token) bool {
			// If all unknown tokens are NOT enclosed, then we don't want to collect an episode title that is enclosed
			// e.g. This avoids collecting "MBS" in "Anime title - 05 (MBS 1080p)" as an episode title
			if !episodeShouldBeEnclosed && tkn.isEnclosed() {
				return false
			}
			// Collect every unknown token, not a keyword, not a separator
			return tkn.isUnknown() && !tkn.isKeyword() && !tkn.isSeparator()
		},
		func(tkn *token) bool {
			// Stop when we encounter an opening bracket or a keyword
			return (tkn.isOpeningBracket() && tkn.getValue() == "[") || tkn.isKeyword()
		})
	if !found {
		return
	}

	// Check anime type in the episode title tokens
	// e.g. "Episode title OVA", we don't remove "OVA" from the title, but add it as an anime type
	for _, tkn := range tkns {
		if kwd, found := p.tokenManager.keywordManager.findStandaloneKeywordByValue(tkn.getValue()); found {
			if kwd.isAnimeType() { // if keyword is an anime type, add it to the metadata
				// set it directly to the struct because the token will be removed from the list
				p.metadata.AnimeType = append(p.metadata.AnimeType, tkn.getValue())
			}
		}
	}

	p.tokenManager.tokens.combineTitle(tkns[0], tkns[len(tkns)-1], metadataEpisodeTitle)

	// If the episode title is between parentheses, then consider it a ReleaseGroup
	// FIXME might lead to false positives
	_newTkn := p.tokenManager.tokens.getFromUUID(tkns[0].UUID)
	if p.tokenManager.tokens.isBetweenParentheses(_newTkn) {
		_newTkn.setMetadataCategory(metadataReleaseGroup)
	}

	return

}

func (p *parser) parseTitleIfAllEnclosed() (foundTitle bool) {
	foundTitle = false

	if !p.tokenManager.tokens.allUnknownTokensAreEnclosed() {
		return
	}

	// Get enclosed tokens before enclosed episode number
	// e.g. "[sub][anime title][01][...]"
	for {
		// Get first episode number token (first because we are going backwards, and it could be a range)
		found, epTkns := p.tokenManager.tokens.findWithMetadataCategory(metadataEpisodeNumber)
		if !found {
			break // Next try
		}
		firstEpTkn := epTkns[0]

		// Get the first and second opening bracket going backwards from the episode number token
		// e.g. [ ... ] `[` <-(!) 01 ][ ... ]
		firstOpeningBracketTkn, found := p.tokenManager.tokens.getFirstOccurrenceBefore(
			p.tokenManager.tokens.getIndexOf(firstEpTkn),
			func(tkn *token) bool {
				return tkn.isOpeningBracket()
			})
		if !found {
			break // Next try
		}

		// FIX for "[sub][anime title] 01 [...]", the second opening bracket will be that of the subgroup
		// so instead, we need to check if the firstEpTkn is enclosed, and if it is, then we get the second opening bracket
		secondOpeningBracketTkn := firstOpeningBracketTkn
		if firstEpTkn.isEnclosed() {
			// Get second opening bracket going backwards
			// e.g. `[` <-(!) ... <- ] <- `[` 01 ][ ... ]
			secondOpeningBracketTkn, found = p.tokenManager.tokens.getFirstOccurrenceBefore(
				p.tokenManager.tokens.getIndexOf(firstOpeningBracketTkn),
				func(tkn *token) bool {
					return tkn.isOpeningBracket()
				})
			if !found {
				break // Next try
			}
		}

		// Get all unknown tokens between the two opening brackets
		// e.g. `[` -> "anime" -> "title" ] -| `[` 01 ][ ... ]
		tkns, found := p.tokenManager.tokens.walkAndCollecIf(
			p.tokenManager.tokens.getIndexOf(secondOpeningBracketTkn)+1,
			func(tkn *token) bool {
				return tkn.isUnknown() && !tkn.isKeyword() && !tkn.isSeparator()
			},
			func(tkn *token) bool {
				// Stop when we encounter the first opening bracket or a keyword
				// e.g. [Mobile_Suit_Gundam_Seed_Destiny_HD_REMASTER][07] -> Mobile Suit Gundam Seed Destiny
				if firstEpTkn.isEnclosed() {
					return tkn.UUID == firstOpeningBracketTkn.UUID || tkn.isKeyword()
				} else {
					return tkn.isClosingBracket()
				}
			})
		if !found {
			break // Next try
		}

		p.tokenManager.tokens.combineTitle(tkns[0], tkns[len(tkns)-1], metadataTitle)
		return true
	}

	// Get the second enclosed group
	// e.g. "[sub][anime title][BDRIP][...]"
	for {
		// Get all opening brackets
		openingBracketTkns, found := p.tokenManager.tokens.filter(func(tkn *token) bool {
			return tkn.isOpeningBracket() && tkn.getValue() != "("
		})
		if !found {
			break // Next try
		}

		if len(openingBracketTkns) < 3 {
			break
		}

		// Get all unknown tokens between the second and third opening brackets
		// e.g. [ sub ] `[` -> "anime" -> "title" -> `]` [ ... ]
		tkns, found := p.tokenManager.tokens.walkAndCollecIf(
			p.tokenManager.tokens.getIndexOf(openingBracketTkns[1])+1,
			func(tkn *token) bool {
				return tkn.isUnknown() && !tkn.isKeyword() && !tkn.isSeparator()
			},
			func(tkn *token) bool {
				// Stop when we encounter the third opening bracket or a keyword
				return tkn.UUID == openingBracketTkns[2].UUID || tkn.isKeyword()
			})

		if !found {
			break // Next try
		}

		p.tokenManager.tokens.combineTitle(tkns[0], tkns[len(tkns)-1], metadataTitle)
		return true
	}

	return
}

// ---------------------------------------------------------------------------------------------------------------------
// Parse title
// ---------------------------------------------------------------------------------------------------------------------

// DEVOTE
// FIXME Evangelion 1.0 You Are [Not] Alone, fails because of "[Not]"'
// One solution would be to check if "[unknown]" is isolated, and if it is, add it to the title and continue until next opening bracket

func (p *parser) parseTitle() {

	// First try to find the title if all unknown tokens are enclosed
	if found := p.parseTitleIfAllEnclosed(); found {
		return // Stop if title found
	}

	// e.g. "[sub] anime title ep01"
	for {
		// Get first episode number token (first because we are going backwards, and it could be a range)
		found, epTkns := p.tokenManager.tokens.findWithMetadataCategory(metadataEpisodeNumber)
		if !found {
			break // Next try
		}
		firstEpTkn := epTkns[0]

		// 1st strategy: Get all unknown tokens before the episode number token until a closing bracket is found or SoF
		// e.g. "[sub `]` |- "anime title" <- `ep01`"
		// e.g. "|- "anime title" <- `ep01`"
		tkns, found := p.tokenManager.tokens.walkBackAndCollecIf(
			p.tokenManager.tokens.getIndexOf(firstEpTkn)-1,
			func(tkn *token) bool {
				return tkn.isUnknown() && !tkn.isKeyword() && !tkn.isSeparator()
			},
			func(tkn *token) bool {
				return tkn.isClosingBracket() && tkn.getValue() == "]"
			})
		if !found {
			break // Next try
		}

		// Check anime type in the title tokens
		// e.g. "Anime title OVA", we don't remove "OVA" from the title, but add it as an anime type
		for _, tkn := range tkns {
			if kwd, found := p.tokenManager.keywordManager.findStandaloneKeywordByValue(tkn.getValue()); found {
				if kwd.isAnimeType() { // if keyword is an anime type, add it to the metadata
					// set it directly to the struct because the token will be removed from the list
					p.metadata.AnimeType = append(p.metadata.AnimeType, tkn.getValue())
				}
			}
		}
		// Reverse the tokens because we were going backwards
		slices.Reverse(tkns)

		p.tokenManager.tokens.combineTitle(tkns[0], tkns[len(tkns)-1], metadataTitle)
		return

	}

	// e.g. [sub] anime title [...]
	for {
		// Get first non-enclosed token
		nonEnclosedTkns, found := p.tokenManager.tokens.filter(func(tkn *token) bool {
			return !tkn.isEnclosed() && tkn.isUnknown() && !tkn.isKeyword()
		})
		if !found {
			break // Next try
		}

		// Collect all unknown tokens from the first non-enclosed token until an opening bracket or keyword is found
		// e.g. "[ignored] collected collected [ignored]"
		tkns, found := p.tokenManager.tokens.walkAndCollecIf(
			p.tokenManager.tokens.getIndexOf(nonEnclosedTkns[0]),
			func(tkn *token) bool {
				return !tkn.isEnclosed() && // not enclosed
					tkn.isUnknown() && // unknown
					!tkn.isKeyword() // not a keyword
				//!tkn.isSeparator() // not a separator
			},
			func(tkn *token) bool {
				return (tkn.isOpeningBracket() && tkn.getValue() == "[") || tkn.isKeyword()
			})
		if !found {
			break // Next try
		}

		// Title should not be after file info metadata like 1080p
		if p.tokenManager.tokens.isTokenAfterFileMetadata(tkns[0]) {
			break // Next try
		}

		// Check anime type
		// e.g. "Anime title OVA", we don't remove "OVA" from the title, but add it as an anime type
		for _, tkn := range tkns {
			if kwd, found := p.tokenManager.keywordManager.findStandaloneKeywordByValue(tkn.getValue()); found {
				if kwd.isAnimeType() { // if keyword is an anime type, add it to the metadata
					// set it directly to the struct because the token will be removed from the list
					p.metadata.AnimeType = append(p.metadata.AnimeType, tkn.getValue())
				}
			}
		}

		p.tokenManager.tokens.combineTitle(tkns[0], tkns[len(tkns)-1], metadataTitle)
		return
	}

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (p *parser) parseReleaseGroup() {

	foundTitle, titleTkns := p.tokenManager.tokens.findWithMetadataCategory(metadataTitle)
	foundReleaseGroup, _ := p.tokenManager.tokens.findWithMetadataCategory(metadataReleaseGroup)
	if foundReleaseGroup {
		return // Stop if release group already found
	}

	// Handle case where all unknown tokens are enclosed
	for {
		if !foundTitle {
			break // Next try
		}

		titleTkn := titleTkns[0]

		// Get all unknown tokens before the title
		unknownTkns, found := p.tokenManager.tokens.walkAndCollecIf(
			0,
			func(tkn *token) bool {
				return tkn.isUnknown() && !tkn.isKeyword() && !tkn.isSeparator() && tkn.isEnclosed()
			},
			func(tkn *token) bool {
				return tkn.UUID == titleTkn.UUID
			})
		if !found {
			break // Next try
		}

		// Found release group
		if len(unknownTkns) == 1 {
			unknownTkns[0].setMetadataCategory(metadataReleaseGroup)
			return
		}

		// Found longer release group
		p.tokenManager.tokens.combineTitle(unknownTkns[0], unknownTkns[len(unknownTkns)-1], metadataReleaseGroup)
		return
	}

	// If we still haven't found a release group, try to find:
	// - the first enclosed group of unknown tokens going backwards
	for {

		// Get all closing brackets
		closingBracketTkns, found := p.tokenManager.tokens.filter(func(tkn *token) bool {
			return tkn.isClosingBracket() && tkn.getValue() != ")"
		})
		if !found {
			break // Next try
		}

		slices.Reverse(closingBracketTkns)

		// Go through all closing brackets going backwards
		// e.g.
		// 		-[ ... `]` <-(3) [              `]` <-(2) [ known `]` <-(1)
		// 		-[ ... `]`       [ |- "unknown" `]` <-(!) [ known `]`
		// 		Release group found: "unknown"
		for _, lastClosingBracket := range closingBracketTkns {

			// Get matching opening bracket going backwards
			// e.g. `[` <- ... <- ... <- ]
			lastOpeningBracket, found := p.tokenManager.tokens.getFirstOccurrenceBefore(
				p.tokenManager.tokens.getIndexOf(lastClosingBracket),
				func(tkn *token) bool {
					return tkn.isOpeningBracket() && isMatchingClosingBracket(tkn.getValue(), lastClosingBracket.getValue())
				})
			if !found {
				continue // Next try
			}

			// Get all tokens between the opening and closing brackets
			unknownTkns, found := p.tokenManager.tokens.walkAndCollecIf(
				p.tokenManager.tokens.getIndexOf(lastOpeningBracket)+1,
				func(tkn *token) bool {
					// Get all non-separator tokens
					return !tkn.isSeparator()
				},
				func(tkn *token) bool {
					return tkn.UUID == lastClosingBracket.UUID
				})
			if !found {
				continue // Next try
			}

			// If we only get a single character, then it's probably not a release group
			if len(unknownTkns) == 1 && len([]rune(unknownTkns[0].getValue())) == 1 {
				continue // Next try
			}
			// If we find a keyword, then it's probably not a release group
			if lo.ContainsBy(unknownTkns, func(tkn *token) bool {
				return tkn.isKeyword() || tkn.isCRC32()
			}) {
				continue // Next try
			}

			// Found release group
			if len(unknownTkns) == 1 {
				unknownTkns[0].setMetadataCategory(metadataReleaseGroup)
				return
			}

			// Found longer release group
			p.tokenManager.tokens.combineTitle(unknownTkns[0], unknownTkns[len(unknownTkns)-1], metadataReleaseGroup)
			return
		}

		break
	}

	for {
		// Get all unknown tokens
		unknownTkns, found := p.tokenManager.tokens.filter(func(tkn *token) bool {
			return tkn.isUnknown() && !tkn.isKeyword() && !tkn.isSeparator() && tkn.isWordKind()
		})
		if !found {
			break // Next try
		}

		unknownTkn := unknownTkns[0]
		if len(unknownTkns) > 1 {
			unknownTkn = unknownTkns[len(unknownTkns)-1]
		}

		if p.tokenManager.tokens.isTokenInFirstHalf(unknownTkn) {
			break
		}

		// Found release group
		unknownTkn.setMetadataCategory(metadataReleaseGroup)
		return
	}

}
