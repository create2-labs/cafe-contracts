package v1

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// Result is the qualification result of one address.
type Result struct {
	Address            string      `json:"address"`
	AccountType        AccountType `json:"account_type"`
	ObservedAt         time.Time   `json:"observed_at"`
	Chains             []Chain     `json:"chains"`
	UnobservedChainIDs []int64     `json:"unobserved_chain_ids"`
	RelatedAddresses   []Related   `json:"related_addresses"`
	Anomalies          []Anomaly   `json:"anomalies"`
	EOA                *EOA        `json:"eoa"`
}

// Related is an address found while qualifying the observed address, with its
// role and the chain where it was found. It is never qualified or scanned.
type Related struct {
	Role    RelatedRole `json:"role"`
	ChainID int64       `json:"chain_id"`
	Address string      `json:"address"`
}

// EOA holds the address-level facts of an address of account type eoa. It is
// nil for any other account type.
type EOA struct {
	KeyExposed bool    `json:"key_exposed"`
	Algorithm  string  `json:"algorithm"`
	NISTLevel  int     `json:"nist_level"`
	RiskScore  float64 `json:"risk_score"`
}

// Chain is the qualification of the address on one observed chain. Nonce is
// nil only on a contract chain or on a delegated chain whose nonce read
// failed. KeyExposed is reported on eoa chains only.
type Chain struct {
	ChainID       int64         `json:"chain_id"`
	Qualification Qualification `json:"qualification"`
	Nonce         *uint64       `json:"nonce,omitempty"`
	KeyExposed    *bool         `json:"key_exposed,omitempty"`
}

// Validate checks the result against the contract: a non-empty address, a
// known account type, a non-zero observation time, a non-empty list of chains
// consistent with the account type, and positive chain identifiers that are
// unique and disjoint between chains and unobserved_chain_ids. A chain is
// delegated when related_addresses holds a delegation_target for it. It also
// checks the address-level facts: the eoa block, the related addresses, and
// the anomalies. Every failure wraps ErrMalformed and names the field path,
// never its value.
func (r Result) Validate() error {
	if r.Address == "" {
		return malformed("address")
	}
	if !r.AccountType.IsValid() {
		return malformed("account_type")
	}
	if r.ObservedAt.IsZero() {
		return malformed("observed_at")
	}
	if len(r.Chains) == 0 {
		return malformed("chains")
	}
	delegated := r.delegatedChains()
	seen := make(map[int64]struct{}, len(r.Chains)+len(r.UnobservedChainIDs))
	hasContract := false
	for _, c := range r.Chains {
		if err := c.validate(r.AccountType, delegated[c.ChainID]); err != nil {
			return err
		}
		if _, dup := seen[c.ChainID]; dup {
			return malformed("chains.chain_id")
		}
		seen[c.ChainID] = struct{}{}
		hasContract = hasContract || c.Qualification == QualificationContract
	}
	if r.AccountType == AccountTypeContract && !hasContract {
		return malformed("chains.qualification")
	}
	for _, id := range r.UnobservedChainIDs {
		if id <= 0 {
			return malformed("unobserved_chain_ids")
		}
		if _, dup := seen[id]; dup {
			return malformed("unobserved_chain_ids")
		}
		seen[id] = struct{}{}
	}
	if err := r.validateEOA(); err != nil {
		return err
	}
	if err := r.validateRelated(); err != nil {
		return err
	}
	return r.validateAnomalies()
}

// delegatedChains returns the identifiers of the chains that carry a
// delegation_target entry. The entries are read without being validated, so
// that a chains.* error precedes a related_addresses.* error.
func (r Result) delegatedChains() map[int64]bool {
	m := make(map[int64]bool, len(r.RelatedAddresses))
	for _, a := range r.RelatedAddresses {
		if a.Role == RelatedRoleDelegationTarget {
			m[a.ChainID] = true
		}
	}
	return m
}

// validateEOA checks the eoa block: required for an eoa result, absent
// otherwise, in range, and consistent with the key_exposed of the chains.
func (r Result) validateEOA() error {
	if r.AccountType != AccountTypeEOA {
		if r.EOA != nil {
			return malformed("eoa")
		}
		return nil
	}
	e := r.EOA
	if e == nil {
		return malformed("eoa")
	}
	if math.IsNaN(e.RiskScore) || e.RiskScore < 0 || e.RiskScore > 1 {
		return malformed("eoa.risk_score")
	}
	if e.NISTLevel < 1 || e.NISTLevel > 5 {
		return malformed("eoa.nist_level")
	}
	// The algorithm is not fixed to one value (crypto-agility).
	if e.Algorithm == "" {
		return malformed("eoa.algorithm")
	}
	exposed := false
	for _, c := range r.Chains {
		exposed = exposed || (c.KeyExposed != nil && *c.KeyExposed)
	}
	if e.KeyExposed != exposed {
		return malformed("eoa.key_exposed")
	}
	return nil
}

// validateRelated checks the related addresses: only for an eoa result, on
// its eoa chains, or for a contract result, on its not_deployed chains; one
// delegation target at most per chain.
func (r Result) validateRelated() error {
	if len(r.RelatedAddresses) == 0 {
		return nil
	}
	var want Qualification
	switch r.AccountType {
	case AccountTypeEOA:
		want = QualificationEOA
	case AccountTypeContract:
		want = QualificationNotDeployed
	default:
		return malformed("related_addresses")
	}
	eligible := make(map[int64]struct{}, len(r.Chains))
	for _, c := range r.Chains {
		if c.Qualification == want {
			eligible[c.ChainID] = struct{}{}
		}
	}
	seen := make(map[int64]struct{}, len(r.RelatedAddresses))
	for _, a := range r.RelatedAddresses {
		if a.Role != RelatedRoleDelegationTarget {
			return malformed("related_addresses.role")
		}
		if _, ok := eligible[a.ChainID]; !ok {
			return malformed("related_addresses.chain_id")
		}
		if a.Address == "" {
			return malformed("related_addresses.address")
		}
		// One entry per chain also excludes a duplicate entry.
		if _, dup := seen[a.ChainID]; dup {
			return malformed("related_addresses.chain_id")
		}
		seen[a.ChainID] = struct{}{}
	}
	return nil
}

// validateAnomalies checks that contract_with_eoa_activity is present exactly
// when a not_deployed chain has a nonce above zero or a delegation_target. An
// unknown nonce is not evidence of activity, but a designator is evidence of
// an EOA, whatever the nonce.
func (r Result) validateAnomalies() error {
	want := false
	if r.AccountType == AccountTypeContract {
		delegated := r.delegatedChains()
		for _, c := range r.Chains {
			want = want || (c.Qualification == QualificationNotDeployed &&
				(delegated[c.ChainID] || (c.Nonce != nil && *c.Nonce > 0)))
		}
	}
	got := false
	for _, a := range r.Anomalies {
		if a != AnomalyContractWithEOAActivity || got {
			return malformed("anomalies")
		}
		got = true
	}
	if got != want {
		return malformed("anomalies")
	}
	return nil
}

// validate checks one chain against the account type of the result. A chain
// without contract code needs its nonce, unless it is delegated: a designator
// chain stays observed when its nonce read fails.
func (c Chain) validate(t AccountType, delegated bool) error {
	if c.ChainID <= 0 {
		return malformed("chains.chain_id")
	}
	if !c.Qualification.IsValid() || !allowed(t, c.Qualification) {
		return malformed("chains.qualification")
	}
	if c.Qualification != QualificationContract && c.Nonce == nil && !delegated {
		return malformed("chains.nonce")
	}
	if c.Qualification == QualificationBlank && c.Nonce != nil && *c.Nonce > 0 {
		return malformed("chains.nonce")
	}
	if c.Qualification != QualificationEOA {
		if c.KeyExposed != nil {
			return malformed("chains.key_exposed")
		}
		return nil
	}
	// A designator exposes the key whatever the nonce; otherwise a nonce above
	// zero means the account has sent a transaction. The nonce is present
	// when the chain is not delegated.
	exposed := delegated || *c.Nonce > 0
	if c.KeyExposed == nil || *c.KeyExposed != exposed {
		return malformed("chains.key_exposed")
	}
	return nil
}

// allowed reports whether a chain qualification is consistent with the
// account type (ADR section 4.1). The contract rule that at least one chain is
// a contract is checked on the whole result.
func allowed(t AccountType, q Qualification) bool {
	switch t {
	case AccountTypeEOA:
		return q == QualificationEOA
	case AccountTypeContract:
		return q == QualificationContract || q == QualificationNotDeployed
	case AccountTypeBlank:
		return q == QualificationBlank
	}
	return false
}

func malformed(field string) error {
	return fmt.Errorf("%w: %s", ErrMalformed, field)
}

// DecodeResult unmarshals and validates a qualification result. Unknown fields
// are ignored so that later additive enrichments stay readable. A malformed
// body yields a zero Result.
func DecodeResult(raw []byte) (Result, error) {
	var r Result
	if err := json.Unmarshal(raw, &r); err != nil {
		// The decoder error can quote body content, so it is not wrapped.
		return Result{}, fmt.Errorf("%w: json", ErrMalformed)
	}
	if len(r.UnobservedChainIDs) == 0 {
		r.UnobservedChainIDs = nil
	}
	if len(r.RelatedAddresses) == 0 {
		r.RelatedAddresses = nil
	}
	if len(r.Anomalies) == 0 {
		r.Anomalies = nil
	}
	if err := r.Validate(); err != nil {
		return Result{}, err
	}
	return r, nil
}
