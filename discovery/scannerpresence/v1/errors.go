package v1

import "errors"

// ErrMalformed is returned when a presence message fails boundary validation.
// Callers can use errors.Is. The text names the field and not the body.
var ErrMalformed = errors.New("discovery/scannerpresence/v1: malformed message")
