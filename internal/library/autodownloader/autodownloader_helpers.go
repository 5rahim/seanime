package autodownloader

import (
	"regexp"
	"seanime/internal/library/anime"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

func (ad *AutoDownloader) isExcludedTermsMatch(torrentName string, rule *anime.AutoDownloaderRule) bool {
	if len(rule.ExcludeTerms) == 0 {
		return true
	}
	torrentNameLower := strings.ToLower(torrentName)
	for _, term := range rule.ExcludeTerms {
		terms := strings.Split(term, ",")
		for _, t := range terms {
			t = strings.TrimSpace(t)
			if strings.Contains(torrentNameLower, strings.ToLower(t)) {
				return false
			}
		}
	}
	return true
}

func (ad *AutoDownloader) isConstraintsMatch(t *NormalizedTorrent, rule *anime.AutoDownloaderRule) bool {
	if t.Seeders > -1 && rule.MinSeeders > 0 && t.Seeders < rule.MinSeeders {
		return false
	}
	if t.Size > 0 && rule.MinSize != "" {
		minSize, err := stringToBytes(rule.MinSize)
		if err == nil && t.Size < minSize {
			return false
		}
	}
	if t.Size > 0 && rule.MaxSize != "" {
		maxSize, err := stringToBytes(rule.MaxSize)
		if err == nil && t.Size > maxSize {
			return false
		}
	}
	return true
}

// isProfileValidChecks checks if the torrent matches the profile's validity conditions (require, block, thresholds)
// It does not calculate scores
func (ad *AutoDownloader) isProfileValidChecks(t *NormalizedTorrent, profile *anime.AutoDownloaderProfile) bool {
	if profile == nil {
		return true
	}

	// Check thresholds
	// Only check if torrent has seeders info
	if t.Seeders > -1 && profile.MinSeeders > 0 && t.Seeders < profile.MinSeeders {
		return false
	}
	// Only check if torrent has size info
	if profile.MinSize != "" && t.Size > 0 {
		minSize, err := stringToBytes(profile.MinSize)
		if err == nil && t.Size < minSize {
			return false
		}
	}
	if profile.MaxSize != "" && t.Size > 0 {
		maxSize, err := stringToBytes(profile.MaxSize)
		if err == nil && t.Size > maxSize {
			return false
		}
	}

	// Check conditions (block & require)
	// Condition ID -> bool
	requiredFound := make(map[string]bool)
	requiredConditions := make([]string, 0)

	// Identify required conditions first
	for _, condition := range profile.Conditions {
		if condition.Action == anime.AutoDownloaderProfileRuleFormatActionRequire {
			requiredConditions = append(requiredConditions, condition.ID)
		}
	}

	torrentNameLower := strings.ToLower(t.Name)

	for _, condition := range profile.Conditions {
		isMatch := false
		if condition.IsRegex {
			re, err := regexp.Compile(condition.Term)
			if err == nil {
				isMatch = re.MatchString(t.Name)
			}
		} else {
			terms := strings.Split(condition.Term, ",")
			for _, term := range terms {
				term = strings.TrimSpace(term)
				if strings.Contains(torrentNameLower, strings.ToLower(term)) {
					isMatch = true
					break
				}
			}
		}

		if isMatch {
			switch condition.Action {
			case anime.AutoDownloaderProfileRuleFormatActionBlock:
				return false // Immediate fail
			case anime.AutoDownloaderProfileRuleFormatActionRequire:
				requiredFound[condition.ID] = true
			}
		}
	}

	// Check if all required conditions were met
	for _, reqID := range requiredConditions {
		if !requiredFound[reqID] {
			return false
		}
	}

	return true
}

func (ad *AutoDownloader) calculateTorrentScore(t *NormalizedTorrent, profile *anime.AutoDownloaderProfile) int {
	if profile == nil {
		return 0
	}

	score := 0
	torrentNameLower := strings.ToLower(t.Name)

	for _, condition := range profile.Conditions {
		if condition.Action != anime.AutoDownloaderProfileRuleFormatActionScore {
			continue
		}

		isMatch := false
		if condition.IsRegex {
			re, err := regexp.Compile(condition.Term)
			if err == nil {
				isMatch = re.MatchString(t.Name)
			}
		} else {
			terms := strings.Split(condition.Term, ",")
			for _, term := range terms {
				term = strings.TrimSpace(term)
				if strings.Contains(torrentNameLower, strings.ToLower(term)) {
					isMatch = true
					break
				}
			}
		}

		if isMatch {
			score += condition.Score
		}
	}

	return score
}

// stringToBytes converts a size string (e.g. "1.5GB", "200MB", "1GiB") to bytes.
// Supports B, KB, MB, GB, TB, KiB, MiB, GiB, TiB.
// All units are treated as binary (1024-based)
func stringToBytes(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0, nil
	}

	if strings.Contains(s, "IB") {
		s = strings.ReplaceAll(s, "IB", "B")
	}

	var multiplier int64 = 1
	var numStr string

	if strings.HasSuffix(s, "TB") {
		multiplier = 1024 * 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(s, "TB")
	} else if strings.HasSuffix(s, "GB") {
		multiplier = 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(s, "GB")
	} else if strings.HasSuffix(s, "MB") {
		multiplier = 1024 * 1024
		numStr = strings.TrimSuffix(s, "MB")
	} else if strings.HasSuffix(s, "KB") {
		multiplier = 1024
		numStr = strings.TrimSuffix(s, "KB")
	} else if strings.HasSuffix(s, "B") {
		numStr = strings.TrimSuffix(s, "B")
	} else {
		numStr = s // Assume raw or default to simple parse attempt
	}

	numStr = strings.TrimSpace(numStr)
	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, err
	}

	return int64(val * float64(multiplier)), nil
}

func (ad *AutoDownloader) isResolutionMatch(quality string, resolutions []string) (ok bool) {
	defer util.HandlePanicInModuleThen("autodownloader/isResolutionMatch", func() {
		ok = false
	})

	if len(resolutions) == 0 {
		return true
	}
	if quality == "" {
		return false
	}

	normalizedQuality := comparison.NormalizeResolution(quality)

	for _, q := range resolutions {
		normalizedRuleRes := comparison.NormalizeResolution(q)
		if comparison.ExtractResolutionInt(normalizedQuality) == comparison.ExtractResolutionInt(normalizedRuleRes) {
			return true
		}
	}
	return false
}

// getReleaseGroupToResolutionsMap groups rules by release group to optimize search queries.
// It resolves resolutions from profiles if the rule doesn't have them explicitly set.
func (ad *AutoDownloader) getReleaseGroupToResolutionsMap(rules []*anime.AutoDownloaderRule, profiles []*anime.AutoDownloaderProfile) map[string][]string {
	res := make(map[string][]string)

	for _, rule := range rules {
		// Determine effective resolutions for this rule
		effectiveResolutions := rule.Resolutions
		if len(effectiveResolutions) == 0 {
			// Fallback to profile resolutions
			for _, p := range profiles {
				// Check global profile or specific assigned profile
				if p.Global || (rule.ProfileID != nil && p.DbID == *rule.ProfileID) {
					effectiveResolutions = append(effectiveResolutions, p.Resolutions...)
				}
			}
		}
		effectiveResolutions = lo.Uniq(effectiveResolutions)

		// Group by release groups
		if len(rule.ReleaseGroups) > 0 {
			for _, rg := range rule.ReleaseGroups {
				if _, ok := res[rg]; !ok {
					res[rg] = make([]string, 0)
				}
				res[rg] = append(res[rg], effectiveResolutions...)
			}
		}
	}

	// Deduplicate resolutions for each group
	for rg, resolutions := range res {
		res[rg] = lo.Uniq(resolutions)
	}

	return res
}
