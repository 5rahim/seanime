package util

import (
	"bufio"
	"encoding/json"
	"net/http"
	"time"
)

var onlineUserAgentList []string

func GetOnlineUserAgents() ([]string, error) {

	link := "https://raw.githubusercontent.com/fake-useragent/fake-useragent/refs/heads/main/src/fake_useragent/data/browsers.jsonl"

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	response, err := client.Get(link)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	type UserAgent struct {
		UserAgent string `json:"useragent"`
	}

	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := scanner.Text()
		var ua UserAgent
		if err := json.Unmarshal([]byte(line), &ua); err != nil {
			return nil, err
		}
		onlineUserAgentList = append(onlineUserAgentList, ua.UserAgent)
	}

	return onlineUserAgentList, nil
}
