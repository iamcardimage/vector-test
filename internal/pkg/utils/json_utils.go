package utils

import (
	"encoding/json"
	"errors"
)

func ExtractUserID(raw json.RawMessage) (int, error) {
	var tmp map[string]any
	if err := json.Unmarshal(raw, &tmp); err != nil {
		return 0, err
	}
	idVal, ok := tmp["id"]
	if !ok {
		return 0, errors.New("id field not found")
	}
	switch v := idVal.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	default:
		return 0, errors.New("invalid id type")
	}
}

func ExtractString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}

	if pi, ok := m["person_info"].(map[string]any); ok {
		if v, ok := pi[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}
