// Code generated by 'yaegi extract github.com/adrg/strutil/metrics'. DO NOT EDIT.

package yaegi_interp

import (
	"github.com/adrg/strutil/metrics"
	"reflect"
)

func init() {
	Symbols["github.com/adrg/strutil/metrics/metrics"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"NewHamming":            reflect.ValueOf(metrics.NewHamming),
		"NewJaccard":            reflect.ValueOf(metrics.NewJaccard),
		"NewJaro":               reflect.ValueOf(metrics.NewJaro),
		"NewJaroWinkler":        reflect.ValueOf(metrics.NewJaroWinkler),
		"NewLevenshtein":        reflect.ValueOf(metrics.NewLevenshtein),
		"NewOverlapCoefficient": reflect.ValueOf(metrics.NewOverlapCoefficient),
		"NewSmithWatermanGotoh": reflect.ValueOf(metrics.NewSmithWatermanGotoh),
		"NewSorensenDice":       reflect.ValueOf(metrics.NewSorensenDice),

		// type definitions
		"Hamming":            reflect.ValueOf((*metrics.Hamming)(nil)),
		"Jaccard":            reflect.ValueOf((*metrics.Jaccard)(nil)),
		"Jaro":               reflect.ValueOf((*metrics.Jaro)(nil)),
		"JaroWinkler":        reflect.ValueOf((*metrics.JaroWinkler)(nil)),
		"Levenshtein":        reflect.ValueOf((*metrics.Levenshtein)(nil)),
		"MatchMismatch":      reflect.ValueOf((*metrics.MatchMismatch)(nil)),
		"OverlapCoefficient": reflect.ValueOf((*metrics.OverlapCoefficient)(nil)),
		"SmithWatermanGotoh": reflect.ValueOf((*metrics.SmithWatermanGotoh)(nil)),
		"SorensenDice":       reflect.ValueOf((*metrics.SorensenDice)(nil)),
		"Substitution":       reflect.ValueOf((*metrics.Substitution)(nil)),

		// interface wrapper definitions
		"_Substitution": reflect.ValueOf((*_github_com_adrg_strutil_metrics_Substitution)(nil)),
	}
}

// _github_com_adrg_strutil_metrics_Substitution is an interface wrapper for Substitution type
type _github_com_adrg_strutil_metrics_Substitution struct {
	IValue   interface{}
	WCompare func(a []rune, idxA int, b []rune, idxB int) float64
	WMax     func() float64
	WMin     func() float64
}

func (W _github_com_adrg_strutil_metrics_Substitution) Compare(a []rune, idxA int, b []rune, idxB int) float64 {
	return W.WCompare(a, idxA, b, idxB)
}
func (W _github_com_adrg_strutil_metrics_Substitution) Max() float64 {
	return W.WMax()
}
func (W _github_com_adrg_strutil_metrics_Substitution) Min() float64 {
	return W.WMin()
}
