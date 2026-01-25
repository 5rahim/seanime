package autoselect

import (
	"cmp"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"seanime/internal/util"
	"slices"
	"strings"

	"github.com/5rahim/habari"
)

const (
	scoreBestReleaseBase   = 200
	scoreResolutionBase    = 100
	scoreResolutionDecay   = 10
	scoreProviderBase      = 5
	scoreProviderDecay     = 1
	scoreReleaseGroupBase  = 50
	scoreReleaseGroupDecay = 5
	scoreCodecBase         = 40
	scoreCodecDecay        = 5
	scoreSourceBase        = 30
	scoreSourceDecay       = 5
	scoreLanguageBase      = 20
	scoreLanguageDecay     = 2
	scoreMultiAudio        = 15
	scoreMultiSubs         = 10
	scoreBatch             = 20
	scoreBestRelease       = 20
)

type candidate struct {
	torrent   *hibiketorrent.AnimeTorrent
	parsed    *habari.Metadata
	lowerName string
	score     int
}

type TorrentWithCacheStatus struct {
	Torrent  *hibiketorrent.AnimeTorrent
	IsCached bool
}

// filterAndSort filters and sorts the torrents based on the profile or defaults.
func (s *AutoSelect) filterAndSort(torrents []*hibiketorrent.AnimeTorrent, profile *anime.AutoSelectProfile, postSearchSort func([]*hibiketorrent.AnimeTorrent) []*TorrentWithCacheStatus) []*hibiketorrent.AnimeTorrent {
	s.log("Filtering and sorting torrents")
	s.logger.Debug().Int("count", len(torrents)).Msg("autoselect: Filtering and sorting torrents")

	if len(torrents) == 0 {
		return torrents
	}

	// Optimize: Parse metadata once
	candidates := make([]*candidate, len(torrents))
	for i, t := range torrents {
		candidates[i] = &candidate{
			torrent:   t,
			parsed:    habari.Parse(t.Name),
			lowerName: strings.ToLower(t.Name),
		}
	}

	// Filter
	candidates = s.filterCandidates(candidates, profile)

	// Sort by profile scores first
	s.sortCandidates(candidates, profile)

	// apply torrent prioritization if provided
	if postSearchSort != nil {
		filteredTorrents := make([]*hibiketorrent.AnimeTorrent, len(candidates))
		for i, c := range candidates {
			filteredTorrents[i] = c.torrent
		}
		return s.smartCachedPrioritization(filteredTorrents, candidates, profile, postSearchSort)
	}

	filteredTorrents := make([]*hibiketorrent.AnimeTorrent, len(candidates))
	for i, c := range candidates {
		if i < 3 {
			s.logger.Debug().Str("name", c.torrent.Name).Int("seeders", c.torrent.Seeders).Int("score", c.score).Str("provider", c.torrent.Provider).Msg("autoselect: Top selection")
		}
		filteredTorrents[i] = c.torrent
	}

	return filteredTorrents
}

// filter is a shim for testing or legacy usage.
func (s *AutoSelect) filter(torrents []*hibiketorrent.AnimeTorrent, profile *anime.AutoSelectProfile) []*hibiketorrent.AnimeTorrent {
	candidates := make([]*candidate, len(torrents))
	for i, t := range torrents {
		candidates[i] = &candidate{
			torrent:   t,
			parsed:    habari.Parse(t.Name),
			lowerName: strings.ToLower(t.Name),
		}
	}
	candidates = s.filterCandidates(candidates, profile)
	ret := make([]*hibiketorrent.AnimeTorrent, len(candidates))
	for i, c := range candidates {
		ret[i] = c.torrent
	}
	return ret
}

// sort is a shim for testing or legacy usage.
func (s *AutoSelect) sort(torrents []*hibiketorrent.AnimeTorrent, profile *anime.AutoSelectProfile) {
	candidates := make([]*candidate, len(torrents))
	for i, t := range torrents {
		candidates[i] = &candidate{
			torrent:   t,
			parsed:    habari.Parse(t.Name),
			lowerName: strings.ToLower(t.Name),
		}
	}
	s.sortCandidates(candidates, profile)
	for i, c := range candidates {
		torrents[i] = c.torrent
	}
}

func containsMultiOrDual(terms []string) bool {
	for _, s := range terms {
		lower := strings.ToLower(s)
		if strings.Contains(lower, "multi") || strings.Contains(lower, "dual") || strings.Contains(lower, "dub") {
			return true
		}
	}
	return false
}

func splitAndClean(items []string) []string {
	var ret []string
	for _, item := range items {
		for _, sub := range strings.Split(item, ",") {
			ret = append(ret, strings.TrimSpace(sub))
		}
	}
	return ret
}

func checkPreference(condition bool, preference anime.AutoSelectPreference) bool {
	if preference == anime.AutoSelectPreferenceOnly && !condition {
		return false
	}
	if preference == anime.AutoSelectPreferenceNever && condition {
		return false
	}
	return true
}

func (s *AutoSelect) filterCandidates(candidates []*candidate, profile *anime.AutoSelectProfile) []*candidate {
	if profile == nil {
		return candidates
	}

	// Pre-process profile constraints
	var excludeTerms []string
	for _, term := range profile.ExcludeTerms {
		excludeTerms = append(excludeTerms, strings.ToLower(term))
	}

	preferredLanguages := splitAndClean(profile.PreferredLanguages)
	preferredCodecs := splitAndClean(profile.PreferredCodecs)
	preferredSources := splitAndClean(profile.PreferredSources)

	// Parse sizes
	var minSize int64 = -1
	if profile.MinSize != "" {
		if val, err := util.StringToBytes(profile.MinSize); err == nil {
			minSize = val
		}
	}
	var maxSize int64 = -1
	if profile.MaxSize != "" {
		if val, err := util.StringToBytes(profile.MaxSize); err == nil {
			maxSize = val
		}
	}

	var filtered []*candidate
	for _, c := range candidates {
		t := c.torrent
		parsed := c.parsed

		// Exclude terms
		if len(excludeTerms) > 0 {
			excluded := false
			for _, term := range excludeTerms {
				if strings.Contains(c.lowerName, term) {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}

		// bad
		if profile.BestReleasePreference == anime.AutoSelectPreferenceOnly && !t.IsBestRelease {
			continue
		}

		if profile.BestReleasePreference == anime.AutoSelectPreferenceNever && t.IsBestRelease {
			continue
		}

		// Language requirement
		if profile.RequireLanguage && len(preferredLanguages) > 0 {
			foundLang := false
			// Check parsed language
			if len(parsed.Language) > 0 {
				for _, lang := range preferredLanguages {
					if slices.ContainsFunc(parsed.Language, func(pl string) bool {
						return strings.EqualFold(pl, lang)
					}) {
						foundLang = true
						break
					}
				}
			} else { // Fallback to string matching
				for _, lang := range preferredLanguages {
					if len(lang) > 3 && strings.Contains(c.lowerName, strings.ToLower(lang)) {
						foundLang = true
						break
					}
				}
			}
			if !foundLang {
				continue
			}
		}

		// Seeders filtering
		if profile.MinSeeders > 0 && t.Seeders < profile.MinSeeders {
			continue
		}

		// Size filtering
		if minSize != -1 && t.Size > 0 && t.Size < minSize {
			continue
		}
		if maxSize != -1 && t.Size > 0 && t.Size > maxSize {
			continue
		}

		// Require codec
		if profile.RequireCodec && len(preferredCodecs) > 0 {
			foundCodec := false
			for _, codec := range preferredCodecs {
				if slices.ContainsFunc(parsed.VideoTerm, func(vt string) bool {
					return strings.EqualFold(vt, codec)
				}) {
					foundCodec = true
					break
				}
				if strings.Contains(c.lowerName, strings.ToLower(codec)) {
					foundCodec = true
					break
				}
			}
			if !foundCodec {
				continue
			}
		}

		// Require source
		if profile.RequireSource && len(preferredSources) > 0 {
			foundSource := false
			for _, source := range preferredSources {
				if slices.ContainsFunc(parsed.Source, func(src string) bool {
					return strings.EqualFold(src, source)
				}) {
					foundSource = true
					break
				}
				if strings.Contains(c.lowerName, strings.ToLower(source)) {
					foundSource = true
					break
				}
			}
			if !foundSource {
				continue
			}
		}

		// Preferences
		if !checkPreference(containsMultiOrDual(parsed.AudioTerm), profile.MultipleAudioPreference) {
			continue
		}
		if !checkPreference(containsMultiOrDual(parsed.Subtitles), profile.MultipleSubsPreference) {
			continue
		}
		if !checkPreference(t.IsBatch, profile.BatchPreference) {
			continue
		}

		filtered = append(filtered, c)
	}
	return filtered
}

func (s *AutoSelect) sortCandidates(candidates []*candidate, profile *anime.AutoSelectProfile) {
	for _, c := range candidates {
		c.score = s.calculateScore(c, profile)
	}

	slices.SortStableFunc(candidates, func(a, b *candidate) int {
		if a.score != b.score {
			return cmp.Compare(b.score, a.score) // Higher score first
		}

		// If the scores are the same, sort by seeders
		return cmp.Compare(b.torrent.Seeders, a.torrent.Seeders)
	})
}

// smartCachedPrioritization applies the postSearchSort (which identifies cached torrents)
// and reorders results to prioritize cached torrents that have similar quality scores.
// It prevents low-quality cached torrents from being prioritized over high quality uncached ones.
func (s *AutoSelect) smartCachedPrioritization(
	torrents []*hibiketorrent.AnimeTorrent,
	candidates []*candidate,
	profile *anime.AutoSelectProfile,
	postSearchSort func([]*hibiketorrent.AnimeTorrent) []*TorrentWithCacheStatus,
) []*hibiketorrent.AnimeTorrent {

	if len(torrents) == 0 {
		return torrents
	}

	// Get torrents with explicit cache status
	torrentsWithStatus := postSearchSort(torrents)

	// Build a map of scores for each torrent
	candidateMap := make(map[string]*candidate, len(candidates))
	for _, c := range candidates {
		candidateMap[c.torrent.InfoHash] = c
	}

	// Find the top score
	topScore := 0
	if len(candidates) > 0 {
		topScore = candidates[0].score
	}

	// cached torrents must be within 30% of the top score
	// this ensures we don't prioritize very low-quality cached torrents
	const scoreThresholdPercent = 0.7 // 70% of top score
	scoreThreshold := int(float64(topScore) * scoreThresholdPercent)

	// separate by cache status and quality
	highQualityCached := make([]*hibiketorrent.AnimeTorrent, 0)
	lowQualityCached := make([]*hibiketorrent.AnimeTorrent, 0)
	uncached := make([]*hibiketorrent.AnimeTorrent, 0)

	for _, tws := range torrentsWithStatus {
		c, ok := candidateMap[tws.Torrent.InfoHash]
		if !ok {
			continue
		}

		if tws.IsCached {
			if c.score >= scoreThreshold {
				highQualityCached = append(highQualityCached, tws.Torrent)
			} else {
				lowQualityCached = append(lowQualityCached, tws.Torrent)
			}
		} else {
			uncached = append(uncached, tws.Torrent)
		}
	}

	// final order: high-quality cached, then uncached, then low-quality cached
	result := make([]*hibiketorrent.AnimeTorrent, 0, len(torrents))
	result = append(result, highQualityCached...)
	result = append(result, uncached...)
	result = append(result, lowQualityCached...)

	for i, t := range result {
		if i < 3 {
			c, ok := candidateMap[t.InfoHash]
			if !ok {
				continue
			}
			s.logger.Debug().Str("name", t.Name).Int("seeders", t.Seeders).Int("score", c.score).Str("provider", t.Provider).Msg("autoselect: Top candidates")
		}
	}

	return result
}

func (s *AutoSelect) calculateScore(c *candidate, profile *anime.AutoSelectProfile) int {
	score := 0
	parsed := c.parsed
	t := c.torrent

	// Boost by provider order

	if profile == nil {
		return score
	}

	//// Awlays best releases as a base line if user doesn't reject them
	//if (profile.BestReleasePreference != anime.AutoSelectPreferenceAvoid && profile.BestReleasePreference != anime.AutoSelectPreferenceNever) &&
	//	c.torrent.IsBestRelease && (t.Seeders == -1 || t.Seeders > 2) {
	//	// boost only if user isn't looking for low resolutions
	//	score += scoreBestReleaseBase
	//}

	// Resolution
	if len(profile.Resolutions) > 0 {
		for i, res := range profile.Resolutions {
			if strings.EqualFold(parsed.VideoResolution, res) {
				score += scoreResolutionBase - (i * scoreResolutionDecay)
				break
			}
		}
	}

	// Providers
	if len(profile.Providers) > 0 {
		for i, provider := range profile.Providers {
			if strings.EqualFold(t.Provider, provider) {
				score += scoreProviderBase - (i * scoreProviderDecay)
				break
			}
		}
	}

	// Release groups
	if len(profile.ReleaseGroups) > 0 {
		for i, group := range profile.ReleaseGroups {
			if strings.EqualFold(parsed.ReleaseGroup, group) {
				score += scoreReleaseGroupBase - (i * scoreReleaseGroupDecay)
				break
			}
		}
	}

	// Codec
	if len(profile.PreferredCodecs) > 0 {
		for i, codecs := range profile.PreferredCodecs {
			for _, codec := range strings.Split(codecs, ",") {
				codec = strings.TrimSpace(codec)
				if slices.ContainsFunc(parsed.VideoTerm, func(vt string) bool {
					return strings.EqualFold(vt, codec)
				}) || slices.ContainsFunc(parsed.AudioTerm, func(at string) bool {
					return strings.EqualFold(at, codec)
				}) {
					score += scoreCodecBase - (i * scoreCodecDecay)
					break
				}
				// Fallback check
				if strings.Contains(c.lowerName, strings.ToLower(codec)) {
					score += scoreCodecBase - (i * scoreCodecDecay)
					break
				}
			}
		}
	}

	// Source
	if len(profile.PreferredSources) > 0 {
		for i, sources := range profile.PreferredSources {
			for _, source := range strings.Split(sources, ",") {
				source = strings.TrimSpace(source)
				if slices.ContainsFunc(parsed.Source, func(src string) bool {
					return strings.EqualFold(src, source)
				}) {
					score += scoreSourceBase - (i * scoreSourceDecay)
					break
				}
				if strings.Contains(c.lowerName, strings.ToLower(source)) {
					score += scoreSourceBase - (i * scoreSourceDecay)
					break
				}
			}
		}
	}

	// Language
	if len(profile.PreferredLanguages) > 0 {
		for i, languages := range profile.PreferredLanguages {
			for _, lang := range strings.Split(languages, ",") {
				lang = strings.TrimSpace(lang)
				if slices.ContainsFunc(parsed.Language, func(pl string) bool {
					return strings.EqualFold(pl, lang)
				}) {
					score += scoreLanguageBase - (i * scoreLanguageDecay)
					break
				}
			}
		}
	}

	// Multiple audio preference (prefer/avoid)
	isMultiAudio := containsMultiOrDual(parsed.AudioTerm)
	if profile.MultipleAudioPreference == anime.AutoSelectPreferencePrefer && isMultiAudio {
		score += scoreMultiAudio
	}
	if profile.MultipleAudioPreference == anime.AutoSelectPreferenceAvoid && isMultiAudio {
		score -= scoreMultiAudio
	}

	// Multiple subs preference (prefer/avoid)
	isMultiSubs := containsMultiOrDual(parsed.Subtitles)
	if profile.MultipleSubsPreference == anime.AutoSelectPreferencePrefer && isMultiSubs {
		score += scoreMultiSubs
	}
	if profile.MultipleSubsPreference == anime.AutoSelectPreferenceAvoid && isMultiSubs {
		score -= scoreMultiSubs
	}

	// Batch preference (prefer/avoid)
	isBatch := t.IsBatch
	if profile.BatchPreference == anime.AutoSelectPreferencePrefer && isBatch {
		score += scoreBatch
	}
	if profile.BatchPreference == anime.AutoSelectPreferenceAvoid && isBatch {
		score -= scoreBatch
	}

	// Best release preference (prefer/avoid)
	isBestRelease := t.IsBestRelease && (t.Seeders == -1 || t.Seeders > 2)
	if profile.BestReleasePreference == anime.AutoSelectPreferencePrefer && isBestRelease {
		score += scoreBestRelease
	}
	if profile.BestReleasePreference == anime.AutoSelectPreferenceAvoid && isBestRelease {
		score -= scoreBestRelease
	}

	return score
}
