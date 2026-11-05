package v1

import "errors"

// ErrRequest wraps every request validation failure. Callers can use
// errors.Is. Error text names the field and never includes the message body
// or a value.
var ErrRequest = errors.New("indexergateway/contractcreation/v1: invalid request")
