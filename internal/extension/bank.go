package extension

import (
	"seanime/internal/util/result"
	"sync"
	"sync/atomic"
)

type UnifiedBank struct {
	extensions  *result.Map[string, BaseExtension]
	subscribers *result.Map[string, *BankSubscriber]
	mu          sync.RWMutex
}

type BankSubscriber struct {
	id                     string
	extensionAddedCh       chan struct{}
	extensionRemovedCh     chan struct{}
	customSourcesChangedCh chan struct{}
	mu                     sync.Mutex
	closeOnce              sync.Once
	closed                 atomic.Bool
}

func NewUnifiedBank() *UnifiedBank {
	return &UnifiedBank{
		extensions:  result.NewMap[string, BaseExtension](),
		subscribers: result.NewMap[string, *BankSubscriber](),
		mu:          sync.RWMutex{},
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
	b.extensions = result.NewMap[string, BaseExtension]()
}

func (b *UnifiedBank) Subscribe(id string) *BankSubscriber {
	sub := &BankSubscriber{
		id:                     id,
		extensionAddedCh:       make(chan struct{}, 100),
		extensionRemovedCh:     make(chan struct{}, 100),
		customSourcesChangedCh: make(chan struct{}, 100),
	}
	b.subscribers.Set(id, sub)
	return sub
}

func (b *UnifiedBank) Unsubscribe(id string) {
	sub, ok := b.subscribers.Get(id)
	if !ok {
		return
	}
	sub.closeOnce.Do(func() {
		sub.closed.Store(true)
		b.mu.Lock()
		close(sub.extensionAddedCh)
		close(sub.extensionRemovedCh)
		close(sub.customSourcesChangedCh)
		b.mu.Unlock()
	})
	b.subscribers.Delete(id)
	return
}

func (b *UnifiedBank) RemoveExternalExtensions() {
	b.mu.Lock()
	defer b.mu.Unlock()

	var ids []string
	b.extensions.Range(func(id string, ext BaseExtension) bool {
		if ext.GetManifestURI() != "builtin" {
			ids = append(ids, id)
		}
		return true
	})

	for _, id := range ids {
		b.Delete(id)
	}
}

func (b *UnifiedBank) Set(id string, ext BaseExtension) {
	// Add the extension to the map
	b.extensions.Set(id, ext)

	// Notify listeners that an extension has been added

	go func() {
		b.subscribers.Range(func(id string, sub *BankSubscriber) bool {
			if sub.closed.Load() {
				return true
			}
			sub.mu.Lock()
			sub.extensionAddedCh <- struct{}{}
			if ext.GetType() == TypeCustomSource {
				sub.customSourcesChangedCh <- struct{}{}
			}
			sub.mu.Unlock()
			return true
		})
	}()
}

func (b *UnifiedBank) Get(id string) (BaseExtension, bool) {
	//b.mu.RLock()
	//defer b.mu.RUnlock()
	return b.extensions.Get(id)
}

func (b *UnifiedBank) Delete(id string) {
	ext, ok := b.extensions.Get(id)
	if !ok {
		return
	}
	// Delete the extension from the map
	b.extensions.Delete(id)

	// Notify listeners that an extension has been removed
	go func() {
		b.subscribers.Range(func(id string, sub *BankSubscriber) bool {
			if sub.closed.Load() {
				return true
			}
			sub.mu.Lock()
			sub.extensionRemovedCh <- struct{}{}
			if ext.GetType() == TypeCustomSource {
				sub.customSourcesChangedCh <- struct{}{}
			}
			sub.mu.Unlock()
			return true
		})
	}()
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

func (b *BankSubscriber) OnExtensionAdded() <-chan struct{} {
	return b.extensionAddedCh
}

func (b *BankSubscriber) OnExtensionRemoved() <-chan struct{} {
	return b.extensionRemovedCh
}

func (b *BankSubscriber) OnCustomSourcesChanged() <-chan struct{} {
	return b.customSourcesChangedCh
}

func (b *BankSubscriber) ID() string {
	return b.id
}

func GetExtension[T BaseExtension](bank *UnifiedBank, id string) (ret T, ok bool) {
	ext, ok := bank.extensions.Get(id)
	if !ok {
		return
	}

	ret, ok = ext.(T)
	return
}

func RangeExtensions[T BaseExtension](bank *UnifiedBank, f func(id string, ext T) bool) {
	bank.extensions.Range(func(id string, ext BaseExtension) bool {
		if typedExt, ok := ext.(T); ok {
			return f(id, typedExt)
		}
		return true
	})
}
