package metadata

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/util/filecache"
)

type (
	// Provider is the main interface for managing metadata.
	Provider struct {
		logger     *zerolog.Logger
		fileCacher *filecache.Cacher
	}

	NewProviderOptions struct {
		Logger     *zerolog.Logger
		FileCacher *filecache.Cacher
	}
)

// NewProvider creates a new metadata provider.
func NewProvider(options *NewProviderOptions) *Provider {
	return &Provider{
		logger:     options.Logger,
		fileCacher: options.FileCacher,
	}
}
