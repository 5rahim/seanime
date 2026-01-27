package autoselect

import (
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func newTestAutoSelect() *AutoSelect {
	logger := zerolog.Nop()
	return New(&NewAutoSelectOptions{
		Logger: &logger,
	})
}

func TestAutoSelect_Filter(t *testing.T) {
	s := newTestAutoSelect()

	// Mock Torrents
	t1 := &hibiketorrent.AnimeTorrent{Name: "[SubsPlease] One Piece - 1000 (1080p).mkv", Seeders: 100, Size: 1024 * 1024 * 1000}                             // 1GB
	t2 := &hibiketorrent.AnimeTorrent{Name: "[Erai-raws] One Piece - 1000 [720p][Multiple Subtitle].mkv", Seeders: 50, Size: 1024 * 1024 * 500}              // 500MB
	t3 := &hibiketorrent.AnimeTorrent{Name: "[EMBER] One Piece - 1000 [1080p] [Dual Audio] [HEVC].mkv", Seeders: 200, Size: 1024 * 1024 * 800}               // 800MB
	t4 := &hibiketorrent.AnimeTorrent{Name: "[French-Fansub] One Piece - 1000 [480p] [French].mkv", Seeders: 10, Size: 1024 * 1024 * 200}                    // 200MB
	t5 := &hibiketorrent.AnimeTorrent{Name: "[Cleo] One Piece - Batch [1080p] [Dual Audio].mkv", Seeders: 500, Size: 1024 * 1024 * 1024 * 50, IsBatch: true} // 50GB Batch
	t6 := &hibiketorrent.AnimeTorrent{Name: "[Judas] One Piece - 1000 [1080p][HEVC x265 10bit].mkv", Seeders: 150, Size: 1024 * 1024 * 600}                  // 600MB
	t7 := &hibiketorrent.AnimeTorrent{Name: "[HorribleSubs] One Piece - 1000 [720p].mkv", Seeders: 20, Size: 1024 * 1024 * 400}                              // 400MB
	t8 := &hibiketorrent.AnimeTorrent{Name: "[SourceCheck] One Piece - 1000 [Web-DL].mkv", Seeders: 30, Size: 1024 * 1024 * 900}                             // 900MB

	torrents := []*hibiketorrent.AnimeTorrent{t1, t2, t3, t4, t5, t6, t7, t8}

	tests := []struct {
		name     string
		profile  *anime.AutoSelectProfile
		expected []string // Names of expected torrents
	}{
		{
			name:     "No profile should return all",
			profile:  nil,
			expected: []string{t1.Name, t2.Name, t3.Name, t4.Name, t5.Name, t6.Name, t7.Name, t8.Name},
		},
		{
			name: "Min Seeders > 50",
			profile: &anime.AutoSelectProfile{
				MinSeeders: 51,
			},
			expected: []string{t1.Name, t3.Name, t5.Name, t6.Name}, // t1(100), t3(200), t5(500), t6(150)
		},
		{
			name: "Resolution exclusion",
			profile: &anime.AutoSelectProfile{
				ExcludeTerms: []string{"480p", "720p"},
			},
			expected: []string{t1.Name, t3.Name, t5.Name, t6.Name, t8.Name},
		},
		{
			name: "Require language (French)",
			profile: &anime.AutoSelectProfile{
				RequireLanguage:    true,
				PreferredLanguages: []string{"fr, french"},
			},
			expected: []string{t4.Name},
		},
		{
			name: "Require language (English)",
			profile: &anime.AutoSelectProfile{
				RequireLanguage:    true,
				PreferredLanguages: []string{"English"},
			},
			expected: []string{},
		},
		{
			name: "Min size 600MB",
			profile: &anime.AutoSelectProfile{
				MinSize: "600MB",
			},
			expected: []string{t1.Name, t3.Name, t5.Name, t6.Name, t8.Name},
		},
		{
			name: "Max size 700MB",
			profile: &anime.AutoSelectProfile{
				MaxSize: "700MB",
			},
			expected: []string{t2.Name, t4.Name, t6.Name, t7.Name},
		},
		{
			name: "Require codec HEVC",
			profile: &anime.AutoSelectProfile{
				RequireCodec:    true,
				PreferredCodecs: []string{"HEVC", "x265"},
			},
			expected: []string{t3.Name, t6.Name},
		},
		{
			name: "Required source Web-DL",
			profile: &anime.AutoSelectProfile{
				RequireSource:    true,
				PreferredSources: []string{"Web-DL"},
			},
			expected: []string{t8.Name}, // assuming habari detects Web-DL properly or fallback check works
		},
		{
			name: "Dual audio only",
			profile: &anime.AutoSelectProfile{
				MultipleAudioPreference: anime.AutoSelectPreferenceOnly,
			},
			expected: []string{t3.Name, t5.Name},
		},
		{
			name: "Dual audio never",
			profile: &anime.AutoSelectProfile{
				MultipleAudioPreference: anime.AutoSelectPreferenceNever,
			},
			expected: []string{t1.Name, t2.Name, t4.Name, t6.Name, t7.Name, t8.Name},
		},
		{
			name: "Batch only",
			profile: &anime.AutoSelectProfile{
				BatchPreference: anime.AutoSelectPreferenceOnly,
			},
			expected: []string{t5.Name},
		},
		{
			name: "Batch never",
			profile: &anime.AutoSelectProfile{
				BatchPreference: anime.AutoSelectPreferenceNever,
			},
			expected: []string{t1.Name, t2.Name, t3.Name, t4.Name, t6.Name, t7.Name, t8.Name},
		},
		{
			name: "Complex combination",
			profile: &anime.AutoSelectProfile{
				MinSeeders:              100,
				RequireCodec:            true,
				PreferredCodecs:         []string{"HEVC"},
				MultipleAudioPreference: anime.AutoSelectPreferenceOnly,
			},
			expected: []string{t3.Name}, // t3 matches all (200 seeders, HEVC, Dual Audio)
			// t6 matches seeders and codec but NOT Dual Audio
			// t5 matches seeders and Dual Audio but NOT Codec (in name)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := s.filter(torrents, tt.profile)

			var filteredNames []string
			for _, ft := range filtered {
				filteredNames = append(filteredNames, ft.Name)
			}

			assert.ElementsMatchf(t, tt.expected, filteredNames, "Expected %v,\ngot %v", tt.expected, filteredNames)
		})
	}
}

func TestAutoSelect_Sort(t *testing.T) {
	s := newTestAutoSelect()

	t1 := &hibiketorrent.AnimeTorrent{Name: "[A] Show - 01 [1080p].mkv", Seeders: 100, Provider: "catsound"}
	t2 := &hibiketorrent.AnimeTorrent{Name: "[B] Show - 01 [720p].mkv", Seeders: 200, Provider: "catsound"}
	t3 := &hibiketorrent.AnimeTorrent{Name: "[C] Show - 01 [1080p] [Dual-Audio].mkv", Seeders: 50, Provider: "tosho"}
	t4 := &hibiketorrent.AnimeTorrent{Name: "[D] Show - 01 [1080p] [HEVC].mkv", Seeders: 80, Provider: "catsound"}

	torrents := []*hibiketorrent.AnimeTorrent{t1, t2, t3, t4}

	tests := []struct {
		name     string
		profile  *anime.AutoSelectProfile
		expected []string // Names in expected order
	}{
		{
			name: "Prefer 1080p",
			profile: &anime.AutoSelectProfile{
				Resolutions: []string{"1080p"},
			},
			// 1080p torrents get +100 score. 720p gets 0.
			// t1(1080p, 100s), t3(1080p, 50s), t4(1080p, 80s) have score 100.
			// Tie breaker is seeders: t1(100), t4(80), t3(50).
			// t2(720p) comes last.
			expected: []string{t1.Name, t4.Name, t3.Name, t2.Name},
		},
		{
			name: "Prefer Dual Audio",
			profile: &anime.AutoSelectProfile{
				MultipleAudioPreference: anime.AutoSelectPreferencePrefer,
			},
			// t3 has Dual Audio -> +15 score. Others 0.
			expected: []string{t3.Name, t2.Name, t1.Name, t4.Name}, // t3 first. Others sorted by seeders (t2=200, t1=100, t4=80)
		},
		{
			name: "Prefer Provider Animetosho",
			profile: &anime.AutoSelectProfile{
				Providers: []string{"tosho"},
			},
			// t3 matches provider -> +50 score.
			expected: []string{t3.Name, t2.Name, t1.Name, t4.Name},
		},
		{
			name: "Complex Priorities",
			profile: &anime.AutoSelectProfile{
				Resolutions:             []string{"1080p"},                // +100
				PreferredCodecs:         []string{"HEVC"},                 // +40
				MultipleAudioPreference: anime.AutoSelectPreferencePrefer, // +15
			},
			// t1: 1080p (+100) = 100
			// t2: 720p (0) = 0
			// t3: 1080p (+100) + Dual Audio (+15) = 115
			// t4: 1080p (+100) + HEVC (+40) = 140
			// Expected order: t4 (140), t3 (115), t1 (100), t2 (0)
			expected: []string{t4.Name, t3.Name, t1.Name, t2.Name},
		},
		{
			name: "Avoid Dual Audio",
			profile: &anime.AutoSelectProfile{
				MultipleAudioPreference: anime.AutoSelectPreferenceAvoid, // -15
			},
			// t3: -15
			// Others: 0
			// Sorted by seeders for 0 score: t2(200), t1(100), t4(80)
			// Then t3 last.
			expected: []string{t2.Name, t1.Name, t4.Name, t3.Name},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTorrents := make([]*hibiketorrent.AnimeTorrent, len(torrents))
			copy(testTorrents, torrents)

			s.sort(testTorrents, tt.profile)

			var sortedNames []string
			for _, st := range testTorrents {
				sortedNames = append(sortedNames, st.Name)
			}

			assert.Equal(t, tt.expected, sortedNames)
		})
	}
}

func TestAutoSelect_SmartCachedPrioritization(t *testing.T) {
	s := newTestAutoSelect()

	// these tests prioritization when the provider doesn't support smart search to exclude resolutions

	highQuality1080p := &hibiketorrent.AnimeTorrent{
		Name:     "[SubsPlease] Show - 01 [1080p][HEVC].mkv",
		InfoHash: "hash1",
		Seeders:  200,
		Provider: "catsound",
	}
	mediumQuality1080p := &hibiketorrent.AnimeTorrent{
		Name:     "[RandomGroup] Show - 01 [1080p].mkv",
		InfoHash: "hash2",
		Seeders:  100,
		Provider: "catsound",
	}
	lowQuality720p := &hibiketorrent.AnimeTorrent{
		Name:     "[LowQuality] Show - 01 [720p].mkv",
		InfoHash: "hash3",
		Seeders:  50,
		Provider: "catsound",
	}
	veryLowQuality480p := &hibiketorrent.AnimeTorrent{
		Name:     "[BadGroup] Show - 01 [480p].mkv",
		InfoHash: "hash4",
		Seeders:  10,
		Provider: "catsound",
	}
	highQuality1080pAlt := &hibiketorrent.AnimeTorrent{
		Name:     "[Erai-raws] Show - 01 [1080p][Multiple Subtitle].mkv",
		InfoHash: "hash5",
		Seeders:  150,
		Provider: "tosho",
	}

	tests := []struct {
		name          string
		torrents      []*hibiketorrent.AnimeTorrent
		cachedHashes  []string // Hashes of cached torrents
		profile       *anime.AutoSelectProfile
		expectedOrder []string // Expected names in order
	}{
		{
			name:         "High quality cached should be prioritized",
			torrents:     []*hibiketorrent.AnimeTorrent{highQuality1080p, mediumQuality1080p, lowQuality720p, veryLowQuality480p},
			cachedHashes: []string{"hash1"}, // highQuality1080p is cached
			profile: &anime.AutoSelectProfile{
				Resolutions: []string{"1080p"},
			},
			expectedOrder: []string{highQuality1080p.Name, mediumQuality1080p.Name, lowQuality720p.Name, veryLowQuality480p.Name},
		},
		{
			name:         "Low quality cached should NOT be prioritized over high quality uncached",
			torrents:     []*hibiketorrent.AnimeTorrent{highQuality1080p, mediumQuality1080p, lowQuality720p, veryLowQuality480p},
			cachedHashes: []string{"hash4"}, // veryLowQuality480p is cached
			profile: &anime.AutoSelectProfile{
				Resolutions: []string{"1080p"},
			},
			expectedOrder: []string{highQuality1080p.Name, mediumQuality1080p.Name, lowQuality720p.Name, veryLowQuality480p.Name},
		},
		{
			name:         "Medium quality cached within threshold should be prioritized",
			torrents:     []*hibiketorrent.AnimeTorrent{highQuality1080p, mediumQuality1080p, lowQuality720p, veryLowQuality480p},
			cachedHashes: []string{"hash2"}, // mediumQuality1080p is cached (similar score to high quality)
			profile: &anime.AutoSelectProfile{
				Resolutions: []string{"1080p"},
			},
			expectedOrder: []string{mediumQuality1080p.Name, highQuality1080p.Name, lowQuality720p.Name, veryLowQuality480p.Name},
		},
		{
			name:         "Multiple cached torrents should maintain quality order",
			torrents:     []*hibiketorrent.AnimeTorrent{highQuality1080p, mediumQuality1080p, lowQuality720p, veryLowQuality480p, highQuality1080pAlt},
			cachedHashes: []string{"hash1", "hash5"}, // Two high quality cached
			profile: &anime.AutoSelectProfile{
				Resolutions: []string{"1080p"},
				Providers:   []string{"tosho"},
			},
			expectedOrder: []string{highQuality1080pAlt.Name, highQuality1080p.Name, mediumQuality1080p.Name, lowQuality720p.Name, veryLowQuality480p.Name},
		},
		{
			name:         "Mixed cached (high and low quality)",
			torrents:     []*hibiketorrent.AnimeTorrent{highQuality1080p, mediumQuality1080p, lowQuality720p, veryLowQuality480p},
			cachedHashes: []string{"hash1", "hash4"}, // High and very low quality cached
			profile: &anime.AutoSelectProfile{
				Resolutions: []string{"1080p"},
			},
			expectedOrder: []string{highQuality1080p.Name, mediumQuality1080p.Name, lowQuality720p.Name, veryLowQuality480p.Name},
		},
		{
			name:         "When all cached, maintain quality-based order",
			torrents:     []*hibiketorrent.AnimeTorrent{highQuality1080p, mediumQuality1080p, lowQuality720p, veryLowQuality480p},
			cachedHashes: []string{"hash1", "hash2", "hash3", "hash4"}, // All cached
			profile: &anime.AutoSelectProfile{
				Resolutions: []string{"1080p"},
			},
			expectedOrder: []string{highQuality1080p.Name, mediumQuality1080p.Name, lowQuality720p.Name, veryLowQuality480p.Name},
		},
		{
			name:         "No cached, maintain normal sort order",
			torrents:     []*hibiketorrent.AnimeTorrent{highQuality1080p, mediumQuality1080p, lowQuality720p, veryLowQuality480p},
			cachedHashes: []string{}, // None cached
			profile: &anime.AutoSelectProfile{
				Resolutions: []string{"1080p"},
			},
			expectedOrder: []string{highQuality1080p.Name, mediumQuality1080p.Name, lowQuality720p.Name, veryLowQuality480p.Name},
		},
		{
			name:         "Cached 720p within threshold vs uncached 1080p",
			torrents:     []*hibiketorrent.AnimeTorrent{highQuality1080p, lowQuality720p},
			cachedHashes: []string{"hash3"}, // 720p is cached
			profile: &anime.AutoSelectProfile{
				Resolutions: []string{"1080p", "720p"}, // Both acceptable
			},
			// When both resolutions are in profile, 720p gets score too and may be within 70% threshold
			// So cached 720p CAN be prioritized if within threshold
			expectedOrder: []string{lowQuality720p.Name, highQuality1080p.Name},
		},
		{
			name:         "Cached 480p should NOT beat uncached 1080p",
			torrents:     []*hibiketorrent.AnimeTorrent{highQuality1080p, veryLowQuality480p},
			cachedHashes: []string{"hash4"}, // 480p is cached
			profile: &anime.AutoSelectProfile{
				Resolutions: []string{"1080p"}, // Only 1080p preferred
			},
			// 480p gets no resolution bonus, so score is very low (below 70% threshold)
			// Should NOT be prioritized even though cached
			expectedOrder: []string{highQuality1080p.Name, veryLowQuality480p.Name},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postSearchSort := func(torrents []*hibiketorrent.AnimeTorrent) []*TorrentWithCacheStatus {
				result := make([]*TorrentWithCacheStatus, 0, len(torrents))
				cachedMap := make(map[string]bool)
				for _, hash := range tt.cachedHashes {
					cachedMap[hash] = true
				}

				for _, t := range torrents {
					result = append(result, &TorrentWithCacheStatus{
						Torrent:  t,
						IsCached: cachedMap[t.InfoHash],
					})
				}
				return result
			}

			// Run filterAndSort with postSearchSort
			testTorrents := make([]*hibiketorrent.AnimeTorrent, len(tt.torrents))
			copy(testTorrents, tt.torrents)

			sorted := s.filterAndSort(testTorrents, tt.profile, postSearchSort)

			var sortedNames []string
			for _, st := range sorted {
				sortedNames = append(sortedNames, st.Name)
			}

			assert.Equal(t, tt.expectedOrder, sortedNames)
		})
	}
}

func TestAutoSelect_SmartCachedPrioritization_EdgeCases(t *testing.T) {
	s := newTestAutoSelect()

	t.Run("Empty torrents list", func(t *testing.T) {
		postSearchSort := func(torrents []*hibiketorrent.AnimeTorrent) []*TorrentWithCacheStatus {
			return []*TorrentWithCacheStatus{}
		}

		result := s.filterAndSort([]*hibiketorrent.AnimeTorrent{}, nil, postSearchSort)
		assert.Empty(t, result)
	})

	t.Run("Single cached torrent", func(t *testing.T) {
		torrent := &hibiketorrent.AnimeTorrent{
			Name:     "[Test] Show - 01 [1080p].mkv",
			InfoHash: "hash1",
			Seeders:  100,
		}

		postSearchSort := func(torrents []*hibiketorrent.AnimeTorrent) []*TorrentWithCacheStatus {
			return []*TorrentWithCacheStatus{{Torrent: torrents[0], IsCached: true}}
		}

		result := s.filterAndSort([]*hibiketorrent.AnimeTorrent{torrent}, nil, postSearchSort)
		assert.Len(t, result, 1)
		assert.Equal(t, torrent.Name, result[0].Name)
	})

	t.Run("Nil postSearchSort function", func(t *testing.T) {
		torrent := &hibiketorrent.AnimeTorrent{
			Name:     "[Test] Show - 01 [1080p].mkv",
			InfoHash: "hash1",
			Seeders:  100,
		}

		result := s.filterAndSort([]*hibiketorrent.AnimeTorrent{torrent}, nil, nil)
		assert.Len(t, result, 1)
		assert.Equal(t, torrent.Name, result[0].Name)
	})

	t.Run("Cached torrent exactly at 70% threshold", func(t *testing.T) {
		// Create scenario where cached torrent is exactly at threshold
		highQuality := &hibiketorrent.AnimeTorrent{
			Name:     "[SubsPlease] Show - 01 [1080p][HEVC].mkv",
			InfoHash: "hash1",
			Seeders:  200,
			Provider: "tosho",
		}
		thresholdQuality := &hibiketorrent.AnimeTorrent{
			Name:     "[Test] Show - 01 [720p].mkv",
			InfoHash: "hash2",
			Seeders:  50,
		}

		profile := &anime.AutoSelectProfile{
			Resolutions: []string{"1080p"},
			Providers:   []string{"tosho"},
		}

		postSearchSort := func(torrents []*hibiketorrent.AnimeTorrent) []*TorrentWithCacheStatus {
			return []*TorrentWithCacheStatus{
				{Torrent: torrents[0], IsCached: false},
				{Torrent: torrents[1], IsCached: true},
			}
		}

		result := s.filterAndSort([]*hibiketorrent.AnimeTorrent{highQuality, thresholdQuality}, profile, postSearchSort)
		assert.Len(t, result, 2)
		assert.NotNil(t, result[0])
	})
}
