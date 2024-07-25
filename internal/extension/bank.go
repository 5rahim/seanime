package extension

import (
	"seanime/internal/util/result"
	"sync"
)

type Bank[T BaseExtension] struct {
	extensions         *result.Map[string, T]
	extensionAddedCh   chan struct{}
	extensionRemovedCh chan struct{}
	mu                 sync.Mutex
}

func NewBank[T BaseExtension]() *Bank[T] {
	return &Bank[T]{
		extensions:         result.NewResultMap[string, T](),
		extensionAddedCh:   make(chan struct{}),
		extensionRemovedCh: make(chan struct{}),
		mu:                 sync.Mutex{},
	}
}

func (b *Bank[T]) Set(id string, ext T) {
	// Add the extension to the map
	b.extensions.Set(id, ext)

	// Notify listeners that an extension has been added
	b.mu.Lock()
	defer b.mu.Unlock()

	close(b.extensionAddedCh)
	b.extensionAddedCh = make(chan struct{})
}

func (b *Bank[T]) Get(id string) (T, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.extensions.Get(id)
}

func (b *Bank[T]) Delete(id string) {
	// Delete the extension from the map
	b.extensions.Delete(id)

	// Notify listeners that an extension has been removed
	b.mu.Lock()
	defer b.mu.Unlock()

	close(b.extensionRemovedCh)
	b.extensionRemovedCh = make(chan struct{})
}

func (b *Bank[T]) GetAll() *result.Map[string, T] {
	return b.extensions
}

func (b *Bank[T]) Range(f func(id string, ext T) bool) {
	b.extensions.Range(f)
}

func (b *Bank[T]) OnExtensionAdded() <-chan struct{} {
	return b.extensionAddedCh
}

func (b *Bank[T]) OnExtensionRemoved() <-chan struct{} {
	return b.extensionRemovedCh
}
