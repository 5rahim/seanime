package goja_runtime

import (
	"context"
	"fmt"
	"runtime"
	"seanime/internal/util/result"
	"sync"
	"sync/atomic"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

// Manager manages a shared pool of Goja runtimes for all extensions.
type Manager struct {
	pluginPools *result.Map[string, *Pool]
	basePool    *Pool
	logger      *zerolog.Logger
}

type Pool struct {
	sp      sync.Pool
	factory func() *goja.Runtime
	logger  *zerolog.Logger
	size    int32
	metrics metrics
}

// metrics holds counters for pool stats.
type metrics struct {
	prewarmed   atomic.Int64
	created     atomic.Int64
	reused      atomic.Int64
	timeouts    atomic.Int64
	invocations atomic.Int64
}

func NewManager(logger *zerolog.Logger) *Manager {
	return &Manager{
		logger: logger,
	}
}

// GetOrCreatePrivatePool returns the pool for the given extension.
func (m *Manager) GetOrCreatePrivatePool(extID string, initFn func() *goja.Runtime) (*Pool, error) {
	if m.pluginPools == nil {
		m.pluginPools = result.NewResultMap[string, *Pool]()
	}

	pool, ok := m.pluginPools.Get(extID)
	if !ok {
		pool = newPool(5, initFn, m.logger)
		m.pluginPools.Set(extID, pool)
	}
	return pool, nil
}

func (m *Manager) DeletePluginPool(extID string) {
	m.logger.Trace().Msgf("plugin: Deleting pool for extension %s", extID)
	if m.pluginPools == nil {
		return
	}

	// Get the pool first to interrupt all runtimes
	if pool, ok := m.pluginPools.Get(extID); ok {
		// Drain the pool and interrupt all runtimes
		m.logger.Debug().Msgf("plugin: Interrupting all runtimes in pool for extension %s", extID)

		interruptedCount := 0
		for {
			// Get a runtime without using a context to avoid blocking
			runtimeV := pool.sp.Get()
			if runtimeV == nil {
				break // No more runtimes in the pool or error occurred
			}

			runtime, ok := runtimeV.(*goja.Runtime)
			if !ok {
				break
			}

			// Interrupt the runtime
			runtime.ClearInterrupt()
			interruptedCount++
		}

		m.logger.Debug().Msgf("plugin: Interrupted %d runtimes in pool for extension %s", interruptedCount, extID)
	}

	// Delete the pool
	m.pluginPools.Delete(extID)
	runtime.GC()
}

// GetOrCreateSharedPool returns the shared pool.
func (m *Manager) GetOrCreateSharedPool(initFn func() *goja.Runtime) (*Pool, error) {
	if m.basePool == nil {
		m.basePool = newPool(15, initFn, m.logger)
	}
	return m.basePool, nil
}

func (m *Manager) Run(ctx context.Context, extID string, fn func(*goja.Runtime) error) error {
	pool, ok := m.pluginPools.Get(extID)
	if !ok {
		return fmt.Errorf("plugin pool not found for extension ID: %s", extID)
	}
	runtime, err := pool.Get(ctx)
	pool.metrics.invocations.Add(1)
	if err != nil {
		return err
	}
	defer pool.Put(runtime)
	return fn(runtime)
}

func (m *Manager) RunShared(ctx context.Context, fn func(*goja.Runtime) error) error {
	runtime, err := m.basePool.Get(ctx)
	if err != nil {
		return err
	}
	defer m.basePool.Put(runtime)
	return fn(runtime)
}

func (m *Manager) GetLogger() *zerolog.Logger {
	return m.logger
}

func (m *Manager) PrintPluginPoolMetrics(extID string) {
	if m.pluginPools == nil {
		return
	}
	pool, ok := m.pluginPools.Get(extID)
	if !ok {
		return
	}
	stats := pool.Stats()
	m.logger.Trace().
		Int64("prewarmed", stats["prewarmed"]).
		Int64("created", stats["created"]).
		Int64("reused", stats["reused"]).
		Int64("timeouts", stats["timeouts"]).
		Int64("invocations", stats["invocations"]).
		Msg("goja runtime: VM Pool Metrics")
}

func (m *Manager) PrintBasePoolMetrics() {
	if m.basePool == nil {
		return
	}
	stats := m.basePool.Stats()
	m.logger.Trace().
		Int64("prewarmed", stats["prewarmed"]).
		Int64("created", stats["created"]).
		Int64("reused", stats["reused"]).
		Int64("invocations", stats["invocations"]).
		Int64("timeouts", stats["timeouts"]).
		Msg("goja runtime: Base VM Pool Metrics")
}

// newPool creates a new Pool using sync.Pool, pre-warming it with size items.
func newPool(size int32, initFn func() *goja.Runtime, logger *zerolog.Logger) *Pool {
	p := &Pool{
		factory: initFn,
		logger:  logger,
		size:    size,
	}

	// p.sp.New = func() interface{} {
	// 	runtime := initFn()
	// 	p.metrics.created.Add(1)
	// 	return runtime
	// }

	p.sp.New = func() any {
		return nil
	}

	// Pre-warm the pool
	logger.Trace().Int32("size", size).Msg("goja runtime: Pre-warming pool")
	for i := int32(0); i < size; i++ {
		r := initFn()
		p.sp.Put(r)
		p.metrics.prewarmed.Add(1)
	}

	return p
}

// Get retrieves a runtime from the pool or creates a new one. It respects the context for cancellation.
func (p *Pool) Get(ctx context.Context) (*goja.Runtime, error) {
	v := p.sp.Get()
	if v == nil {
		// If sync.Pool.New returned nil or context canceled, try factory manually.
		select {
		case <-ctx.Done():
			p.metrics.timeouts.Add(1)
			return nil, ctx.Err()
		default:
		}
		runtime := p.factory()
		p.metrics.created.Add(1)
		return runtime, nil
	}
	p.metrics.reused.Add(1)
	return v.(*goja.Runtime), nil
}

// Put returns a runtime to the pool after clearing its interrupt.
func (p *Pool) Put(runtime *goja.Runtime) {
	if runtime == nil {
		return
	}
	runtime.ClearInterrupt()
	p.sp.Put(runtime)
}

// Stats returns pool metrics as a map.
func (p *Pool) Stats() map[string]int64 {
	return map[string]int64{
		"prewarmed":   p.metrics.prewarmed.Load(),
		"invocations": p.metrics.invocations.Load(),
		"created":     p.metrics.created.Load(),
		"reused":      p.metrics.reused.Load(),
		"timeouts":    p.metrics.timeouts.Load(),
	}
}
