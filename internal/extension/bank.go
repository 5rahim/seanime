package extension

import (
	"seanime/internal/util/result"
	"sync"
)

type UnifiedBank struct {
	extensions         *result.Map[string, BaseExtension]
	extensionAddedCh   chan struct{}
	extensionRemovedCh chan struct{}
	mu                 sync.RWMutex
}

func NewUnifiedBank() *UnifiedBank {
	return &UnifiedBank{
		extensions:         result.NewResultMap[string, BaseExtension](),
		extensionAddedCh:   make(chan struct{}),
		extensionRemovedCh: make(chan struct{}),
		mu:                 sync.RWMutex{},
	}
}

func (b *UnifiedBank) Lock() {
	b.mu.Lock()
}

func (b *UnifiedBank) Unlock() {
	b.mu.Unlock()
}

func (b *UnifiedBank) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.extensions = result.NewResultMap[string, BaseExtension]()
	close(b.extensionAddedCh)
	close(b.extensionRemovedCh)
	b.extensionAddedCh = make(chan struct{})
	b.extensionRemovedCh = make(chan struct{})
}

func (b *UnifiedBank) RemoveExternalExtensions() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.extensions.Range(func(id string, ext BaseExtension) bool {
		if ext.GetManifestURI() != "builtin" {
			b.extensions.Delete(id)
		}
		return true
	})
}

func (b *UnifiedBank) Set(id string, ext BaseExtension) {
	// Add the extension to the map
	b.extensions.Set(id, ext)

	// Notify listeners that an extension has been added
	b.mu.Lock()
	defer b.mu.Unlock()

	close(b.extensionAddedCh)
	b.extensionAddedCh = make(chan struct{})
}

func (b *UnifiedBank) Get(id string) (BaseExtension, bool) {
	//b.mu.RLock()
	//defer b.mu.RUnlock()
	return b.extensions.Get(id)
}

func (b *UnifiedBank) Delete(id string) {
	// Delete the extension from the map
	b.extensions.Delete(id)

	// Notify listeners that an extension has been removed
	b.mu.Lock()
	defer b.mu.Unlock()

	close(b.extensionRemovedCh)
	b.extensionRemovedCh = make(chan struct{})
}

func (b *UnifiedBank) GetExtensionMap() *result.Map[string, BaseExtension] {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.extensions
}

func (b *UnifiedBank) Range(f func(id string, ext BaseExtension) bool) {
	// No need to lock
	b.extensions.Range(f)
}

func (b *UnifiedBank) OnExtensionAdded() <-chan struct{} {
	return b.extensionAddedCh
}

func (b *UnifiedBank) OnExtensionRemoved() <-chan struct{} {
	return b.extensionRemovedCh
}

func GetExtension[T BaseExtension](bank *UnifiedBank, id string) (ret T, ok bool) {
	// No need to lock
	ext, ok := bank.extensions.Get(id)
	if !ok {
		return
	}

	ret, ok = ext.(T)
	return
}

func RangeExtensions[T BaseExtension](bank *UnifiedBank, f func(id string, ext T) bool) {
	// No need to lock
	bank.extensions.Range(func(id string, ext BaseExtension) bool {
		if typedExt, ok := ext.(T); ok {
			return f(id, typedExt)
		}
		return true
	})
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Bank[T BaseExtension] struct {
	extensions         *result.Map[string, T]
	extensionAddedCh   chan struct{}
	extensionRemovedCh chan struct{}
	mu                 sync.RWMutex
}

func NewBank[T BaseExtension]() *Bank[T] {
	return &Bank[T]{
		extensions:         result.NewResultMap[string, T](),
		extensionAddedCh:   make(chan struct{}),
		extensionRemovedCh: make(chan struct{}),
		mu:                 sync.RWMutex{},
	}
}

func (b *Bank[T]) Lock() {
	b.mu.Lock()
}

func (b *Bank[T]) Unlock() {
	b.mu.Unlock()
}

func (b *Bank[T]) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.extensions = result.NewResultMap[string, T]()
	close(b.extensionAddedCh)
	close(b.extensionRemovedCh)
	b.extensionAddedCh = make(chan struct{})
	b.extensionRemovedCh = make(chan struct{})
}

func (b *Bank[T]) RemoveExternalExtensions() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.extensions.Range(func(id string, ext T) bool {
		if ext.GetManifestURI() != "builtin" {
			b.extensions.Delete(id)
		}
		return true
	})
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
	b.mu.RLock()
	defer b.mu.RUnlock()

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

func (b *Bank[T]) GetExtensionMap() *result.Map[string, T] {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.extensions
}

func (b *Bank[T]) Range(f func(id string, ext T) bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	b.extensions.Range(f)
}

func (b *Bank[T]) OnExtensionAdded() <-chan struct{} {
	return b.extensionAddedCh
}

func (b *Bank[T]) OnExtensionRemoved() <-chan struct{} {
	return b.extensionRemovedCh
}
