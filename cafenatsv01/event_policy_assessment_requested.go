package cafenatsv01

import (
	"encoding/json"
	"fmt"
	"strings"

	walletobsv01 "github.com/create2-labs/cafe-contracts/observation/wallet/v01"
)

// PolicyAssessmentRequested is the explicit asynchronous command that starts CPM assessment.
// It carries a self-sufficient observation snapshot (full cafe.discovery.wallet.observed event)
// so CPM does not need to read Discovery persistence. event_id is the primary idempotence key.
//
// Assessment payload wire v0.2 (ADR amendement / CPM-P9a-contracts): crypto_policy_id + scan
// context (observation). selection_request and layer-B user constraints are rejected.
type PolicyAssessmentRequested struct {
	EnvelopeV01
	Subject SubjectRef                       `json:"subject"`
	Payload PolicyAssessmentRequestedPayload `json:"payload"`
}

// PolicyAssessmentRequestedPayload is the assessment command body (wire v0.2).
type PolicyAssessmentRequestedPayload struct {
	// CryptoPolicyID is the catalogue Crypto Policy id (stable intention; required_posture lives on the CP).
	CryptoPolicyID string `json:"crypto_policy_id"`
	// Observation is a full cafe.discovery.wallet.observed v0.1 event (scan context).
	Observation walletobsv01.Event `json:"observation"`
	// ClientRequestID is optional; correlates to a user or API request id (not the idempotence key).
	ClientRequestID string `json:"client_request_id,omitempty"`
}

// Legacy / couche-B keys that must not appear on assessment payload v0.2.
var assessmentPayloadForbiddenKeys = map[string]struct{}{
	"selection_request":           {},
	"allow_new_wallet":            {},
	"address_continuity_required": {},
	"key_rotation_model":          {},
	"target_posture":              {},
	"target_chain_ids":            {},
	"require_multichain":          {},
	"recovery_required":           {},
	"minimum_maturity":            {},
	"allow_research":              {},
	"allowed_provider_modes":      {},
	"preferred_families":          {},
	"preferred_providers":         {},
	"require_bundler_available":   {},
	"require_paymaster_available": {},
	"approval_mode":               {},
}

var policyAssessmentRequestProducers = map[string]struct{}{
	ProducerCafeCryptoBackend: {},
	ProducerCafeDiscovery:     {},
	ProducerCafeEdge:          {},
}

// UnmarshalJSON rejects legacy selection_request / layer-B fields, then decodes v0.2 fields.
func (p *PolicyAssessmentRequestedPayload) UnmarshalJSON(data []byte) error {
	if p == nil {
		return ErrPayloadInvalid
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for key := range raw {
		if _, forbidden := assessmentPayloadForbiddenKeys[key]; forbidden {
			return fmt.Errorf("%w: %s", ErrLegacyAssessmentField, key)
		}
	}
	for key := range raw {
		switch key {
		case "crypto_policy_id", "observation", "client_request_id":
		default:
			return fmt.Errorf("%w: unknown field %s", ErrPayloadInvalid, key)
		}
	}

	var decoded struct {
		CryptoPolicyID  string             `json:"crypto_policy_id"`
		Observation     walletobsv01.Event `json:"observation"`
		ClientRequestID string             `json:"client_request_id,omitempty"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*p = PolicyAssessmentRequestedPayload{
		CryptoPolicyID:  decoded.CryptoPolicyID,
		Observation:     decoded.Observation,
		ClientRequestID: decoded.ClientRequestID,
	}
	return nil
}

// Validate checks envelope, producer, subject, crypto_policy_id, and nested observation.
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
	if strings.TrimSpace(e.Payload.CryptoPolicyID) == "" {
		return ErrCryptoPolicyIDRequired
	}
	if err := e.Payload.Observation.Validate(); err != nil {
		return err
	}
	return nil
}
