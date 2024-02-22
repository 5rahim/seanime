package animetosho

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/mmcdole/gofeed/json"
)

type FeedTranslator struct {
}

func (t *FeedTranslator) Translate(feed interface{}) (*gofeed.Feed, error) {
	json, found := feed.(*json.Feed)
	if !found {
		return nil, fmt.Errorf("feed did not match expected type of *json.Feed")
	}

	result := &gofeed.Feed{}
	result.Title = json.Title

	return result, nil
}

func NewFeedTranslator() *FeedTranslator {
	t := &FeedTranslator{}
	return t
}
