package v1

import "errors"

// ErrMalformed is returned when a scan request fails boundary validation.
// Callers can use errors.Is. The text names the field and not the body.
var ErrMalformed = errors.New("discovery/scanrequest/v1: malformed message")
