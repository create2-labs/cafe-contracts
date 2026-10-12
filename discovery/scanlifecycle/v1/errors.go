package v1

import "errors"

// ErrMalformed is a boundary failure. errors.Is works. The text names the field, not the body.
var ErrMalformed = errors.New("discovery/scanlifecycle/v1: malformed message")
