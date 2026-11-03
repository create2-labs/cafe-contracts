package v1

import "errors"

// ErrMalformed is returned when an enrichment lifecycle message fails boundary
// validation. Callers can use errors.Is. The text names the field and not the
// body.
var ErrMalformed = errors.New("discovery/enrichmentlifecycle/v1: malformed message")
