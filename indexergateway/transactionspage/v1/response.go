package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// Transaction is one history row. Hash and From keep the order of the array.
// This package does not sort the page and does not judge its order.
type Transaction struct {
	Hash string `json:"hash"`
	From string `json:"from"`
}

// Error is the error object of a response envelope.
// Code is one of the six v1 codes. Message is a generic string of 1 to
// MaxErrorMessageBytes and does not name a provider.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Response is the discriminated page envelope.
// A success has RequestID, Transactions, and NextCursor, and no Error.
// An error has RequestID and Error, and neither Transactions nor NextCursor.
// Transactions is [] for an empty page: never null and never omitted.
// NextCursor is always present on success and may be empty.
type Response struct {
	RequestID    string
	Transactions []Transaction
	NextCursor   string
	Error        *Error
}

// Validate checks an in-memory envelope.
// request_id may be any string of at most MaxRequestIDBytes. It is not
// required to be a UUID. The empty string is valid only when error.code is
// invalid_request. A success page or any other code rejects it.
func (r Response) Validate() error {
	if r.Error != nil {
		if len(r.Transactions) > 0 || r.NextCursor != "" {
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
	if len(r.Transactions) > MaxTransactions {
		return fmt.Errorf("%w: transactions", ErrResponse)
	}
	if len(r.NextCursor) > MaxCursorBytes {
		return fmt.Errorf("%w: next_cursor", ErrResponse)
	}
	return validateResponseID(r.RequestID, "")
}

// ValidateForRequest checks the envelope and requires request_id to equal
// the identifier the caller sent. The reader uses this with its own UUID, so
// an empty request_id never passes that check.
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

// MarshalJSON writes either a page or an error object.
// A page always includes transactions, using [] when the slice is nil, and
// always includes next_cursor. An error object includes neither.
// Array order is the order of Transactions.
func (r Response) MarshalJSON() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	if r.Error != nil {
		payload := struct {
			RequestID string `json:"request_id"`
			Error     *Error `json:"error"`
		}{
			RequestID: r.RequestID,
			Error:     r.Error,
		}
		return json.Marshal(payload)
	}
	txs := r.Transactions
	if txs == nil {
		txs = []Transaction{}
	}
	payload := struct {
		RequestID    string        `json:"request_id"`
		Transactions []Transaction `json:"transactions"`
		NextCursor   string        `json:"next_cursor"`
	}{
		RequestID:    r.RequestID,
		Transactions: txs,
		NextCursor:   r.NextCursor,
	}
	return json.Marshal(payload)
}

// DecodeResponse unmarshals and validates a response envelope.
// Unreadable JSON is an invalid response. On failure the result is zero.
func DecodeResponse(raw []byte) (Response, error) {
	var resp Response
	if err := json.Unmarshal(raw, &resp); err != nil {
		if errors.Is(err, ErrResponse) {
			return Response{}, err
		}
		return Response{}, fmt.Errorf("%w: json", ErrResponse)
	}
	return resp, nil
}

// UnmarshalJSON rejects envelopes that mix a page and an error, omit page
// fields, set transactions to null, exceed the page bounds, or break the
// request_id rules.
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
	if err := json.Unmarshal(data, &raw); err != nil {
		return Response{}, fmt.Errorf("%w: json", ErrResponse)
	}
	id, err := optionalString(raw, "request_id")
	if err != nil {
		return Response{}, fmt.Errorf("%w: request_id", ErrResponse)
	}
	_, hasTx := raw["transactions"]
	_, hasCursor := raw["next_cursor"]
	errRaw, hasErr := raw["error"]
	if hasErr {
		if hasTx || hasCursor {
			return Response{}, fmt.Errorf("%w: mixed", ErrResponse)
		}
		body, err := decodeError(errRaw)
		if err != nil {
			return Response{}, err
		}
		resp := Response{RequestID: id, Error: &body}
		if err := resp.Validate(); err != nil {
			return Response{}, err
		}
		return resp, nil
	}
	if !hasTx || !hasCursor {
		return Response{}, fmt.Errorf("%w: page", ErrResponse)
	}
	txs, err := decodeTransactions(raw["transactions"])
	if err != nil {
		return Response{}, err
	}
	cursor, err := jsonString(raw["next_cursor"])
	if err != nil {
		return Response{}, fmt.Errorf("%w: next_cursor", ErrResponse)
	}
	resp := Response{RequestID: id, Transactions: txs, NextCursor: cursor}
	if err := resp.Validate(); err != nil {
		return Response{}, err
	}
	return resp, nil
}

func decodeError(raw json.RawMessage) (Error, error) {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return Error{}, fmt.Errorf("%w: error", ErrResponse)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
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

func decodeTransactions(raw json.RawMessage) ([]Transaction, error) {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, fmt.Errorf("%w: transactions", ErrResponse)
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return nil, fmt.Errorf("%w: transactions", ErrResponse)
	}
	txs := []Transaction{}
	if err := json.Unmarshal(trimmed, &txs); err != nil {
		return nil, fmt.Errorf("%w: transactions", ErrResponse)
	}
	if txs == nil {
		txs = []Transaction{}
	}
	return txs, nil
}

func optionalString(raw map[string]json.RawMessage, key string) (string, error) {
	value, ok := raw[key]
	if !ok {
		return "", nil
	}
	return jsonString(value)
}

func jsonString(raw json.RawMessage) (string, error) {
	if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return "", errNotString
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", errNotString
	}
	return s, nil
}

var errNotString = errors.New("indexergateway/transactionspage/v1: json string required")
