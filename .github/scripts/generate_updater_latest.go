package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	DownloadUrl = "https://github.com/5rahim/seanime/releases/latest/download/"
)

func main() {
	// Retrieve version from environment variable
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "1.0.0" // Default to '1.0.0' if not set
	}

	// Define the asset filenames
	assets := map[string]struct {
		Asset  string
		AppZip string
		Sig    string
	}{
		"MacOS_arm64": {
			Asset: fmt.Sprintf("seanime-desktop-%s_MacOS_arm64.app.tar.gz", version),
			Sig:   fmt.Sprintf("seanime-desktop-%s_MacOS_arm64.app.tar.gz.sig", version),
		},
		"MacOS_x86_64": {
			Asset: fmt.Sprintf("seanime-desktop-%s_MacOS_x86_64.app.tar.gz", version),
			Sig:   fmt.Sprintf("seanime-desktop-%s_MacOS_x86_64.app.tar.gz.sig", version),
		},
		"Linux_x86_64": {
			Asset: fmt.Sprintf("seanime-desktop-%s_Linux_x86_64.AppImage", version),
			Sig:   fmt.Sprintf("seanime-desktop-%s_Linux_x86_64.AppImage.sig", version),
		},
		"Windows_x86_64": {
			AppZip: fmt.Sprintf("seanime-desktop-%s_Windows_x86_64.exe", version),
			Sig:    fmt.Sprintf("seanime-desktop-%s_Windows_x86_64.sig", version),
		},
	}

	// Function to generate URL based on asset names
	generateURL := func(filename string) string {
		return fmt.Sprintf("%s%s", DownloadUrl, filename)
	}

	// Prepare the JSON structure
	latestJSON := map[string]interface{}{
		"version":  version,
		"pub_date": time.Now().Format(time.RFC3339), // Change to the actual publish date
		"platforms": map[string]map[string]string{
			"linux-x86_64": {
				"url":       generateURL(assets["Linux_x86_64"].Asset),
				"signature": getContent(assets["Linux_x86_64"].Sig),
			},
			"windows-x86_64": {
				"url":       generateURL(assets["Windows_x86_64"].AppZip),
				"signature": getContent(assets["Windows_x86_64"].Sig),
			},
			"darwin-x86_64": {
				"url":       generateURL(assets["MacOS_x86_64"].Asset),
				"signature": getContent(assets["MacOS_x86_64"].Sig),
			},
			"darwin-aarch64": {
				"url":       generateURL(assets["MacOS_arm64"].Asset),
				"signature": getContent(assets["MacOS_arm64"].Sig),
			},
		},
	}

	// Remove non-existent assets
	for platform, asset := range latestJSON["platforms"].(map[string]map[string]string) {
		if asset["signature"] == "" {
			delete(latestJSON["platforms"].(map[string]map[string]string), platform)
		}
	}

	// Write to latest.json
	outputPath := filepath.Join(".", "latest.json")
	file, err := os.Create(outputPath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(latestJSON); err != nil {
		fmt.Println("Error writing JSON to file:", err)
		return
	}

	fmt.Printf("Generated %s successfully.\n", outputPath)
}

func getContent(filename string) string {
	fileContent, err := os.ReadFile(filepath.Join(".", filename))
	if err != nil {
		return ""
	}
	return string(fileContent)
}
