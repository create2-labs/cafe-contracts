package v01

import "time"

// Envelope is the shared header for CAFE wire events.
type Envelope struct {
	EventID       string    `json:"event_id"`
	EventType     string    `json:"event_type"`
	EventVersion  string    `json:"event_version"`
	OccurredAt    time.Time `json:"occurred_at"`
	CorrelationID string    `json:"correlation_id"`
	CausationID   string    `json:"causation_id"`
	Producer      string    `json:"producer"`
}

// Validate checks minimal shape constraints for envelope v0.1.
func (e *Envelope) Validate() error {
	if e.EventID == "" {
		return ErrEventIDRequired
	}
	if e.EventType == "" {
		return ErrEventTypeRequired
	}
	if e.EventVersion == "" {
		return ErrEventVersionRequired
	}
	if e.OccurredAt.IsZero() {
		return ErrOccurredAtRequired
	}
	if e.Producer == "" {
		return ErrProducerRequired
	}
	return nil
}

// ValidateVersioned enforces minimal validation plus an expected event version.
func (e *Envelope) ValidateVersioned(expectedVersion string) error {
	if err := e.Validate(); err != nil {
		return err
	}
	if expectedVersion != "" && e.EventVersion != expectedVersion {
		return ErrEventVersionMismatch
	}
	return nil
}
