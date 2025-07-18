package manga_providers

import (
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type ScannedChapterFile struct {
	Chapter      []string // can be a single chapter or a range of chapters
	MangaTitle   string   // typically comes before the chapter number
	ChapterTitle string   // typically comes after the chapter number
	Volume       []string // typically comes after the chapter number
	IsPDF        bool
}

type TokenType int

const (
	TokenUnknown TokenType = iota
	TokenText
	TokenNumber
	TokenKeyword
	TokenSeparator
	TokenEnclosed
	TokenFileExtension
)

// Token represents a parsed token from the filename
type Token struct {
	Type      TokenType
	Value     string
	Position  int
	IsChapter bool
	IsVolume  bool
}

// Lexer handles the tokenization of the filename
type Lexer struct {
	input        string
	position     int
	tokens       []Token
	currentToken int
}

var ChapterKeywords = []string{
	"ch", "chp", "chapter", "chap", "c",
}

var VolumeKeywords = []string{
	"v", "vol", "volume",
}

var SeparatorChars = []rune{
	' ', '-', '_', '.', '[', ']', '(', ')', '{', '}', '~',
}

var ImageExtensions = map[string]struct{}{
	".png":  {},
	".jpg":  {},
	".jpeg": {},
	".gif":  {},
	".webp": {},
	".bmp":  {},
	".tiff": {},
	".tif":  {},
}

// NewLexer creates a new lexer instance
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:        strings.TrimSpace(input),
		tokens:       make([]Token, 0),
		currentToken: 0,
	}
}

// Tokenize breaks down the input into tokens
func (l *Lexer) Tokenize() []Token {
	l.position = 0
	l.tokens = make([]Token, 0)

	for l.position < len(l.input) {
		if l.isWhitespace(l.current()) {
			l.skipWhitespace()
			continue
		}

		if l.isEnclosedStart(l.current()) {
			l.readEnclosed()
			continue
		}

		if l.isSeparator(l.current()) {
			l.readSeparator()
			continue
		}

		if l.isDigit(l.current()) {
			l.readNumber()
			continue
		}

		if l.isLetter(l.current()) {
			l.readText()
			continue
		}

		// Skip unknown characters
		l.position++
	}

	l.classifyTokens()
	return l.tokens
}

// current returns the current character
func (l *Lexer) current() rune {
	if l.position >= len(l.input) {
		return 0
	}
	return rune(l.input[l.position])
}

// peek returns the next character without advancing
func (l *Lexer) peek() rune {
	if l.position+1 >= len(l.input) {
		return 0
	}
	return rune(l.input[l.position+1])
}

// advance moves to the next character
func (l *Lexer) advance() {
	l.position++
}

// isWhitespace checks if character is whitespace
func (l *Lexer) isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// isSeparator checks if character is a separator
func (l *Lexer) isSeparator(r rune) bool {
	for _, sep := range SeparatorChars {
		if r == sep {
			return true
		}
	}
	return false
}

// isEnclosedStart checks if character starts an enclosed section
func (l *Lexer) isEnclosedStart(r rune) bool {
	return r == '[' || r == '(' || r == '{'
}

// isDigit checks if character is a digit
func (l *Lexer) isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// isLetter checks if character is a letter
func (l *Lexer) isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// skipWhitespace skips all whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.position < len(l.input) && l.isWhitespace(l.current()) {
		l.advance()
	}
}

// readEnclosed reads content within brackets/parentheses
func (l *Lexer) readEnclosed() {
	start := l.position
	openChar := l.current()
	var closeChar rune

	switch openChar {
	case '[':
		closeChar = ']'
	case '(':
		closeChar = ')'
	case '{':
		closeChar = '}'
	default:
		l.advance()
		return
	}

	l.advance() // Skip opening character
	startContent := l.position

	for l.position < len(l.input) && l.current() != closeChar {
		l.advance()
	}

	if l.position < len(l.input) {
		content := l.input[startContent:l.position]
		l.advance() // Skip closing character

		// Only add if content is meaningful
		if len(strings.TrimSpace(content)) > 0 {
			l.addToken(TokenEnclosed, content, start)
		}
	}
}

// readSeparator reads separator characters
func (l *Lexer) readSeparator() {
	start := l.position
	value := string(l.current())
	l.advance()
	l.addToken(TokenSeparator, value, start)
}

// readNumber reads numeric values (including decimals)
func (l *Lexer) readNumber() {
	start := l.position

	for l.position < len(l.input) && (l.isDigit(l.current()) || l.current() == '.') {
		// Stop if we hit a file extension
		if l.current() == '.' && l.position+1 < len(l.input) {
			// Check if this is followed by common file extensions
			remaining := l.input[l.position+1:]
			if strings.HasPrefix(remaining, "cbz") || strings.HasPrefix(remaining, "cbr") ||
				strings.HasPrefix(remaining, "pdf") || strings.HasPrefix(remaining, "epub") {
				break
			}
		}
		l.advance()
	}

	value := l.input[start:l.position]
	l.addToken(TokenNumber, value, start)
}

// readText reads alphabetic text
func (l *Lexer) readText() {
	start := l.position

	for l.position < len(l.input) && (l.isLetter(l.current()) || l.isDigit(l.current())) {
		l.advance()
	}

	value := l.input[start:l.position]
	lowerValue := strings.ToLower(value) // Use lowercase for keyword checking

	// Check if this might be a concatenated keyword that continues with a decimal
	if l.startsWithKeyword(lowerValue) && l.position < len(l.input) && l.current() == '.' {
		// Look ahead to see if there are more digits after the decimal
		tempPos := l.position + 1
		if tempPos < len(l.input) && l.isDigit(rune(l.input[tempPos])) {
			// Read the decimal part
			l.advance() // consume the '.'
			for l.position < len(l.input) && l.isDigit(l.current()) {
				l.advance()
			}
			// Update value to include decimal part
			value = l.input[start:l.position]
			lowerValue = strings.ToLower(value)
		}
	}

	// Check for concatenated keywords like "ch001", "c001", "chp001", "c12.5"
	if l.containsKeywordPrefix(lowerValue) {
		l.splitKeywordAndNumber(lowerValue, value, start) // Pass both versions
	} else {
		l.addToken(TokenText, value, start) // Use original case
	}
}

// startsWithKeyword checks if text starts with any known keyword
func (l *Lexer) startsWithKeyword(text string) bool {
	for _, keyword := range ChapterKeywords {
		if strings.HasPrefix(text, keyword) {
			return true
		}
	}
	for _, keyword := range VolumeKeywords {
		if strings.HasPrefix(text, keyword) {
			return true
		}
	}
	return false
}

// containsKeywordPrefix checks if text starts with a known keyword
func (l *Lexer) containsKeywordPrefix(text string) bool {
	chKeywords := ChapterKeywords
	// Sort by length descending to match longer keywords first
	slices.SortFunc(chKeywords, func(a, b string) int {
		return len(b) - len(a) // Sort by length descending
	})
	for _, keyword := range ChapterKeywords {
		if strings.HasPrefix(text, keyword) && len(text) > len(keyword) {
			remaining := text[len(keyword):]
			// Check if remaining part is numeric (including decimals)
			if len(remaining) == 0 {
				return false
			}
			return l.isValidNumberPart(remaining)
		}
	}
	for _, keyword := range VolumeKeywords {
		if strings.HasPrefix(text, keyword) && len(text) > len(keyword) {
			remaining := text[len(keyword):]
			// Check if remaining part is numeric (including decimals)
			if len(remaining) == 0 {
				return false
			}
			return l.isValidNumberPart(remaining)
		}
	}
	return false
}

// isValidNumberPart checks if string is valid number (including decimals)
func (l *Lexer) isValidNumberPart(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Don't allow starting with decimal
	if s[0] == '.' {
		return false
	}

	hasDecimal := false
	for _, r := range s {
		if r == '.' {
			if hasDecimal {
				return false // Multiple decimals not allowed
			}
			hasDecimal = true
		} else if !l.isDigit(r) {
			return false
		}
	}
	return true
}

// splitKeywordAndNumber splits concatenated keyword and number tokens
func (l *Lexer) splitKeywordAndNumber(lowerText, originalText string, position int) {
	for _, keyword := range ChapterKeywords {
		if strings.HasPrefix(lowerText, keyword) && len(lowerText) > len(keyword) {
			// Use original case for the keyword part
			originalKeyword := originalText[:len(keyword)]
			l.addKeywordToken(originalKeyword, position, true, false)

			// Extract number part (keeping original case/formatting)
			numberPart := originalText[len(keyword):]
			l.addToken(TokenNumber, numberPart, position+len(keyword))
			return
		}
	}
	for _, keyword := range VolumeKeywords {
		if strings.HasPrefix(lowerText, keyword) && len(lowerText) > len(keyword) {
			// Use original case for the keyword part
			originalKeyword := originalText[:len(keyword)]
			l.addKeywordToken(originalKeyword, position, false, true)

			// Extract number part (keeping original case/formatting)
			numberPart := originalText[len(keyword):]
			l.addToken(TokenNumber, numberPart, position+len(keyword))
			return
		}
	}
}

// addKeywordToken adds a keyword token with flags
func (l *Lexer) addKeywordToken(value string, position int, isChapter, isVolume bool) {
	l.tokens = append(l.tokens, Token{
		Type:      TokenKeyword,
		Value:     value,
		Position:  position,
		IsChapter: isChapter,
		IsVolume:  isVolume,
	})
}

// addToken adds a token to the list
func (l *Lexer) addToken(tokenType TokenType, value string, position int) {
	l.tokens = append(l.tokens, Token{
		Type:     tokenType,
		Value:    value,
		Position: position,
	})
}

// classifyTokens identifies chapter and volume keywords
func (l *Lexer) classifyTokens() {
	for i := range l.tokens {
		token := &l.tokens[i]

		// Check for chapter keywords (case insensitive)
		lowerValue := strings.ToLower(token.Value)
		for _, keyword := range ChapterKeywords {
			if lowerValue == keyword {
				token.Type = TokenKeyword
				token.IsChapter = true
				break
			}
		}

		// Check for volume keywords (case insensitive)
		for _, keyword := range VolumeKeywords {
			if lowerValue == keyword {
				token.Type = TokenKeyword
				token.IsVolume = true
				break
			}
		}

		// Check for file extensions
		if strings.Contains(lowerValue, "pdf") || strings.Contains(lowerValue, "cbz") ||
			strings.Contains(lowerValue, "cbr") || strings.Contains(lowerValue, "epub") {
			token.Type = TokenFileExtension
		}
	}
}

// Parser handles the semantic analysis of tokens
type Parser struct {
	tokens []Token
	result *ScannedChapterFile
}

// NewParser creates a new parser instance
func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
		result: &ScannedChapterFile{
			Chapter: make([]string, 0),
			Volume:  make([]string, 0),
		},
	}
}

// Parse performs semantic analysis on the tokens
func (p *Parser) Parse() *ScannedChapterFile {
	p.extractChapters()
	p.extractVolumes()
	p.extractTitles()
	p.checkPDF()

	return p.result
}

// extractChapters finds and extracts chapter numbers
func (p *Parser) extractChapters() {
	for i, token := range p.tokens {
		if token.IsChapter {
			// Look for numbers after chapter keyword
			for j := i + 1; j < len(p.tokens) && j < i+3; j++ {
				nextToken := p.tokens[j]
				if nextToken.Type == TokenNumber {
					p.addChapterNumber(nextToken.Value)
					break
				} else if nextToken.Type == TokenSeparator {
					continue
				} else {
					break
				}
			}
		} else if token.Type == TokenNumber && !token.IsVolume {
			// Standalone number might be a chapter
			if p.isLikelyChapterNumber(token, i) {
				p.addChapterNumber(token.Value)
			}
		}
	}

	// Handle ranges by looking for dash-separated numbers
	p.handleChapterRanges()
}

// handleChapterRanges processes chapter ranges like "1-2" or "001-002"
func (p *Parser) handleChapterRanges() {
	for i := 0; i < len(p.tokens)-2; i++ {
		if p.tokens[i].Type == TokenNumber &&
			p.tokens[i+1].Type == TokenSeparator && p.tokens[i+1].Value == "-" &&
			p.tokens[i+2].Type == TokenNumber {

			// Check if first number is already a chapter
			firstIsChapter := false
			for _, ch := range p.result.Chapter {
				if ch == p.tokens[i].Value {
					firstIsChapter = true
					break
				}
			}

			if firstIsChapter {
				// Add the second number as a chapter too
				p.result.Chapter = append(p.result.Chapter, p.tokens[i+2].Value)
			}
		}
	}
}

// extractVolumes finds and extracts volume numbers
func (p *Parser) extractVolumes() {
	for i, token := range p.tokens {
		if token.IsVolume {
			// Look for numbers after volume keyword
			for j := i + 1; j < len(p.tokens) && j < i+3; j++ {
				nextToken := p.tokens[j]
				if nextToken.Type == TokenNumber {
					p.result.Volume = append(p.result.Volume, nextToken.Value)
					break
				} else if nextToken.Type == TokenSeparator {
					continue
				} else {
					break
				}
			}
		}
	}
}

// extractTitles finds manga title and chapter title
func (p *Parser) extractTitles() {
	// Find first chapter keyword or number position
	chapterPos := -1
	for i, token := range p.tokens {
		if token.IsChapter || (token.Type == TokenNumber && p.isLikelyChapterNumber(token, i)) {
			chapterPos = i
			break
		}
	}

	if chapterPos > 0 {
		// Everything before chapter is likely manga title
		titleParts := make([]string, 0)
		for i := 0; i < chapterPos; i++ {
			token := p.tokens[i]
			if token.Type == TokenText && !token.IsVolume && !p.isIgnoredToken(token) {
				titleParts = append(titleParts, token.Value)
			} else if token.Type == TokenNumber && p.isNumberInTitle(token, i, chapterPos) {
				// Include numbers that are part of the title (but not volume indicators)
				titleParts = append(titleParts, token.Value)
			}
		}
		if len(titleParts) > 0 {
			p.result.MangaTitle = strings.Join(titleParts, " ")
		}

		// Look for chapter title after chapter number
		p.extractChapterTitle(chapterPos)
	} else {
		// No clear chapter indicator, check if this is a "number - title" pattern
		if len(p.result.Chapter) > 0 && p.hasChapterTitlePattern() {
			p.extractChapterTitleFromPattern()
		} else {
			// Treat most text as manga title
			p.extractFallbackTitle()
		}
	}
}

// hasChapterTitlePattern checks for "number - title" pattern
func (p *Parser) hasChapterTitlePattern() bool {
	for i := 0; i < len(p.tokens)-2; i++ {
		if p.tokens[i].Type == TokenNumber &&
			p.tokens[i+1].Type == TokenSeparator && p.tokens[i+1].Value == "-" &&
			i+2 < len(p.tokens) && p.tokens[i+2].Type == TokenText {
			return true
		}
	}
	return false
}

// extractChapterTitleFromPattern extracts title from "number - title" pattern
func (p *Parser) extractChapterTitleFromPattern() {
	for i := 0; i < len(p.tokens)-2; i++ {
		if p.tokens[i].Type == TokenNumber &&
			p.tokens[i+1].Type == TokenSeparator && p.tokens[i+1].Value == "-" {

			// Collect text after the dash
			titleParts := make([]string, 0)
			for j := i + 2; j < len(p.tokens); j++ {
				token := p.tokens[j]
				if token.Type == TokenText && !p.isIgnoredToken(token) {
					titleParts = append(titleParts, token.Value)
				} else if token.Type == TokenFileExtension {
					break
				}
			}
			if len(titleParts) > 0 {
				p.result.ChapterTitle = strings.Join(titleParts, " ")
			}
			break
		}
	}
}

// extractFallbackTitle extracts title when no clear chapter indicators
func (p *Parser) extractFallbackTitle() {
	titleParts := make([]string, 0)
	for _, token := range p.tokens {
		if token.Type == TokenText && !p.isIgnoredToken(token) {
			titleParts = append(titleParts, token.Value)
		}
	}
	if len(titleParts) > 0 {
		p.result.MangaTitle = strings.Join(titleParts, " ")
	}
}

// addChapterNumber adds a chapter number, handling ranges
func (p *Parser) addChapterNumber(value string) {
	// Check for range indicators in the surrounding tokens
	if strings.Contains(value, "-") {
		parts := strings.Split(value, "-")
		for _, part := range parts {
			if part != "" {
				p.result.Chapter = append(p.result.Chapter, strings.TrimSpace(part))
			}
		}
	} else {
		p.result.Chapter = append(p.result.Chapter, value)
	}
}

// isLikelyChapterNumber determines if a number token is likely a chapter
func (p *Parser) isLikelyChapterNumber(token Token, position int) bool {
	// If we already have chapters from keywords, be more strict
	if len(p.result.Chapter) > 0 {
		return false
	}

	// Check context - numbers at the start of filename are likely chapters
	if position < 3 {
		return true
	}

	// Check if preceded by common patterns
	if position > 0 {
		prevToken := p.tokens[position-1]
		if prevToken.Type == TokenSeparator && (prevToken.Value == "-" || prevToken.Value == " ") {
			return true
		}
	}

	return false
}

// isNumberInTitle determines if a number token should be part of the title
func (p *Parser) isNumberInTitle(token Token, position int, chapterPos int) bool {
	// Don't include numbers that are right before the chapter position
	if position == chapterPos-1 {
		return false
	}

	// Check if this number looks like it's associated with volume
	if position > 0 {
		prevToken := p.tokens[position-1]
		if prevToken.IsVolume {
			return false // This number belongs to volume
		}
	}

	// Small numbers (like 05, 2) that appear early in the title are likely part of title
	if position < 5 {
		if val := token.Value; len(val) <= 2 {
			// Check if this number looks like part of a title (e.g., "Title 05")
			return true
		}
	}
	return false
}

// isIgnoredToken checks if token should be ignored in titles
func (p *Parser) isIgnoredToken(token Token) bool {
	ignoredWords := []string{"digital", "group", "scan", "scans", "team", "raw", "raws"}
	for _, word := range ignoredWords {
		if token.Value == word {
			return true
		}
	}

	// Check for version indicators that shouldn't be in volume
	if strings.HasPrefix(token.Value, "v") && len(token.Value) > 1 {
		remaining := token.Value[1:]
		// If it's just "v" + digit, it might be version, not volume
		if len(remaining) > 0 && remaining[0] >= '0' && remaining[0] <= '9' {
			// Check context - if preceded by a number, it's likely a version
			return true
		}
	}

	return false
}

// checkPDF sets the PDF flag if file is a PDF
func (p *Parser) checkPDF() {
	for _, token := range p.tokens {
		if token.Type == TokenFileExtension && strings.Contains(token.Value, "pdf") {
			p.result.IsPDF = true
			break
		}
	}
}

// scanChapterFilename scans the filename and returns a chapter entry if it is a chapter.
func scanChapterFilename(filename string) (res *ScannedChapterFile, ok bool) {
	// Create lexer and tokenize
	lexer := NewLexer(filename)
	tokens := lexer.Tokenize()

	// Create parser and parse
	parser := NewParser(tokens)
	res = parser.Parse()

	return res, true
}

func isFileImage(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	_, ok := ImageExtensions[ext]
	return ok
}

// extractChapterTitle finds chapter title after chapter number
func (p *Parser) extractChapterTitle(startPos int) {
	// Skip to after chapter number
	numberPos := -1
	for i := startPos; i < len(p.tokens); i++ {
		if p.tokens[i].Type == TokenNumber {
			numberPos = i
			break
		}
	}

	if numberPos == -1 {
		return
	}

	// Look for dash separator followed by text
	for i := numberPos + 1; i < len(p.tokens); i++ {
		token := p.tokens[i]
		if token.Type == TokenSeparator && token.Value == "-" {
			// Found dash, collect text after it
			titleParts := make([]string, 0)
			for j := i + 1; j < len(p.tokens); j++ {
				nextToken := p.tokens[j]
				if nextToken.Type == TokenText && !p.isIgnoredToken(nextToken) {
					titleParts = append(titleParts, nextToken.Value)
				} else if nextToken.Type == TokenFileExtension {
					break
				}
			}
			if len(titleParts) > 0 {
				p.result.ChapterTitle = strings.Join(titleParts, " ")
			}
			break
		}
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ScannedPageFile struct {
	Number   float64
	Filename string
	Ext      string
}

func parsePageFilename(filename string) (res *ScannedPageFile, ok bool) {
	res = &ScannedPageFile{
		Filename: filename,
	}

	filename = strings.ToLower(filename)
	res.Ext = filepath.Ext(filename)
	filename = strings.TrimSuffix(filename, res.Ext)

	if len(filename) == 0 {
		return res, false
	}

	// Find number at the start
	// check if first rune is a digit
	numStr := ""
	if !unicode.IsDigit(rune(filename[0])) {
		// walk until non-digit
		for i := 0; i < len(filename); i++ {
			if !unicode.IsDigit(rune(filename[i])) && rune(filename[i]) != '.' {
				break
			}
			numStr += string(filename[i])
		}
		if len(numStr) > 0 {
			res.Number, _ = strconv.ParseFloat(numStr, 64)
			return res, true
		}
	}

	// walk until first digit
	numStr = ""
	firstDigitIdx := strings.IndexFunc(filename, unicode.IsDigit)
	if firstDigitIdx != -1 {
		numStr += string(filename[firstDigitIdx])
		// walk until first non-digit or end
		for i := firstDigitIdx + 1; i < len(filename); i++ {
			if !unicode.IsDigit(rune(filename[i])) && rune(filename[i]) != '.' {
				break
			}
			numStr += string(filename[i])
		}
		if len(numStr) > 0 {
			res.Number, _ = strconv.ParseFloat(numStr, 64)
			return res, true
		}
	}

	return res, false
}
