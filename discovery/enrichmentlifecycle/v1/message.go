package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	enrichmentrequestv1 "github.com/create2-labs/cafe-contracts/discovery/enrichmentrequest/v1"
)

// Started is one enrichment.started message.
type Started struct {
	EnrichmentID string                     `json:"enrichment_id"`
	ScanID       string                     `json:"scan_id"`
	UserID       string                     `json:"user_id"`
	Family       enrichmentrequestv1.Family `json:"family"`
	StartedAt    string                     `json:"started_at"`
}

// Failed is one enrichment.failed message.
type Failed struct {
	EnrichmentID string                     `json:"enrichment_id"`
	ScanID       string                     `json:"scan_id"`
	UserID       string                     `json:"user_id"`
	Family       enrichmentrequestv1.Family `json:"family"`
	Error        string                     `json:"error"`
	CompletedAt  string                     `json:"completed_at"`
}

// Validate checks the message against the contract. Every failure wraps
// ErrMalformed and names the field, never its value.
func (m Started) Validate() error {
	if err := validateCommon(m.EnrichmentID, m.ScanID, m.UserID, m.Family); err != nil {
		return err
	}
	if m.StartedAt == "" {
		return malformed("started_at")
	}
	return nil
}

// Validate checks the message against the contract. Every failure wraps
// ErrMalformed and names the field, never its value.
func (m Failed) Validate() error {
	if err := validateCommon(m.EnrichmentID, m.ScanID, m.UserID, m.Family); err != nil {
		return err
	}
	if m.Error == "" {
		return malformed("error")
	}
	if m.CompletedAt == "" {
		return malformed("completed_at")
	}
	return nil
}

// DecodeStarted unmarshals and validates enrichment.started. Identifiers are
// accepted in any case and canonicalized to lowercase. Unknown fields are
// ignored. A malformed body yields a zero Started.
func DecodeStarted(raw []byte) (Started, error) {
	var m Started
	if err := unmarshal(raw, &m); err != nil {
		return Started{}, err
	}
	m.EnrichmentID = lowerIfUUID(m.EnrichmentID)
	m.ScanID = lowerIfUUID(m.ScanID)
	m.UserID = lowerIfUUID(m.UserID)
	if err := m.Validate(); err != nil {
		return Started{}, err
	}
	return m, nil
}

// DecodeFailed unmarshals and validates enrichment.failed. Identifiers are
// accepted in any case and canonicalized to lowercase. Unknown fields are
// ignored. A malformed body yields a zero Failed.
func DecodeFailed(raw []byte) (Failed, error) {
	var m Failed
	if err := unmarshal(raw, &m); err != nil {
		return Failed{}, err
	}
	m.EnrichmentID = lowerIfUUID(m.EnrichmentID)
	m.ScanID = lowerIfUUID(m.ScanID)
	m.UserID = lowerIfUUID(m.UserID)
	if err := m.Validate(); err != nil {
		return Failed{}, err
	}
	return m, nil
}

func validateCommon(enrichmentID, scanID, userID string, family enrichmentrequestv1.Family) error {
	for _, f := range []struct{ name, id string }{
		{"enrichment_id", enrichmentID},
		{"scan_id", scanID},
		{"user_id", userID},
	} {
		if err := validateID(f.name, f.id); err != nil {
			return err
		}
	}
	if !family.IsValid() {
		return malformed("family")
	}
	return nil
}

func unmarshal(raw []byte, v any) error {
	if err := json.Unmarshal(raw, v); err != nil {
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) && typeErr.Field != "" {
			return malformed(typeErr.Field)
		}
		return malformed("body")
	}
	return nil
}

func malformed(field string) error {
	return fmt.Errorf("%w: %s", ErrMalformed, field)
}

const nilUUID = "00000000-0000-0000-0000-000000000000"

// validateID requires the canonical lowercase form and rejects the nil UUID.
func validateID(field, s string) error {
	if !isUUID(s) || s != strings.ToLower(s) || s == nilUUID {
		return malformed(field)
	}
	return nil
}

func lowerIfUUID(s string) string {
	if isUUID(s) {
		return strings.ToLower(s)
	}
	return s
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
