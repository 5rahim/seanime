package qbittorrent_transfer

import (
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

func (c Client) GetTransferInfo() (*qbittorrent_model.TransferInfo, error) {
	var res qbittorrent_model.TransferInfo
	if err := qbittorrent_util.GetInto(c.Client, &res, c.BaseUrl+"/info", nil); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c Client) AlternativeSpeedLimitsEnabled() (bool, error) {
	var res int
	if err := qbittorrent_util.GetInto(c.Client, &res, c.BaseUrl+"/speedLimitsMode", nil); err != nil {
		return false, err
	}
	return res == 1, nil
}

func (c Client) ToggleAlternativeSpeedLimits() error {
	if err := qbittorrent_util.Post(c.Client, c.BaseUrl+"/toggleSpeedLimitsMode", nil); err != nil {
		return err
	}
	return nil
}

func (c Client) GetGlobalDownloadLimit() (int, error) {
	var res int
	if err := qbittorrent_util.GetInto(c.Client, &res, c.BaseUrl+"/downloadLimit", nil); err != nil {
		return 0, err
	}
	return res, nil
}

func (c Client) SetGlobalDownloadLimit(limit int) error {
	params := url.Values{}
	params.Add("limit", strconv.Itoa(limit))
	endpoint := c.BaseUrl + "/setDownloadLimit?" + params.Encode()
	if err := qbittorrent_util.Post(c.Client, endpoint, nil); err != nil {
		return err
	}
	return nil
}

func (c Client) GetGlobalUploadLimit() (int, error) {
	var res int
	if err := qbittorrent_util.GetInto(c.Client, &res, c.BaseUrl+"/uploadLimit", nil); err != nil {
		return 0, err
	}
	return res, nil
}

func (c Client) SetGlobalUploadLimit(limit int) error {
	params := url.Values{}
	params.Add("limit", strconv.Itoa(limit))
	endpoint := c.BaseUrl + "/setUploadLimit?" + params.Encode()
	if err := qbittorrent_util.Post(c.Client, endpoint, nil); err != nil {
		return err
	}
	return nil
}
