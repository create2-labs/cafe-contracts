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

// optionalRaw keeps a JSON value as it arrived. Null and a missing field are omitted,
// the same way an empty interface{} with omitempty is serialized today.
func optionalRaw(raw map[string]json.RawMessage, key string) (json.RawMessage, error) {
	value, ok := raw[key]
	if !ok {
		return nil, nil
	}
	trimmed := bytes.TrimSpace(value)
	if bytes.Equal(trimmed, []byte("null")) {
		return nil, nil
	}
	if !json.Valid(trimmed) {
		return nil, fmt.Errorf("%w: %s", ErrMalformed, key)
	}
	out := make(json.RawMessage, len(trimmed))
	copy(out, trimmed)
	return out, nil
}

func optionalString(raw map[string]json.RawMessage, key string) (string, error) {
	value, ok := raw[key]
	if !ok {
		return "", nil
	}
	if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return "", fmt.Errorf("%w: %s", ErrMalformed, key)
	}
	var s string
	if err := json.Unmarshal(bytes.TrimSpace(value), &s); err != nil {
		return "", fmt.Errorf("%w: %s", ErrMalformed, key)
	}
	return s, nil
}

func requiredString(raw map[string]json.RawMessage, key string) (string, error) {
	s, err := optionalString(raw, key)
	if err != nil {
		return "", err
	}
	if _, ok := raw[key]; !ok || s == "" {
		return "", fmt.Errorf("%w: %s", ErrMalformed, key)
	}
	return s, nil
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

func setTarget(endpoint, address *string, raw map[string]json.RawMessage) error {
	var err error
	if *endpoint, err = optionalString(raw, "endpoint"); err != nil {
		return err
	}
	*address, err = optionalString(raw, "address")
	return err
}

type head struct {
	ScanID, Kind, UserID, Endpoint, Address string
}

func readHead(raw map[string]json.RawMessage, withScanID bool) (head, error) {
	var h head
	var err error
	if withScanID {
		if h.ScanID, err = uuidField(raw, "scan_id"); err != nil {
			return head{}, err
		}
	}
	if h.Kind, err = requiredString(raw, "kind"); err != nil {
		return head{}, err
	}
	if h.UserID, err = uuidField(raw, "user_id"); err != nil {
		return head{}, err
	}
	if err = setTarget(&h.Endpoint, &h.Address, raw); err != nil {
		return head{}, err
	}
	return h, nil
}

func applyHead(scanID, kind, userID, endpoint, address *string, h head) {
	*scanID, *kind, *userID, *endpoint, *address = h.ScanID, h.Kind, h.UserID, h.Endpoint, h.Address
}

func validateID(field, s string) error {
	canonical, ok := canonicalUUID(s)
	if !ok || s != canonical {
		return fmt.Errorf("%w: %s", ErrMalformed, field)
	}
	return nil
}

func validateScan(scanID, kind, userID string) error {
	if err := validateID("scan_id", scanID); err != nil {
		return err
	}
	if kind == "" {
		return fmt.Errorf("%w: kind", ErrMalformed)
	}
	return validateID("user_id", userID)
}

func text(field, s string) error {
	if s == "" {
		return fmt.Errorf("%w: %s", ErrMalformed, field)
	}
	return nil
}

func canonicalUUID(s string) (string, bool) {
	if len(s) != 36 {
		return "", false
	}
	for i, c := range s {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return "", false
			}
		default:
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
				return "", false
			}
		}
	}
	return strings.ToLower(s), true
}
