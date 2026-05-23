package models

import (
	"errors"
	"strings"

	"github.com/goccy/go-json"
)

func SetSettingsPath(settings *Settings, path string, value interface{}) (*Settings, error) {
	if settings == nil {
		return nil, errors.New("settings is nil")
	}

	base, err := settingsToMap(settings)
	if err != nil {
		return nil, err
	}

	if err := setSettingsMapPath(base, path, value); err != nil {
		return nil, err
	}

	return mapToSettings(base)
}

func PatchSettings(settings *Settings, patch map[string]interface{}) (*Settings, error) {
	if settings == nil {
		return nil, errors.New("settings is nil")
	}

	base, err := settingsToMap(settings)
	if err != nil {
		return nil, err
	}

	mergeSettingsMap(base, patch)

	return mapToSettings(base)
}

func settingsToMap(in interface{}) (map[string]interface{}, error) {
	bytes, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	var ret map[string]interface{}
	if err := json.Unmarshal(bytes, &ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func mapToSettings(in map[string]interface{}) (*Settings, error) {
	bytes, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	var ret Settings
	if err := json.Unmarshal(bytes, &ret); err != nil {
		return nil, err
	}

	return &ret, nil
}

func mergeSettingsMap(dst map[string]interface{}, src map[string]interface{}) {
	for key, value := range src {
		srcMap, srcOk := value.(map[string]interface{})
		dstMap, dstOk := dst[key].(map[string]interface{})
		if srcOk && dstOk {
			mergeSettingsMap(dstMap, srcMap)
			continue
		}

		dst[key] = value
	}
}

func setSettingsMapPath(settings map[string]interface{}, path string, value interface{}) error {
	parts := splitSettingsPath(path)
	if len(parts) == 0 {
		return errors.New("settings path is empty")
	}

	curr := settings
	for _, part := range parts[:len(parts)-1] {
		next, ok := curr[part].(map[string]interface{})
		if !ok {
			next = map[string]interface{}{}
			curr[part] = next
		}

		curr = next
	}

	curr[parts[len(parts)-1]] = value
	return nil
}

func splitSettingsPath(path string) []string {
	parts := strings.Split(path, ".")
	ret := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			ret = append(ret, part)
		}
	}

	return ret
}
