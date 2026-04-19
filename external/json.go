package external

import (
	"encoding/json"
	"fmt"
	"os"
)

func loadJSON(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	result := make(map[string]any)
	flatten("", raw, result)
	return result, nil
}

func flatten(prefix string, m map[string]any, result map[string]any) {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch val := v.(type) {
		case map[string]any:
			flatten(key, val, result)
		case []any:
			for i, item := range val {
				itemKey := fmt.Sprintf("%s.%d", key, i)
				if sub, ok := item.(map[string]any); ok {
					flatten(itemKey, sub, result)
				} else {
					result[itemKey] = item
				}
			}
		default:
			result[key] = v
		}
	}
}
