package seanime_parser

import (
	"bytes"
	"github.com/samber/lo"
	"slices"
)

var brackets = [][]rune{
	{'(', ')'},
	{'[', ']'},
	{'{', '}'},
	{'\u300C', '\u300D'},
	{'\u300E', '\u300F'},
	{'\u3010', '\u3011'},
	{'\uFF08', '\uFF09'},
}
var separators = []rune{'-', '+', '~', '&', '\u2010', '\u2011', '\u2012', '\u2013', '\u2014', '\u2015'}
var delimiters = []rune{' ', '_', '.', '|', ','}

// tokenize
//
// "[hello] world" -> "[", "hello", "]", " ", "world"
func tokenize(filename string) []*token {

	type runeVal struct {
		openingBracket rune
		closingBracket rune
		separator      rune
		delimiter      rune
		unknown        rune
	}

	// Get all runes
	runes := []rune(filename)

	runeVars := make([]*runeVal, 0)

	for _, r := range runes {
		// Identify brackets
		foundBracket := false
	inner:
		for _, bracketDuo := range brackets {
			if slices.Contains(bracketDuo, r) {
				if bracketDuo[0] == r {
					runeVars = append(runeVars, &runeVal{openingBracket: r})
				} else {
					runeVars = append(runeVars, &runeVal{closingBracket: r})
				}
				foundBracket = true
				break inner
			}
		}
		if foundBracket {
			continue
		}
		idx := lo.IndexOf(separators, r)
		if idx != -1 {
			runeVars = append(runeVars, &runeVal{separator: r})
			continue
		}
		idx = lo.IndexOf(delimiters, r)
		if idx != -1 {
			runeVars = append(runeVars, &runeVal{delimiter: r})
			continue
		}
		runeVars = append(runeVars, &runeVal{unknown: r})
	}

	rawTkns := make([]*token, 0)

	// Merge unknown runeVal
	unknownBuff := bytes.NewBufferString("")
	isUnknownSeq := false
	for _, r := range runeVars {
		if r.unknown != 0 {
			if isUnknownSeq {
				unknownBuff.WriteString(string(r.unknown))
			} else {
				unknownBuff.Reset()
				unknownBuff.WriteString(string(r.unknown))
				isUnknownSeq = true
			}
		} else {
			// End the unknown sequence
			if isUnknownSeq {
				isUnknownSeq = false
				rawTkns = append(rawTkns, newToken(unknownBuff.String()))
			}
			if r.separator != 0 {
				tkn := newToken(string(r.separator))
				tkn.Category = tokenCatSeparator
				rawTkns = append(rawTkns, tkn)
			}
			if r.delimiter != 0 {
				tkn := newToken(string(r.delimiter))
				tkn.Category = tokenCatDelimiter
				rawTkns = append(rawTkns, tkn)
			}
			if r.openingBracket != 0 {
				tkn := newToken(string(r.openingBracket))
				tkn.Category = tokenCatOpeningBracket
				rawTkns = append(rawTkns, tkn)
			}
			if r.closingBracket != 0 {
				tkn := newToken(string(r.closingBracket))
				tkn.Category = tokenCatClosingBracket
				rawTkns = append(rawTkns, tkn)
			}
		}
	}

	if isUnknownSeq {
		rawTkns = append(rawTkns, newToken(unknownBuff.String()))
	}

	enclosedTknsBuff := make([]*token, 0)
	openingBracket := ""

	// Find enclosed tokens
	for _, tkn := range rawTkns {
		if tkn.isOpeningBracket() {
			if openingBracket != "" {

			} else {
				openingBracket = tkn.getValue()
			}
			continue
		}

		// Count parenthesis as enclosed tokens if we are already in an enclosed token sequence
		// and if the opening bracket is not a parenthesis
		tknIsParenthesis := (tkn.getValue() == "(" || tkn.getValue() == ")") && openingBracket != "("

		if (tkn.isUnknown() || tknIsParenthesis) && openingBracket != "" {
			enclosedTknsBuff = append(enclosedTknsBuff, tkn)
			continue
		}

		// we encounter a closing bracket
		if tkn.isClosingBracket() && openingBracket != "" {
			// confirm that is it a matching closing bracket
			if isMatchingClosingBracket(openingBracket, tkn.getValue()) {
				openingBracket = ""

				// Update tokens
				for _, enclosedTkn := range enclosedTknsBuff {
					enclosedTkn.Enclosed = true
				}
				enclosedTknsBuff = make([]*token, 0)

				continue
			}
		}
	}

	identifyTokenKinds(rawTkns)

	return rawTkns

}

func identifyTokenKinds(tkns []*token) {
	for _, tkn := range tkns {

		if isCRC32(tkn.getValue()) {
			tkn.setKind(tokenKindCrc32)
			continue
		}

		if isResolution(tkn.getValue()) {
			tkn.setKind(tokenKindPossibleVideoRes)
			continue
		}

		if isNumber(tkn.getValue()) {
			if isYearNumber(tkn.getValue()) {
				tkn.setKind(tokenKindYear)
				continue
			}
			tkn.setKind(tokenKindNumber)
			continue
		}

		if isNumberLike(tkn.getValue()) {
			tkn.setKind(tokenKindNumberLike)
			continue
		}

		if isOrdinalNumber(tkn.getValue()) {
			tkn.setKind(tokenKindOrdinalNumber)
			continue
		}

		if len([]rune(tkn.getValue())) > 1 {
			tkn.setKind(tokenKindWord)
		} else {
			tkn.setKind(tokenKindCharacter)
		}

	}
}

func isMatchingClosingBracket(openingBracket string, bracket string) bool {
	for _, bracketDuo := range brackets {
		if string(bracketDuo[0]) == openingBracket {
			return string(bracketDuo[1]) == bracket
		}
	}
	return false
}

func isMatchingOpeningBracket(closingBracket string, bracket string) bool {
	for _, bracketDuo := range brackets {
		if string(bracketDuo[1]) == closingBracket {
			return string(bracketDuo[0]) == bracket
		}
	}
	return false
}
