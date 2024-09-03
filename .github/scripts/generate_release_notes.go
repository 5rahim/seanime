package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	const inFile = "CHANGELOG.md"
	const outFile = "whats-new.md"

	// Get the path to the changelog
	changelogPath := filepath.Join(".", inFile)

	// Read the changelog content
	content, err := os.ReadFile(changelogPath)
	if err != nil {
		fmt.Println("Error reading changelog:", err)
		return
	}

	// Convert the content to a string
	changelog := string(content)

	// Extract everything between the first and second "## " headers
	sections := strings.Split(changelog, "## ")
	if len(sections) < 2 {
		fmt.Println("Not enough headers found in the changelog.")
		return
	}

	// We only care about the first section
	changelog = sections[1]

	// Remove everything after the next header (if any)
	changelog = strings.Split(changelog, "## ")[0]

	// Remove the first line (which is the title of the first section)
	lines := strings.Split(changelog, "\n")
	if len(lines) > 1 {
		changelog = strings.Join(lines[1:], "\n")
	}

	// Trim newlines
	changelog = strings.TrimSpace(changelog)

	// Write the extracted content to the output file
	outPath := filepath.Join(".", outFile)
	if err := os.WriteFile(outPath, []byte(changelog), 0644); err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Printf("Changelog content written to %s\n", outPath)
}
