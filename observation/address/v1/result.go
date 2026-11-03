package v1

import (
	"encoding/json"
	"fmt"
	"time"
)

// Result is the qualification result of one address.
type Result struct {
	Address            string      `json:"address"`
	AccountType        AccountType `json:"account_type"`
	ObservedAt         time.Time   `json:"observed_at"`
	Chains             []Chain     `json:"chains"`
	UnobservedChainIDs []int64     `json:"unobserved_chain_ids"`
}

// Chain is the qualification of the address on one observed chain. Nonce is
// nil when it was not read. KeyExposed is nil when not reported; it is
// reported on eoa chains.
type Chain struct {
	ChainID       int64         `json:"chain_id"`
	Qualification Qualification `json:"qualification"`
	Nonce         *uint64       `json:"nonce,omitempty"`
	KeyExposed    *bool         `json:"key_exposed,omitempty"`
}

// Validate checks the result against the contract: a non-empty address, a
// known account type, a non-zero observation time, a non-empty list of chains
// consistent with the account type, and positive chain identifiers that are
// unique and disjoint between chains and unobserved_chain_ids. Every failure
// wraps ErrMalformed and names the field, never its value.
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
	seen := make(map[int64]struct{}, len(r.Chains)+len(r.UnobservedChainIDs))
	hasContract := false
	for _, c := range r.Chains {
		if err := c.validate(r.AccountType); err != nil {
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
	return nil
}

// validate checks one chain against the account type of the result.
func (c Chain) validate(t AccountType) error {
	if c.ChainID <= 0 {
		return malformed("chains.chain_id")
	}
	if !c.Qualification.IsValid() || !allowed(t, c.Qualification) {
		return malformed("chains.qualification")
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
	// A nonce above zero means the account has sent a transaction, so its
	// public key is exposed. An unknown nonce is not evidence of exposure.
	exposed := c.Nonce != nil && *c.Nonce > 0
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
	if err := r.Validate(); err != nil {
		return Result{}, err
	}
	return r, nil
}
