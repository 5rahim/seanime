package util

import (
	"bufio"
	"encoding/json"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var (
	userAgentList []string
	uaMu          sync.RWMutex
)

func getOnlineUserAgents() ([]string, error) {
	link := "https://raw.githubusercontent.com/fake-useragent/fake-useragent/refs/heads/main/src/fake_useragent/data/browsers.jsonl"

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	response, err := client.Get(link)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var agents []string
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
		agents = append(agents, ua.UserAgent)
	}

	return agents, nil
}

func GetRandomUserAgent() string {
	uaMu.RLock()
	defer uaMu.RUnlock()

	if len(userAgentList) > 0 {
		return userAgentList[rand.Intn(len(userAgentList))]
	}
	return UserAgentList[rand.Intn(len(UserAgentList))]
}
