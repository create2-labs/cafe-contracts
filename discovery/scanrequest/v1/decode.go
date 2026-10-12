package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

func objectFields(data []byte) (map[string]json.RawMessage, error) {
	dec := json.NewDecoder(bytes.NewReader(bytes.TrimSpace(data)))
	var raw map[string]json.RawMessage
	if err := dec.Decode(&raw); err != nil || raw == nil || dec.More() {
		return nil, fmt.Errorf("%w: json", ErrMalformed)
	}
	return raw, nil
}

func requiredString(raw map[string]json.RawMessage, key string) (string, error) {
	value, ok := raw[key]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrMalformed, key)
	}
	var s string
	if err := json.Unmarshal(bytes.TrimSpace(value), &s); err != nil {
		return "", fmt.Errorf("%w: %s", ErrMalformed, key)
	}
	return s, nil
}

// boolField reads a JSON bool. A missing field is false, matching encoding/json
// on the current TLSScanMessage, which has no omitempty. Null and other types
// are rejected.
func boolField(raw map[string]json.RawMessage, key string) (bool, error) {
	value, ok := raw[key]
	if !ok {
		return false, nil
	}
	switch string(bytes.TrimSpace(value)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("%w: %s", ErrMalformed, key)
	}
}

func uuidField(raw map[string]json.RawMessage, key string) (string, error) {
	s, err := requiredString(raw, key)
	if err != nil {
		return "", err
	}
	if canonical, ok := canonicalUUID(s); ok {
		return canonical, nil
	}
	return s, nil
}

func validateID(field, s string) error {
	canonical, ok := canonicalUUID(s)
	if !ok || s != canonical {
		return fmt.Errorf("%w: %s", ErrMalformed, field)
	}
	return nil
}

func canonicalUUID(s string) (string, bool) {
	if !isUUID(s) {
		return "", false
	}
	return strings.ToLower(s), true
}

func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		if !isHex(c) {
			return false
		}
	}
	return true
}

func isHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}
