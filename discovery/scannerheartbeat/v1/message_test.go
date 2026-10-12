package v1

import (
	"encoding/json"
	"errors"
	"io/fs"
	"strings"
	"testing"
)

func TestHeartbeat(t *testing.T) {
	if SubjectScannerHeartbeatTLS != "scanner.heartbeat.tls" ||
		SubjectScannerHeartbeatWallet != "scanner.heartbeat.wallet" ||
		KindTLS != "tls" || KindWallet != "wallet" {
		t.Fatal(SubjectScannerHeartbeatTLS, SubjectScannerHeartbeatWallet, KindTLS, KindWallet)
	}
	entries, err := fs.ReadDir(Vectors, "testdata")
	if err != nil {
		t.Fatal(err)
	}
	var sawWallet, sawTLS, sawBad bool
	for _, entry := range entries {
		name := entry.Name()
		raw, err := fs.ReadFile(Vectors, "testdata/"+name)
		if err != nil {
			t.Fatal(err)
		}
		switch {
		case name == "ok_wallet.json":
			sawWallet = true
			roundTrip(t, name, raw, KindWallet)
		case name == "ok_tls.json":
			sawTLS = true
			roundTrip(t, name, raw, KindTLS)
		case strings.HasPrefix(name, "bad_"):
			sawBad = true
			reject(t, name, raw)
		default:
			t.Fatalf("unclassified fixture %s", name)
		}
	}
	if !sawWallet || !sawTLS || !sawBad {
		t.Fatal(sawWallet, sawTLS, sawBad)
	}

	kept := `{"scanner_id":"scanner-1","kind":"other","timestamp":"not-a-clock"}`
	msg, err := Decode([]byte(kept))
	if err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(msg)
	if err != nil || string(out) != kept {
		t.Fatalf("%s %v", out, err)
	}
	raw := []byte(`{"scanner_id":"SCANNER-1","kind":"wallet","timestamp":"2026-10-08T08:00:30Z","ignored":1}`)
	msg, err = Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	out, err = json.Marshal(msg)
	if err != nil || msg.ScannerID != "SCANNER-1" || strings.Contains(string(out), "ignored") {
		t.Fatalf("%s %v", out, err)
	}
	const token = "body-must-not-leak"
	err = decodeErr([]byte(`{"scanner_id":"` + token + `","kind":"","timestamp":"2026-10-08T08:00:30Z"}`))
	if err == nil || !errors.Is(err, ErrMalformed) || strings.Contains(err.Error(), token) {
		t.Fatal(err)
	}
	if err := (Message{}).Validate(); err == nil || !errors.Is(err, ErrMalformed) {
		t.Fatal(err)
	}
	for _, body := range []string{
		`{"kind":"wallet","timestamp":"2026-10-08T08:00:30Z"}`,
		`{"scanner_id":"scanner-1","timestamp":"2026-10-08T08:00:30Z"}`,
		`{"scanner_id":"scanner-1","kind":"wallet"}`,
		`{"scanner_id":null,"kind":"wallet","timestamp":"2026-10-08T08:00:30Z"}`,
		`{"scanner_id":"scanner-1","kind":true,"timestamp":"2026-10-08T08:00:30Z"}`,
		`{"scanner_id":"scanner-1","kind":"wallet","timestamp":1}`,
		`{"scanner_id":"scanner-1","kind":"wallet","timestamp":""}`,
		`true`,
		``,
		`{}{}`,
	} {
		reject(t, "shape", []byte(body))
	}
}

func roundTrip(t *testing.T, name string, raw []byte, kind string) {
	t.Helper()
	msg, err := Decode(raw)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if msg.Kind != kind || msg.ScannerID == "" || msg.Timestamp == "" {
		t.Fatalf("%s: %+v", name, msg)
	}
	out, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != strings.TrimSpace(string(raw)) {
		t.Fatalf("%s round trip:\n%s\nvs\n%s", name, out, raw)
	}
}

func reject(t *testing.T, name string, raw []byte) {
	t.Helper()
	if err := decodeErr(raw); err == nil || !errors.Is(err, ErrMalformed) {
		t.Fatalf("%s: %v", name, err)
	}
}

func decodeErr(raw []byte) error {
	_, err := Decode(raw)
	return err
}
