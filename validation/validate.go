// Package validation provides small, reusable checks for wire contract fields.
// Domain-specific rules stay in each versioned contract package; this package only
// offers generic building blocks.
package validation

import (
	"errors"
	"fmt"
)

// ErrInvalid indicates that wire data does not satisfy a contract rule.
var ErrInvalid = errors.New("invalid wire contract data")

// FieldError attaches a field name (or path) to a validation error.
type FieldError struct {
	Field string
	Err   error
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("%s: %v", e.Field, e.Err)
}

func (e *FieldError) Unwrap() error {
	return e.Err
}

// RequireNonEmpty returns a *FieldError wrapping ErrInvalid when s is empty.
func RequireNonEmpty(field, s string) error {
	if s == "" {
		return &FieldError{Field: field, Err: ErrInvalid}
	}
	return nil
}
