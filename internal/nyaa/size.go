package nyaa

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"regexp"
	"strconv"
	"strings"
)

func (t *Torrent) GetSizeInBytes() int64 {
	bytes, _ := stringToBytes(t.Size)
	return bytes
}

func stringToBytes(str string) (int64, error) {
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

	spew.Dump(match)

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
