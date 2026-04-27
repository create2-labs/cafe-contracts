package v01

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestGoldenFixture_ValidateAndRoundTrip(t *testing.T) {
	data := readFixture(t, "event_envelope_v01.json")

	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if err := env.ValidateVersioned(EventVersionV01); err != nil {
		t.Fatalf("ValidateVersioned: %v", err)
	}

	out, err := json.Marshal(&env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var again Envelope
	if err := json.Unmarshal(out, &again); err != nil {
		t.Fatalf("round-trip unmarshal: %v", err)
	}
	if err := again.ValidateVersioned(EventVersionV01); err != nil {
		t.Fatalf("round-trip validate: %v", err)
	}
	if !reflect.DeepEqual(env, again) {
		t.Fatalf("round-trip mismatch:\n%+v\nvs\n%+v", env, again)
	}
}

func TestValidate_RequiresFields(t *testing.T) {
	env := validEnvelope()
	env.EventID = ""
	if err := env.Validate(); !errors.Is(err, ErrEventIDRequired) {
		t.Fatalf("expected ErrEventIDRequired, got %v", err)
	}

	env = validEnvelope()
	env.EventType = ""
	if err := env.Validate(); !errors.Is(err, ErrEventTypeRequired) {
		t.Fatalf("expected ErrEventTypeRequired, got %v", err)
	}

	env = validEnvelope()
	env.EventVersion = ""
	if err := env.Validate(); !errors.Is(err, ErrEventVersionRequired) {
		t.Fatalf("expected ErrEventVersionRequired, got %v", err)
	}

	env = validEnvelope()
	env.OccurredAt = time.Time{}
	if err := env.Validate(); !errors.Is(err, ErrOccurredAtRequired) {
		t.Fatalf("expected ErrOccurredAtRequired, got %v", err)
	}

	env = validEnvelope()
	env.Producer = ""
	if err := env.Validate(); !errors.Is(err, ErrProducerRequired) {
		t.Fatalf("expected ErrProducerRequired, got %v", err)
	}
}

func TestValidateVersioned_WrongVersion(t *testing.T) {
	env := validEnvelope()
	env.EventVersion = "v0.2"
	if err := env.ValidateVersioned(EventVersionV01); !errors.Is(err, ErrEventVersionMismatch) {
		t.Fatalf("expected ErrEventVersionMismatch, got %v", err)
	}
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return b
}

func validEnvelope() Envelope {
	return Envelope{
		EventID:       "evt_common_0001",
		EventType:     "cafe.discovery.wallet.observed",
		EventVersion:  EventVersionV01,
		OccurredAt:    time.Date(2026, 4, 27, 10, 0, 0, 0, time.UTC),
		CorrelationID: "corr_0001",
		CausationID:   "cause_0001",
		Producer:      "cafe-discovery",
	}
}
