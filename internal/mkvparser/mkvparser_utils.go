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

// UTF8ToASSText
//
// note: needs testing
func UTF8ToASSText(text string) string {
	// Convert HTML entities to actual characters
	type tags struct {
		values  []string
		replace string
	}
	t := []tags{
		{values: []string{"&lt;"}, replace: "<"},
		{values: []string{"&gt;"}, replace: ">"},
		{values: []string{"&amp;"}, replace: "&"},
		{values: []string{"&nbsp;"}, replace: "\\h"},
		{values: []string{"&quot;"}, replace: "\""},
		{values: []string{"&#39;"}, replace: "'"},
		{values: []string{"&apos;"}, replace: "'"},
		{values: []string{"&laquo;"}, replace: "«"},
		{values: []string{"&raquo;"}, replace: "»"},
		{values: []string{"&ndash;"}, replace: "-"},
		{values: []string{"&mdash;"}, replace: "—"},
		{values: []string{"&hellip;"}, replace: "…"},
		{values: []string{"&copy;"}, replace: "©"},
		{values: []string{"&reg;"}, replace: "®"},
		{values: []string{"&trade;"}, replace: "™"},
		{values: []string{"&euro;"}, replace: "€"},
		{values: []string{"&pound;"}, replace: "£"},
		{values: []string{"&yen;"}, replace: "¥"},
		{values: []string{"&dollar;"}, replace: "$"},
		{values: []string{"&cent;"}, replace: "¢"},
		//
		{values: []string{"\r\n", "\n", "\r", "<br>", "<br/>", "<br />", "<BR>", "<BR/>", "<BR />"}, replace: "\\N"},
		{values: []string{"<b>", "<B>", "<strong>"}, replace: "{\\b1}"},
		{values: []string{"</b>", "</B>", "</strong>"}, replace: "{\\b0}"},
		{values: []string{"<i>", "<I>", "<em>"}, replace: "{\\i1}"},
		{values: []string{"</i>", "</I>", "</em>"}, replace: "{\\i0}"},
		{values: []string{"<u>", "<U>"}, replace: "{\\u1}"},
		{values: []string{"</u>", "</U>"}, replace: "{\\u0}"},
		{values: []string{"<s>", "<S>", "<strike>", "<del>"}, replace: "{\\s1}"},
		{values: []string{"</s>", "</S>", "</strike>", "</del>"}, replace: "{\\s0}"},
		{values: []string{"<center>", "<CENTER>"}, replace: "{\\an8}"},
		{values: []string{"</center>", "</CENTER>"}, replace: ""},
		{values: []string{"<ruby>", "<rt>"}, replace: "{\\ruby1}"},
		{values: []string{"</ruby>", "</rt>"}, replace: "{\\ruby0}"},
		{values: []string{"<p>", "<P>", "<div>", "<DIV>"}, replace: ""},
		{values: []string{"</p>", "</P>", "</div>", "</DIV>"}, replace: "\\N"},
	}

	for _, tag := range t {
		for _, value := range tag.values {
			text = strings.ReplaceAll(text, value, tag.replace)
		}
	}

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

	return text
}
