package report

import (
	"encoding/json"
	"fmt"
	"strings"
)

const refPrefix = "$ref:"

// minRecordSize is the minimum string length to consider for deduplication
const minRecordSize = 64

// Deduplicate replaces repeated large values with pointer references ("$ref:N")
// and stores the originals in a Records map.
func (ir *IssueReport) Deduplicate() {
	// value (as string key) -> ref string
	seen := make(map[string]string)
	idx := 0

	ir.Records = make(map[string]interface{})

	// internString deduplicates string values.
	internString := func(val string) string {
		if len(val) < minRecordSize {
			return val
		}
		if ref, ok := seen[val]; ok {
			return ref
		}
		key := fmt.Sprintf("%d", idx)
		idx++
		ref := refPrefix + key
		seen[val] = ref
		ir.Records[key] = val
		return ref
	}

	// internRaw deduplicates json.RawMessage values.
	// Returns the (possibly replaced) raw message and whether it was interned.
	internRaw := func(val json.RawMessage) json.RawMessage {
		s := string(val)
		if len(s) < minRecordSize {
			return val
		}
		if ref, ok := seen[s]; ok {
			// Return quoted ref string as JSON
			b, _ := json.Marshal(ref)
			return b
		}
		key := fmt.Sprintf("%d", idx)
		idx++
		ref := refPrefix + key
		seen[s] = ref
		// Store as raw JSON (preserving the original type)
		ir.Records[key] = json.RawMessage(val)
		b, _ := json.Marshal(ref)
		return b
	}

	for _, log := range ir.NetworkLogs {
		log.DataPreview = internString(log.DataPreview)
		log.Body = internString(log.Body)
		log.URL = internString(log.URL)
		log.PageURL = internString(log.PageURL)
	}

	for _, log := range ir.ReactQueryLogs {
		log.DataPreview = internString(log.DataPreview)
		log.PageURL = internString(log.PageURL)
	}

	for _, log := range ir.ConsoleLogs {
		log.Content = internString(log.Content)
		log.PageURL = internString(log.PageURL)
	}

	for _, log := range ir.ClickLogs {
		log.PageURL = internString(log.PageURL)
		if log.ClassName != nil {
			v := internString(*log.ClassName)
			log.ClassName = &v
		}
	}

	for _, log := range ir.NavigationLogs {
		log.From = internString(log.From)
		log.To = internString(log.To)
	}

	for _, log := range ir.WebSocketLogs {
		log.Payload = internRaw(log.Payload)
	}

	// remove records that are only referenced once (no savings)
	refCount := make(map[string]int)
	for _, ref := range seen {
		refCount[ref] = 0
	}

	countRefStr := func(val string) {
		if strings.HasPrefix(val, refPrefix) {
			refCount[val]++
		}
	}

	countRefRaw := func(val json.RawMessage) {
		// Check if the raw message is a quoted $ref:N string
		var s string
		if json.Unmarshal(val, &s) == nil && strings.HasPrefix(s, refPrefix) {
			refCount[s]++
		}
	}

	for _, log := range ir.NetworkLogs {
		countRefStr(log.DataPreview)
		countRefStr(log.Body)
		countRefStr(log.URL)
		countRefStr(log.PageURL)
	}
	for _, log := range ir.ReactQueryLogs {
		countRefStr(log.DataPreview)
		countRefStr(log.PageURL)
	}
	for _, log := range ir.ConsoleLogs {
		countRefStr(log.Content)
		countRefStr(log.PageURL)
	}
	for _, log := range ir.ClickLogs {
		countRefStr(log.PageURL)
		if log.ClassName != nil {
			countRefStr(*log.ClassName)
		}
	}
	for _, log := range ir.NavigationLogs {
		countRefStr(log.From)
		countRefStr(log.To)
	}
	for _, log := range ir.WebSocketLogs {
		countRefRaw(log.Payload)
	}

	// inline back any value that only appears once
	unreferenceStr := func(val string) string {
		if !strings.HasPrefix(val, refPrefix) {
			return val
		}
		if refCount[val] <= 1 {
			key := strings.TrimPrefix(val, refPrefix)
			original, ok := ir.Records[key]
			if ok {
				delete(ir.Records, key)
				return original.(string)
			}
		}
		return val
	}

	unreferenceRaw := func(val json.RawMessage) json.RawMessage {
		var s string
		if json.Unmarshal(val, &s) != nil || !strings.HasPrefix(s, refPrefix) {
			return val
		}
		if refCount[s] <= 1 {
			key := strings.TrimPrefix(s, refPrefix)
			original, ok := ir.Records[key]
			if ok {
				delete(ir.Records, key)
				// original is a json.RawMessage
				if raw, ok := original.(json.RawMessage); ok {
					return raw
				}
			}
		}
		return val
	}

	for _, log := range ir.NetworkLogs {
		log.DataPreview = unreferenceStr(log.DataPreview)
		log.Body = unreferenceStr(log.Body)
		log.URL = unreferenceStr(log.URL)
		log.PageURL = unreferenceStr(log.PageURL)
	}
	for _, log := range ir.ReactQueryLogs {
		log.DataPreview = unreferenceStr(log.DataPreview)
		log.PageURL = unreferenceStr(log.PageURL)
	}
	for _, log := range ir.ConsoleLogs {
		log.Content = unreferenceStr(log.Content)
		log.PageURL = unreferenceStr(log.PageURL)
	}
	for _, log := range ir.ClickLogs {
		log.PageURL = unreferenceStr(log.PageURL)
		if log.ClassName != nil {
			v := unreferenceStr(*log.ClassName)
			log.ClassName = &v
		}
	}
	for _, log := range ir.NavigationLogs {
		log.From = unreferenceStr(log.From)
		log.To = unreferenceStr(log.To)
	}
	for _, log := range ir.WebSocketLogs {
		log.Payload = unreferenceRaw(log.Payload)
	}

	// If no records remain, omit it
	if len(ir.Records) == 0 {
		ir.Records = nil
	}
}
