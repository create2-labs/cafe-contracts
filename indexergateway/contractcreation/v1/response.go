package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/create2-labs/cafe-contracts/address"
)

// Error is the error object of a response envelope.
// Code is one of the seven v1 codes. Message is a generic string of 1 to
// MaxErrorMessageBytes and does not name a provider.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Response is the discriminated creation envelope.
// A found response has Creator and optionally CreationTxHash. A not found
// response has none of Creator, CreationTxHash, and Error. An error has
// Error and neither Creator nor CreationTxHash.
type Response struct {
	RequestID      string `json:"request_id"`
	Creator        string `json:"creator,omitempty"`
	CreationTxHash string `json:"creation_tx_hash,omitempty"`
	Error          *Error `json:"error,omitempty"`
}

// Validate checks an in-memory envelope.
// request_id may be any string of at most MaxRequestIDBytes. It is not
// required to be a UUID. The empty string is valid only when error.code is
// invalid_request. Creator needs the 0x or 0X prefix.
func (r Response) Validate() error {
	if r.Error != nil {
		if r.Creator != "" || r.CreationTxHash != "" {
			return fmt.Errorf("%w: mixed", ErrResponse)
		}
		if !ValidCode(r.Error.Code) {
			return fmt.Errorf("%w: error.code", ErrResponse)
		}
		n := len(r.Error.Message)
		if n < 1 || n > MaxErrorMessageBytes {
			return fmt.Errorf("%w: error.message", ErrResponse)
		}
		return validateResponseID(r.RequestID, r.Error.Code)
	}
	if r.Creator == "" {
		if r.CreationTxHash != "" {
			return fmt.Errorf("%w: creation_tx_hash", ErrResponse)
		}
	} else if !hasHexPrefix(r.Creator) || !address.IsValidHexAddress(r.Creator) {
		return fmt.Errorf("%w: creator", ErrResponse)
	}
	return validateResponseID(r.RequestID, "")
}

// ValidateForRequest checks the envelope, then that its request_id echoes
// requestID. The reader uses its own UUID, so an empty request_id never passes.
func (r Response) ValidateForRequest(requestID string) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if r.RequestID != requestID {
		return fmt.Errorf("%w: request_id", ErrResponse)
	}
	return nil
}

func validateResponseID(id, code string) error {
	if len(id) > MaxRequestIDBytes {
		return fmt.Errorf("%w: request_id", ErrResponse)
	}
	if id == "" && code != CodeInvalidRequest {
		return fmt.Errorf("%w: request_id", ErrResponse)
	}
	return nil
}

// MarshalJSON validates the envelope, then writes it. request_id is always
// written. A found response adds creator and, when non-empty,
// creation_tx_hash. A not found response has request_id only. An error adds
// error only.
func (r Response) MarshalJSON() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	type wire Response
	return json.Marshal(wire(r))
}

// DecodeResponse unmarshals and validates a response envelope.
// Unreadable JSON is an invalid response. On failure the result is zero.
func DecodeResponse(raw []byte) (Response, error) {
	return decodeResponse(raw)
}

// UnmarshalJSON decodes and validates a response envelope. Unknown fields are
// ignored. creator is delivered in lowercase. The receiver is left unchanged
// on failure.
func (r *Response) UnmarshalJSON(data []byte) error {
	decoded, err := decodeResponse(data)
	if err != nil {
		return err
	}
	*r = decoded
	return nil
}

func decodeResponse(data []byte) (Response, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil || raw == nil {
		return Response{}, fmt.Errorf("%w: json", ErrResponse)
	}
	var resp Response
	var err error
	if resp.RequestID, err = optionalString(raw, "request_id"); err != nil {
		return Response{}, fmt.Errorf("%w: request_id", ErrResponse)
	}
	if resp.Creator, err = optionalNonEmptyString(raw, "creator"); err != nil {
		return Response{}, err
	}
	if resp.CreationTxHash, err = optionalNonEmptyString(raw, "creation_tx_hash"); err != nil {
		return Response{}, err
	}
	if errRaw, ok := raw["error"]; ok {
		body, err := decodeError(errRaw)
		if err != nil {
			return Response{}, err
		}
		resp.Error = &body
	}
	if err := resp.Validate(); err != nil {
		return Response{}, err
	}
	resp.Creator = strings.ToLower(resp.Creator)
	return resp, nil
}

func decodeError(raw json.RawMessage) (Error, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return Error{}, fmt.Errorf("%w: error", ErrResponse)
	}
	codeRaw, ok := fields["code"]
	if !ok {
		return Error{}, fmt.Errorf("%w: error.code", ErrResponse)
	}
	code, err := jsonString(codeRaw)
	if err != nil || !ValidCode(code) {
		return Error{}, fmt.Errorf("%w: error.code", ErrResponse)
	}
	msgRaw, ok := fields["message"]
	if !ok {
		return Error{}, fmt.Errorf("%w: error.message", ErrResponse)
	}
	msg, err := jsonString(msgRaw)
	if err != nil {
		return Error{}, fmt.Errorf("%w: error.message", ErrResponse)
	}
	return Error{Code: code, Message: msg}, nil
}

// optionalNonEmptyString reads a field that may be absent but, when present,
// must be a non-empty JSON string.
func optionalNonEmptyString(raw map[string]json.RawMessage, key string) (string, error) {
	s, err := optionalString(raw, key)
	_, present := raw[key]
	if err != nil || (present && s == "") {
		return "", fmt.Errorf("%w: %s", ErrResponse, key)
	}
	return s, nil
}

func optionalString(raw map[string]json.RawMessage, key string) (string, error) {
	value, ok := raw[key]
	if !ok {
		return "", nil
	}
	return jsonString(value)
}

func jsonString(raw json.RawMessage) (string, error) {
	var s string
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &s) != nil {
		return "", errNotString
	}
	return s, nil
}

var errNotString = errors.New("indexergateway/contractcreation/v1: json string required")
