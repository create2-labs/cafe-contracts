package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/create2-labs/cafe-contracts/address"
)

// Request is one page read on cafe.indexer.gateway.transactions.page.v1.
// Cursor is opaque: validation checks its size only.
type Request struct {
	RequestID string `json:"request_id"`
	Address   string `json:"address"`
	ChainID   int64  `json:"chain_id"`
	Cursor    string `json:"cursor"`
	Limit     int    `json:"limit"`
	Deadline  string `json:"deadline"`
}

// DecodeRequest unmarshals and validates a request.
// On any failure it returns a zero Request, so a rejected body cannot be logged
// from the result. Use BoundedRequestID to choose the request_id of an
// invalid_request response.
func DecodeRequest(raw []byte) (Request, error) {
	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		if errors.Is(err, ErrRequest) {
			return Request{}, err
		}
		return Request{}, fmt.Errorf("%w: json", ErrRequest)
	}
	if err := req.Validate(); err != nil {
		return Request{}, err
	}
	return req, nil
}

// UnmarshalJSON rejects a non-object body and any field that is not the JSON
// type the contract requires. A missing cursor is the empty string. Null is
// not a string. chain_id and limit must be JSON integers, not fractions.
func (r *Request) UnmarshalJSON(data []byte) error {
	*r = Request{}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("%w: json", ErrRequest)
	}
	var err error
	if r.RequestID, err = requiredStringField(raw, "request_id"); err != nil {
		return err
	}
	if r.Address, err = requiredStringField(raw, "address"); err != nil {
		return err
	}
	if r.ChainID, err = requiredIntField(raw, "chain_id"); err != nil {
		return err
	}
	if r.Cursor, err = optionalStringField(raw, "cursor"); err != nil {
		return err
	}
	limit, err := requiredIntField(raw, "limit")
	if err != nil {
		return err
	}
	r.Limit = int(limit)
	if r.Deadline, err = requiredStringField(raw, "deadline"); err != nil {
		return err
	}
	return nil
}

func requiredStringField(raw map[string]json.RawMessage, key string) (string, error) {
	value, ok := raw[key]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrRequest, key)
	}
	s, err := jsonString(value)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrRequest, key)
	}
	return s, nil
}

func optionalStringField(raw map[string]json.RawMessage, key string) (string, error) {
	value, ok := raw[key]
	if !ok {
		return "", nil
	}
	s, err := jsonString(value)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrRequest, key)
	}
	return s, nil
}

func requiredIntField(raw map[string]json.RawMessage, key string) (int64, error) {
	value, ok := raw[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrRequest, key)
	}
	token := bytes.TrimSpace(value)
	if !isJSONInteger(token) {
		return 0, fmt.Errorf("%w: %s", ErrRequest, key)
	}
	var n int64
	if err := json.Unmarshal(token, &n); err != nil {
		return 0, fmt.Errorf("%w: %s", ErrRequest, key)
	}
	return n, nil
}

func isJSONInteger(token []byte) bool {
	if len(token) == 0 {
		return false
	}
	i := 0
	if token[0] == '-' {
		if len(token) == 1 {
			return false
		}
		i = 1
	}
	if token[i] == '0' {
		return len(token) == i+1
	}
	for ; i < len(token); i++ {
		if token[i] < '0' || token[i] > '9' {
			return false
		}
	}
	return true
}

// Validate checks a structurally valid page request.
// request_id must be a canonical UUID. address must be an EVM address.
// chain_id must be a strictly positive integer. cursor must be at most
// MaxCursorBytes and is not interpreted. limit must be from 1 to 100.
// deadline must be RFC3339Nano.
func (r Request) Validate() error {
	if !isUUID(r.RequestID) {
		return fmt.Errorf("%w: request_id", ErrRequest)
	}
	if !address.IsValidHexAddress(r.Address) {
		return fmt.Errorf("%w: address", ErrRequest)
	}
	if r.ChainID <= 0 {
		return fmt.Errorf("%w: chain_id", ErrRequest)
	}
	if len(r.Cursor) > MaxCursorBytes {
		return fmt.Errorf("%w: cursor", ErrRequest)
	}
	if r.Limit < MinLimit || r.Limit > MaxLimit {
		return fmt.Errorf("%w: limit", ErrRequest)
	}
	if _, err := time.Parse(time.RFC3339Nano, r.Deadline); err != nil {
		return fmt.Errorf("%w: deadline", ErrRequest)
	}
	return nil
}

// BoundedRequestID is the request_id copied into an invalid_request response.
// A JSON string of 1 to MaxRequestIDBytes is returned unchanged, including a
// string that is not a UUID. An empty string, a longer string, a missing
// field, a non-string, or unreadable JSON yields "".
// The original string beyond MaxRequestIDBytes is never returned.
func BoundedRequestID(raw []byte) string {
	var probe struct {
		RequestID json.RawMessage `json:"request_id"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return ""
	}
	if len(bytes.TrimSpace(probe.RequestID)) == 0 {
		return ""
	}
	if bytes.Equal(bytes.TrimSpace(probe.RequestID), []byte("null")) {
		return ""
	}
	var id string
	if err := json.Unmarshal(probe.RequestID, &id); err != nil {
		return ""
	}
	if len(id) == 0 || len(id) > MaxRequestIDBytes {
		return ""
	}
	return id
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
		if !isHexByte(c) {
			return false
		}
	}
	return true
}

func isHexByte(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}
