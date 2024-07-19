package parallel

import (
	"github.com/samber/lo"
	"seanime/internal/util/limiter"
	"sync"
)

// EachTask iterates over elements of collection and invokes the task function for each element.
// `task` is called in parallel.
func EachTask[T any](collection []T, task func(item T, index int)) {
	var wg sync.WaitGroup

	for i, item := range collection {
		wg.Add(1)
		go func(_item T, _i int) {
			defer wg.Done()
			task(_item, _i)
		}(item, i)
	}

	wg.Wait()
}

// EachTaskL is the same as EachTask, but takes a pointer to limiter.Limiter.
func EachTaskL[T any](collection []T, rl *limiter.Limiter, task func(item T, index int)) {
	var wg sync.WaitGroup

	for i, item := range collection {
		wg.Add(1)
		go func(_item T, _i int) {
			defer wg.Done()
			rl.Wait()
			task(_item, _i)
		}(item, i)
	}

	wg.Wait()
}

type SettledResults[T comparable, R any] struct {
	Collection []T
	Fulfilled  map[T]R
	Results    []R
	Rejected   map[T]error
}

// NewSettledResults returns a pointer to a new SettledResults struct.
func NewSettledResults[T comparable, R any](c []T) *SettledResults[T, R] {
	return &SettledResults[T, R]{
		Collection: c,
		Fulfilled:  map[T]R{},
		Rejected:   map[T]error{},
	}
}

// GetFulfilledResults returns a pointer to the slice of fulfilled results and a boolean indicating whether the slice is not nil.
func (sr *SettledResults[T, R]) GetFulfilledResults() (*[]R, bool) {
	if sr.Results != nil {
		return &sr.Results, true
	}
	return nil, false
}

// AllSettled executes the provided task function once, in parallel for each element in the slice passed to NewSettledResults.
// It returns a map of fulfilled results and a map of errors whose keys are the elements of the slice.
func (sr *SettledResults[T, R]) AllSettled(task func(item T, index int) (R, error)) (map[T]R, map[T]error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, item := range sr.Collection {
		wg.Add(1)
		go func(_item T, _i int) {

			res, err := task(_item, _i)

			mu.Lock()
			if err != nil {
				sr.Rejected[_item] = err
			} else {
				sr.Fulfilled[_item] = res
			}
			mu.Unlock()
			wg.Done()

		}(item, i)
	}

	wg.Wait()

	sr.Results = lo.MapToSlice(sr.Fulfilled, func(key T, value R) R {
		return value
	})

	return sr.Fulfilled, sr.Rejected
}
