package v1

import (
	"encoding/json"
	"errors"
	"io/fs"
	"strings"
	"testing"
)

const validID = "6f1c0c1e-9b2a-4d3e-8f7a-1c2b3d4e5f60"
const validAddress = "0x742d35cc6634c0532925a3b844bc454e4438f44e"
const validDeadline = "2026-10-08T08:00:30.000000000Z"

// CA3.2
func TestSubject(t *testing.T) {
	if Subject != "cafe.indexer.gateway.contracts.creation.v1" {
		t.Fatalf("subject = %q", Subject)
	}
	if MaxRequestIDBytes != 128 {
		t.Fatalf("MaxRequestIDBytes = %d", MaxRequestIDBytes)
	}
}

// CA3.2
func TestRequestVectors(t *testing.T) {
	entries, err := fs.ReadDir(Vectors, "testdata")
	if err != nil {
		t.Fatal(err)
	}
	var sawOK, sawBad bool
	for _, entry := range entries {
		name := entry.Name()
		raw, err := fs.ReadFile(Vectors, "testdata/"+name)
		if err != nil {
			t.Fatal(err)
		}
		switch {
		case strings.HasPrefix(name, "ok_request"):
			sawOK = true
			req, err := DecodeRequest(raw)
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			out, err := json.Marshal(req)
			if err != nil {
				t.Fatal(err)
			}
			if string(out) != strings.TrimSpace(string(raw)) {
				t.Fatalf("%s round trip:\n%s\nvs\n%s", name, out, raw)
			}
		case strings.HasPrefix(name, "bad_request"):
			sawBad = true
			if _, err := DecodeRequest(raw); err == nil || !errors.Is(err, ErrRequest) {
				t.Fatalf("%s: got %v", name, err)
			}
		}
	}
	if !sawOK || !sawBad {
		t.Fatalf("fixtures missing: ok=%v bad=%v", sawOK, sawBad)
	}
}

// CA3.2: boundary values that stay valid.
func TestDecodeRequestAcceptsBoundary(t *testing.T) {
	req := validRequest()
	req.RequestID = "6F1C0C1E-9B2A-4D3E-8F7A-1C2B3D4E5F60"
	req.Address = "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
	req.ChainID = 9223372036854775807
	req.Deadline = "2026-10-08T10:00:30+02:00"
	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := req
	want.Address = validAddress
	if got != want {
		t.Fatalf("changed request: %+v", got)
	}
}

// CA3.2: a 0X address prefix is accepted and decoded to lowercase 0x.
func TestDecodeRequestAcceptsUpperPrefix(t *testing.T) {
	req := validRequest()
	req.Address = "0X742D35CC6634C0532925A3B844BC454E4438F44E"
	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Address != validAddress {
		t.Fatalf("address = %q", got.Address)
	}
}

// CA3.2: unknown fields are ignored.
func TestDecodeRequestIgnoresUnknownField(t *testing.T) {
	raw := []byte(`{"request_id":"` + validID + `","address":"` + validAddress + `","chain_id":1,"deadline":"` + validDeadline + `","extra":{"a":1}}`)
	got, err := DecodeRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != validRequest() {
		t.Fatalf("got %+v", got)
	}
}

// CA3.2: strict decoding and field validation, error names the field only.
func TestDecodeRequestRejects(t *testing.T) {
	const dl = `"deadline":"` + validDeadline + `"`
	cases := []struct {
		name  string
		body  string
		field string
	}{
		{"request_id missing", `{"address":"` + validAddress + `","chain_id":1,` + dl + `}`, "request_id"},
		{"request_id number", `{"request_id":1,"address":"` + validAddress + `","chain_id":1,` + dl + `}`, "request_id"},
		{"request_id null", `{"request_id":null,"address":"` + validAddress + `","chain_id":1,` + dl + `}`, "request_id"},
		{"request_id empty", `{"request_id":"","address":"` + validAddress + `","chain_id":1,` + dl + `}`, "request_id"},
		{"request_id braced", `{"request_id":"{` + validID + `}","address":"` + validAddress + `","chain_id":1,` + dl + `}`, "request_id"},
		{"address missing", `{"request_id":"` + validID + `","chain_id":1,` + dl + `}`, "address"},
		{"address number", `{"request_id":"` + validID + `","address":5,"chain_id":1,` + dl + `}`, "address"},
		{"address short", `{"request_id":"` + validID + `","address":"0xabc","chain_id":1,` + dl + `}`, "address"},
		{"address no prefix", `{"request_id":"` + validID + `","address":"742d35cc6634c0532925a3b844bc454e4438f44e","chain_id":1,` + dl + `}`, "address"},
		{"chain missing", `{"request_id":"` + validID + `","address":"` + validAddress + `",` + dl + `}`, "chain_id"},
		{"chain string", `{"request_id":"` + validID + `","address":"` + validAddress + `","chain_id":"1",` + dl + `}`, "chain_id"},
		{"chain fraction", `{"request_id":"` + validID + `","address":"` + validAddress + `","chain_id":1.0,` + dl + `}`, "chain_id"},
		{"chain overflow", `{"request_id":"` + validID + `","address":"` + validAddress + `","chain_id":9223372036854775808,` + dl + `}`, "chain_id"},
		{"chain zero", `{"request_id":"` + validID + `","address":"` + validAddress + `","chain_id":0,` + dl + `}`, "chain_id"},
		{"chain negative", `{"request_id":"` + validID + `","address":"` + validAddress + `","chain_id":-1,` + dl + `}`, "chain_id"},
		{"deadline missing", `{"request_id":"` + validID + `","address":"` + validAddress + `","chain_id":1}`, "deadline"},
		{"deadline null", `{"request_id":"` + validID + `","address":"` + validAddress + `","chain_id":1,"deadline":null}`, "deadline"},
		{"deadline empty", `{"request_id":"` + validID + `","address":"` + validAddress + `","chain_id":1,"deadline":""}`, "deadline"},
		{"deadline text", `{"request_id":"` + validID + `","address":"` + validAddress + `","chain_id":1,"deadline":"soon-secret"}`, "deadline"},
		{"first invalid in order", `{"request_id":"bad","address":"0xabc","chain_id":0,"deadline":"x"}`, "request_id"},
		{"address before chain", `{"request_id":"` + validID + `","address":"0xabc","chain_id":0,"deadline":"x"}`, "address"},
		{"chain before deadline", `{"request_id":"` + validID + `","address":"` + validAddress + `","chain_id":0,"deadline":"x"}`, "chain_id"},
		{"unreadable", `{`, "json"},
		{"array", `[]`, "json"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeRequest([]byte(tt.body))
			if err == nil || !errors.Is(err, ErrRequest) {
				t.Fatalf("got %v", err)
			}
			if got != (Request{}) {
				t.Fatalf("non-zero request on failure: %+v", got)
			}
			if !strings.HasSuffix(err.Error(), ": "+tt.field) {
				t.Fatalf("error %q does not end with field %q", err, tt.field)
			}
			if strings.Contains(err.Error(), "soon-secret") {
				t.Fatal("error quotes a value")
			}
		})
	}
}

// CA3.2: Validate on a decoded value.
func TestValidate(t *testing.T) {
	if err := validRequest().Validate(); err != nil {
		t.Fatal(err)
	}
	bad := validRequest()
	bad.ChainID = -5
	if err := bad.Validate(); err == nil || !errors.Is(err, ErrRequest) || !strings.HasSuffix(err.Error(), ": chain_id") {
		t.Fatalf("got %v", err)
	}
	bad = validRequest()
	bad.Address = "0xabc"
	bad.Deadline = ""
	if err := bad.Validate(); err == nil || !strings.HasSuffix(err.Error(), ": address") {
		t.Fatalf("got %v", err)
	}
}

// CA3.2
func TestBoundedRequestID(t *testing.T) {
	entries := []struct {
		name string
		want string
	}{
		{"echo_bounded.json", "not-a-uuid"},
		{"echo_empty.json", ""},
		{"echo_missing.json", ""},
		{"echo_number.json", ""},
		{"echo_unreadable.txt", ""},
		{"ok_request.json", validID},
		{"bad_request_id.json", "not-a-uuid"},
	}
	for _, tt := range entries {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := fs.ReadFile(Vectors, "testdata/"+tt.name)
			if err != nil {
				t.Fatal(err)
			}
			if got := BoundedRequestID(raw); got != tt.want {
				t.Fatalf("got %q", got)
			}
			if _, err := DecodeRequest(raw); tt.name != "ok_request.json" && err == nil {
				t.Fatal("expected invalid request")
			}
		})
	}

	id128 := strings.Repeat("b", MaxRequestIDBytes)
	id129 := strings.Repeat("c", MaxRequestIDBytes+1)
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"one byte", `{"request_id":"x"}`, "x"},
		{"128", `{"request_id":"` + id128 + `"}`, id128},
		{"129", `{"request_id":"` + id129 + `"}`, ""},
		{"null", `{"request_id":null}`, ""},
		{"bool", `{"request_id":true}`, ""},
		{"object", `{"request_id":{"id":"a"}}`, ""},
		{"array", `{"request_id":["a"]}`, ""},
		{"json null", `null`, ""},
		{"json string", `"abc"`, ""},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := BoundedRequestID([]byte(tt.raw)); got != tt.want {
				t.Fatalf("got %q", got)
			}
		})
	}
}

// CA3.2: an oversized request_id is neither echoed by the error nor by the helper.
func TestRejectedRequestDoesNotLeakIdentifier(t *testing.T) {
	long := strings.Repeat("d", 200)
	raw := []byte(`{"request_id":"` + long + `","address":"0xabc","chain_id":1,"deadline":"` + validDeadline + `"}`)
	_, err := DecodeRequest(raw)
	if err == nil || !errors.Is(err, ErrRequest) {
		t.Fatal(err)
	}
	if strings.Contains(err.Error(), long) {
		t.Fatal("error contains request_id")
	}
	if BoundedRequestID(raw) != "" {
		t.Fatal("oversized id was echoed")
	}
}

func validRequest() Request {
	return Request{RequestID: validID, Address: validAddress, ChainID: 1, Deadline: validDeadline}
}
