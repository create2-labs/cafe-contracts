package v01

import "errors"

// Validation errors for event envelope v0.1.
const errorPrefix = "eventenvelope " + EventVersionV01 + ": "

var (
	ErrEventIDRequired      = errors.New(errorPrefix + "event_id is required")
	ErrEventTypeRequired    = errors.New(errorPrefix + "event_type is required")
	ErrEventVersionRequired = errors.New(errorPrefix + "event_version is required")
	ErrEventVersionMismatch = errors.New(errorPrefix + "event_version mismatch")
	ErrOccurredAtRequired   = errors.New(errorPrefix + "occurred_at is required")
	ErrProducerRequired     = errors.New(errorPrefix + "producer is required")
)
