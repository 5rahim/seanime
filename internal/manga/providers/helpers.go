package manga_providers

import "strings"

// GetNormalizedChapter returns a normalized chapter string.
func (ch *ChapterDetails) GetNormalizedChapter() string {
	chapter := ch.Chapter
	// Trim padding zeros
	unpaddedChStr := strings.TrimLeft(chapter, "0")
	if unpaddedChStr == "" {
		unpaddedChStr = "0"
	}
	return unpaddedChStr
}
