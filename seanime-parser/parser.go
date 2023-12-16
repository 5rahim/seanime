package seanime_parser

import (
	"github.com/samber/lo"
	"strings"
)

type parser struct {
	filename     string
	tokenManager *tokenManager
	metadata     *Metadata
}

func newParser(filename string) *parser {
	return &parser{
		filename:     filename,
		tokenManager: newTokenManager(filename),
		metadata:     &Metadata{},
	}
}

func Parse(filename string) *Metadata {

	p := newParser(filename)
	p.parse()
	p.cleanUp()

	return p.metadata

}

func ParseAndDebug(filename string) (*Metadata, *tokens) {

	p := newParser(filename)
	p.parse()
	p.cleanUp()

	return p.metadata, p.tokenManager.tokens

}

func (p *parser) parse() {

	p.parseFileExtension()

	p.parseKeywords("normal")

	p.parseSeason()

	p.parseEpisode()

	p.parseEpisodeTitle()

	p.parseTitle()

	p.parseReleaseGroup()

	p.parseKeywords("")

	p.writeFormattedTitle()

	p.collectMetadata()

}

func (p *parser) parseFileExtension() {

	if *p.tokenManager.tokens == nil || len(*p.tokenManager.tokens) < 2 {
		return
	}
	// Get last token
	lastTkn := (*p.tokenManager.tokens)[len(*p.tokenManager.tokens)-1]
	dotTkn := (*p.tokenManager.tokens)[len(*p.tokenManager.tokens)-2]
	kwd, found := p.tokenManager.keywordManager.findStandaloneKeywordByValue(lastTkn.getValue())
	if dotTkn.getValue() == "." && found && (kwd.isFileExtension() || kwd.isVideoTerm()) {
		lastTkn.setMetadataCategory(metadataFileExtension)
	}
}

func (p *parser) parseKeywords(priority string) {

	for _, tkn := range *p.tokenManager.tokens {

		if tkn.isKeyword() || !tkn.isUnknown() { // Don't bother if token is already a keyword
			continue // Skip to next token
		}

		// Identify keyword
		_ = p.identifyKeyword(tkn, priority)

	}

}

// identifyKeyword identifies STANDALONE and multi-PART keywords for the given token
func (p *parser) identifyKeyword(tkn *token, priority string) bool {

	if tkn.Kind == tokenKindCrc32 {
		tkn.setIdentifiedKeywordCategory(keywordCatFileChecksum, keywordKindStandalone)
		return true
	}

	if tkn.Kind == tokenKindPossibleVideoRes {
		tkn.setIdentifiedKeywordCategory(keywordCatVideoResolution, keywordKindStandalone)
		return true
	}

	if tkn.Kind == tokenKindYear && tkn.isEnclosed() {
		tkn.setIdentifiedKeywordCategory(keywordCatYear, keywordKindStandalone)
		return true
	}

	// Check if token is a known pre-defined keyword prefix (e.g. "Blu" for "Blu-ray")
	keywordParts, found := p.tokenManager.keywordManager.findKeywordPartGroups(tkn.getValue())
	foundParts := false
	if found {
		foundParts = false
		for _, keywordGroup := range keywordParts {
			if retTkns, found := p.tokenManager.tokens.peekValuesAfter(p.tokenManager.tokens.getIndexOf(tkn), keywordGroup.seqParts); found {
				// Update token value
				seqPartsStr := ""
				for _, t := range retTkns {
					seqPartsStr += t.getValue()
				}
				tkn.setValue(mergeValues(tkn.getValue(), []string{seqPartsStr}))
				tkn.setIdentifiedKeywordCategory(keywordGroup.category, keywordKindStandalone)
				tkn.setKind(tokenKindWord)
				// Remove subsequent tokens
				for _, retTkn := range retTkns {
					p.tokenManager.tokens.removeAt(p.tokenManager.tokens.getIndexOf(retTkn))
				}
				foundParts = true
				break
			}
		}
	}

	if foundParts {
		return true
	}

	// Check if token is a known pre-defined STANDALONE keyword (e.g. "60FPS")
	if len(tkn.getValue()) > 1 {
		if keyword, found := p.tokenManager.keywordManager.findStandaloneKeywordByValue(tkn.getValue()); found {

			// When the priority is "normal", we only want to identify STANDALONE keywords that are not an anime type
			// That is because those are prone to false positives
			if priority == "normal" && keyword.isAnimeType() {
				return false
			}

			// when the priority is "normal", we only want to identify STANDALONE keywords that are not ambiguous
			// e.g. Do not flag "Ita" as a keyword in "Bokura Ga Ita"
			if priority == "normal" && p.tokenManager.keywordManager.isKeywordAmbiguous(keyword) {
				return false
			}

			tkn.setIdentifiedKeywordCategory(keyword.category, keyword.kind)
			return true
		}
	}

	return false

}

// collectMetadata collects the metadata elements from the parsed tokens.
// de-duplicates elements
func (p *parser) collectMetadata() {

	p.metadata.FileName = p.filename

	for _, tkn := range *p.tokenManager.tokens {

		switch tkn.IdentifiedKeywordCategory {
		case keywordCatYear:
			p.metadata.Year = tkn.getValue()
		case keywordCatReleaseVersion:
			p.metadata.ReleaseVersion = append(p.metadata.ReleaseVersion, strings.Replace(tkn.getNormalizedValue(), "V", "", 1))
		case keywordCatFileChecksum:
			p.metadata.FileChecksum = tkn.getValue()
		case keywordCatVideoResolution:
			p.metadata.VideoResolution = tkn.getValue()
		case keywordCatReleaseGroup:
			p.metadata.ReleaseGroup = tkn.getValue()
		case keywordCatAudioTerm:
			p.metadata.AudioTerm = append(p.metadata.AudioTerm, tkn.getValue())
		case keywordCatAnimeType:
			p.metadata.AnimeType = append(p.metadata.AnimeType, tkn.getValue())
		case keywordCatVideoTerm:
			p.metadata.VideoTerm = append(p.metadata.VideoTerm, tkn.getValue())
		case keywordCatDeviceCompat:
			p.metadata.DeviceCompatibility = append(p.metadata.DeviceCompatibility, tkn.getValue())
		case keywordCatLanguage:
			p.metadata.Language = append(p.metadata.Language, tkn.getValue())
		case keywordCatSubtitles:
			p.metadata.Subtitles = append(p.metadata.Subtitles, tkn.getValue())
		case keywordCatSource:
			p.metadata.Source = append(p.metadata.Source, tkn.getValue())
		//case keywordCatFileExtension:
		//p.metadata.FileExtension = tkn.getValue()
		case keywordCatReleaseInformation:
			p.metadata.ReleaseInformation = append(p.metadata.ReleaseInformation, tkn.getValue())
		default:
		}

		switch tkn.MetadataCategory {
		case metadataFileExtension:
			p.metadata.FileExtension = tkn.getValue()
		case metadataTitle:
			p.metadata.Title = tkn.getValue()
		case metadataEpisodeTitle:
			p.metadata.EpisodeTitle = tkn.getValue()
		case metadataEpisodeNumber:
			p.metadata.EpisodeNumber = append(p.metadata.EpisodeNumber, tkn.getValue())
		case metadataOtherEpisodeNumber:
			p.metadata.OtherEpisodeNumber = append(p.metadata.OtherEpisodeNumber, tkn.getValue())
		case metadataEpisodeNumberAlt:
			p.metadata.EpisodeNumberAlt = append(p.metadata.EpisodeNumberAlt, tkn.getValue())
		case metadataSeason:
			p.metadata.SeasonNumber = append(p.metadata.SeasonNumber, tkn.getValue())
		case metadataPart:
			p.metadata.PartNumber = append(p.metadata.PartNumber, tkn.getValue())
		case metadataVolumeNumber:
			p.metadata.VolumeNumber = append(p.metadata.VolumeNumber, tkn.getValue())
		case metadataAnimeType:
			p.metadata.AnimeType = append(p.metadata.AnimeType, tkn.getValue())
		case metadataAudioTerm:
			p.metadata.AudioTerm = append(p.metadata.AudioTerm, tkn.getValue())
		case metadataDeviceCompat:
			p.metadata.DeviceCompatibility = append(p.metadata.DeviceCompatibility, tkn.getValue())
		case metadataLanguage:
			p.metadata.Language = append(p.metadata.Language, tkn.getValue())
		case metadataSubtitles:
			p.metadata.Subtitles = append(p.metadata.Subtitles, tkn.getValue())
		case metadataReleaseGroup:
			p.metadata.ReleaseGroup = tkn.getValue()
		case metadataReleaseVersion:
			p.metadata.ReleaseVersion = append(p.metadata.ReleaseVersion, tkn.getValue())
		case metadataSource:
			p.metadata.Source = append(p.metadata.Source, tkn.getValue())
		case metadataVideoResolution:
			p.metadata.VideoResolution = tkn.getValue()
		case metadataVideoTerm:
			p.metadata.VideoTerm = append(p.metadata.VideoTerm, tkn.getValue())
		default:
		}
	}

	if len(p.metadata.EpisodeNumber) == 0 && len(p.metadata.OtherEpisodeNumber) > 0 {
		p.metadata.EpisodeNumber = p.metadata.OtherEpisodeNumber
		p.metadata.OtherEpisodeNumber = nil
	}

}

func (p *parser) writeFormattedTitle() {
	// Get title
	found, titleTkns := p.tokenManager.tokens.findWithMetadataCategory(metadataTitle)
	if !found {
		return
	}
	titleTkn := titleTkns[0]
	title := titleTkn.getValue()

	// Add content between parentheses
parenLoop:
	for {
		// Get opening parenthesis directly after after title token
		if openingParenTkn, found, _ := p.tokenManager.tokens.getTokenAfterSD(titleTkn); found && openingParenTkn.isOpeningBracket() && openingParenTkn.getValue() == "(" {
			// Get closing parenthesis after opening parenthesis
			if closingParenTkn, found := p.tokenManager.tokens.getFirstOccurrenceAfter(p.tokenManager.tokens.getIndexOf(openingParenTkn), func(tkn *token) bool {
				return tkn.getValue() == ")"
			}); found &&
				closingParenTkn.isClosingBracket() && closingParenTkn.getValue() == ")" {
				// Get tokens between parentheses
				if inbetweenTkns, found := p.tokenManager.tokens.getFromTo(p.tokenManager.tokens.getIndexOf(openingParenTkn)+1, p.tokenManager.tokens.getIndexOf(closingParenTkn)); found {
					/** Check **/
					// Check if tokens between parentheses are only years or unknowns
					_inbetweenTkns := make([]*token, 0)
					for _, tkn := range inbetweenTkns {
						if !tkn.isYear() && !tkn.isUnknown() {
							break parenLoop
						}
						_inbetweenTkns = append(_inbetweenTkns, tkn)
					}
					if len(_inbetweenTkns) == 0 {
						break parenLoop
					}
					/** End check */
					title += " ("
					for idx, tkn := range _inbetweenTkns {
						title += tkn.getValue()
						if idx < len(_inbetweenTkns)-1 {
							title += " "
						}
					}
					title += ")"
				}
			}
		}
		break
	}
	// Get anime types but only if they are standalone or movie types
	found, animeTypeTkns := p.tokenManager.tokens.findWithKeywordCategory(keywordCatAnimeType)
	if !found {
		p.metadata.FormattedTitle = title
		return
	}
	animeTypeTkns = lo.Filter(animeTypeTkns, func(tkn *token, _ int) bool {
		return tkn.isStandaloneKeyword() || strings.Contains(tkn.getNormalizedValue(), "MOVIE")
	})
	// Get other episode numbers
	_, otherEpisodeNumberTkns := p.tokenManager.tokens.findWithMetadataCategory(metadataOtherEpisodeNumber)

	includesMovieToken := false

	for idx, tkn := range animeTypeTkns {
		title += " " + tkn.getValue()
		if otherEpisodeNumberTkns != nil && idx < len(otherEpisodeNumberTkns) {
			title += " " + otherEpisodeNumberTkns[idx].getValue()
		}
		// check movie token
		if strings.Contains(tkn.getNormalizedValue(), "MOVIE") {
			includesMovieToken = true
		}
	}

	/* Include episode title if movie */
	// Get title
	if found, epTitleTkns := p.tokenManager.tokens.findWithMetadataCategory(metadataEpisodeTitle); found && includesMovieToken {
		title += " " + epTitleTkns[0].getValue()
	}
	/* end */

	p.metadata.FormattedTitle = title

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (p *parser) cleanUp() {

	if p.metadata.EpisodeNumber != nil {
		ret, vers := cleanNumbers(p.metadata.EpisodeNumber)
		p.metadata.EpisodeNumber = ret
		if len(vers) > 0 {
			p.metadata.ReleaseVersion = append(p.metadata.ReleaseVersion, vers)
		}
	}
	if p.metadata.SeasonNumber != nil {
		ret, vers := cleanNumbers(p.metadata.SeasonNumber)
		p.metadata.SeasonNumber = ret
		if len(vers) > 0 {
			p.metadata.ReleaseVersion = append(p.metadata.ReleaseVersion, vers)
		}
	}
	if p.metadata.PartNumber != nil {
		ret, vers := cleanNumbers(p.metadata.PartNumber)
		p.metadata.PartNumber = ret
		if len(vers) > 0 {
			p.metadata.ReleaseVersion = append(p.metadata.ReleaseVersion, vers)
		}
	}
	if p.metadata.VolumeNumber != nil {
		ret, vers := cleanNumbers(p.metadata.VolumeNumber)
		p.metadata.VolumeNumber = ret
		if len(vers) > 0 {
			p.metadata.ReleaseVersion = append(p.metadata.ReleaseVersion, vers)
		}
	}
}

func cleanNumbers(numbers []string) ([]string, string) {
	var cleaned []string
	var vers string
	for _, number := range numbers {
		num, v := cleanNumber(number)
		cleaned = append(cleaned, num)
		vers = v
	}
	return cleaned, vers
}

func cleanNumber(number string) (string, string) {
	if isDigitsOnly(number) {
		return number, ""
	}
	//sepIdx := strings.IndexByte(number, '.')
	//if sepIdx != -1 {
	//	number = number[:sepIdx]
	//	return number, "2"
	//}
	sepIdx := strings.IndexByte(strings.ToLower(number), 'v')
	if sepIdx != -1 {
		s := number
		number = number[:sepIdx]
		pre, ok := strings.CutPrefix(strings.ToLower(s), number+"v")
		if !ok {
			pre = ""
		}
		return number, pre
	}
	sepIdx = strings.IndexByte(number, '\'')
	if sepIdx != -1 {
		number = number[:sepIdx]
		return number, "2"
	}
	for _, letter := range []rune{'a', 'b', 'c'} {
		sepIdx = strings.IndexByte(strings.ToLower(number), byte(letter))
		if sepIdx != -1 {
			number = number[:sepIdx]
			return number, string(letter)
		}
	}
	return number, ""
}
