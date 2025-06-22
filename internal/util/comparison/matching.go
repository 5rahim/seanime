// Package comparison contains helpers related to comparison, comparison and filtering of media titles.
package comparison

import (
	"github.com/adrg/strutil/metrics"
)

// LevenshteinResult is a struct that holds a string and its Levenshtein distance compared to another string.
type LevenshteinResult struct {
	OriginalValue *string
	Value         *string
	Distance      int
}

// CompareWithLevenshtein compares a string to a slice of strings and returns a slice of LevenshteinResult containing the Levenshtein distance for each string.
func CompareWithLevenshtein(v *string, vals []*string) []*LevenshteinResult {
	return CompareWithLevenshteinCleanFunc(v, vals, func(val string) string {
		return val
	})
}
func CompareWithLevenshteinCleanFunc(v *string, vals []*string, cleanFunc func(val string) string) []*LevenshteinResult {

	lev := metrics.NewLevenshtein()
	lev.CaseSensitive = false
	//lev.DeleteCost = 1

	res := make([]*LevenshteinResult, len(vals))

	for _, val := range vals {
		res = append(res, &LevenshteinResult{
			OriginalValue: v,
			Value:         val,
			Distance:      lev.Distance(cleanFunc(*v), cleanFunc(*val)),
		})
	}

	return res
}

// FindBestMatchWithLevenshtein returns the best match from a slice of strings as a reference to a LevenshteinResult.
// It also returns a boolean indicating whether the best match was found.
func FindBestMatchWithLevenshtein(v *string, vals []*string) (*LevenshteinResult, bool) {
	res := CompareWithLevenshtein(v, vals)

	if len(res) == 0 {
		return nil, false
	}

	var bestResult *LevenshteinResult
	for _, result := range res {
		if bestResult == nil || result.Distance < bestResult.Distance {
			bestResult = result
		}
	}

	return bestResult, true
}

//----------------------------------------------------------------------------------------------------------------------

// JaroWinklerResult is a struct that holds a string and its JaroWinkler distance compared to another string.
type JaroWinklerResult struct {
	OriginalValue *string
	Value         *string
	Rating        float64
}

// CompareWithJaroWinkler compares a string to a slice of strings and returns a slice of JaroWinklerResult containing the JaroWinkler distance for each string.
func CompareWithJaroWinkler(v *string, vals []*string) []*JaroWinklerResult {

	jw := metrics.NewJaroWinkler()
	jw.CaseSensitive = false

	res := make([]*JaroWinklerResult, len(vals))

	for _, val := range vals {
		res = append(res, &JaroWinklerResult{
			OriginalValue: v,
			Value:         val,
			Rating:        jw.Compare(*v, *val),
		})
	}

	return res
}

// FindBestMatchWithJaroWinkler returns the best match from a slice of strings as a reference to a JaroWinklerResult.
// It also returns a boolean indicating whether the best match was found.
func FindBestMatchWithJaroWinkler(v *string, vals []*string) (*JaroWinklerResult, bool) {
	res := CompareWithJaroWinkler(v, vals)

	if len(res) == 0 {
		return nil, false
	}

	var bestResult *JaroWinklerResult
	for _, result := range res {
		if bestResult == nil || result.Rating > bestResult.Rating {
			bestResult = result
		}
	}

	return bestResult, true
}

//----------------------------------------------------------------------------------------------------------------------

// JaccardResult is a struct that holds a string and its Jaccard distance compared to another string.
type JaccardResult struct {
	OriginalValue *string
	Value         *string
	Rating        float64
}

// CompareWithJaccard compares a string to a slice of strings and returns a slice of JaccardResult containing the Jaccard distance for each string.
func CompareWithJaccard(v *string, vals []*string) []*JaccardResult {

	jw := metrics.NewJaccard()
	jw.CaseSensitive = false
	jw.NgramSize = 1

	res := make([]*JaccardResult, len(vals))

	for _, val := range vals {
		res = append(res, &JaccardResult{
			OriginalValue: v,
			Value:         val,
			Rating:        jw.Compare(*v, *val),
		})
	}

	return res
}

// FindBestMatchWithJaccard returns the best match from a slice of strings as a reference to a JaccardResult.
// It also returns a boolean indicating whether the best match was found.
func FindBestMatchWithJaccard(v *string, vals []*string) (*JaccardResult, bool) {
	res := CompareWithJaccard(v, vals)

	if len(res) == 0 {
		return nil, false
	}

	var bestResult *JaccardResult
	for _, result := range res {
		if bestResult == nil || result.Rating > bestResult.Rating {
			bestResult = result
		}
	}

	return bestResult, true
}

//----------------------------------------------------------------------------------------------------------------------

type SorensenDiceResult struct {
	OriginalValue *string
	Value         *string
	Rating        float64
}

func CompareWithSorensenDice(v *string, vals []*string) []*SorensenDiceResult {

	dice := metrics.NewSorensenDice()
	dice.CaseSensitive = false

	res := make([]*SorensenDiceResult, len(vals))

	for _, val := range vals {
		res = append(res, &SorensenDiceResult{
			OriginalValue: v,
			Value:         val,
			Rating:        dice.Compare(*v, *val),
		})
	}

	return res
}

func FindBestMatchWithSorensenDice(v *string, vals []*string) (*SorensenDiceResult, bool) {
	res := CompareWithSorensenDice(v, vals)

	if len(res) == 0 {
		return nil, false
	}

	var bestResult *SorensenDiceResult
	for _, result := range res {
		if bestResult == nil || result.Rating > bestResult.Rating {
			bestResult = result
		}
	}

	return bestResult, true
}

func EliminateLeastSimilarValue(arr []string) []string {
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
