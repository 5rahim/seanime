package qbittorrent

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"seanime/internal/torrent_clients/qbittorrent/application"
	"seanime/internal/torrent_clients/qbittorrent/log"
	"seanime/internal/torrent_clients/qbittorrent/rss"
	"seanime/internal/torrent_clients/qbittorrent/search"
	"seanime/internal/torrent_clients/qbittorrent/sync"
	"seanime/internal/torrent_clients/qbittorrent/torrent"
	"seanime/internal/torrent_clients/qbittorrent/transfer"
	"strings"

	"github.com/rs/zerolog"
	"golang.org/x/net/publicsuffix"
)

type Client struct {
	baseURL          string
	logger           *zerolog.Logger
	client           *http.Client
	Username         string
	Password         string
	Port             int
	Host             string
	Path             string
	DisableBinaryUse bool
	Tags             string
	Category         string
	Application      qbittorrent_application.Client
	Log              qbittorrent_log.Client
	RSS              qbittorrent_rss.Client
	Search           qbittorrent_search.Client
	Sync             qbittorrent_sync.Client
	Torrent          qbittorrent_torrent.Client
	Transfer         qbittorrent_transfer.Client
}

type NewClientOptions struct {
	Logger           *zerolog.Logger
	Username         string
	Password         string
	Port             int
	Host             string
	Path             string
	DisableBinaryUse bool
	Tags             string
	Category         string
}

func NewClient(opts *NewClientOptions) *Client {

	host := opts.Host
	scheme := "http"
	if strings.HasPrefix(host, "https://") {
		scheme = "https"
		host = strings.TrimPrefix(host, "https://")
	} else if strings.HasPrefix(host, "http://") {
		host = strings.TrimPrefix(host, "http://")
	}
	opts.Host = host

	var baseURL string
	if opts.Port > 0 {
		baseURL = fmt.Sprintf("%s://%s:%d/api/v2", scheme, host, opts.Port)
	} else {
		baseURL = fmt.Sprintf("%s://%s/api/v2", scheme, host)
	}

	client := &http.Client{}
	return &Client{
		baseURL:          baseURL,
		logger:           opts.Logger,
		client:           client,
		Username:         opts.Username,
		Password:         opts.Password,
		Port:             opts.Port,
		Path:             opts.Path,
		DisableBinaryUse: opts.DisableBinaryUse,
		Host:             opts.Host,
		Tags:             opts.Tags,
		Category:         opts.Category,
		Application: qbittorrent_application.Client{
			BaseUrl: baseURL + "/app",
			Client:  client,
			Logger:  opts.Logger,
		},
		Log: qbittorrent_log.Client{
			BaseUrl: baseURL + "/log",
			Client:  client,
			Logger:  opts.Logger,
		},
		RSS: qbittorrent_rss.Client{
			BaseUrl: baseURL + "/rss",
			Client:  client,
			Logger:  opts.Logger,
		},
		Search: qbittorrent_search.Client{
			BaseUrl: baseURL + "/search",
			Client:  client,
			Logger:  opts.Logger,
		},
		Sync: qbittorrent_sync.Client{
			BaseUrl: baseURL + "/sync",
			Client:  client,
			Logger:  opts.Logger,
		},
		Torrent: qbittorrent_torrent.Client{
			BaseUrl: baseURL + "/torrents",
			Client:  client,
			Logger:  opts.Logger,
		},
		Transfer: qbittorrent_transfer.Client{
			BaseUrl: baseURL + "/transfer",
			Client:  client,
			Logger:  opts.Logger,
		},
	}
}

func (c *Client) Login() error {
	endpoint := c.baseURL + "/auth/login"
	data := url.Values{}
	data.Add("username", c.Username)
	data.Add("password", c.Password)
	request, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	request.Header.Add("content-type", "application/x-www-form-urlencoded")
	resp, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Err(err).Msg("failed to close login response body")
		}
	}()
	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid status %s", resp.Status)
	}
	if len(resp.Cookies()) < 1 {
		return fmt.Errorf("no cookies in login response")
	}
	apiURL, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return err
	}
	jar.SetCookies(apiURL, []*http.Cookie{resp.Cookies()[0]})
	c.client.Jar = jar
	return nil
}

func (c *Client) Logout() error {
	endpoint := c.baseURL + "/auth/logout"
	request, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(request)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid status %s", resp.Status)
	}
	return nil
}
