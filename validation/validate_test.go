package validation

import (
	"errors"
	"testing"
)

func TestRequireNonEmpty(t *testing.T) {
	if err := RequireNonEmpty("id", "x"); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	err := RequireNonEmpty("id", "")
	var fe *FieldError
	if !errors.As(err, &fe) {
		t.Fatalf("want *FieldError, got %T: %v", err, err)
	}
	if fe.Field != "id" || !errors.Is(err, ErrInvalid) {
		t.Fatalf("field or cause: %v", err)
	}
}
