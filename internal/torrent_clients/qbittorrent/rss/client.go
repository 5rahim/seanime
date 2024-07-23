package qbittorrent_rss

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"net/http"
	"net/url"
	qbittorrent_model "seanime/internal/torrent_clients/qbittorrent/model"
	qbittorrent_util "seanime/internal/torrent_clients/qbittorrent/util"
)

type Client struct {
	BaseUrl string
	Client  *http.Client
	Logger  *zerolog.Logger
}

func (c Client) AddFolder(folder string) error {
	params := url.Values{}
	params.Add("path", folder)
	if err := qbittorrent_util.Post(c.Client, c.BaseUrl+"/addFolder?"+params.Encode(), nil); err != nil {
		return err
	}
	return nil
}

func (c Client) AddFeed(link string, folder string) error {
	params := url.Values{}
	params.Add("path", folder)
	if folder != "" {
		params.Add("path", folder)
	}
	if err := qbittorrent_util.Post(c.Client, c.BaseUrl+"/addFeed?"+params.Encode(), nil); err != nil {
		return err
	}
	return nil
}

func (c Client) RemoveItem(folder string) error {
	params := url.Values{}
	params.Add("path", folder)
	if err := qbittorrent_util.Post(c.Client, c.BaseUrl+"/removeItem?"+params.Encode(), nil); err != nil {
		return err
	}
	return nil
}

func (c Client) MoveItem(currentFolder, destinationFolder string) error {
	params := url.Values{}
	params.Add("itemPath", currentFolder)
	params.Add("destPath", destinationFolder)
	if err := qbittorrent_util.Post(c.Client, c.BaseUrl+"/moveItem?"+params.Encode(), nil); err != nil {
		return err
	}
	return nil
}

func (c Client) AddRule(name string, def qbittorrent_model.RuleDefinition) error {
	params := url.Values{}
	b, err := json.Marshal(def)
	if err != nil {
		return err
	}
	params.Add("ruleName", name)
	params.Add("ruleDef", string(b))
	if err := qbittorrent_util.Post(c.Client, c.BaseUrl+"/setRule?"+params.Encode(), nil); err != nil {
		return err
	}
	return nil
}

func (c Client) RenameRule(old, new string) error {
	params := url.Values{}
	params.Add("ruleName", old)
	params.Add("newRuleName", new)
	if err := qbittorrent_util.Post(c.Client, c.BaseUrl+"/renameRule?"+params.Encode(), nil); err != nil {
		return err
	}
	return nil
}

func (c Client) RemoveRule(name string) error {
	params := url.Values{}
	params.Add("ruleName", name)
	if err := qbittorrent_util.Post(c.Client, c.BaseUrl+"/removeRule?"+params.Encode(), nil); err != nil {
		return err
	}
	return nil
}

func (c Client) GetRules() (map[string]qbittorrent_model.RuleDefinition, error) {
	var res map[string]qbittorrent_model.RuleDefinition
	if err := qbittorrent_util.GetInto(c.Client, &res, c.BaseUrl+"/rules", nil); err != nil {
		return nil, err
	}
	return res, nil
}
