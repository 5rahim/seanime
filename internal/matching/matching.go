package matching

import (
	"github.com/adrg/strutil/metrics"
	"github.com/samber/lo"
)

// LevenshteinResult is a struct that holds a string and its Levenshtein distance compared to another string.
type LevenshteinResult struct {
	Value    string
	Distance int
}

// CompareWithLevenstein compares a string to a slice of strings and returns a slice of LevenshteinResult containing the Levenshtein distance for each string.
func CompareWithLevenstein(v string, vals []string) []*LevenshteinResult {

	lev := metrics.NewLevenshtein()
	lev.CaseSensitive = false

	res := make([]*LevenshteinResult, len(vals))

	for _, val := range vals {
		res = append(res, &LevenshteinResult{
			Value:    val,
			Distance: lev.Distance(v, val),
		})
	}

	return res
}

// FindBestMatchWithLevenstein returns the best match from a slice of strings as a reference to a LevenshteinResult.
// It also returns a boolean indicating whether the best match was found.
func FindBestMatchWithLevenstein(v string, vals []string) (*LevenshteinResult, bool) {
	res := CompareWithLevenstein(v, vals)

	n := lo.Reduce(res, func(prev *LevenshteinResult, curr *LevenshteinResult, index int) *LevenshteinResult {
		if prev == nil || curr == nil {
			return curr
		}
		if prev.Distance < curr.Distance {
			return prev
		} else {
			return curr
		}
	}, &LevenshteinResult{})

	if n == nil {
		return nil, false
	}

	return n, true
}
