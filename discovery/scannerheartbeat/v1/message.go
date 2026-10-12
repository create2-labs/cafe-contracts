package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// Message is one scanner heartbeat: scanner_id, kind, and timestamp.
type Message struct {
	ScannerID string `json:"scanner_id"`
	Kind      string `json:"kind"`
	Timestamp string `json:"timestamp"`
}

// Decode unmarshals and validates a heartbeat. A malformed body yields a zero Message.
func Decode(raw []byte) (Message, error) {
	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		if errors.Is(err, ErrMalformed) {
			return Message{}, err
		}
		return Message{}, fmt.Errorf("%w: json", ErrMalformed)
	}
	if err := msg.Validate(); err != nil {
		return Message{}, err
	}
	return msg, nil
}

// UnmarshalJSON rejects a non-object body and a field that is not a non-empty string.
func (m *Message) UnmarshalJSON(data []byte) error {
	*m = Message{}
	raw, err := objectFields(data)
	if err != nil {
		return err
	}
	if m.ScannerID, err = requiredString(raw, "scanner_id"); err != nil {
		return err
	}
	if m.Kind, err = requiredString(raw, "kind"); err != nil {
		return err
	}
	m.Timestamp, err = requiredString(raw, "timestamp")
	return err
}

// Validate checks the three required strings.
func (m Message) Validate() error {
	for _, field := range []struct{ name, value string }{
		{"scanner_id", m.ScannerID},
		{"kind", m.Kind},
		{"timestamp", m.Timestamp},
	} {
		if field.value == "" {
			return fmt.Errorf("%w: %s", ErrMalformed, field.name)
		}
	}
	return nil
}

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
	if err := json.Unmarshal(bytes.TrimSpace(value), &s); err != nil || s == "" {
		return "", fmt.Errorf("%w: %s", ErrMalformed, key)
	}
	return s, nil
}
