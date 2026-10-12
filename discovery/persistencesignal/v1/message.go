package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// Ready is the persistence.ready body. The published value is {}.
type Ready struct{}

// Decode accepts a JSON object, including {}, and rejects an unreadable body.
// A malformed body yields a zero Ready.
func Decode(raw []byte) (Ready, error) {
	var msg Ready
	if err := json.Unmarshal(raw, &msg); err != nil {
		if errors.Is(err, ErrMalformed) {
			return Ready{}, err
		}
		return Ready{}, fmt.Errorf("%w: json", ErrMalformed)
	}
	return msg, nil
}

// UnmarshalJSON accepts any JSON object. Null, a non-object, and trailing
// data are rejected. Fields are not kept: Marshal emits {}.
func (m *Ready) UnmarshalJSON(data []byte) error {
	*m = Ready{}
	dec := json.NewDecoder(bytes.NewReader(bytes.TrimSpace(data)))
	var raw map[string]json.RawMessage
	if err := dec.Decode(&raw); err != nil || raw == nil || dec.More() {
		return fmt.Errorf("%w: json", ErrMalformed)
	}
	return nil
}
