package plugin_ui

import (
	"strings"
	"sync"
	"unsafe"
)

var (
	dangerousAttributes = map[string]struct{}{
		"onclick": {}, "ondblclick": {}, "onmousedown": {}, "onmouseup": {},
		"onmouseover": {}, "onmousemove": {}, "onmouseout": {}, "onkeydown": {},
		"onkeyup": {}, "onkeypress": {}, "onload": {}, "onunload": {},
		"onabort": {}, "onerror": {}, "onresize": {}, "onscroll": {},
		"onblur": {}, "onchange": {}, "onfocus": {}, "onreset": {},
		"onselect": {}, "onsubmit": {}, "ondrag": {}, "ondragend": {},
		"ondragenter": {}, "ondragleave": {}, "ondragover": {}, "ondragstart": {},
		"ondrop": {}, "onanimationstart": {}, "onanimationend": {},
		"onanimationiteration": {}, "ontransitionend": {},
	}

	dangerousProperties = map[string]struct{}{
		"innerhtml": {}, "outerhtml": {},
		"onclick": {}, "ondblclick": {}, "onmousedown": {}, "onmouseup": {},
		"onmouseover": {}, "onmousemove": {}, "onmouseout": {}, "onkeydown": {},
		"onkeyup": {}, "onkeypress": {}, "onload": {}, "onunload": {},
		"onabort": {}, "onerror": {}, "onresize": {}, "onscroll": {},
		"onblur": {}, "onchange": {}, "onfocus": {}, "onreset": {},
		"onselect": {}, "onsubmit": {}, "ondrag": {}, "ondragend": {},
		"ondragenter": {}, "ondragleave": {}, "ondragover": {}, "ondragstart": {},
		"ondrop": {},
	}

	eventHandlerKeywords = []string{
		"onclick", "ondblclick", "onmousedown", "onmouseup", "onmouseover", "onmousemove", "onmouseout",
		"onkeydown", "onkeyup", "onkeypress",
		"onload", "onunload", "onabort", "onerror",
		"onresize", "onscroll",
		"onblur", "onchange", "onfocus", "onreset", "onselect", "onsubmit",
		"ondrag", "ondragend", "ondragenter", "ondragleave", "ondragover", "ondragstart", "ondrop",
	}
)

// bufferPool reuses byte slices for lowercasing
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 64)
	},
}

// unsafeString converts a byte slice to a string without allocation.
// DEVNOTE: The byte slice must not be modified while the string is in use.
func unsafeString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// toLowerBuffer writes the lowercase version of s into buf and returns buf.
// buf is grown if needed.
func toLowerBuffer(s string, buf []byte) []byte {
	// grow if needed
	if cap(buf) < len(s) {
		buf = make([]byte, 0, len(s))
	}
	buf = buf[:len(s)]

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		buf[i] = c
	}
	return buf
}

// isDangerousAttribute checks if an attribute name is dangerous (event handler)
func isDangerousAttribute(name string) bool {
	// This covers the case where the input is lowercase
	if _, ok := dangerousAttributes[name]; ok {
		return true
	}

	// fast fail: the string has no uppercase letters, step 1 covered it
	hasUpper := false
	for i := 0; i < len(name); i++ {
		if name[i] >= 'A' && name[i] <= 'Z' {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return false
	}

	// lower using pool + unsafe lookup
	bufPtr := bufferPool.Get().([]byte)
	defer bufferPool.Put(bufPtr) // Return to pool

	lowerBuf := toLowerBuffer(name, bufPtr)

	_, ok := dangerousAttributes[unsafeString(lowerBuf)]
	return ok
}

// isDangerousProperty checks if a property name is dangerous
func isDangerousProperty(name string) bool {
	if _, ok := dangerousProperties[name]; ok {
		return true
	}

	// fast fail: no uppercase
	hasUpper := false
	for i := 0; i < len(name); i++ {
		if name[i] >= 'A' && name[i] <= 'Z' {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return false
	}

	// lower using pool + unsafe lookup
	bufPtr := bufferPool.Get().([]byte)
	defer bufferPool.Put(bufPtr)

	lowerBuf := toLowerBuffer(name, bufPtr)
	_, ok := dangerousProperties[unsafeString(lowerBuf)]
	return ok
}

// containsDangerousHTML checks if HTML contains script tags or event handlers
func containsDangerousHTML(html string) bool {
	if len(html) == 0 {
		return false
	}

	// the string doesn't contain '<', 'j' (javascript), or '=' (event handlers usually have =),
	hasTrigger := false
	for i := 0; i < len(html); i++ {
		b := html[i]
		// Check for '<', '=', or 'j'/'J' (for javascript:) or 's'/'S' (for script)
		// can be a bit loose here to be fast
		if b == '<' || b == '=' || b == 'j' || b == 'J' {
			hasTrigger = true
			break
		}
	}
	if !hasTrigger {
		return false
	}

	bufPtr := bufferPool.Get().([]byte)
	defer bufferPool.Put(bufPtr)

	// convert the whole string to lowercase
	lowerBuf := toLowerBuffer(html, bufPtr)
	lowerHTML := unsafeString(lowerBuf)

	if strings.Contains(lowerHTML, "<script") {
		return true
	}
	if strings.Contains(lowerHTML, "javascript:") {
		return true
	}

	// check event handlers
	for _, handler := range eventHandlerKeywords {
		// we check for "onclick=" or "onclick ="
		idx := strings.Index(lowerHTML, handler)
		if idx != -1 {
			rest := lowerHTML[idx+len(handler):]
			if len(rest) > 0 {
				if rest[0] == '=' {
					return true
				}
				if rest[0] == ' ' {
					// Handle "onclick =" case
					trimmed := strings.TrimLeft(rest, " ")
					if len(trimmed) > 0 && trimmed[0] == '=' {
						return true
					}
				}
			}
		}
	}

	return false
}

// isDangerousAttributeValue checks if an attribute value contains dangerous content
func isDangerousAttributeValue(name, value string) bool {
	if len(value) < 11 { // shortest dangerous is "data:..." or "javascript:"
		return false
	}

	nameLen := len(name)
	isTargetAttr := false

	// check for "src" or "href" without allocation
	if nameLen == 3 && (name[0]|0x20 == 's') && (name[1]|0x20 == 'r') && (name[2]|0x20 == 'c') {
		isTargetAttr = true
	} else if nameLen == 4 && (name[0]|0x20 == 'h') && (name[1]|0x20 == 'r') && (name[2]|0x20 == 'e') && (name[3]|0x20 == 'f') {
		isTargetAttr = true
	}

	if !isTargetAttr {
		return false
	}

	start := 0
	for start < len(value) && value[start] <= ' ' {
		start++
	}
	valTrimmed := value[start:]

	if hasPrefixCaseInsensitive(valTrimmed, "javascript:") {
		return true
	}

	if hasPrefixCaseInsensitive(valTrimmed, "data:") {
		// If data protocol, check for html content type
		bufPtr := bufferPool.Get().([]byte)
		defer bufferPool.Put(bufPtr)

		lowerVal := toLowerBuffer(valTrimmed, bufPtr)
		return strings.Contains(unsafeString(lowerVal), "text/html")
	}

	return false
}

func hasPrefixCaseInsensitive(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		sb := s[i]
		pb := prefix[i]
		if (sb | 0x20) != (pb | 0x20) {
			return false
		}
	}
	return true
}

// containsDangerousCSS checks if CSS contains dangerous expressions
func containsDangerousCSS(css string) bool {
	if len(css) == 0 {
		return false
	}

	bufPtr := bufferPool.Get().([]byte)
	defer bufferPool.Put(bufPtr)

	lowerBuf := toLowerBuffer(css, bufPtr)
	lowerCSS := unsafeString(lowerBuf)

	if strings.Contains(lowerCSS, "javascript:") {
		return true
	}
	if strings.Contains(lowerCSS, "expression(") {
		return true
	}
	if strings.Contains(lowerCSS, "-moz-binding") {
		return true
	}
	return false
}
