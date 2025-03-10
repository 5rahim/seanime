package goja_bindings

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/dop251/goja"
)

var (
	_ goja.FieldNameMapper = (*DefaultFieldMapper)(nil)
)

// DefaultFieldMapper provides custom mapping between Go and JavaScript methods names.
//
// It is similar to the builtin "uncapFieldNameMapper" but also converts
// all uppercase identifiers to their lowercase equivalent (eg. "GET" -> "get").
type DefaultFieldMapper struct {
}

// FieldName implements the [FieldNameMapper.FieldName] interface method.
func (u DefaultFieldMapper) FieldName(_ reflect.Type, f reflect.StructField) string {
	return f.Name
}

// MethodName implements the [FieldNameMapper.MethodName] interface method.
func (u DefaultFieldMapper) MethodName(_ reflect.Type, m reflect.Method) string {
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
