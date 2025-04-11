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
		extensionAddedCh:   make(chan struct{}, 100),
		extensionRemovedCh: make(chan struct{}, 100),
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

	b.extensionAddedCh <- struct{}{}
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

	b.extensionRemovedCh <- struct{}{}
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