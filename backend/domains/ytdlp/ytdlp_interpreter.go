package ytdlp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// GetJSONValue extracts values from JSON using dot notation paths
// Examples: "name", "address.street", "hobbies.0", "users.1.name"
func GetJSONValue(jsonStr string, path string) (interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return getValueFromPath(data, path)
}

// getValueFromPath navigates through the data structure using dot notation
func getValueFromPath(data interface{}, path string) (interface{}, error) {
	if path == "" {
		return data, nil
	}

	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			// Navigate object property
			val, exists := v[part]
			if !exists {
				return nil, fmt.Errorf("key '%s' not found in object at path '%s'", part, strings.Join(parts[:i+1], "."))
			}
			current = val
		case []interface{}:
			// Navigate array index
			index, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid array index '%s' at path '%s'", part, strings.Join(parts[:i+1], "."))
			}
			if index < 0 || index >= len(v) {
				return nil, fmt.Errorf("array index %d out of bounds (length %d) at path '%s'", index, len(v), strings.Join(parts[:i+1], "."))
			}
			current = v[index]
		default:
			return nil, fmt.Errorf("cannot navigate further: expected object or array at path '%s', got %T", strings.Join(parts[:i], "."), current)
		}
	}

	return current, nil
}

// Helper functions for type-safe extraction
func GetString(jsonStr, path string) (string, error) {
	val, err := GetJSONValue(jsonStr, path)
	if err != nil {
		return "", err
	}

	if str, ok := val.(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("value at path '%s' is not a string, got %T", path, val)
}

func GetFloat(jsonStr, path string) (float64, error) {
	val, err := GetJSONValue(jsonStr, path)
	if err != nil {
		return 0, err
	}

	if num, ok := val.(float64); ok {
		return num, nil
	}

	return 0, fmt.Errorf("value at path '%s' is not a float64, got %T", path, val)
}

func GetBool(jsonStr, path string) (bool, error) {
	val, err := GetJSONValue(jsonStr, path)
	if err != nil {
		return false, err
	}

	if b, ok := val.(bool); ok {
		return b, nil
	}

	return false, fmt.Errorf("value at path '%s' is not a bool, got %T", path, val)
}
