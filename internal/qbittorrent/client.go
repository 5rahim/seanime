package qbittorrent

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime-server/internal/qbittorrent/application"
	"github.com/seanime-app/seanime-server/internal/qbittorrent/log"
	"github.com/seanime-app/seanime-server/internal/qbittorrent/rss"
	"github.com/seanime-app/seanime-server/internal/qbittorrent/search"
	"github.com/seanime-app/seanime-server/internal/qbittorrent/sync"
	"github.com/seanime-app/seanime-server/internal/qbittorrent/torrent"
	"github.com/seanime-app/seanime-server/internal/qbittorrent/transfer"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
)

type Client struct {
	baseURL     string
	logger      *zerolog.Logger
	client      *http.Client
	Application qbittorrent_application.Client
	Log         qbittorrent_log.Client
	RSS         qbittorrent_rss.Client
	Search      qbittorrent_search.Client
	Sync        qbittorrent_sync.Client
	Torrent     qbittorrent_torrent.Client
	Transfer    qbittorrent_transfer.Client
}

func NewClient(baseURL string, logger *zerolog.Logger) *Client {
	baseURL = baseURL + "/api/v2"
	client := &http.Client{}
	return &Client{
		baseURL: baseURL,
		logger:  logger,
		client:  client,
		Application: qbittorrent_application.Client{
			BaseUrl: baseURL + "/app",
			Client:  client,
			Logger:  logger,
		},
		Log: qbittorrent_log.Client{
			BaseUrl: baseURL + "/log",
			Client:  client,
			Logger:  logger,
		},
		RSS: qbittorrent_rss.Client{
			BaseUrl: baseURL + "/rss",
			Client:  client,
			Logger:  logger,
		},
		Search: qbittorrent_search.Client{
			BaseUrl: baseURL + "/search",
			Client:  client,
			Logger:  logger,
		},
		Sync: qbittorrent_sync.Client{
			BaseUrl: baseURL + "/sync",
			Client:  client,
			Logger:  logger,
		},
		Torrent: qbittorrent_torrent.Client{
			BaseUrl: baseURL + "/torrents",
			Client:  client,
			Logger:  logger,
		},
		Transfer: qbittorrent_transfer.Client{
			BaseUrl: baseURL + "/transfer",
			Client:  client,
			Logger:  logger,
		},
	}
}

func (c *Client) Login(username, password string) error {
	endpoint := c.baseURL + "/auth/login"
	data := url.Values{}
	data.Add("username", username)
	data.Add("password", password)
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

func (c Client) Logout() error {
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
