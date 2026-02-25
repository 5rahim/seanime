package goja_bindings

import (
	"fmt"
	"regexp"
	"seanime/internal/library/scanner"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"strconv"
	"strings"

	"github.com/dop251/goja"
)

func BindScannerUtils(vm *goja.Runtime) error {
	scannerUtils := vm.NewObject()
	_ = scannerUtils.Set("normalizeTitle", normalizeTitleFunc(vm))
	_ = scannerUtils.Set("extractPartNumber", extractPartNumberFunc(vm))
	_ = scannerUtils.Set("extractSeasonNumber", extractSeasonNumberFunc(vm))
	_ = scannerUtils.Set("extractYear", extractYearFunc(vm))
	_ = scannerUtils.Set("compareTitles", compareTitlesFunc(vm))
	_ = scannerUtils.Set("findBestMatch", findBestMatchFunc(vm))
	_ = scannerUtils.Set("getSignificantTokens", getSignificantTokensFunc(vm))
	_ = scannerUtils.Set("buildSearchQuery", buildSearchQueryFunc(vm))
	_ = scannerUtils.Set("buildAdvancedQuery", buildAdvancedQueryFunc(vm))
	_ = scannerUtils.Set("sanitizeQuery", sanitizeQueryFunc(vm))
	_ = scannerUtils.Set("buildSeasonQuery", buildSeasonQueryFunc(vm))
	_ = scannerUtils.Set("buildPartQuery", buildPartQueryFunc(vm))
	_ = scannerUtils.Set("buildSmartSearchTitles", buildSmartSearchTitlesFunc(vm))
	_ = vm.Set("$scannerUtils", scannerUtils)

	return nil
}

func normalizeTitleFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	return func(call goja.FunctionCall) goja.Value {
		defer func() {
			if r := recover(); r != nil {
				panic(vm.ToValue("selection is nil"))
			}
		}()

		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: normalizeTitle requires at least 1 argument"))
		}

		str, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue(vm.NewTypeError("argument is not a string")))
		}

		normalized := scanner.NormalizeTitle(str)
		return vm.ToValue(normalized)
	}
}

func extractPartNumberFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	return func(call goja.FunctionCall) goja.Value {
		defer func() {
			if r := recover(); r != nil {
				panic(vm.ToValue("selection is nil"))
			}
		}()

		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: extractPartNumber requires at least 1 argument"))
		}

		str, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue(vm.NewTypeError("argument is not a string")))
		}

		part := scanner.ExtractPartNumber(str)
		return vm.ToValue(part)
	}
}

func extractYearFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	return func(call goja.FunctionCall) goja.Value {
		defer func() {
			if r := recover(); r != nil {
				panic(vm.ToValue("selection is nil"))
			}
		}()

		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: extractYear requires at least 1 argument"))
		}

		str, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue(vm.NewTypeError("argument is not a string")))
		}

		year := scanner.ExtractYear(str)
		return vm.ToValue(year)
	}
}

func compareTitlesFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	return func(call goja.FunctionCall) goja.Value {
		defer func() {
			if r := recover(); r != nil {
				panic(vm.ToValue("selection is nil"))
			}
		}()

		if len(call.Arguments) < 2 {
			panic(vm.ToValue("TypeError: compareTitles requires 2 arguments"))
		}

		str1, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue(vm.NewTypeError("argument 1 is not a string")))
		}

		str2, ok := call.Argument(1).Export().(string)
		if !ok {
			panic(vm.ToValue(vm.NewTypeError("argument 2 is not a string")))
		}

		norm1 := scanner.NormalizeTitle(str1)
		norm2 := scanner.NormalizeTitle(str2)

		ratio := scanner.WeightedTokenMatchRatio(norm1.Tokens, norm2.Tokens)
		return vm.ToValue(ratio)
	}
}

func getSignificantTokensFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	return func(call goja.FunctionCall) goja.Value {
		defer func() {
			if r := recover(); r != nil {
				panic(vm.ToValue("selection is nil"))
			}
		}()

		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: getSignificantTokens requires at least 1 argument"))
		}

		str, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue(vm.NewTypeError("argument is not a string")))
		}

		norm := scanner.NormalizeTitle(str)
		sigTokens := scanner.GetSignificantTokens(norm.Tokens)
		return vm.ToValue(sigTokens)
	}
}

func findBestMatchFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	return func(call goja.FunctionCall) goja.Value {
		defer func() {
			if r := recover(); r != nil {
				panic(vm.ToValue("selection is nil"))
			}
		}()

		if len(call.Arguments) < 2 {
			panic(vm.ToValue("TypeError: findBestMatch requires 2 arguments"))
		}

		target, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue(vm.NewTypeError("argument 1 is not a string")))
		}

		candidatesVar := call.Argument(1).Export()
		candidates, ok := candidatesVar.([]interface{})
		if !ok {
			panic(vm.ToValue(vm.NewTypeError("argument 2 is not an array")))
		}

		normTarget := scanner.NormalizeTitle(target)

		bestMatch := ""
		bestScore := -1.0

		for _, candObj := range candidates {
			cand, ok := candObj.(string)
			if !ok {
				continue
			}
			normCand := scanner.NormalizeTitle(cand)
			score := scanner.WeightedTokenMatchRatio(normTarget.Tokens, normCand.Tokens)
			if score > bestScore {
				bestScore = score
				bestMatch = cand
			}
		}

		return vm.ToValue(bestMatch)
	}
}

// ------------------------------------------------------------------------------------------------
// Search query building helpers
// ------------------------------------------------------------------------------------------------

var searchSyntaxChars = regexp.MustCompile(`[()[\]{}|"'~*?\\^!]`)
var multiSpaceRegex = regexp.MustCompile(`\s{2,}`)

// sanitizeForSearch removes characters that could interfere with search engine syntax
func sanitizeForSearch(s string) string {
	s = searchSyntaxChars.ReplaceAllString(s, " ")
	s = multiSpaceRegex.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// buildSearchQueryFunc creates a clean search query from a title.
// It normalizes the title, removes noise/format words, and returns a compact string
func buildSearchQueryFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	return func(call goja.FunctionCall) goja.Value {
		defer func() {
			if r := recover(); r != nil {
				panic(vm.ToValue("selection is nil"))
			}
		}()

		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: buildSearchQuery requires at least 1 argument"))
		}

		str, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue(vm.NewTypeError("argument is not a string")))
		}

		norm := scanner.NormalizeTitle(str)
		// Use the denoised title (significant words only) for a compact query
		query := norm.DenoisedTitle
		if query == "" {
			query = norm.CleanBaseTitle
		}
		if query == "" {
			query = norm.Normalized
		}
		query = sanitizeForSearch(query)
		return vm.ToValue(query)
	}
}

// buildAdvancedQueryFunc builds an advanced boolean query grouping multiple
// alternative titles with (title1 | title2 | ...) syntax.
// Arguments: titles []string
func buildAdvancedQueryFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	return func(call goja.FunctionCall) goja.Value {
		defer func() {
			if r := recover(); r != nil {
				panic(vm.ToValue("selection is nil"))
			}
		}()

		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: buildAdvancedQuery requires at least 1 argument"))
		}

		titlesVar := call.Argument(0).Export()
		titles, ok := titlesVar.([]interface{})
		if !ok {
			panic(vm.ToValue(vm.NewTypeError("argument 1 is not an array")))
		}

		seen := make(map[string]struct{}, len(titles))
		var parts []string
		for _, tObj := range titles {
			t, ok := tObj.(string)
			if !ok || t == "" {
				continue
			}
			norm := scanner.NormalizeTitle(t)
			q := norm.DenoisedTitle
			if q == "" {
				q = norm.CleanBaseTitle
			}
			if q == "" {
				q = norm.Normalized
			}
			q = sanitizeForSearch(q)
			if q == "" {
				continue
			}
			if _, dup := seen[q]; dup {
				continue
			}
			seen[q] = struct{}{}
			parts = append(parts, q)
		}

		var result string
		switch len(parts) {
		case 0:
			result = ""
		case 1:
			result = parts[0]
		default:
			result = "(" + strings.Join(parts, " | ") + ")"
		}

		return vm.ToValue(result)
	}
}

// sanitizeQueryFunc strips special search syntax characters from a raw string
// so it can be safely embedded in a search query.
func sanitizeQueryFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	return func(call goja.FunctionCall) goja.Value {
		defer func() {
			if r := recover(); r != nil {
				panic(vm.ToValue("selection is nil"))
			}
		}()

		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: sanitizeQuery requires at least 1 argument"))
		}

		str, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue(vm.NewTypeError("argument is not a string")))
		}

		return vm.ToValue(sanitizeForSearch(str))
	}
}

// buildSeasonQueryFunc builds a query string that includes multiple season
// identifier formats after the base title.
// Arguments: title string, season int
// e.g. buildSeasonQuery("Overlord", 2) → "(Overlord S02 | Overlord S2 | Overlord Season 2 | Overlord 2nd Season)"
func buildSeasonQueryFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	return func(call goja.FunctionCall) goja.Value {
		defer func() {
			if r := recover(); r != nil {
				panic(vm.ToValue("selection is nil"))
			}
		}()

		if len(call.Arguments) < 2 {
			panic(vm.ToValue("TypeError: buildSeasonQuery requires 2 arguments"))
		}

		title, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue(vm.NewTypeError("argument 1 is not a string")))
		}

		seasonRaw := call.Argument(1).Export()
		season, err := toInt(seasonRaw)
		if err != nil {
			panic(vm.ToValue(vm.NewTypeError("argument 2 is not a number")))
		}

		// Normalize the title
		norm := scanner.NormalizeTitle(title)
		base := norm.DenoisedTitle
		if base == "" {
			base = norm.CleanBaseTitle
		}
		if base == "" {
			base = norm.Normalized
		}
		base = sanitizeForSearch(base)

		if season <= 0 || season == 1 {
			// Season 1 or unspecified: just return the base title
			return vm.ToValue(base)
		}

		variants := []string{
			fmt.Sprintf("%s S%02d", base, season),
			fmt.Sprintf("%s S%d", base, season),
			fmt.Sprintf("%s Season %d", base, season),
			fmt.Sprintf("%s %s Season", base, util.IntegerToOrdinal(season)),
		}

		result := "(" + strings.Join(variants, " | ") + ")"
		return vm.ToValue(result)
	}
}

func toInt(v interface{}) (int, error) {
	switch n := v.(type) {
	case int:
		return n, nil
	case int64:
		return int(n), nil
	case float64:
		return int(n), nil
	case string:
		return strconv.Atoi(n)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", v)
	}
}

// extractSeasonNumberFunc exposes comparison.ExtractSeasonNumber to JS.
// This handles: "Season X", "SXX", "Xnd Season", roman numerals (II, III...),
// trailing numbers (Konosuba 2), Japanese patterns (2期), etc.
func extractSeasonNumberFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	return func(call goja.FunctionCall) goja.Value {
		defer func() {
			if r := recover(); r != nil {
				panic(vm.ToValue("selection is nil"))
			}
		}()

		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: extractSeasonNumber requires at least 1 argument"))
		}

		str, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue(vm.NewTypeError("argument is not a string")))
		}

		season := comparison.ExtractSeasonNumber(str)
		return vm.ToValue(season)
	}
}

// buildPartQueryFunc builds a query string with multiple part identifier formats.
// Arguments: title string, part int
// e.g. buildPartQuery("Re:Zero", 2) → "(Re Zero Part 2 | Re Zero Part II | Re Zero 2nd Cour)"
func buildPartQueryFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	return func(call goja.FunctionCall) goja.Value {
		defer func() {
			if r := recover(); r != nil {
				panic(vm.ToValue("selection is nil"))
			}
		}()

		if len(call.Arguments) < 2 {
			panic(vm.ToValue("TypeError: buildPartQuery requires 2 arguments"))
		}

		title, ok := call.Argument(0).Export().(string)
		if !ok {
			panic(vm.ToValue(vm.NewTypeError("argument 1 is not a string")))
		}

		partRaw := call.Argument(1).Export()
		part, err := toInt(partRaw)
		if err != nil {
			panic(vm.ToValue(vm.NewTypeError("argument 2 is not a number")))
		}

		// Normalize the title
		norm := scanner.NormalizeTitle(title)
		base := norm.DenoisedTitle
		if base == "" {
			base = norm.CleanBaseTitle
		}
		if base == "" {
			base = norm.Normalized
		}
		base = sanitizeForSearch(base)

		if part <= 0 || part == 1 {
			return vm.ToValue(base)
		}

		// Map part numbers to roman numerals
		romanNumerals := map[int]string{2: "II", 3: "III", 4: "IV", 5: "V", 6: "VI", 7: "VII", 8: "VIII", 9: "IX"}

		variants := []string{
			fmt.Sprintf("%s Part %d", base, part),
		}

		if roman, ok := romanNumerals[part]; ok {
			variants = append(variants, fmt.Sprintf("%s Part %s", base, roman))
		}

		variants = append(variants, fmt.Sprintf("%s %s Cour", base, util.IntegerToOrdinal(part)))

		result := "(" + strings.Join(variants, " | ") + ")"
		return vm.ToValue(result)
	}
}

// buildSmartSearchTitlesFunc processes all media titles (romaji, english, synonyms)
// and returns an object with:
//
//	titles: []string cleaned, deduplicated search-ready title variants
//	season: int      detected season number (-1 if none)
//	part: int        detected part number (-1 if none)
//
// 1. Extracts season/part from all titles
// 2. Normalizes titles (macrons, possessives, separators, etc.)
// 3. Strips season/part indicators and roman numerals from search variants
// 4. Generates shortened colon-split variants for long titles
// 5. Deduplicates all variants
//
// Arguments: titles []string (all titles: romaji, english, synonyms)
func buildSmartSearchTitlesFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	return func(call goja.FunctionCall) goja.Value {
		defer func() {
			if r := recover(); r != nil {
				panic(vm.ToValue("selection is nil"))
			}
		}()

		if len(call.Arguments) < 1 {
			panic(vm.ToValue("TypeError: buildSmartSearchTitles requires at least 1 argument"))
		}

		titlesVar := call.Argument(0).Export()
		rawTitles, ok := titlesVar.([]interface{})
		if !ok {
			panic(vm.ToValue(vm.NewTypeError("argument 1 is not an array")))
		}

		season := -1
		part := -1
		seen := make(map[string]struct{}, len(rawTitles)*2)
		var cleanTitles []string

		// Helper to add a title if not duplicate
		addTitle := func(t string) {
			t = strings.TrimSpace(t)
			if t == "" {
				return
			}
			key := strings.ToLower(t)
			if _, dup := seen[key]; dup {
				return
			}
			seen[key] = struct{}{}
			cleanTitles = append(cleanTitles, t)
		}

		for _, tObj := range rawTitles {
			t, ok := tObj.(string)
			if !ok || t == "" {
				continue
			}

			// Extract season
			if s := comparison.ExtractSeasonNumber(t); s > 0 && season <= 0 {
				season = s
			}

			// Extract part
			if p := scanner.ExtractPartNumber(t); p > 0 && part <= 0 {
				part = p
			}

			// Normalize and get the denoised title for search
			norm := scanner.NormalizeTitle(t)

			// Use denoised (no noise words) as primary search variant
			searchTitle := norm.DenoisedTitle
			if searchTitle == "" {
				searchTitle = norm.CleanBaseTitle
			}
			if searchTitle == "" {
				searchTitle = norm.Normalized
			}
			searchTitle = sanitizeForSearch(searchTitle)
			addTitle(searchTitle)

			// Also add the clean base title (includes noise words but strips format/roman numerals)
			if norm.CleanBaseTitle != "" {
				cleanBase := sanitizeForSearch(norm.CleanBaseTitle)
				addTitle(cleanBase)
			}

			// Generate colon-split shortened variant from the original title
			colonIdx := strings.IndexRune(t, ':')
			if colonIdx > 0 {
				prefix := strings.TrimSpace(t[:colonIdx])
				if len(prefix) >= 5 { // Only if meaningful length
					prefixNorm := scanner.NormalizeTitle(prefix)
					prefixSearch := prefixNorm.DenoisedTitle
					if prefixSearch == "" {
						prefixSearch = prefixNorm.CleanBaseTitle
					}
					if prefixSearch == "" {
						prefixSearch = prefixNorm.Normalized
					}
					prefixSearch = sanitizeForSearch(prefixSearch)
					addTitle(prefixSearch)
				}
			}

			// Also try dash-split for titles like "Shingeki no Kyojin - The Final Season"
			dashIdx := strings.Index(t, " - ")
			if dashIdx > 0 {
				prefix := strings.TrimSpace(t[:dashIdx])
				if len(prefix) >= 5 {
					prefixNorm := scanner.NormalizeTitle(prefix)
					prefixSearch := prefixNorm.DenoisedTitle
					if prefixSearch == "" {
						prefixSearch = prefixNorm.CleanBaseTitle
					}
					if prefixSearch == "" {
						prefixSearch = prefixNorm.Normalized
					}
					prefixSearch = sanitizeForSearch(prefixSearch)
					addTitle(prefixSearch)
				}
			}
		}

		// Build result object
		result := vm.NewObject()
		_ = result.Set("titles", cleanTitles)
		_ = result.Set("season", season)
		_ = result.Set("part", part)
		return result
	}
}
