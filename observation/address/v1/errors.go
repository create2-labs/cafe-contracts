package v1

import "errors"

// ErrMalformed is returned when a qualification result fails boundary
// validation. Callers can use errors.Is. The text names the field and not the
// body.
var ErrMalformed = errors.New("observation/address/v1: malformed result")
