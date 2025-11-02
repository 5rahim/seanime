package transmission

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/hekmon/transmissionrpc/v3"
	"github.com/rs/zerolog"
)

type (
	Transmission struct {
		Client *transmissionrpc.Client
		Path   string
		Logger *zerolog.Logger
	}

	NewTransmissionOptions struct {
		Path     string
		Logger   *zerolog.Logger
		Username string
		Password string
		Host     string // Default: 127.0.0.1
		Port     int
	}
)

func New(options *NewTransmissionOptions) (*Transmission, error) {
	// Set default host
	if options.Host == "" {
		options.Host = "127.0.0.1"
	}

	baseUrl := fmt.Sprintf("http://%s:%s@%s:%d/transmission/rpc",
		options.Username,
		url.QueryEscape(options.Password),
		options.Host,
		options.Port,
	)

	if strings.HasPrefix(options.Host, "https://") {
		options.Host = strings.TrimPrefix(options.Host, "https://")
		baseUrl = fmt.Sprintf("https://%s:%s@%s:%d/transmission/rpc",
			options.Username,
			url.QueryEscape(options.Password),
			options.Host,
			options.Port,
		)
	}

	baseUrl = strings.Replace(baseUrl, ":0/transmission/rpc", "/transmission/rpc", 1)

	_url, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	client, _ := transmissionrpc.New(_url, nil)
	return &Transmission{
		Client: client,
		Path:   options.Path,
		Logger: options.Logger,
	}, nil
}
