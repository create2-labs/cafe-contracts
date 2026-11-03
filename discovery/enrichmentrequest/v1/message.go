package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Request is one enrichment request.
type Request struct {
	EnrichmentID string  `json:"enrichment_id"`
	ScanID       string  `json:"scan_id"`
	UserID       string  `json:"user_id"`
	Address      string  `json:"address"`
	Family       Family  `json:"family"`
	Chains       []Chain `json:"chains"`
}

// Chain is one chain of a request. Nonce is required for the history family.
type Chain struct {
	ChainID int64   `json:"chain_id"`
	Nonce   *uint64 `json:"nonce,omitempty"`
}

// Validate checks the request against the contract. Every failure wraps
// ErrMalformed and names the field path, never its value.
func (r Request) Validate() error {
	for _, f := range []struct{ name, id string }{
		{"enrichment_id", r.EnrichmentID},
		{"scan_id", r.ScanID},
		{"user_id", r.UserID},
	} {
		if err := validateID(f.name, f.id); err != nil {
			return err
		}
	}
	if r.Address == "" {
		return malformed("address")
	}
	if !r.Family.IsValid() {
		return malformed("family")
	}
	if len(r.Chains) == 0 {
		return malformed("chains")
	}
	seen := make(map[int64]struct{}, len(r.Chains))
	for i, c := range r.Chains {
		if c.ChainID <= 0 {
			return malformed(fmt.Sprintf("chains[%d].chain_id", i))
		}
		if _, dup := seen[c.ChainID]; dup {
			return malformed(fmt.Sprintf("chains[%d].chain_id", i))
		}
		seen[c.ChainID] = struct{}{}
		switch r.Family {
		case FamilyHistory:
			if c.Nonce == nil || *c.Nonce == 0 {
				return malformed(fmt.Sprintf("chains[%d].nonce", i))
			}
		case FamilyDeployer:
			if c.Nonce != nil {
				return malformed(fmt.Sprintf("chains[%d].nonce", i))
			}
		}
	}
	return nil
}

// DecodeRequest unmarshals and validates an enrichment request. Identifiers
// are accepted in any case and canonicalized to lowercase. Unknown fields are
// ignored. A malformed body yields a zero Request.
func DecodeRequest(raw []byte) (Request, error) {
	var r Request
	if err := json.Unmarshal(raw, &r); err != nil {
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) && typeErr.Field != "" {
			return Request{}, malformed(typeErr.Field)
		}
		return Request{}, malformed("body")
	}
	r.EnrichmentID = lowerIfUUID(r.EnrichmentID)
	r.ScanID = lowerIfUUID(r.ScanID)
	r.UserID = lowerIfUUID(r.UserID)
	if err := r.Validate(); err != nil {
		return Request{}, err
	}
	return r, nil
}

func malformed(field string) error {
	return fmt.Errorf("%w: %s", ErrMalformed, field)
}

const nilUUID = "00000000-0000-0000-0000-000000000000"

// validateID requires the canonical lowercase form and rejects the nil UUID.
func validateID(field, s string) error {
	canonical, ok := canonicalUUID(s)
	if !ok || s != canonical || s == nilUUID {
		return malformed(field)
	}
	return nil
}

func lowerIfUUID(s string) string {
	if canonical, ok := canonicalUUID(s); ok {
		return canonical
	}
	return s
}

func canonicalUUID(s string) (string, bool) {
	if !isUUID(s) {
		return "", false
	}
	return strings.ToLower(s), true
}

func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		if !isHex(c) {
			return false
		}
	}
	return true
}

func isHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}
