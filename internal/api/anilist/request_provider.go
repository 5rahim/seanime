package anilist

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"seanime/internal/constants"
	"strings"
	"sync"
)

const (
	OfficialRequestProviderName = "official"
	CustomRequestProviderName   = "custom"
)

type CustomClientConfig struct {
	Name          string            `json:"name" goja:"name"`
	Endpoint      string            `json:"endpoint" goja:"endpoint"`
	Token         string            `json:"token" goja:"token"`
	Headers       map[string]string `json:"headers" goja:"headers"`
	Authenticated bool              `json:"authenticated" goja:"authenticated"`
}

type RequestProvider interface {
	Name() string
	ApiUrl() string
	HttpClient() *http.Client
	PrepareRequest(ctx context.Context, req *http.Request, token string) error
	IsAuthenticated(token string) bool
}

type requestProviderRegistry struct {
	mu      sync.RWMutex
	current RequestProvider
}

var globalRequestProviders = &requestProviderRegistry{}

func init() {
	UseOfficialAPI()
}

func SetRequestProvider(provider RequestProvider) error {
	return globalRequestProviders.set(provider)
}

func UseOfficialAPI() {
	globalRequestProviders.set(anilistApiProvider{})
}

func UseCustomAPI(config CustomClientConfig) error {
	provider, err := newCustomRequestProvider(config)
	if err != nil {
		return err
	}

	return globalRequestProviders.set(provider)
}

func CurrentRequestProviderName() string {
	return globalRequestProviders.currentName()
}

func CurrentRequestProvider() RequestProvider {
	return globalRequestProviders.currentProvider()
}

func currentRequestProvider() RequestProvider {
	return globalRequestProviders.currentProvider()
}

func (r *requestProviderRegistry) set(provider RequestProvider) error {
	if provider == nil {
		return errors.New("anilist: request provider is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.current = provider
	return nil
}

func (r *requestProviderRegistry) currentName() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.current == nil {
		return OfficialRequestProviderName
	}

	return r.current.Name()
}

func (r *requestProviderRegistry) currentProvider() RequestProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.current != nil {
		return r.current
	}

	return anilistApiProvider{}
}

func requestProviderHTTPClient(provider RequestProvider) *http.Client {
	if provider == nil {
		return http.DefaultClient
	}

	client := provider.HttpClient()
	if client == nil {
		return http.DefaultClient
	}

	return client
}

type anilistApiProvider struct{}

func (anilistApiProvider) Name() string {
	return OfficialRequestProviderName
}

func (anilistApiProvider) ApiUrl() string {
	return constants.AnilistApiUrl
}

func (anilistApiProvider) HttpClient() *http.Client {
	return http.DefaultClient
}

func (anilistApiProvider) PrepareRequest(_ context.Context, req *http.Request, token string) error {
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return nil
}

func (anilistApiProvider) IsAuthenticated(token string) bool {
	return strings.TrimSpace(token) != ""
}

type customRequestProvider struct {
	name          string
	endpoint      string
	token         string
	headers       map[string]string
	authenticated bool
}

func newCustomRequestProvider(config CustomClientConfig) (*customRequestProvider, error) {
	endpoint := strings.TrimSpace(config.Endpoint)
	if endpoint == "" {
		return nil, errors.New("anilist: custom api endpoint is empty")
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("anilist: custom api endpoint must be absolute")
	}

	name := strings.TrimSpace(config.Name)
	if name == "" {
		name = CustomRequestProviderName
	}

	headers := make(map[string]string, len(config.Headers))
	for key, value := range config.Headers {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		headers[key] = strings.TrimSpace(value)
	}

	return &customRequestProvider{
		name:          name,
		endpoint:      endpoint,
		token:         strings.TrimSpace(config.Token),
		headers:       headers,
		authenticated: config.Authenticated,
	}, nil
}

func (p *customRequestProvider) Name() string {
	return p.name
}

func (p *customRequestProvider) ApiUrl() string {
	return p.endpoint
}

func (p *customRequestProvider) HttpClient() *http.Client {
	return http.DefaultClient
}

func (p *customRequestProvider) PrepareRequest(_ context.Context, req *http.Request, token string) error {
	authToken := p.token
	if authToken == "" {
		authToken = strings.TrimSpace(token)
	}
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	for key, value := range p.headers {
		req.Header.Set(key, value)
	}

	return nil
}

func (p *customRequestProvider) IsAuthenticated(token string) bool {
	if p.authenticated || p.token != "" || strings.TrimSpace(token) != "" {
		return true
	}

	return p.headers["Authorization"] != "" || p.headers["authorization"] != ""
}
