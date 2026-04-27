package cafenatsv01

import (
	dwv01 "github.com/create2-labs/cafe-contracts/discoverywalletobserved/v01"
)

// PolicyAssessmentRequested is the explicit asynchronous command that starts CPM assessment.
// It carries a self-sufficient observation snapshot (full discovery.wallet.observed event)
// so CPM does not need to read Discovery persistence. event_id is the primary idempotence key.
type PolicyAssessmentRequested struct {
	EnvelopeV01
	Subject SubjectRef                       `json:"subject"`
	Payload PolicyAssessmentRequestedPayload `json:"payload"`
}

// PolicyAssessmentRequestedPayload is the command body.
type PolicyAssessmentRequestedPayload struct {
	// Observation is a full discovery.wallet.observed v0.1 event (contract in discoverywalletobserved/v01).
	Observation dwv01.Event `json:"observation"`
	// SelectionRequest drives policy selection; same JSON as CPM PolicySelectionRequest.
	SelectionRequest PolicySelectionRequestWire `json:"selection_request"`
	// ClientRequestID is optional; correlates to a user or API request id (not the idempotence key).
	ClientRequestID string `json:"client_request_id,omitempty"`
}

var policyAssessmentRequestProducers = map[string]struct{}{
	ProducerCafeCryptoBackend: {},
	ProducerCafeDiscovery:     {},
	ProducerCafeEdge:          {},
}

// Validate checks envelope, producer, subject, nested observation, and selection_request.
func (e *PolicyAssessmentRequested) Validate() error {
	if e == nil {
		return ErrEventIDRequired
	}
	if err := requireEnvelopeV01(e.EnvelopeV01, EventTypePolicyAssessmentRequested); err != nil {
		return err
	}
	if _, ok := policyAssessmentRequestProducers[e.Producer]; !ok {
		return ErrProducer
	}
	if err := requireSubjectWallet(e.Subject); err != nil {
		return err
	}
	if err := e.Payload.Observation.Validate(); err != nil {
		return err
	}
	e.Payload.SelectionRequest.Normalize()
	if err := e.Payload.SelectionRequest.Validate(); err != nil {
		return err
	}
	return nil
}
