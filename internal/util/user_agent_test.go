package util

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGetOnlineUserAgents(t *testing.T) {
	userAgents, err := getOnlineUserAgents()
	if err != nil {
		t.Fatalf("Failed to get online user agents: %v", err)
	}
	t.Logf("Online user agents: %v", userAgents)
}

func TestTransformUserAgentJsonlToSliceFile(t *testing.T) {

	jsonlFilePath := filepath.Join("data", "user_agents.jsonl")

	jsonlFile, err := os.Open(jsonlFilePath)
	if err != nil {
		t.Fatalf("Failed to open JSONL file: %v", err)
	}
	defer jsonlFile.Close()

	sliceFilePath := filepath.Join("user_agent_list.go")
	sliceFile, err := os.Create(sliceFilePath)
	if err != nil {
		t.Fatalf("Failed to create slice file: %v", err)
	}
	defer sliceFile.Close()

	sliceFile.WriteString("package util\n\nvar UserAgentList = []string{\n")

	type UserAgent struct {
		UserAgent string `json:"useragent"`
	}

	scanner := bufio.NewScanner(jsonlFile)
	for scanner.Scan() {
		line := scanner.Text()
		var ua UserAgent
		if err := json.Unmarshal([]byte(line), &ua); err != nil {
			t.Fatalf("Failed to unmarshal line: %v", err)
		}
		sliceFile.WriteString(fmt.Sprintf("\t\"%s\",\n", ua.UserAgent))
	}
	sliceFile.WriteString("}\n")

	if err := scanner.Err(); err != nil {
		t.Fatalf("Failed to read JSONL file: %v", err)
	}

	t.Logf("User agent list generated successfully: %s", sliceFilePath)
}
