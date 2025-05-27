package mkvparser

import (
	"bytes"
	"io"
	"strings"
)

// ReadIsMkvOrWebm reads the first 1KB of the stream to determine if it is a Matroska or WebM file.
// It returns the mime type and a boolean indicating if it is a Matroska or WebM file.
// It seeks to the beginning of the stream before and after reading.
func ReadIsMkvOrWebm(r io.ReadSeeker) (string, bool) {
	// Go to the beginning of the stream
	_, err := r.Seek(0, io.SeekStart)
	if err != nil {
		return "", false
	}
	defer r.Seek(0, io.SeekStart)

	return isMkvOrWebm(r)
}

func isMkvOrWebm(r io.Reader) (string, bool) {
	header := make([]byte, 1024) // Read the first 1KB to be safe
	n, err := r.Read(header)
	if err != nil {
		return "", false
	}

	// Check for EBML magic bytes
	if !bytes.HasPrefix(header, []byte{0x1A, 0x45, 0xDF, 0xA3}) {
		return "", false
	}

	// Look for the DocType tag (0x42 82) and check the string
	docTypeTag := []byte{0x42, 0x82}
	idx := bytes.Index(header, docTypeTag)
	if idx == -1 || idx+3 >= n {
		return "", false
	}

	size := int(header[idx+2]) // Size of DocType field
	if idx+3+size > n {
		return "", false
	}

	docType := string(header[idx+3 : idx+3+size])
	switch docType {
	case "matroska":
		return "video/x-matroska", true
	case "webm":
		return "video/webm", true
	default:
		return "", false
	}
}

func UTF8ToASS(text string) string {
	// Convert HTML entities to actual characters
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&nbsp;", "\\h")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&apos;", "'")
	text = strings.ReplaceAll(text, "&laquo;", "«")
	text = strings.ReplaceAll(text, "&raquo;", "»")
	text = strings.ReplaceAll(text, "&ndash;", "-")
	text = strings.ReplaceAll(text, "&mdash;", "—")
	text = strings.ReplaceAll(text, "&hellip;", "…")

	// Convert line breaks
	text = strings.ReplaceAll(text, "\r\n", "\\N")
	text = strings.ReplaceAll(text, "\n", "\\N")
	text = strings.ReplaceAll(text, "\r", "\\N")
	text = strings.ReplaceAll(text, "<br>", "\\N")
	text = strings.ReplaceAll(text, "<br/>", "\\N")
	text = strings.ReplaceAll(text, "<br />", "\\N")
	text = strings.ReplaceAll(text, "<BR>", "\\N")
	text = strings.ReplaceAll(text, "<BR/>", "\\N")
	text = strings.ReplaceAll(text, "<BR />", "\\N")

	// Convert basic HTML tags to ASS tags
	text = strings.ReplaceAll(text, "<b>", "{\\b1}")
	text = strings.ReplaceAll(text, "</b>", "{\\b0}")
	text = strings.ReplaceAll(text, "<B>", "{\\b1}")
	text = strings.ReplaceAll(text, "</B>", "{\\b0}")
	text = strings.ReplaceAll(text, "<strong>", "{\\b1}")
	text = strings.ReplaceAll(text, "</strong>", "{\\b0}")

	// Italic tags
	text = strings.ReplaceAll(text, "<i>", "{\\i1}")
	text = strings.ReplaceAll(text, "</i>", "{\\i0}")
	text = strings.ReplaceAll(text, "<I>", "{\\i1}")
	text = strings.ReplaceAll(text, "</I>", "{\\i0}")
	text = strings.ReplaceAll(text, "<em>", "{\\i1}")
	text = strings.ReplaceAll(text, "</em>", "{\\i0}")

	// Underline tags
	text = strings.ReplaceAll(text, "<u>", "{\\u1}")
	text = strings.ReplaceAll(text, "</u>", "{\\u0}")
	text = strings.ReplaceAll(text, "<U>", "{\\u1}")
	text = strings.ReplaceAll(text, "</U>", "{\\u0}")

	// Strikethrough tags
	text = strings.ReplaceAll(text, "<s>", "{\\s1}")
	text = strings.ReplaceAll(text, "</s>", "{\\s0}")
	text = strings.ReplaceAll(text, "<S>", "{\\s1}")
	text = strings.ReplaceAll(text, "</S>", "{\\s0}")
	text = strings.ReplaceAll(text, "<strike>", "{\\s1}")
	text = strings.ReplaceAll(text, "</strike>", "{\\s0}")
	text = strings.ReplaceAll(text, "<del>", "{\\s1}")
	text = strings.ReplaceAll(text, "</del>", "{\\s0}")

	// Font tags with color and size
	if strings.Contains(text, "<font") || strings.Contains(text, "<FONT") {
		// Process font tags with attributes
		for strings.Contains(text, "<font") || strings.Contains(text, "<FONT") {
			var tagStart int
			if idx := strings.Index(text, "<font"); idx != -1 {
				tagStart = idx
			} else {
				tagStart = strings.Index(text, "<FONT")
			}

			if tagStart == -1 {
				break
			}

			tagEnd := strings.Index(text[tagStart:], ">")
			if tagEnd == -1 {
				break
			}
			tagEnd += tagStart

			// Extract the font tag content
			fontTag := text[tagStart : tagEnd+1]
			replacement := ""

			// Handle color attribute
			if colorStart := strings.Index(fontTag, "color=\""); colorStart != -1 {
				colorStart += 7 // length of 'color="'
				if colorEnd := strings.Index(fontTag[colorStart:], "\""); colorEnd != -1 {
					color := fontTag[colorStart : colorStart+colorEnd]
					// Convert HTML color to ASS format
					if strings.HasPrefix(color, "#") {
						if len(color) == 7 { // #RRGGBB format
							color = "&H" + color[5:7] + color[3:5] + color[1:3] + "&" // Convert to ASS BGR format
						}
					}
					replacement += "{\\c" + color + "}"
				}
			}

			// Handle size attribute
			if sizeStart := strings.Index(fontTag, "size=\""); sizeStart != -1 {
				sizeStart += 6 // length of 'size="'
				if sizeEnd := strings.Index(fontTag[sizeStart:], "\""); sizeEnd != -1 {
					size := fontTag[sizeStart : sizeStart+sizeEnd]
					replacement += "{\\fs" + size + "}"
				}
			}

			// Handle face/family attribute
			if faceStart := strings.Index(fontTag, "face=\""); faceStart != -1 {
				faceStart += 6 // length of 'face="'
				if faceEnd := strings.Index(fontTag[faceStart:], "\""); faceEnd != -1 {
					face := fontTag[faceStart : faceStart+faceEnd]
					replacement += "{\\fn" + face + "}"
				}
			}

			// Replace the opening font tag
			text = text[:tagStart] + replacement + text[tagEnd+1:]

			// Find and remove the corresponding closing tag
			if closeStart := strings.Index(text, "</font>"); closeStart != -1 {
				text = text[:closeStart] + "{\\r}" + text[closeStart+7:]
			} else if closeStart = strings.Index(text, "</FONT>"); closeStart != -1 {
				text = text[:closeStart] + "{\\r}" + text[closeStart+7:]
			}
		}
	}

	// Handle alignment tags
	text = strings.ReplaceAll(text, "<center>", "{\\an8}")
	text = strings.ReplaceAll(text, "</center>", "")
	text = strings.ReplaceAll(text, "<CENTER>", "{\\an8}")
	text = strings.ReplaceAll(text, "</CENTER>", "")

	// Handle ruby/furigana text
	text = strings.ReplaceAll(text, "<ruby>", "{\\ruby1}")
	text = strings.ReplaceAll(text, "</ruby>", "{\\ruby0}")
	text = strings.ReplaceAll(text, "<rt>", "{\\rt}")
	text = strings.ReplaceAll(text, "</rt>", "{\\rt0}")

	// Clean up any
	text = strings.ReplaceAll(text, "<p>", "")
	text = strings.ReplaceAll(text, "</p>", "\\N")
	text = strings.ReplaceAll(text, "<P>", "")
	text = strings.ReplaceAll(text, "</P>", "\\N")
	text = strings.ReplaceAll(text, "<div>", "")
	text = strings.ReplaceAll(text, "</div>", "\\N")
	text = strings.ReplaceAll(text, "<DIV>", "")
	text = strings.ReplaceAll(text, "</DIV>", "\\N")

	return text
}
