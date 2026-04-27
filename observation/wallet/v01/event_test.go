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
	data := readFixture(t, "cafe_discovery_wallet_observed_v01.json")

	var ev Event
	if err := json.Unmarshal(data, &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if err := ev.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	if ev.EventID != "evt_disc_0001" {
		t.Fatalf("event_id: got %q", ev.EventID)
	}
	if ev.Subject.Type != "wallet" || ev.Subject.ID == "" {
		t.Fatalf("subject: %+v", ev.Subject)
	}
	if len(ev.Payload.ChainIDs) != 2 || ev.Payload.ChainIDs[0] != 1 || ev.Payload.ChainIDs[1] != 8453 {
		t.Fatalf("chain_ids: %v", ev.Payload.ChainIDs)
	}

	out, err := json.Marshal(&ev)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var again Event
	if err := json.Unmarshal(out, &again); err != nil {
		t.Fatalf("round-trip unmarshal: %v", err)
	}
	if err := again.Validate(); err != nil {
		t.Fatalf("round-trip validate: %v", err)
	}
	if !reflect.DeepEqual(ev, again) {
		t.Fatalf("round-trip mismatch:\n%+v\nvs\n%+v", ev, again)
	}
}

func TestValidate_InvalidEventType(t *testing.T) {
	ev := validMinimalEvent()
	ev.EventType = "wrong"
	err := ev.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrEventType) {
		t.Fatalf("got %v", err)
	}
}

func TestValidate_InvalidAlgorithm(t *testing.T) {
	ev := validMinimalEvent()
	ev.Payload.CurrentAlgorithm = "not-listed"
	err := ev.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrAlgorithmID) {
		t.Fatalf("got %v", err)
	}
}

func TestValidate_HybridAlgorithmAccepted(t *testing.T) {
	ev := validMinimalEvent()
	ev.Payload.CurrentAlgorithm = "hybrid_custom"
	if err := ev.Validate(); err != nil {
		t.Fatal(err)
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

func validMinimalEvent() Event {
	return Event{
		EventID:       "e1",
		EventType:     EventTypeWalletObserved,
		EventVersion:  EventVersion,
		OccurredAt:    time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC),
		CorrelationID: "c1",
		CausationID:   "a1",
		Producer:      ProducerCafeDiscovery,
		Subject: Subject{
			Type: "wallet",
			ID:   "wallet:0xabc",
		},
		Payload: Payload{
			ChainIDs:         []int64{1},
			AccountKind:      "eoa",
			CurrentAlgorithm: "secp256k1_ecrecover",
			PublicKeyExposed: false,
			IsMultichain:     false,
			ObservedAt:       time.Date(2026, 4, 17, 9, 59, 58, 0, time.UTC),
			CurrentPQPosture: "unknown",
		},
	}
}
