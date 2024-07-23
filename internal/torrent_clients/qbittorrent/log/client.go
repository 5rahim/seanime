package qbittorrent_log

import (
	"github.com/google/go-querystring/query"
	"github.com/rs/zerolog"
	"net/http"
	"net/url"
	"seanime/internal/torrent_clients/qbittorrent/model"
	"seanime/internal/torrent_clients/qbittorrent/util"
	"strconv"
)

type Client struct {
	BaseUrl string
	Client  *http.Client
	Logger  *zerolog.Logger
}

func (c Client) GetLog(options *qbittorrent_model.GetLogOptions) ([]*qbittorrent_model.LogEntry, error) {
	endpoint := c.BaseUrl + "/main"
	if options != nil {
		params, err := query.Values(options)
		if err != nil {
			return nil, err
		}
		endpoint += "?" + params.Encode()
	}
	var res []*qbittorrent_model.LogEntry
	if err := qbittorrent_util.GetInto(c.Client, &res, endpoint, nil); err != nil {
		return nil, err
	}
	return res, nil
}

func (c Client) GetPeerLog(lastKnownID int) ([]*qbittorrent_model.PeerLogEntry, error) {
	params := url.Values{}
	params.Add("last_known_id", strconv.Itoa(lastKnownID))
	endpoint := c.BaseUrl + "/peers?" + params.Encode()
	var res []*qbittorrent_model.PeerLogEntry
	if err := qbittorrent_util.GetInto(c.Client, &res, endpoint, nil); err != nil {
		return nil, err
	}
	return res, nil
}
