package onlinestream

import (
	"github.com/seanime-app/seanime/internal/onlinestream/providers"
)

type (
	ProviderEpisodes struct {
		Provider string
		Episodes []*Episode
	}

	Episode struct {
		Server onlinestream_providers.Server
		*onlinestream_providers.ProviderEpisode
		Sources []*onlinestream_providers.ProviderEpisodeSource
	}
)
