package main

import (
	"net/http"
	"time"

	bypass "github.com/5rahim/hibike/pkg/util/bypass"
	"github.com/rs/zerolog"
	onlinestream "seanime/internal/extension/hibike/onlinestream"
)

type (
	Provider struct {
		url    string
		client *http.Client
		logger *zerolog.Logger
	}
)

func NewProvider(logger *zerolog.Logger) onlinestream.Provider {
	c := &http.Client{
		Timeout: 60 * time.Second,
	}
	c.Transport = bypass.AddCloudFlareByPass(c.Transport)
	return &Provider{
		url:    "https://example.com",
		client: c,
		logger: logger,
	}
}

func (p *Provider) Search(query string, dub bool) ([]*onlinestream.SearchResult, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Provider) FindEpisode(id string) ([]*onlinestream.EpisodeDetails, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Provider) FindEpisodeServer(episode *onlinestream.EpisodeDetails, server string) (*onlinestream.EpisodeServer, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Provider) GetEpisodeServers() []string {
	return []string{"server1", "server2"}
}
