package goja_runtime

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

// Manager manages a shared pool of Goja runtimes for all extensions.
type Manager struct {
	pool   *Pool
	logger *zerolog.Logger
	size   int32
}

func NewManager(logger *zerolog.Logger, size int32) *Manager {
	return &Manager{
		logger: logger,
		size:   size,
	}
}

// GetOrCreatePool returns the shared pool.
func (m *Manager) GetOrCreatePool(initFn func() *goja.Runtime) (*Pool, error) {
	if m.pool == nil {
		m.pool = newPool(m.size, initFn, m.logger)
	}
	return m.pool, nil
}

func (m *Manager) Run(ctx context.Context, fn func(*goja.Runtime) error) error {
	runtime, err := m.pool.Get(ctx)
	if err != nil {
		return err
	}
	return fn(runtime)
}

func (m *Manager) GetLogger() *zerolog.Logger {
	return m.logger
}

func (m *Manager) PrintMetrics() {
	if m.pool == nil {
		return
	}
	stats := m.pool.Stats()
	m.logger.Trace().
		Int64("created", stats["created"]).
		Int64("reused", stats["reused"]).
		Int64("timeouts", stats["timeouts"]).
		Msg("goja runtime: VM Pool Metrics")
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
	created  atomic.Int64
	reused   atomic.Int64
	timeouts atomic.Int64
}

// newPool creates a new Pool using sync.Pool, pre-warming it with size items.
func newPool(size int32, initFn func() *goja.Runtime, logger *zerolog.Logger) *Pool {
	p := &Pool{
		factory: initFn,
		logger:  logger,
		size:    size,
	}

	p.sp.New = func() interface{} {
		runtime := initFn()
		p.metrics.created.Add(1)
		return runtime
	}

	for i := int32(0); i < size; i++ {
		r := initFn()
		p.sp.Put(r)
		p.metrics.created.Add(1)
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
		"created":  p.metrics.created.Load(),
		"reused":   p.metrics.reused.Load(),
		"timeouts": p.metrics.timeouts.Load(),
	}
}
