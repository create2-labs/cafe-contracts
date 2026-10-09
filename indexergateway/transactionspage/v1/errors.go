package v1

import "errors"

// Sentinel errors for boundary validation. Callers can use errors.Is.
// Error text names the field and never includes the message body, a cursor,
// a URL, or a secret.
var (
	ErrRequest  = errors.New("indexergateway/transactionspage/v1: invalid request")
	ErrResponse = errors.New("indexergateway/transactionspage/v1: invalid response")
)
