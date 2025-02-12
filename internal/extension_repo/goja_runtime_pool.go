package extension_repo

import (
	"seanime/internal/goja/goja_bindings"
	"sync"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	"github.com/rs/zerolog"
)

type RuntimePool struct {
	pool    sync.Pool
	logger  *zerolog.Logger
	maxSize int
}

func NewRuntimePool(maxSize int, logger *zerolog.Logger) *RuntimePool {
	return &RuntimePool{
		pool: sync.Pool{
			New: func() interface{} {
				vm := goja.New()
				vm.SetParserOptions(parser.WithDisableSourceMaps)

				// Initialize bindings
				if err := goja_bindings.BindFetch(vm); err != nil {
					logger.Error().Err(err).Msg("Failed to bind fetch")
				}
				if err := goja_bindings.BindConsole(vm, logger); err != nil {
					logger.Error().Err(err).Msg("Failed to bind console")
				}
				if err := goja_bindings.BindFormData(vm); err != nil {
					logger.Error().Err(err).Msg("Failed to bind form data")
				}
				if err := goja_bindings.BindDocument(vm); err != nil {
					logger.Error().Err(err).Msg("Failed to bind document")
				}
				if err := goja_bindings.BindCrypto(vm); err != nil {
					logger.Error().Err(err).Msg("Failed to bind crypto")
				}
				if err := goja_bindings.BindTorrentUtils(vm); err != nil {
					logger.Error().Err(err).Msg("Failed to bind torrent utils")
				}

				return vm
			},
		},
		maxSize: maxSize,
		logger:  logger,
	}
}

func (p *RuntimePool) Get() *goja.Runtime {
	return p.pool.Get().(*goja.Runtime)
}

func (p *RuntimePool) Put(vm *goja.Runtime) {
	if vm != nil {
		vm.ClearInterrupt()
		p.pool.Put(vm)
	}
}
