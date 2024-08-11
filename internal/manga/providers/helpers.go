package manga_providers

import "strings"

// GetNormalizedChapter returns a normalized chapter string.
// e.g. "0001" -> "1"
func GetNormalizedChapter(chapter string) string {
	// Trim padding zeros
	unpaddedChStr := strings.TrimLeft(chapter, "0")
	if unpaddedChStr == "" {
		unpaddedChStr = "0"
	}
	return unpaddedChStr
}
