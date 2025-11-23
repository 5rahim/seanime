package util

import (
	"sync"
)

// Ref allows swapping the underlying value at runtime.
// Used for interfaces and other values that need to be updated dynamically.
type Ref[T comparable] struct {
	mu      sync.RWMutex
	current T
}

func NewRef[T comparable](initial T) *Ref[T] {
	return &Ref[T]{
		current: initial,
	}
}

// Get returns the current value safely.
func (s *Ref[T]) Get() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// Set updates the value safely.
func (s *Ref[T]) Set(newValue T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = newValue
}

func (s *Ref[T]) IsPresent() bool {
	if s == nil {
		return false
	}
	var zero T
	return s.Get() != zero
}

func (s *Ref[T]) IsAbsent() bool {
	if s == nil {
		return true
	}
	var zero T
	return s.Get() == zero
}
