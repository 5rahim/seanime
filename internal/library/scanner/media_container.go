package scanner

import (
	"seanime/internal/library/anime"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"strings"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type (
	MediaContainerOptions struct {
		AllMedia   []*anime.NormalizedMedia
		ScanLogger *ScanLogger
	}

	// MediaContainer holds all the NormalizedMedia that will be used by the Matcher.
	// It creates an inverted index for fast candidate lookup based on title tokens.
	// Note: It doesn't care that the NormalizedMedia are not fully fetched.
	// Before v3.5, it was used to flatten relations into NormalizedMedia.
	MediaContainer struct {
		NormalizedMedia       []*anime.NormalizedMedia
		NormalizedTitlesCache map[int][]*NormalizedTitle // mediaId -> normalized titles
		ScanLogger            *ScanLogger
		// Inverted Index for fast candidate lookup
		// Token -> media that contain this token in their title
		TokenIndex map[string][]*anime.NormalizedMedia
		engTitles  []*string // legacy
		romTitles  []*string // legacy
		synonyms   []*string // legacy
	}
)

// NewMediaContainer creates a new MediaContainer from a list of NormalizedMedia that will be used by the Matcher.
func NewMediaContainer(opts *MediaContainerOptions) *MediaContainer {
	mc := new(MediaContainer)
	mc.ScanLogger = opts.ScanLogger

	mc.NormalizedMedia = opts.AllMedia

	// pre-compute normalized titles for all media
	mc.NormalizedTitlesCache = make(map[int][]*NormalizedTitle, len(mc.NormalizedMedia))

	// Initialize token index
	mc.TokenIndex = make(map[string][]*anime.NormalizedMedia)

	for _, m := range mc.NormalizedMedia {
		normalized := make([]*NormalizedTitle, 0)

		// Keep track of which tokens this media has been added to to avoid duplicates
		seenTokens := make(map[string]struct{})

		addTitle := func(t *string, isMain bool) {
			if t != nil && *t != "" {
				norm := NormalizeTitle(*t)
				norm.IsMain = isMain
				normalized = append(normalized, norm)

				// Populate index
				tokens := GetSignificantTokens(norm.Tokens)
				for _, token := range tokens {
					if _, ok := seenTokens[token]; !ok {
						mc.TokenIndex[token] = append(mc.TokenIndex[token], m)
						seenTokens[token] = struct{}{}
					}
				}
				// Also index compound tokens (adjacent pairs concatenated)
				// e.g. "re" + "zero" -> "rezero" so that "ReZero" can find "Re Zero kara..."
				for i := 0; i < len(tokens)-1; i++ {
					// only concatenate short tokens to avoid too much noise
					if len(tokens[i]) <= 5 && len(tokens[i+1]) <= 5 {
						compound := tokens[i] + tokens[i+1]
						if _, ok := seenTokens[compound]; !ok {
							mc.TokenIndex[compound] = append(mc.TokenIndex[compound], m)
							seenTokens[compound] = struct{}{}
						}
					}
				}
			}
		}

		if m.Title != nil {
			addTitle(m.Title.Romaji, true)
			addTitle(m.Title.English, true)
			addTitle(m.Title.Native, false)
			addTitle(m.Title.UserPreferred, true)
		}
		if m.Synonyms != nil {
			for _, syn := range m.Synonyms {
				if !util.IsMostlyLatinString(*syn) {
					continue
				}
				addTitle(syn, false)
			}
		}

		mc.NormalizedTitlesCache[m.ID] = normalized
		seenTokens = nil
	}

	// ------------------------------------------

	// Legacy stuff (used for legacy matching)
	// should've been maps instead of slices
	engTitles := make([]*string, 0, len(mc.NormalizedMedia))
	romTitles := make([]*string, 0, len(mc.NormalizedMedia))
	synonymsSlice := make([]*string, 0, len(mc.NormalizedMedia)*2)

	for _, m := range mc.NormalizedMedia {
		if m.Title.English != nil && len(*m.Title.English) > 0 {
			engTitles = append(engTitles, m.Title.English)
		}
		if m.Title.Romaji != nil && len(*m.Title.Romaji) > 0 {
			romTitles = append(romTitles, m.Title.Romaji)
		}
		if m.Synonyms != nil {
			for _, syn := range m.Synonyms {
				if syn != nil && comparison.ValueContainsSeason(*syn) {
					synonymsSlice = append(synonymsSlice, syn)
				}
			}
		}
	}

	mc.engTitles = engTitles
	mc.romTitles = romTitles
	mc.synonyms = synonymsSlice

	// ------------------------------------------

	if mc.ScanLogger != nil {
		mc.ScanLogger.LogMediaContainer(zerolog.InfoLevel).
			Any("mediaCount", len(mc.NormalizedMedia)).
			Any("legacyTitleCount", len(mc.engTitles)+len(mc.romTitles)+len(mc.synonyms)).
			Any("tokenIndexSize", len(mc.TokenIndex)).
			Msg("Created media container")
	}

	return mc
}

// Legacy helper function
func (mc *MediaContainer) GetMediaFromTitleOrSynonym(title *string) (*anime.NormalizedMedia, bool) {
	if title == nil {
		return nil, false
	}
	t := strings.ToLower(*title)
	res, found := lo.Find(mc.NormalizedMedia, func(m *anime.NormalizedMedia) bool {
		if m.HasEnglishTitle() && t == strings.ToLower(*m.Title.English) {
			return true
		}
		if m.HasRomajiTitle() && t == strings.ToLower(*m.Title.Romaji) {
			return true
		}
		if m.HasSynonyms() {
			for _, syn := range m.Synonyms {
				if t == strings.ToLower(*syn) {
					return true
				}
			}
		}
		return false
	})

	return res, found
}
