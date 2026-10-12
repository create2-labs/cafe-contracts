package v1

import (
	"encoding/json"
	"errors"
	"io/fs"
	"strings"
	"testing"
)

func TestSubjects(t *testing.T) {
	if SubjectWalletScan != "cafe.discovery.wallet.scan" ||
		SubjectTLSScan != "cafe.discovery.tls.scan" ||
		SubjectScanRequestedWallet != "scan.requested.wallet" ||
		SubjectScanRequestedTLS != "scan.requested.tls" {
		t.Fatalf("subjects %s %s %s %s", SubjectWalletScan, SubjectTLSScan, SubjectScanRequestedWallet, SubjectScanRequestedTLS)
	}
	if NilID != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("nil id %s", NilID)
	}
}

func TestScanRequestVectors(t *testing.T) {
	entries, err := fs.ReadDir(Vectors, "testdata")
	if err != nil {
		t.Fatal(err)
	}
	var sawWallet, sawTLS, sawDefault, sawBad bool
	for _, entry := range entries {
		name := entry.Name()
		raw, err := fs.ReadFile(Vectors, "testdata/"+name)
		if err != nil {
			t.Fatal(err)
		}
		switch {
		case strings.HasPrefix(name, "ok_wallet"):
			sawWallet = true
			msg, err := DecodeWallet(raw)
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			out, err := json.Marshal(msg)
			if err != nil {
				t.Fatal(err)
			}
			if string(out) != strings.TrimSpace(string(raw)) {
				t.Fatalf("%s round trip:\n%s\nvs\n%s", name, out, raw)
			}
			if _, err := DecodeTLS(raw); err == nil {
				t.Fatalf("%s decoded as TLS", name)
			}
		case name == "ok_tls.json":
			sawTLS = true
			assertTLSRoundTrip(t, name, raw, false, "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee")
		case name == "ok_tls_default.json":
			sawDefault = true
			assertTLSRoundTrip(t, name, raw, true, NilID)
		case strings.HasPrefix(name, "bad_wallet"):
			sawBad = true
			if _, err := DecodeWallet(raw); err == nil || !errors.Is(err, ErrMalformed) {
				t.Fatalf("%s: %v", name, err)
			}
		case strings.HasPrefix(name, "bad_tls"), strings.HasPrefix(name, "bad_"):
			sawBad = true
			if _, err := DecodeTLS(raw); err == nil || !errors.Is(err, ErrMalformed) {
				t.Fatalf("%s tls: %v", name, err)
			}
			if _, err := DecodeWallet(raw); err == nil || !errors.Is(err, ErrMalformed) {
				t.Fatalf("%s wallet: %v", name, err)
			}
		default:
			t.Fatalf("unclassified fixture %s", name)
		}
	}
	if !sawWallet || !sawTLS || !sawDefault || !sawBad {
		t.Fatalf("fixtures missing: wallet=%v tls=%v default=%v bad=%v", sawWallet, sawTLS, sawDefault, sawBad)
	}
}

func assertTLSRoundTrip(t *testing.T, name string, raw []byte, isDefault bool, userID string) {
	t.Helper()
	msg, err := DecodeTLS(raw)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if msg.IsDefault != isDefault || msg.UserID != userID || msg.Endpoint != "https://example.com" {
		t.Fatalf("%s: %+v", name, msg)
	}
	out, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != strings.TrimSpace(string(raw)) {
		t.Fatalf("%s round trip:\n%s\nvs\n%s", name, out, raw)
	}
	if _, err := DecodeWallet(raw); err == nil {
		t.Fatalf("%s decoded as wallet", name)
	}
}

func TestCanonicalUUIDAndDefaults(t *testing.T) {
	raw := []byte(`{"scan_id":"6F1C0C1E-9B2A-4D3E-8F7A-1C2B3D4E5F60","user_id":"` + NilID + `","address":"0xabc","ignored":1}`)
	msg, err := DecodeWallet(raw)
	if err != nil {
		t.Fatal(err)
	}
	if msg.ScanID != "6f1c0c1e-9b2a-4d3e-8f7a-1c2b3d4e5f60" || msg.UserID != NilID || msg.Address != "0xabc" {
		t.Fatalf("%+v", msg)
	}
	out, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "ignored") || strings.Contains(string(out), "6F1C") {
		t.Fatalf("marshal %s", out)
	}

	omitted := []byte(`{"scan_id":"6f1c0c1e-9b2a-4d3e-8f7a-1c2b3d4e5f60","user_id":"` + NilID + `","endpoint":"https://example.com"}`)
	tlsMsg, err := DecodeTLS(omitted)
	if err != nil {
		t.Fatal(err)
	}
	if tlsMsg.IsDefault {
		t.Fatal("missing is_default decoded as true")
	}
	encoded, err := json.Marshal(tlsMsg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"is_default":false`) {
		t.Fatalf("marshal %s", encoded)
	}
}

func TestRejectedShapes(t *testing.T) {
	bodies := []string{
		`{"scan_id":"{6f1c0c1e-9b2a-4d3e-8f7a-1c2b3d4e5f60}","user_id":"aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee","address":"0xabc"}`,
		`{"scan_id":"6f1c0c1e9b2a4d3e8f7a1c2b3d4e5f60","user_id":"aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee","address":"0xabc"}`,
		`{"scan_id":"","user_id":"aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee","endpoint":"https://example.com","is_default":"true"}`,
		`{"scan_id":"6f1c0c1e-9b2a-4d3e-8f7a-1c2b3d4e5f60","user_id":"aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee","endpoint":"https://example.com","is_default":null}`,
	}
	for _, body := range bodies {
		if _, err := DecodeWallet([]byte(body)); err == nil || !errors.Is(err, ErrMalformed) {
			t.Fatalf("wallet %s: %v", body, err)
		}
		if _, err := DecodeTLS([]byte(body)); err == nil || !errors.Is(err, ErrMalformed) {
			t.Fatalf("tls %s: %v", body, err)
		}
	}
}

func TestErrorOmitsBody(t *testing.T) {
	const token = "body-must-not-leak"
	raw := []byte(`{"scan_id":"bad","user_id":"bad","address":"` + token + `"}`)
	_, err := DecodeWallet(raw)
	if err == nil || !errors.Is(err, ErrMalformed) || strings.Contains(err.Error(), token) {
		t.Fatalf("%v", err)
	}
	if err := (Wallet{ScanID: "6F1C0C1E-9B2A-4D3E-8F7A-1C2B3D4E5F60", UserID: NilID, Address: "0xabc"}).Validate(); err == nil || !errors.Is(err, ErrMalformed) {
		t.Fatalf("uppercase: %v", err)
	}
}
