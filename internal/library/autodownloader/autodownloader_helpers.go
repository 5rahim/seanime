package autodownloader

import (
	"regexp"
	"seanime/internal/library/anime"
	"seanime/internal/util"
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
		minSize, err := util.StringToBytes(rule.MinSize)
		if err == nil && t.Size < minSize {
			return false
		}
	}
	if t.Size > 0 && rule.MaxSize != "" {
		maxSize, err := util.StringToBytes(rule.MaxSize)
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
		minSize, err := util.StringToBytes(profile.MinSize)
		if err == nil && t.Size < minSize {
			return false
		}
	}
	if profile.MaxSize != "" && t.Size > 0 {
		maxSize, err := util.StringToBytes(profile.MaxSize)
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

	normalizedQuality := util.NormalizeResolution(quality)

	for _, q := range resolutions {
		normalizedRuleRes := util.NormalizeResolution(q)
		if util.ExtractResolutionInt(normalizedQuality) == util.ExtractResolutionInt(normalizedRuleRes) {
			return true
		}
	}
	return false
}

// getReleaseGroupToResolutionsMap groups rules by release group to optimize search queries.
// It resolves resolutions from profiles if the rule doesn't have them explicitly set.
// It also resolves release groups from profiles if the rule doesn't have them explicitly set.
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

		// Determine effective release groups for this rule
		effectiveReleaseGroups := rule.ReleaseGroups
		if len(effectiveReleaseGroups) == 0 {
			// Fallback to profile release groups
			for _, p := range profiles {
				// Check global profile or specific assigned profile
				if p.Global || (rule.ProfileID != nil && p.DbID == *rule.ProfileID) {
					effectiveReleaseGroups = append(effectiveReleaseGroups, p.ReleaseGroups...)
				}
			}
		}
		effectiveReleaseGroups = lo.Uniq(effectiveReleaseGroups)

		if len(effectiveResolutions) == 0 {
			// Returns "-" if no resolutions were found
			// The rule will just fetch by release group only
			effectiveResolutions = []string{"-"}
		}

		// Group by release groups
		for _, rg := range effectiveReleaseGroups {
			if _, ok := res[rg]; !ok {
				res[rg] = make([]string, 0)
			}
			res[rg] = append(res[rg], effectiveResolutions...)
		}
	}

	// Deduplicate resolutions for each group
	for rg, resolutions := range res {
		res[rg] = lo.Uniq(resolutions)
	}

	return res
}
