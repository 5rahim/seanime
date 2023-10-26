// Package comparison contains helpers related to comparison, comparison and filtering of media titles.
package comparison

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

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type SorensenDiceResult struct {
	Value  string
	Rating float64
}

func CompareWithSorensenDice(v string, vals []string) []*SorensenDiceResult {

	lev := metrics.NewSorensenDice()
	lev.CaseSensitive = false

	res := make([]*SorensenDiceResult, len(vals))

	for _, val := range vals {
		res = append(res, &SorensenDiceResult{
			Value:  val,
			Rating: lev.Compare(v, val),
		})
	}

	return res
}

func FindBestMatchWithSorensenDice(v string, vals []string) (*SorensenDiceResult, bool) {
	res := CompareWithSorensenDice(v, vals)

	n := lo.Reduce(res, func(prev *SorensenDiceResult, curr *SorensenDiceResult, index int) *SorensenDiceResult {
		if prev == nil || curr == nil {
			return curr
		}
		if prev.Rating > curr.Rating {
			return prev
		} else {
			return curr
		}
	}, &SorensenDiceResult{})

	if n == nil {
		return nil, false
	}

	return n, true
}

func EliminateLestSimilarValue(arr []string) []string {
	if len(arr) < 3 {
		return arr
	}

	sd := metrics.NewSorensenDice()
	sd.CaseSensitive = false

	leastSimilarIndex := -1
	leastSimilarScore := 2.0

	for i := 0; i < len(arr); i++ {
		totalSimilarity := 0.0

		for j := 0; j < len(arr); j++ {
			if i != j {
				score := sd.Compare(arr[i], arr[j])
				totalSimilarity += score
			}
		}

		if totalSimilarity < leastSimilarScore {
			leastSimilarScore = totalSimilarity
			leastSimilarIndex = i
		}
	}

	if leastSimilarIndex != -1 {
		arr = append(arr[:leastSimilarIndex], arr[leastSimilarIndex+1:]...)
	}

	return arr

}
