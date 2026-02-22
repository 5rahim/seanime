package scanner

import (
	"sync"
)

// Object pools to reduce allocs during scanning

var stringSlicePool = sync.Pool{
	New: func() interface{} {
		return new(make([]string, 0, 16))
	},
}

func getStringSlice() *[]string {
	return stringSlicePool.Get().(*[]string)
}

func putStringSlice(s *[]string) {
	*s = (*s)[:0]
	stringSlicePool.Put(s)
}

// tokenSetPool provides reusable maps for token set operations
var tokenSetPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]struct{}, 16)
	},
}

func getTokenSet() map[string]struct{} {
	return tokenSetPool.Get().(map[string]struct{})
}
func putTokenSet(m map[string]struct{}) {
	// clear the map
	for k := range m {
		delete(m, k)
	}
	tokenSetPool.Put(m)
}
