package v01

import "time"

// Event is the normalized cafe.discovery.wallet.observed envelope (contract v0.1).
// It is the canonical shape CPM consumes from Discovery over NATS or APIs.
type Event struct {
	EventID       string    `json:"event_id"`
	EventType     string    `json:"event_type"`
	EventVersion  string    `json:"event_version"`
	OccurredAt    time.Time `json:"occurred_at"`
	CorrelationID string    `json:"correlation_id"`
	CausationID   string    `json:"causation_id"`
	Producer      string    `json:"producer"`
	Subject       Subject   `json:"subject"`
	Payload       Payload   `json:"payload"`
}

// Subject identifies the wallet (or future subject types) for this observation.
type Subject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// Payload holds observation fields exported to CPM.
type Payload struct {
	ChainIDs         []int64   `json:"chain_ids"`
	AccountKind      string    `json:"account_kind"`
	CurrentAlgorithm string    `json:"current_algorithm"`
	PublicKeyExposed bool      `json:"public_key_exposed"`
	IsMultichain     bool      `json:"is_multichain"`
	ObservedAt       time.Time `json:"observed_at"`
	CurrentPQPosture string    `json:"current_pq_posture"`
}

// Validate checks normative vocabulary and envelope fields for a v0.1 event.
// It does not enforce CPM business policy rules beyond exported vocabulary.
func (e *Event) Validate() error {
	if e.EventID == "" {
		return ErrEventIDRequired
	}
	if e.EventType != EventTypeWalletObserved {
		return ErrEventType
	}
	if e.EventVersion != EventVersion {
		return ErrEventVersion
	}
	if e.Producer != ProducerCafeDiscovery {
		return ErrProducer
	}
	st := SubjectType(e.Subject.Type)
	if !st.IsValid() {
		return ErrSubjectType
	}
	if e.Subject.ID == "" {
		return ErrSubjectIDRequired
	}
	if !AccountKind(e.Payload.AccountKind).IsValid() {
		return ErrAccountKind
	}
	if !IsValidAlgorithmID(e.Payload.CurrentAlgorithm) {
		return ErrAlgorithmID
	}
	if !CurrentPQPosture(e.Payload.CurrentPQPosture).IsValid() {
		return ErrCurrentPQPosture
	}
	return nil
}
