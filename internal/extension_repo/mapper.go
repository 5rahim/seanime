package extension_repo

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/dop251/goja"
)

var (
	_ goja.FieldNameMapper = (*FieldMapper)(nil)
)

// FieldMapper provides custom mapping between Go and JavaScript property names.
//
// It is similar to the builtin "uncapFieldNameMapper" but also converts
// all uppercase identifiers to their lowercase equivalent (eg. "GET" -> "get").
// It also checks for JSON tags and uses them if they exist.
type FieldMapper struct {
}

// FieldName implements the [FieldNameMapper.FieldName] interface method.
func (u FieldMapper) FieldName(t reflect.Type, f reflect.StructField) string {
	// First check for a JSON tag
	if jsonTag := f.Tag.Get("json"); jsonTag != "" {
		// Split by comma to handle cases like `json:"name,omitempty"`
		parts := strings.Split(jsonTag, ",")
		// If the JSON tag isn't "-" (which means don't include this field)
		if parts[0] != "-" && parts[0] != "" {
			return parts[0]
		}
	}
	// Fall back to default conversion
	return convertGoToJSName(f.Name)
}

// MethodName implements the [FieldNameMapper.MethodName] interface method.
func (u FieldMapper) MethodName(_ reflect.Type, m reflect.Method) string {
	return convertGoToJSName(m.Name)
}

var nameExceptions = map[string]string{"OAuth2": "oauth2"}

func convertGoToJSName(name string) string {
	if v, ok := nameExceptions[name]; ok {
		return v
	}

	startUppercase := make([]rune, 0, len(name))

	for _, c := range name {
		if c != '_' && !unicode.IsUpper(c) && !unicode.IsDigit(c) {
			break
		}

		startUppercase = append(startUppercase, c)
	}

	totalStartUppercase := len(startUppercase)

	// all uppercase eg. "JSON" -> "json"
	if len(name) == totalStartUppercase {
		return strings.ToLower(name)
	}

	// eg. "JSONField" -> "jsonField"
	if totalStartUppercase > 1 {
		return strings.ToLower(name[0:totalStartUppercase-1]) + name[totalStartUppercase-1:]
	}

	// eg. "GetField" -> "getField"
	if totalStartUppercase == 1 {
		return strings.ToLower(name[0:1]) + name[1:]
	}

	return name
}
