package seanime_parser

import (
	"fmt"
	"strings"
)

type tokens []*token

type tokenManager struct {
	tokens         *tokens
	keywordManager *keywordManager
	filename       string
}

func newTokenManager(filename string) *tokenManager {
	tm := tokenManager{
		tokens:         &tokens{},
		filename:       filename,
		keywordManager: newKeywordManager(),
	}

	tm.tokens.setTokens(tokenize(strings.TrimSpace(filename)))

	tm.mergeDecimals()

	return &tm
}

func (tm *tokenManager) mergeDecimals() {
	for _, tkn := range *tm.tokens {
		if !tkn.isNumberKind() {
			continue
		}

		_, _ = tm.tokens.checkNumberWithDecimal(tkn)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// combineTitle combines all tokens between tknBegin and tknEnd into a single, well formatted title token.
// e.g. "Violet" "." "Evergarden" -> "Violet Evergarden"
func (t *tokens) combineTitle(tknBegin *token, tknEnd *token, category metadataCategory) (*token, bool) {
	// Get all the tokens between tknBeing and tknEnd
	// If all delimiters are the same, replace with space
	// If all delimiters are different, keep the minority the same
	tkns, found := t.getFromToInc(t.getIndexOf(tknBegin), t.getIndexOf(tknEnd))
	if !found {
		return nil, false
	}

	tknsIncludeOpeningParenthesis := false
	for _, tkn := range tkns {
		if tkn.getValue() == "(" {
			tknsIncludeOpeningParenthesis = true
		}
	}

	for _, tkn := range tkns {
		if (tkn.isOpeningBracket() || tkn.isClosingBracket()) && tkn.getValue() != "(" && tkn.getValue() != ")" {
			tkn.setValue("")
		}
	}

	// check if next token is closing parenthesis
	if nextTkn, found, _ := t.getTokenAfterSD(tknEnd); found &&
		nextTkn.isClosingBracket() &&
		nextTkn.getValue() == ")" &&
		tknsIncludeOpeningParenthesis {
		tknEnd = nextTkn
		tkns = append(tkns, nextTkn)
	}

	// Check if all delimiters are the same
	delimiters := make(map[string]int)
	for _, tkn := range tkns {
		if tkn.isDelimiter() {
			delimiters[tkn.getValue()]++
		}
	}
	if len(delimiters) == 1 {
		// Replace all delimiters with space
		for _, tkn := range tkns {
			if tkn.isDelimiter() {
				tkn.setValue(" ")
			}
		}
	} else {
		for _, tkn := range tkns {
			if tkn.isDelimiter() {
				// Replace delimiter with space if it's the majority
				if delimiters[tkn.getValue()] > len(tkns)/2 {
					tkn.setValue(" ")
				}
			}
		}
	}

	for _, tkn := range tkns {
		if tkn.getValue() == "_" {
			tkn.setValue(" ")
		}
	}

	allValues := ""
	for _, tkn := range tkns {
		allValues += tkn.getValue()
	}
	combinedTkn := token{
		UUID:             tknBegin.UUID,
		Value:            allValues,
		Kind:             tokenKindWord,
		Category:         tokenCatKnown,
		MetadataCategory: category,
		Enclosed:         tknBegin.isEnclosed(),
	}
	t.overwriteAt(t.getIndexOf(tknBegin), combinedTkn)

	start := t.getIndexOf(tknBegin) + 1
	end := t.getIndexOf(tknEnd)
	*t = append((*t)[:start], (*t)[end+1:]...)

	return &combinedTkn, true
}

// checkNumberWithDecimal checks if a token (number) is followed by a decimal point and a number.
// If it is, it will merge the tokens into a single token and return it.
func (t *tokens) checkNumberWithDecimal(tkn *token) (*token, bool) {
	if tkn == nil || !tkn.isNumberKind() {
		return nil, false
	}

	// Check if token is followed by a decimal point and a number
	dotTkn, ok := t.getTokenAfter(tkn)
	if !ok || !dotTkn.isDotDelimiter() {
		return nil, false
	}

	numTkn, ok := t.getTokenAfter(dotTkn)
	if !ok || !numTkn.isNumberKind() {
		return nil, false
	}

	delTkn, ok := t.getTokenAfter(numTkn)
	if (!ok || !(delTkn.isDelimiter() || !delTkn.isOpeningBracket() || !delTkn.isClosingBracket() || !delTkn.isSeparator())) && !t.isLastToken(numTkn) { // Delimiter or end of tokens
		return nil, false
	}

	// Merge tokens
	tkn.setValue(tkn.getValue() + "." + numTkn.getValue())
	tkn.setKind(tokenKindNumberLike)

	// Remove dot and number tokens
	t.removeAt(t.getIndexOf(dotTkn))
	t.removeAt(t.getIndexOf(numTkn))

	return tkn, true
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// getTokenAfter returns the token that comes after the specified token, along with a boolean indicating if the token was found.
func (t *tokens) getTokenAfter(tkn *token) (*token, bool) {
	index := t.getIndexOf(tkn)
	if index == -1 {
		return nil, false
	}
	return t.getAtSafe(index + 1)
}

// getTokenAfterSD returns the token that comes after the specified token, along with a boolean indicating if the token was found.
// It searches for the next non-delimiter token in the tokens slice starting from the index of the specified token.
// It also returns the number of skipped delimiter tokens.
func (t *tokens) getTokenAfterSD(tkn *token) (*token, bool, int) {
	index := t.getIndexOf(tkn)
	if index == -1 {
		return nil, false, 0
	}
	skipped := 0
	for i := index + 1; i < len(*t); i++ {
		if !(*t)[i].isDelimiter() {
			return (*t)[i], true, skipped
		} else {
			skipped++
		}
	}
	return nil, false, skipped
}

// getTokenBefore returns the token that comes before the specified token, along with a boolean indicating if the token was found.
func (t *tokens) getTokenBefore(tkn *token) (*token, bool) {
	index := t.getIndexOf(tkn)
	if index == -1 {
		return nil, false
	}
	return t.getAtSafe(index - 1)
}

// getTokenBeforeSD returns the token that comes before the specified token, along with a boolean indicating if the token was found.
// It searches for the previous non-delimiter token in the tokens slice starting from the index of the specified token.
// It also returns the number of skipped delimiter tokens.
func (t *tokens) getTokenBeforeSD(tkn *token) (*token, bool, int) {
	index := t.getIndexOf(tkn)
	if index == -1 {
		return nil, false, 0
	}
	skipped := 0
	for i := index - 1; i >= 0; i-- {
		if !(*t)[i].isDelimiter() {
			return (*t)[i], true, skipped
		} else {
			skipped++
		}
	}
	return nil, false, skipped
}

// isBetweenParentheses checks if the specified token is between parentheses
func (t *tokens) isBetweenParentheses(tkn *token) bool {
	index := t.getIndexOf(tkn)
	if index == -1 {
		return false
	}
	prevTkn, found, _ := t.getTokenBeforeSD(tkn)
	leftP := found && prevTkn.isOpeningBracket() && prevTkn.getValue() == "("
	nextTkn, found, _ := t.getTokenAfterSD(tkn)
	rightP := found && nextTkn.isClosingBracket() && nextTkn.getValue() == ")"
	return leftP && rightP
}

// isIsolated checks if the specified token is surrounded by delimiters that are not "."
// e.g. tkn.Value = 01
// e.g. " 01.mkv" -> true, " 01[" -> false, " 01 " -> true
// e.g. "1.01 " -> false
func (t *tokens) isIsolated(tkn *token) bool {
	index := t.getIndexOf(tkn)
	if index == -1 {
		return false
	}

	prevTkn, found := t.getTokenBefore(tkn)
	// Previous token should be non-existent OR a delimiter that is not "."
	isolatedOnTheLeft := !found || (prevTkn.isDelimiter() && prevTkn.getValue() != ".")

	if found {
		prevPrevTkn, found := t.getTokenBefore(prevTkn)
		// Previous previous token should be non-existent OR a non-number token
		isolatedOnTheLeft = !found || (prevTkn.getValue() == "." && !prevPrevTkn.isNumberKind()) || prevTkn.getValue() != "."
	}

	nextTkn, found := t.getTokenAfter(tkn)
	isolatedOnTheRight := !found || (nextTkn.isDelimiter() || nextTkn.isOpeningBracket())
	return isolatedOnTheLeft && isolatedOnTheRight
}

// isTokenInFirstHalf checks if the specified token is in the first half of the tokens list.
// It returns true if the token is found and its index is less than or equal to half the length of the list,
// otherwise it returns false.
func (t *tokens) isTokenInFirstHalf(tkn *token) bool {
	index := t.getIndexOf(tkn)
	if index == -1 {
		return false
	}
	return index <= len(*t)/2
}

// isTokenAfterFileMetadata checks if the specified token comes after file info metadata
// deprecated
func (t *tokens) isTokenAfterFileMetadata(tkn *token) bool {
	//return false
	index := t.getIndexOf(tkn)
	if index == -1 {
		return false
	}
	isAfter := false

	for idx, _tkn := range *t {
		// Check if token is after file info metadata token
		// and if the file info token is not in the first half of the tokens list
		if _tkn.isFileInfoMetadata() && idx != index && idx < index && !t.isTokenInFirstHalf(_tkn) {
			isAfter = true
		}
	}

	return isAfter
}

func (t *tokens) getIndexOf(tkn *token) int {
	for i, _tkn := range *t {
		if _tkn.UUID == tkn.UUID {
			return i
		}
	}
	return -1
}

func (t *tokens) isLastToken(tkn *token) bool {
	if tkn == nil {
		return false
	}
	return t.getIndexOf(tkn) == len(*t)-1
}

func (t *tokens) isFirstToken(tkn *token) bool {
	if tkn == nil {
		return false
	}
	return t.getIndexOf(tkn) == 0
}

// e.g. "-{tkn}" or "- {tkn}
func (t *tokens) foundDashSeparatorBefore(tkn *token) bool {
	// Check if token before previous token is a dash separator
	if prevPrevTkn, found, _ := t.getTokenBeforeSD(tkn); found {
		if prevPrevTkn.isDashSeparator() {
			return true
		}
	}
	return false
}

// e.g. "{tkn}-" or "{tkn}-
func (t *tokens) foundDashSeparatorAfter(tkn *token) bool {
	// Check if token before previous token is a dash separator
	if prevPrevTkn, found, _ := t.getTokenAfterSD(tkn); found {
		if prevPrevTkn.isDashSeparator() {
			return true
		}
	}
	return false
}

// e.g. "01-{tkn}" or "1 ~ {tkn}"
func (t *tokens) checkEpisodeRangeBefore(tkn *token) ([]*token, bool) {
	tkns, found, nSkipped := t.getCategorySequenceBefore(t.getIndexOf(tkn), []tokenCategory{
		tokenCatSeparator,
		tokenCatUnknown,
	}, true)
	if !found || !tkns[1].isNumberOrLikeKind() {
		return nil, false
	}
	if tkns[1].isKeyword() {
		return nil, false
	}
	// Avoid this case "11 - {tkn}"
	// Unless the number is zero padded e.g. "01 - {tkn}"
	if nSkipped > 0 && tkns[0].isDashSeparator() && !isNumberZeroPadded(tkns[1].getValue()) {
		return nil, false
	}
	return tkns, true
}

// e.g. "01-{tkn}" or "1 - {tkn}"
// When rangeWithDelimiters is true, the function will ignore delimiters when checking for a number range
// So, "01 - {tkn}" and "01-{tkn} will return true,
// When it's false, the function will return false for "01 - {tkn}" and true for "01-{tkn}
//
// Returns [0] separator, [1] number or false
func (t *tokens) checkNumberRangeBefore(tkn *token, rangeWithDelimiters bool) ([]*token, bool) {
	tkns, found, _ := t.getCategorySequenceBefore(t.getIndexOf(tkn), []tokenCategory{
		tokenCatSeparator,
		tokenCatUnknown,
	}, rangeWithDelimiters)
	if !found || !tkns[1].isNumberOrLikeKind() {
		return nil, false
	}
	if tkns[1].isKeyword() {
		return nil, false
	}
	return tkns, true
}

// e.g. "{tkn}-02" or "{tkn} - 02"
// When rangeWithDelimiters is true, the function will ignore delimiters when checking for a number range
// So, "01 - {tkn}" and "01-{tkn} will return true,
// When it's false, the function will return false for "01 - {tkn}" and true for "01-{tkn}
//
// Returns [0] separator, [1] number or false
func (t *tokens) checkNumberRangeAfter(tkn *token, rangeWithDelimiters bool) ([]*token, bool) {
	tkns, found, _ := t.getCategorySequenceAfter(t.getIndexOf(tkn), []tokenCategory{
		tokenCatSeparator,
		tokenCatUnknown,
	}, rangeWithDelimiters)
	if !found || !tkns[1].isNumberKind() {
		return nil, false
	}
	return tkns, true
}

// e.g. "[abc][def][ghi].mkv"
func (t *tokens) allUnknownTokensAreEnclosed() bool {
	for _, tkn := range *t {
		if tkn.isFileExt() {
			continue
		}
		if tkn.isUnknown() && !tkn.isEnclosed() && len([]rune(tkn.getValue())) > 1 {
			return false
		}
	}
	return true
}

func (t *tokens) foundFileInfoMetadata() bool {
	for _, tkn := range *t {
		if tkn.isFileInfoMetadata() {
			return true
		}
	}
	return false
}

// collectUntil collects all tokens encountered until `pred` is true
func (t *tokens) collectUntil(start int, pred func(tkn *token) bool) ([]*token, bool) {
	if start+1 > len(*t)-1 {
		return nil, false
	}
	collec := make([]*token, 0)
	for idx, tkn := range (*t)[start:] {
		if pred(tkn) {
			break
		}
		// check if it's the end
		if fileExtTkn, found := t.getAtSafe(idx + 1); found && fileExtTkn.isFileExt() {
			break
		}
		collec = append(collec, tkn)
	}
	if len(collec) == 0 {
		return nil, false
	}
	return collec, true
}

// walkAndCollecIf collects tokens that satisfy `pred` until `stopIf` returns true
//
// Example: Walk until the end
//
//	tkns, found := walkAndCollecIf(0, func(tkn){ return tkn.isUnknown() }, func(tkn) { return false })
func (t *tokens) walkAndCollecIf(start int, pred func(tkn *token) bool, stopIf func(tkn *token) bool) ([]*token, bool) {
	if start+1 > len(*t)-1 {
		return nil, false
	}
	collec := make([]*token, 0)
	for _, tkn := range (*t)[start:] {
		if stopIf(tkn) {
			break
		}
		if pred(tkn) {
			collec = append(collec, tkn)
		}
	}
	if len(collec) == 0 {
		return nil, false
	}
	return collec, true
}

func (t *tokens) walkBackAndCollecIf(start int, pred func(tkn *token) bool, stopIf func(tkn *token) bool) ([]*token, bool) {
	if start-1 < 0 {
		return nil, false
	}
	collec := make([]*token, 0)
	for i := start; i >= 0; i-- {
		tkn := (*t)[i]
		if stopIf(tkn) {
			break
		}
		if pred(tkn) {
			collec = append(collec, tkn)
		}
	}
	if len(collec) == 0 {
		return nil, false
	}
	return collec, true
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (t *tokens) setTokens(tkns []*token) {
	*t = tkns
}

func (t *tokens) insertAt(index int, tkn token) {
	if index < 0 || index > len(*t) {
		return
	}
	*t = append((*t)[:index], append([]*token{&tkn}, (*t)[index:]...)...)
}

func (t *tokens) insertAtEnd(tkn token) {
	*t = append(*t, &tkn)
}

func (t *tokens) insertAtStart(tkn token) {
	*t = append([]*token{&tkn}, *t...)
}

func (t *tokens) insertManyAt(index int, tkns []*token) {
	if index < 0 || index > len(*t) {
		return
	}
	*t = append((*t)[:index], append(tkns, (*t)[index:]...)...)
}

func (t *tokens) insertAfter(index int, tkn token) {
	if index < 0 || index > len(*t) {
		return
	}
	*t = append((*t)[:index+1], append([]*token{&tkn}, (*t)[index+1:]...)...)
}

func (t *tokens) insertManyAfter(index int, tkns []*token) {
	if index < 0 || index > len(*t) {
		return
	}
	*t = append((*t)[:index+1], append(tkns, (*t)[index+1:]...)...)
}

func (t *tokens) removeAt(index int) {
	if index < 0 || index > len(*t) {
		return
	}
	*t = append((*t)[:index], (*t)[index+1:]...)
}

func (t *tokens) overwriteAt(index int, tkn token) {
	(*t)[index] = &tkn
}

func (t *tokens) overwriteManyAt(index int, tkns []*token) {
	*t = append((*t)[:index], append(tkns, (*t)[index+len(tkns):]...)...)
}

func (t *tokens) overwriteAndInsertManyAt(index int, tkns []*token) {
	*t = append((*t)[:index], (*t)[index+1:]...)
	// Then insert new elements at index
	// append takes a slice and follows that with a variadic parameter hence the need for ...
	*t = append((*t)[:index], append(tkns, (*t)[index:]...)...)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (t *tokens) getAtSafe(index int) (*token, bool) {
	if index < 0 || index > len(*t)-1 {
		return nil, false
	}
	return (*t)[index], true
}
func (t *tokens) getAt(index int) *token {
	return (*t)[index]
}

func (t *tokens) getFromUUID(uuid string) *token {
	for _, tkn := range *t {
		if tkn.UUID == uuid {
			return tkn
		}
	}
	return nil
}

func (t *tokens) getFromUUIDSafe(uuid string) (*token, bool) {
	for _, tkn := range *t {
		if tkn.UUID == uuid {
			return tkn, true
		}
	}
	return nil, false
}

func (t *tokens) getFromUUIDs(uuids []string) []*token {
	tkns := make([]*token, 0)
	for _, uuid := range uuids {
		tkn := t.getFromUUID(uuid)
		if tkn != nil {
			tkns = append(tkns, tkn)
		}
	}
	return tkns
}

func (t *tokens) getFrom(index int) []*token {
	if index < 0 || index > len(*t) {
		return []*token{}
	}
	return (*t)[index:]
}

func (t *tokens) getTo(index int) []*token {
	if index < 0 || index > len(*t) {
		return []*token{}
	}
	return (*t)[:index]
}

func (t *tokens) getToInc(index int) []*token {
	if index < 0 || index+1 > len(*t) {
		return []*token{}
	}
	return (*t)[:index+1]
}

func (t *tokens) getFromTo(start int, end int) ([]*token, bool) {
	// check indices
	if start < 0 || end < 0 || start > end || start > len(*t) || end > len(*t) {
		return []*token{}, false
	}
	return (*t)[start:end], true
}

func (t *tokens) getFromToInc(start int, end int) ([]*token, bool) {
	// check indices
	if start < 0 || end < 0 || start > end || start+1 > len(*t) || end+1 > len(*t) {
		return []*token{}, false
	}
	return (*t)[start : end+1], true
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (t *tokens) filter(pred func(tkn *token) bool) ([]*token, bool) {
	collec := make([]*token, 0)
	for _, tkn := range *t {
		if pred(tkn) {
			collec = append(collec, tkn)
		}
	}
	if len(collec) == 0 {
		return nil, false
	}
	return collec, true
}

func (t *tokens) getFirstOccurrenceAfter(start int, pred func(tkn *token) bool) (*token, bool) {
	if start < 0 {
		start = -1
	}
	if start+1 > len(*t) {
		return nil, false
	}
	for i := start + 1; i < len(*t); i++ {
		if pred((*t)[i]) {
			return (*t)[i], true
		}
	}
	return nil, false
}

func (t *tokens) getFirstOccurrenceBefore(start int, pred func(tkn *token) bool) (*token, bool) {
	if start > len(*t) {
		start = len(*t) + 1
	}
	if start < 0 {
		return nil, false
	}
	for i := start - 1; i >= 0; i-- {
		if pred((*t)[i]) {
			return (*t)[i], true
		}
	}
	return nil, false
}

// getCategorySequenceAfter returns the sequence of tokens in the given categories after the specified start index,
// along with a boolean indicating if the sequence was found.
// The skipDelimiters parameter determines whether to skip delimiter tokens when collecting the sequence.
func (t *tokens) getCategorySequenceAfter(start int, categories []tokenCategory, skipDelimiters bool) ([]*token, bool, int) {
	if start < 0 {
		start = -1
	}
	if start+1 > len(*t) {
		return []*token{}, false, 0
	}

	nbSkipped := 0

	var collec []*token
	var cursor int
	for i := start + 1; i < len(*t); i++ {
		if len(collec) == len(categories) {
			break
		}
		if skipDelimiters && (*t)[i].isDelimiter() {
			nbSkipped += 1
			continue
		}
		if (*t)[i].isCategory(categories[cursor]) {
			collec = append(collec, (*t)[i])
			cursor++
		} else {
			break
		}
	}

	if len(collec) == len(categories) {
		return collec, true, nbSkipped
	}

	return []*token{}, false, 0
}

func (t *tokens) getCategorySequenceAfterInc(start int, categories []tokenCategory, skipDelimiters bool) ([]*token, bool, int) {
	return t.getCategorySequenceAfter(start-1, categories, skipDelimiters)
}

// getCategorySequenceBefore returns the sequence of tokens in the given categories before the specified start index,
// along with a boolean indicating if the sequence was found.
// The skipDelimiters parameter determines whether to skip delimiter tokens when collecting the sequence.
func (t *tokens) getCategorySequenceBefore(start int, categories []tokenCategory, skipDelimiters bool) ([]*token, bool, int) {
	if start > len(*t) {
		start = len(*t) + 1
	}
	if start < 0 {
		return []*token{}, false, 0
	}

	nbSkipped := 0

	var collec []*token
	var cursor int
	for i := start - 1; i >= 0; i-- {
		if len(collec) == len(categories) {
			break
		}
		if skipDelimiters && (*t)[i].isDelimiter() {
			nbSkipped += 1
			continue
		}
		if (*t)[i].isCategory(categories[cursor]) {
			collec = append(collec, (*t)[i])
			cursor++
		} else {
			break
		}
	}

	if len(collec) == len(categories) {
		return collec, true, nbSkipped
	}

	return []*token{}, false, 0
}

func (t *tokens) getCategorySequenceBeforeInc(start int, categories []tokenCategory, skipDelimiters bool) ([]*token, bool, int) {
	return t.getCategorySequenceBefore(start+1, categories, skipDelimiters)
}

func (t *tokens) iterate(iterationFunc func(tkn *token, idx int)) {
	for idx, tkn := range *t {
		iterationFunc(tkn, idx)
	}
}

////////////////////

func (t *tokens) peekValuesAfter(start int, strs []string) ([]*token, bool) {

	if start+1+len(strs) > len(*t) {
		return nil, false
	}

	_tkns := (*t)[start+1 : start+1+len(strs)]

	var collec []*token
	for i := 0; i < len(strs); i++ {
		if strings.ToUpper(_tkns[i].getValue()) == strings.ToUpper(strs[i]) {
			uuid, ok := t.getFromUUIDSafe(_tkns[i].UUID)
			if !ok {
				break
			}
			collec = append(collec, uuid)
		} else {
			break
		}
	}

	if len(collec) == len(strs) {
		return collec, true
	}

	return nil, false
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (t *tokens) findWithMetadataCategory(cat metadataCategory) (bool, []*token) {
	_tkns := make([]*token, 0)
	for _, tkn := range *t {
		if tkn.MetadataCategory == cat {
			_tkns = append(_tkns, tkn)
		}
	}
	if len(_tkns) > 0 {
		return true, _tkns
	}
	return false, nil
}

func (t *tokens) findWithTokenCategory(cat tokenCategory) (bool, []*token) {
	_tkns := make([]*token, 0)
	for _, tkn := range *t {
		if tkn.isCategory(cat) {
			_tkns = append(_tkns, tkn)
		}
	}
	if len(_tkns) > 0 {
		return true, _tkns
	}
	return false, nil
}

func (t *tokens) findWithKeywordCategory(cat keywordCategory) (bool, []*token) {
	_tkns := make([]*token, 0)
	for _, tkn := range *t {
		if tkn.IdentifiedKeywordCategory == cat && tkn.isKeyword() {
			_tkns = append(_tkns, tkn)
		}
	}
	if len(_tkns) > 0 {
		return true, _tkns
	}
	return false, nil
}

func (t *tokens) sPrint() string {
	str := "["
	for idx, tkn := range *t {
		str += "\"" + tkn.getValue()
		if idx < len(*t)-1 {
			str += "\", "
		} else {
			str += "\""
		}
	}
	str += "]"
	return str
}

func (t *tokens) Sdump() string {
	str := "\n"
	for _, tkn := range *t {
		str += fmt.Sprintf("%-12s\t%v, kw: %v, %v, m: %v, enclosed: %v\n",
			"\""+tkn.getValue()+"\"",
			tkn.getCategory(),
			tkn.IdentifiedKeywordCategory,
			tkn.getKind(),
			tkn.MetadataCategory,
			tkn.isEnclosed(),
		)
	}
	str += "\n"
	return str
}
