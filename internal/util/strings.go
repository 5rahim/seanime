package util

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/dustin/go-humanize"
)

func Bytes(size uint64) string {
	switch runtime.GOOS {
	case "darwin":
		return humanize.Bytes(size)
	default:
		return humanize.IBytes(size)
	}
}

func Decode(s string) string {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return ""
	}
	return string(decoded)
}

func GenerateCryptoID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func IsMostlyLatinString(str string) bool {
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

// ToHumanReadableSpeed converts an integer representing bytes per second to a human-readable format using binary notation
func ToHumanReadableSpeed(bytesPerSecond int) string {
	if bytesPerSecond <= 0 {
		return `0 KiB/s`
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
	return fmt.Sprintf("%.1f %ciB/s", float64(bytesPerSecond)/float64(div), "KMGTPE"[exp])
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

func Pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

// NormalizePath normalizes a path by converting it to lowercase and replacing backslashes with forward slashes
// Warning: Do not use the returned string for anything filesystem related, only for comparison
func NormalizePath(path string) (ret string) {
	return strings.ToLower(filepath.ToSlash(path))
}

func Base64EncodeStr(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func Base64DecodeStr(str string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func IsBase64(s string) bool {
	// 1. Check if string is empty
	if len(s) == 0 {
		return false
	}

	// 2. Check if length is valid (must be multiple of 4)
	if len(s)%4 != 0 {
		return false
	}

	// 3. Check for valid padding
	padding := strings.Count(s, "=")
	if padding > 2 {
		return false
	}

	// 4. Check if padding is at the end only
	if padding > 0 && !strings.HasSuffix(s, strings.Repeat("=", padding)) {
		return false
	}

	// 5. Check if string contains only valid base64 characters
	validChars := regexp.MustCompile("^[A-Za-z0-9+/]*=*$")
	if !validChars.MatchString(s) {
		return false
	}

	// 6. Try to decode - this is the final verification
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

var snakecaseSplitRegex = regexp.MustCompile(`[\W_]+`)

func Snakecase(str string) string {
	var result strings.Builder

	// split at any non word character and underscore
	words := snakecaseSplitRegex.Split(str, -1)

	for _, word := range words {
		if word == "" {
			continue
		}

		if result.Len() > 0 {
			result.WriteString("_")
		}

		for i, c := range word {
			if unicode.IsUpper(c) && i > 0 &&
				// is not a following uppercase character
				!unicode.IsUpper(rune(word[i-1])) {
				result.WriteString("_")
			}

			result.WriteRune(c)
		}
	}

	return strings.ToLower(result.String())
}

// randomStringWithAlphabet generates a cryptographically random string
// with the specified length and characters set.
//
// It panics if for some reason rand.Int returns a non-nil error.
func RandomStringWithAlphabet(length int, alphabet string) string {
	b := make([]byte, length)
	max := big.NewInt(int64(len(alphabet)))

	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err)
		}
		b[i] = alphabet[n.Int64()]
	}

	return string(b)
}
