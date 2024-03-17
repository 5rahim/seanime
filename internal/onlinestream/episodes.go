package onlinestream

import (
	"github.com/seanime-app/seanime/internal/onlinestream/providers"
)

const (
	ProviderGogoanime Provider = "gogoanime"
	ProviderZoro      Provider = "zoro"
)

type (
	Provider string

	ProviderEpisodes struct {
		Provider Provider
		Episodes []*Episode
	}

	Episode struct {
		Server onlinestream_providers.Server
		*onlinestream_providers.ProviderEpisode
		Sources []*onlinestream_providers.ProviderEpisodeSource
	}
)
