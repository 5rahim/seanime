package util

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// DetectImageFormatAndDimensions attempts to decode the image dimensions and format.
// If the std decoder fails, it uses a fallback parsing to extract metadata.
// If a URL is provided, it can be used to help guess the format if bytes detection fails.
func DetectImageFormatAndDimensions(buf []byte, url string) (width, height int, format string, err error) {
	// 1. Try std
	var config image.Config
	config, format, err = image.DecodeConfig(bytes.NewReader(buf))
	if err == nil {
		return config.Width, config.Height, format, nil
	}

	// 1.5. Try AVIF
	if isAvif(buf) {
		format = "avif"
		width, height, err = parseAvifDimensions(buf)
		if err == nil {
			return width, height, format, nil
		}
		// Fallback if dimensions parsing failed, but format is detected
		return 0, 0, format, nil
	}

	// 2. Fallback for JPEG
	if isJpeg(buf) {
		format = "jpeg"
		width, height, err = parseJpegDimensions(buf)
		if err == nil {
			return width, height, format, nil
		}
	}

	// 3. Fallback format guessing if still guessable
	format = guessImageFormat(buf)
	if format == "" && url != "" {
		format = guessFormatFromURL(url)
	}

	if format != "" {
		// can't guess the format, but don't fail
		return 0, 0, format, nil
	}

	return 0, 0, "", fmt.Errorf("failed to decode image format: %w", err)
}

func isJpeg(buf []byte) bool {
	return len(buf) >= 3 && buf[0] == 0xFF && buf[1] == 0xD8 && buf[2] == 0xFF
}

func guessImageFormat(buf []byte) string {
	if len(buf) < 4 {
		return ""
	}
	if isAvif(buf) {
		return "avif"
	}
	// PNG: \x89PNG
	if buf[0] == 0x89 && buf[1] == 0x50 && buf[2] == 0x4E && buf[3] == 0x47 {
		return "png"
	}
	// GIF: GIF8
	if len(buf) >= 6 && buf[0] == 'G' && buf[1] == 'I' && buf[2] == 'F' && buf[3] == '8' {
		return "gif"
	}
	// WEBP: RIFF....WEBP
	if len(buf) >= 12 && buf[0] == 'R' && buf[1] == 'I' && buf[2] == 'F' && buf[3] == 'F' &&
		buf[8] == 'W' && buf[9] == 'E' && buf[10] == 'B' && buf[11] == 'P' {
		return "webp"
	}
	// BMP: BM
	if buf[0] == 'B' && buf[1] == 'M' {
		return "bmp"
	}
	// TIFF
	if (buf[0] == 'I' && buf[1] == 'I' && buf[2] == '*') || (buf[0] == 'M' && buf[1] == 'M' && buf[2] == '*') {
		return "tiff"
	}
	return ""
}

func guessFormatFromURL(url string) string {
	url = strings.ToLower(url)
	if strings.Contains(url, ".avif") {
		return "avif"
	}
	if strings.Contains(url, ".png") {
		return "png"
	}
	if strings.Contains(url, ".jpg") || strings.Contains(url, ".jpeg") {
		return "jpeg"
	}
	if strings.Contains(url, ".webp") {
		return "webp"
	}
	if strings.Contains(url, ".gif") {
		return "gif"
	}
	if strings.Contains(url, ".bmp") {
		return "bmp"
	}
	if strings.Contains(url, ".tiff") {
		return "tiff"
	}
	return ""
}

func parseJpegDimensions(data []byte) (width, height int, err error) {
	if len(data) < 4 {
		return 0, 0, fmt.Errorf("invalid jpeg data")
	}
	i := 2
	for i < len(data)-1 {
		if data[i] != 0xFF {
			i++
			continue
		}
		// skip padding
		for i < len(data) && data[i] == 0xFF {
			i++
		}
		if i >= len(data) {
			break
		}
		marker := data[i]
		i++

		if marker == 0x00 {
			continue
		}
		if marker == 0xD9 { // EOI
			break
		}

		// markers without payload sizes
		if (marker >= 0xD0 && marker <= 0xD7) || marker == 0x01 {
			continue
		}

		if i+2 > len(data) {
			break
		}
		length := int(data[i])<<8 | int(data[i+1])

		// SOF0 (0xC0) through SOF15 (0xCF), except DHT (0xC4), JPG (0xC8), DAC (0xCC)
		isSOF := (marker >= 0xC0 && marker <= 0xCF) && marker != 0xC4 && marker != 0xC8 && marker != 0xCC
		if isSOF {
			if length >= 7 && i+7 <= len(data) {
				// SOF payload structure:
				// data[i+2]: precision
				// data[i+3], data[i+4]: height
				// data[i+5], data[i+6]: width
				height = int(data[i+3])<<8 | int(data[i+4])
				width = int(data[i+5])<<8 | int(data[i+6])
				return width, height, nil
			}
			break
		}

		i += length
	}
	return 0, 0, fmt.Errorf("SOF marker not found")
}

func isAvif(buf []byte) bool {
	if len(buf) < 12 {
		return false
	}
	// Box type 'ftyp' at offset 4
	if buf[4] != 'f' || buf[5] != 't' || buf[6] != 'y' || buf[7] != 'p' {
		return false
	}
	// Major brand 'avif' or 'avis' at offset 8
	brand := string(buf[8:12])
	return brand == "avif" || brand == "avis"
}

func parseAvifDimensions(buf []byte) (width, height int, err error) {
	idx := bytes.Index(buf, []byte("ispe"))
	if idx == -1 {
		return 0, 0, fmt.Errorf("ispe box not found")
	}
	// 'ispe' (4 bytes) + version/flags (4 bytes) + width (4 bytes) + height (4 bytes) = 16 bytes
	if idx+16 > len(buf) {
		return 0, 0, fmt.Errorf("truncated ispe box")
	}

	width = int(buf[idx+8])<<24 | int(buf[idx+9])<<16 | int(buf[idx+10])<<8 | int(buf[idx+11])
	height = int(buf[idx+12])<<24 | int(buf[idx+13])<<16 | int(buf[idx+14])<<8 | int(buf[idx+15])

	return width, height, nil
}
