package v1

import (
	"fmt"

	enrichmentrequestv1 "github.com/create2-labs/cafe-contracts/discovery/enrichmentrequest/v1"
	addressv1 "github.com/create2-labs/cafe-contracts/observation/address/v1"
)

// Outcome is the outcome of an enrichment on one chain. A chain is answered
// when it is observed or not_observed.
type Outcome string

// Per-chain outcomes.
const (
	OutcomeObserved    Outcome = "observed"
	OutcomeNotObserved Outcome = "not_observed"
	OutcomeError       Outcome = "error"
)

// IsValid reports whether o is a known per-chain outcome.
func (o Outcome) IsValid() bool {
	switch o {
	case OutcomeObserved, OutcomeNotObserved, OutcomeError:
		return true
	}
	return false
}

// PublicKeyRecovery is the address-level outcome of the ECDSA public-key
// recovery of a history enrichment, with the values wallet scans present today.
type PublicKeyRecovery string

// Public-key recovery outcomes.
const (
	PublicKeyRecoveryNotRequired PublicKeyRecovery = "not_required"
	PublicKeyRecoveryRecovered   PublicKeyRecovery = "recovered"
	PublicKeyRecoveryUnresolved  PublicKeyRecovery = "unresolved"
)

// IsValid reports whether r is a known public-key recovery outcome.
func (r PublicKeyRecovery) IsValid() bool {
	switch r {
	case PublicKeyRecoveryNotRequired, PublicKeyRecoveryRecovered, PublicKeyRecoveryUnresolved:
		return true
	}
	return false
}

// ChainOutcome is the outcome of an enrichment on one chain.
type ChainOutcome struct {
	ChainID int64   `json:"chain_id"`
	Outcome Outcome `json:"outcome"`
	// Code is the stable code of an error outcome.
	Code string `json:"code,omitempty"`
	// TransactionHash is the recovering transaction of a history chain.
	TransactionHash string `json:"transaction_hash,omitempty"`
	// Deployer and CreationTxHash are the facts of an observed deployer chain.
	Deployer       string `json:"deployer,omitempty"`
	CreationTxHash string `json:"creation_tx_hash,omitempty"`
}

// Completed is one enrichment.completed message.
type Completed struct {
	EnrichmentID     string                     `json:"enrichment_id"`
	ScanID           string                     `json:"scan_id"`
	UserID           string                     `json:"user_id"`
	Family           enrichmentrequestv1.Family `json:"family"`
	Chains           []ChainOutcome             `json:"chains"`
	RelatedAddresses []addressv1.Related        `json:"related_addresses"`
	// Address-level recovery fields of a history enrichment.
	PublicKey         string            `json:"public_key,omitempty"`
	TransactionHash   string            `json:"transaction_hash,omitempty"`
	ExposedNetwork    string            `json:"exposed_network,omitempty"`
	PublicKeyRecovery PublicKeyRecovery `json:"public_key_recovery,omitempty"`
	CompletedAt       string            `json:"completed_at"`
}

// Validate checks the message against the contract. Every failure wraps
// ErrMalformed and names the field, never its value.
func (m Completed) Validate() error {
	if err := validateCommon(m.EnrichmentID, m.ScanID, m.UserID, m.Family); err != nil {
		return err
	}
	if err := m.validateChains(); err != nil {
		return err
	}
	if err := m.validateRelated(); err != nil {
		return err
	}
	if err := m.validateRecovery(); err != nil {
		return err
	}
	if m.CompletedAt == "" {
		return malformed("completed_at")
	}
	return nil
}

// validateChains checks the identifiers, outcomes, codes, and the fields each
// outcome allows for the family.
func (m Completed) validateChains() error {
	if len(m.Chains) == 0 {
		return malformed("chains")
	}
	seen := make(map[int64]struct{}, len(m.Chains))
	for i, c := range m.Chains {
		path := func(f string) error { return malformed(fmt.Sprintf("chains[%d].%s", i, f)) }
		if c.ChainID <= 0 {
			return path("chain_id")
		}
		if _, dup := seen[c.ChainID]; dup {
			return path("chain_id")
		}
		seen[c.ChainID] = struct{}{}
		if !c.Outcome.IsValid() {
			return path("outcome")
		}
		if (c.Outcome == OutcomeError) != (c.Code != "") {
			return path("code")
		}
		if f := m.chainField(c); f != "" {
			return path(f)
		}
	}
	return nil
}

// chainField returns the name of the first field of c that the family and
// the outcome require or forbid wrongly, or "" when c is consistent.
func (m Completed) chainField(c ChainOutcome) string {
	observed := c.Outcome == OutcomeObserved
	if m.Family == enrichmentrequestv1.FamilyHistory {
		switch {
		case observed != (c.TransactionHash != ""):
			return "transaction_hash"
		case c.Deployer != "":
			return "deployer"
		case c.CreationTxHash != "":
			return "creation_tx_hash"
		}
		return ""
	}
	switch {
	case c.TransactionHash != "":
		return "transaction_hash"
	case observed != (c.Deployer != ""):
		return "deployer"
	case !observed && c.CreationTxHash != "":
		return "creation_tx_hash"
	}
	return ""
}

// validateRelated checks related_addresses: empty for history, exactly the
// deployer of each observed chain for deployer.
func (m Completed) validateRelated() error {
	if m.Family == enrichmentrequestv1.FamilyHistory {
		if len(m.RelatedAddresses) != 0 {
			return malformed("related_addresses")
		}
		return nil
	}
	deployers := make(map[int64]string, len(m.Chains))
	for _, c := range m.Chains {
		if c.Outcome == OutcomeObserved {
			deployers[c.ChainID] = c.Deployer
		}
	}
	if len(m.RelatedAddresses) != len(deployers) {
		return malformed("related_addresses")
	}
	seen := make(map[int64]struct{}, len(deployers))
	for i, r := range m.RelatedAddresses {
		want, ok := deployers[r.ChainID]
		_, dup := seen[r.ChainID]
		if r.Role != addressv1.RelatedRoleDeployer || !ok || dup || r.Address == "" || r.Address != want {
			return malformed(fmt.Sprintf("related_addresses[%d]", i))
		}
		seen[r.ChainID] = struct{}{}
	}
	return nil
}

// validateRecovery checks the address-level recovery fields: required for
// history, forbidden for deployer.
func (m Completed) validateRecovery() error {
	if m.Family != enrichmentrequestv1.FamilyHistory {
		if m.PublicKeyRecovery != "" {
			return malformed("public_key_recovery")
		}
		return m.forbidRecoveryFacts()
	}
	switch m.PublicKeyRecovery {
	case PublicKeyRecoveryUnresolved:
		return m.forbidRecoveryFacts()
	case PublicKeyRecoveryRecovered:
		if m.PublicKey == "" {
			return malformed("public_key")
		}
		if m.ExposedNetwork == "" {
			return malformed("exposed_network")
		}
		for _, c := range m.Chains {
			if c.Outcome == OutcomeObserved && c.TransactionHash == m.TransactionHash && m.TransactionHash != "" {
				return nil
			}
		}
		return malformed("transaction_hash")
	default:
		return malformed("public_key_recovery")
	}
}

// forbidRecoveryFacts rejects the first address-level recovery fact that is
// set, naming its field.
func (m Completed) forbidRecoveryFacts() error {
	switch {
	case m.PublicKey != "":
		return malformed("public_key")
	case m.TransactionHash != "":
		return malformed("transaction_hash")
	case m.ExposedNetwork != "":
		return malformed("exposed_network")
	}
	return nil
}

// DecodeCompleted unmarshals and validates enrichment.completed. Identifiers
// are accepted in any case and canonicalized to lowercase. Unknown fields are
// ignored. A malformed body yields a zero Completed.
func DecodeCompleted(raw []byte) (Completed, error) {
	var m Completed
	if err := unmarshal(raw, &m); err != nil {
		return Completed{}, err
	}
	m.EnrichmentID = lowerIfUUID(m.EnrichmentID)
	m.ScanID = lowerIfUUID(m.ScanID)
	m.UserID = lowerIfUUID(m.UserID)
	if err := m.Validate(); err != nil {
		return Completed{}, err
	}
	return m, nil
}
