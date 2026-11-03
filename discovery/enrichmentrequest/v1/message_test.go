package v1

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const (
	idE  = "6ba7b810-9dad-51d1-80b4-00c04fd430c8"
	idS  = "11111111-1111-4111-8111-111111111111"
	idU  = "22222222-2222-4222-8222-222222222222"
	addr = "0x1111111111111111111111111111111111111111"
	nilU = "00000000-0000-0000-0000-000000000000"
)

func u(n uint64) *uint64 { return &n }

func hist() Request {
	return Request{EnrichmentID: idE, ScanID: idS, UserID: idU, Address: addr, Family: FamilyHistory,
		Chains: []Chain{{ChainID: 10, Nonce: u(7)}, {ChainID: 1, Nonce: u(3)}}}
}

func dep() Request {
	return Request{EnrichmentID: idE, ScanID: idS, UserID: idU, Address: addr, Family: FamilyDeployer,
		Chains: []Chain{{ChainID: 1}, {ChainID: 10}}}
}

func mod(r Request, f func(*Request)) Request { f(&r); return r }

func mustMalformed(t *testing.T, name string, err error) {
	t.Helper()
	if !errors.Is(err, ErrMalformed) {
		t.Fatalf("%s: want ErrMalformed, got %v", name, err)
	}
}

// OC-P09 note: closed family enum.
func TestFamilyIsValid(t *testing.T) {
	for f, want := range map[Family]bool{FamilyHistory: true, FamilyDeployer: true, "": false, "History": false, "other": false} {
		if got := f.IsValid(); got != want {
			t.Errorf("Family(%q).IsValid() = %v, want %v", f, got, want)
		}
	}
}

// CA4.1, ADR 6.3: valid requests, order of chains kept.
func TestValidateAccepts(t *testing.T) {
	for name, r := range map[string]Request{"history": hist(), "deployer": dep()} {
		if err := r.Validate(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// CA4.1, OC-P09 note: each rule rejected once.
func TestValidateRejects(t *testing.T) {
	cases := map[string]Request{
		"enrichment_id uppercase": mod(hist(), func(r *Request) { r.EnrichmentID = strings.ToUpper(idE) }),
		"enrichment_id not uuid":  mod(hist(), func(r *Request) { r.EnrichmentID = "abc" }),
		"enrichment_id nil":       mod(hist(), func(r *Request) { r.EnrichmentID = nilU }),
		"enrichment_id empty":     mod(hist(), func(r *Request) { r.EnrichmentID = "" }),
		"scan_id non-canonical":   mod(hist(), func(r *Request) { r.ScanID = strings.ReplaceAll(idS, "-", "") }),
		"scan_id nil":             mod(hist(), func(r *Request) { r.ScanID = nilU }),
		"user_id non-canonical":   mod(hist(), func(r *Request) { r.UserID = "{" + idU + "}" }),
		"user_id nil":             mod(hist(), func(r *Request) { r.UserID = nilU }),
		"address empty":           mod(hist(), func(r *Request) { r.Address = "" }),
		"family empty":            mod(hist(), func(r *Request) { r.Family = "" }),
		"family unknown":          mod(hist(), func(r *Request) { r.Family = "other" }),
		"chains nil":              mod(hist(), func(r *Request) { r.Chains = nil }),
		"chains empty":            mod(hist(), func(r *Request) { r.Chains = []Chain{} }),
		"chain id zero":           mod(dep(), func(r *Request) { r.Chains = []Chain{{ChainID: 0}} }),
		"chain id negative":       mod(dep(), func(r *Request) { r.Chains = []Chain{{ChainID: -1}} }),
		"chain id duplicate":      mod(dep(), func(r *Request) { r.Chains = []Chain{{ChainID: 1}, {ChainID: 10}, {ChainID: 1}} }),
		"history no nonce":        mod(hist(), func(r *Request) { r.Chains = []Chain{{ChainID: 1}} }),
		"history nonce zero":      mod(hist(), func(r *Request) { r.Chains = []Chain{{ChainID: 1, Nonce: u(0)}} }),
		"deployer nonce":          mod(dep(), func(r *Request) { r.Chains = []Chain{{ChainID: 1, Nonce: u(2)}} }),
		"deployer nonce zero":     mod(dep(), func(r *Request) { r.Chains = []Chain{{ChainID: 1, Nonce: u(0)}} }),
	}
	for name, r := range cases {
		mustMalformed(t, name, r.Validate())
	}
}

// CA4.1: round trip keeps the order of chains.
func TestRoundTrip(t *testing.T) {
	for name, want := range map[string]Request{"history": hist(), "deployer": dep()} {
		raw, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		got, err := DecodeRequest(raw)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: got %+v, want %+v", name, got, want)
		}
	}
}

// OC-P09 note: unknown fields are ignored; a deployer chain omits nonce.
func TestDecodeUnknownFieldsIgnored(t *testing.T) {
	raw := `{"enrichment_id":"` + idE + `","scan_id":"` + idS + `","user_id":"` + idU + `","address":"` + addr +
		`","family":"deployer","chains":[{"chain_id":1,"qualification":"eoa"}],"future":{"a":1}}`
	r, err := DecodeRequest([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if r.Family != FamilyDeployer || len(r.Chains) != 1 || r.Chains[0].Nonce != nil {
		t.Fatalf("unexpected request: %+v", r)
	}
}

// OC-P09 note: malformed shapes wrap ErrMalformed and yield a zero Request.
func TestDecodeMalformedShapes(t *testing.T) {
	for name, raw := range map[string]string{
		"empty": ``, "null": `null`, "array": `[]`, "string": `"x"`, "truncated": `{"family":`,
		"chains object":   `{"enrichment_id":"` + idE + `","chains":{}}`,
		"chain id string": `{"enrichment_id":"` + idE + `","chains":[{"chain_id":"1"}]}`,
		"nonce negative":  `{"enrichment_id":"` + idE + `","chains":[{"chain_id":1,"nonce":-1}]}`,
		"family number":   `{"enrichment_id":"` + idE + `","family":1}`,
	} {
		r, err := DecodeRequest([]byte(raw))
		mustMalformed(t, name, err)
		if !reflect.DeepEqual(r, Request{}) {
			t.Fatalf("%s: non-zero request on error: %+v", name, r)
		}
	}
}

// OC-P09 note: the error names the field, not the body.
func TestErrorOmitsBody(t *testing.T) {
	_, err := DecodeRequest([]byte(`{"enrichment_id":"` + idE + `","scan_id":"` + idS + `","user_id":"` + idU +
		`","address":"` + addr + `","family":"secret-body-marker","chains":[{"chain_id":1}]}`))
	mustMalformed(t, "family", err)
	if strings.Contains(err.Error(), "secret-body-marker") {
		t.Fatalf("error leaks body: %v", err)
	}
}

// CA4.1: ok_* decode and round trip, bad_* are rejected.
func TestRequestVectors(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	var ok, bad int
	for _, e := range entries {
		name := e.Name()
		raw, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		r, err := DecodeRequest(raw)
		switch {
		case strings.HasPrefix(name, "ok_"):
			ok++
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			if err := r.Validate(); err != nil {
				t.Fatalf("%s validate: %v", name, err)
			}
		case strings.HasPrefix(name, "bad_"):
			bad++
			mustMalformed(t, name, err)
			if !reflect.DeepEqual(r, Request{}) {
				t.Fatalf("%s: non-zero request on error", name)
			}
		}
	}
	if ok < 2 || bad < 2 {
		t.Fatalf("vectors: ok=%d bad=%d", ok, bad)
	}
}
