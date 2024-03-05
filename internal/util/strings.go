package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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
func ToHumanReadableSize(bytes int) string { // FIXME incorrect
	if bytes < 0 {
		return "Invalid size"
	}

	// Define the size units
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}

	// Loop through units and divide by 1024 until the size is less than 1024
	size := float64(bytes)
	unitIndex := 0
	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}

	// Use a format string to limit the number of decimal places
	return fmt.Sprintf("%.1f %s", size, units[unitIndex])
}

func StringSizeToBytes(str string) (int64, error) {
	// Regular expression to extract size and unit
	re := regexp.MustCompile(`(?i)^(\d+(\.\d+)?)\s*([KMGT]?i?B)$`)

	match := re.FindStringSubmatch(strings.TrimSpace(str))
	if match == nil {
		return 0, fmt.Errorf("invalid size format: %s", str)
	}

	// Extract the numeric part and convert to float64
	size, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse size: %s", err)
	}

	// Extract the unit and convert to lowercase
	unit := strings.ToLower(match[3])

	// Map units to their respective multipliers
	unitMultipliers := map[string]int64{
		"b":   1,
		"bi":  1,
		"kb":  1024,
		"kib": 1024,
		"mb":  1024 * 1024,
		"mib": 1024 * 1024,
		"gb":  1024 * 1024 * 1024,
		"gib": 1024 * 1024 * 1024,
		"tb":  1024 * 1024 * 1024 * 1024,
		"tib": 1024 * 1024 * 1024 * 1024,
	}

	// Apply the multiplier based on the unit
	multiplier, ok := unitMultipliers[unit]
	if !ok {
		return 0, fmt.Errorf("invalid unit: %s", unit)
	}

	// Calculate the total bytes
	bytes := int64(size * float64(multiplier))
	return bytes, nil
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
	case seconds < 0:
		return "No ETA"
	default:
		return fmt.Sprintf("%d seconds left", seconds)
	}
}
