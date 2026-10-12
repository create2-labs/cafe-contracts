package v1

import "errors"

// ErrMalformed is a boundary failure. errors.Is works. The text does not include the body.
var ErrMalformed = errors.New("discovery/persistencesignal/v1: malformed message")
