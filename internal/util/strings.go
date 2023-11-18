package util

import (
	"fmt"
	"time"
	"unicode"
)

func isMostlyLatinString(str string) bool {
	if len(str) <= 0 {
		return false
	}
	latinLength := 0
	nonLatinLength := 0
	for _, r := range str {
		if isLatinRune(r) {
			latinLength++
		} else {
			nonLatinLength++
		}
	}
	return latinLength > nonLatinLength
}

func isLatinRune(r rune) bool {
	return unicode.In(r, unicode.Latin)
}

// ToHumanReadableSpeed converts an integer representing bytes per second to a human-readable format
func ToHumanReadableSpeed(bytesPerSecond int) string {
	if bytesPerSecond <= 0 {
		return `0 KB/s`
	}

	const unit = 1024
	if bytesPerSecond < unit {
		return fmt.Sprintf("%d B/s", bytesPerSecond)
	}
	div, exp := int64(unit), 0
	for n := int64(bytesPerSecond) / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB/s", float64(bytesPerSecond)/float64(div), "KMGTPE"[exp])
}

// ToHumanReadableSize converts total size in bytes to a human-readable string
func ToHumanReadableSize(totalSizeBytes int) string { // FIXME incorrect
	const unit = 1024

	if totalSizeBytes < unit {
		return fmt.Sprintf("%d B", totalSizeBytes)
	}

	div, exp := int64(unit), 0
	for n := int64(totalSizeBytes) / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	sizeInUnits := float64(totalSizeBytes) / float64(div)
	unitSymbol := "BKMGTPE"[exp]

	return fmt.Sprintf("%.1f %cB", sizeInUnits, unitSymbol)
}

// FormatETA formats an ETA (in seconds) into a human-readable string
func FormatETA(etaInSeconds int) string {
	const noETA = 8640000

	if etaInSeconds == noETA {
		return "No ETA"
	}

	etaDuration := time.Duration(etaInSeconds) * time.Second

	hours := int(etaDuration.Hours())
	minutes := int(etaDuration.Minutes()) % 60
	seconds := int(etaDuration.Seconds()) % 60

	switch {
	case hours > 0:
		return fmt.Sprintf("%d hours left", hours)
	case minutes > 0:
		return fmt.Sprintf("%d minutes left", minutes)
	default:
		return fmt.Sprintf("%d seconds left", seconds)
	}
}
