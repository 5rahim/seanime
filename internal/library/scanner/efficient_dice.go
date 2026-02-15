package scanner

import (
	"strings"
	"sync"
)

// EfficientDice is Sorensen-Dice implementation that minimizes allocs
type EfficientDice struct {
	ngramSize      int
	caseSensitive  bool
	ngramBufferA   []uint64
	ngramBufferB   []uint64
	ngramCountMapA map[uint64]int16
	ngramCountMapB map[uint64]int16
}

var EfficientDicePool = sync.Pool{
	New: func() interface{} {
		return &EfficientDice{
			ngramSize:      2,
			caseSensitive:  false,
			ngramBufferA:   make([]uint64, 0, 64),
			ngramBufferB:   make([]uint64, 0, 64),
			ngramCountMapA: make(map[uint64]int16, 64),
			ngramCountMapB: make(map[uint64]int16, 64),
		}
	},
}

func GetEfficientDice() *EfficientDice {
	return EfficientDicePool.Get().(*EfficientDice)
}
func PutEfficientDice(d *EfficientDice) {
	d.reset()
	EfficientDicePool.Put(d)
}

// reset clears internal state for reuse
func (d *EfficientDice) reset() {
	d.ngramBufferA = d.ngramBufferA[:0]
	d.ngramBufferB = d.ngramBufferB[:0]
	for k := range d.ngramCountMapA {
		delete(d.ngramCountMapA, k)
	}
	for k := range d.ngramCountMapB {
		delete(d.ngramCountMapB, k)
	}
}

// Compare calculates the Sorensen-Dice coefficient between two strings
func (d *EfficientDice) Compare(a, b string) float64 {
	if a == "" && b == "" {
		return 1.0
	}
	if a == "" || b == "" {
		return 0.0
	}

	// normalize case
	if !d.caseSensitive {
		a = strings.ToLower(a)
		b = strings.ToLower(b)
	}

	// bigrams for both strings
	d.generateNgrams(a, &d.ngramBufferA, d.ngramCountMapA)
	d.generateNgrams(b, &d.ngramBufferB, d.ngramCountMapB)

	if len(d.ngramBufferA) == 0 && len(d.ngramBufferB) == 0 {
		return 1.0
	}
	if len(d.ngramBufferA) == 0 || len(d.ngramBufferB) == 0 {
		return 0.0
	}

	// calculate intersection size
	intersection := 0
	for ngram, countA := range d.ngramCountMapA {
		if countB, exists := d.ngramCountMapB[ngram]; exists {
			// take minimum count for multiset intersection
			if countA < countB {
				intersection += int(countA)
			} else {
				intersection += int(countB)
			}
		}
	}

	// coefficient
	totalA := len(d.ngramBufferA)
	totalB := len(d.ngramBufferB)

	return float64(2*intersection) / float64(totalA+totalB)
}

// generateNgrams generates bigrams from a string, stores them as uint64 hashes
func (d *EfficientDice) generateNgrams(s string, buffer *[]uint64, countMap map[uint64]int16) {
	*buffer = (*buffer)[:0]
	for k := range countMap {
		delete(countMap, k)
	}

	runes := []rune(s)
	if len(runes) < d.ngramSize {
		// single character strings
		if len(runes) == 1 {
			hash := uint64(runes[0])
			*buffer = append(*buffer, hash)
			countMap[hash] = 1
		}
		return
	}

	for i := 0; i <= len(runes)-d.ngramSize; i++ {
		// encode bigram as uint64: first rune in high 32 bits, second in low 32 bits
		hash := (uint64(runes[i]) << 32) | uint64(runes[i+1])
		*buffer = append(*buffer, hash)
		countMap[hash]++
	}
}

func CompareStrings(a, b string) float64 {
	d := GetEfficientDice()
	defer PutEfficientDice(d)
	return d.Compare(a, b)
}
