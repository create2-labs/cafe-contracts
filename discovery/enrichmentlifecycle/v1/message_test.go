package v1

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	enrichmentrequestv1 "github.com/create2-labs/cafe-contracts/discovery/enrichmentrequest/v1"
)

const (
	idE  = "6ba7b810-9dad-51d1-80b4-00c04fd430c8"
	idS  = "11111111-1111-4111-8111-111111111111"
	idU  = "22222222-2222-4222-8222-222222222222"
	nilU = "00000000-0000-0000-0000-000000000000"
	ts   = "2026-11-03T10:00:00Z"
)

var families = []enrichmentrequestv1.Family{enrichmentrequestv1.FamilyHistory, enrichmentrequestv1.FamilyDeployer}

func started(f enrichmentrequestv1.Family) Started {
	return Started{EnrichmentID: idE, ScanID: idS, UserID: idU, Family: f, StartedAt: ts}
}

func failed(f enrichmentrequestv1.Family) Failed {
	return Failed{EnrichmentID: idE, ScanID: idS, UserID: idU, Family: f, Error: "rpc unavailable", CompletedAt: ts}
}

func mustMalformed(t *testing.T, name string, err error) {
	t.Helper()
	if !errors.Is(err, ErrMalformed) {
		t.Fatalf("%s: want ErrMalformed, got %v", name, err)
	}
}

// OC-P10 note, ADR 6.4, CA4.4: subjects and stream.
func TestConstants(t *testing.T) {
	for got, want := range map[string]string{
		SubjectEnrichmentStarted:   "enrichment.started",
		SubjectEnrichmentCompleted: "enrichment.completed",
		SubjectEnrichmentFailed:    "enrichment.failed",
		StreamEnrichmentLifecycle:  "ENRICHMENT_LIFECYCLE",
	} {
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

// CA3.4, CA4.4: one valid message per family.
func TestValidateAccepts(t *testing.T) {
	for _, f := range families {
		if err := started(f).Validate(); err != nil {
			t.Errorf("started %s: %v", f, err)
		}
		if err := failed(f).Validate(); err != nil {
			t.Errorf("failed %s: %v", f, err)
		}
	}
}

// OC-P10 note: each rule rejected once per message.
func TestValidateRejects(t *testing.T) {
	for _, c := range []struct {
		name string
		mod  func(e, s, u, fam, a, b *string)
	}{
		{"enrichment_id uppercase", func(e, _, _, _, _, _ *string) { *e = strings.ToUpper(idE) }},
		{"enrichment_id not uuid", func(e, _, _, _, _, _ *string) { *e = "abc" }},
		{"enrichment_id nil", func(e, _, _, _, _, _ *string) { *e = nilU }},
		{"enrichment_id empty", func(e, _, _, _, _, _ *string) { *e = "" }},
		{"scan_id non-canonical", func(_, s, _, _, _, _ *string) { *s = strings.ReplaceAll(idS, "-", "") }},
		{"scan_id nil", func(_, s, _, _, _, _ *string) { *s = nilU }},
		{"user_id non-canonical", func(_, _, u, _, _, _ *string) { *u = "{" + idU + "}" }},
		{"user_id nil", func(_, _, u, _, _, _ *string) { *u = nilU }},
		{"family empty", func(_, _, _, f, _, _ *string) { *f = "" }},
		{"family unknown", func(_, _, _, f, _, _ *string) { *f = "other" }},
		{"family wrong case", func(_, _, _, f, _, _ *string) { *f = "History" }},
		{"timestamp empty", func(_, _, _, _, a, _ *string) { *a = "" }},
		{"error empty", func(_, _, _, _, _, b *string) { *b = "" }},
	} {
		var fam, tsv, errText = "history", ts, "boom"
		e, s, u := idE, idS, idU
		c.mod(&e, &s, &u, &fam, &tsv, &errText)
		f := enrichmentrequestv1.Family(fam)
		st := Started{EnrichmentID: e, ScanID: s, UserID: u, Family: f, StartedAt: tsv}
		fa := Failed{EnrichmentID: e, ScanID: s, UserID: u, Family: f, Error: errText, CompletedAt: tsv}
		switch c.name {
		case "error empty":
			mustMalformed(t, "failed "+c.name, fa.Validate())
		case "timestamp empty":
			mustMalformed(t, "started started_at empty", st.Validate())
			mustMalformed(t, "failed completed_at empty", fa.Validate())
		default:
			mustMalformed(t, "started "+c.name, st.Validate())
			mustMalformed(t, "failed "+c.name, fa.Validate())
		}
	}
}

// CA3.4, CA4.4: round trip per family.
func TestRoundTrip(t *testing.T) {
	for _, f := range families {
		rawS, _ := json.Marshal(started(f))
		gotS, err := DecodeStarted(rawS)
		if err != nil || !reflect.DeepEqual(gotS, started(f)) {
			t.Errorf("started %s: got %+v, err %v", f, gotS, err)
		}
		rawF, _ := json.Marshal(failed(f))
		gotF, err := DecodeFailed(rawF)
		if err != nil || !reflect.DeepEqual(gotF, failed(f)) {
			t.Errorf("failed %s: got %+v, err %v", f, gotF, err)
		}
	}
}

// OC-P10 note: unknown fields are ignored.
func TestDecodeUnknownFieldsIgnored(t *testing.T) {
	pre := `{"enrichment_id":"` + idE + `","scan_id":"` + idS + `","user_id":"` + idU + `","family":"history","future":{"a":1},`
	if got, err := DecodeStarted([]byte(pre + `"started_at":"` + ts + `"}`)); err != nil || got != started("history") {
		t.Errorf("started: %+v, %v", got, err)
	}
	if got, err := DecodeFailed([]byte(pre + `"error":"rpc unavailable","completed_at":"` + ts + `"}`)); err != nil || got != failed("history") {
		t.Errorf("failed: %+v, %v", got, err)
	}
}

// OC-P10 note: malformed shapes wrap ErrMalformed and yield a zero value.
func TestDecodeMalformedShapes(t *testing.T) {
	for name, raw := range map[string]string{
		"empty": ``, "null": `null`, "array": `[]`, "string": `"x"`, "truncated": `{"family":`,
		"family number": `{"enrichment_id":"` + idE + `","family":1}`,
		"id number":     `{"enrichment_id":1}`,
		"error number":  `{"enrichment_id":"` + idE + `","error":1}`,
		"missing all":   `{}`,
	} {
		s, err := DecodeStarted([]byte(raw))
		mustMalformed(t, "started "+name, err)
		if s != (Started{}) {
			t.Errorf("started %s: non-zero value on error: %+v", name, s)
		}
		f, err := DecodeFailed([]byte(raw))
		mustMalformed(t, "failed "+name, err)
		if f != (Failed{}) {
			t.Errorf("failed %s: non-zero value on error: %+v", name, f)
		}
	}
}

// OC-P10 note: the error names the field, not the body.
func TestErrorOmitsBody(t *testing.T) {
	pre := `{"enrichment_id":"` + idE + `","scan_id":"` + idS + `","user_id":"` + idU + `","family":"secret-body-marker",`
	_, err := DecodeStarted([]byte(pre + `"started_at":"` + ts + `"}`))
	mustMalformed(t, "started", err)
	if strings.Contains(err.Error(), "secret-body-marker") {
		t.Errorf("started error leaks body: %v", err)
	}
	_, err = DecodeFailed([]byte(pre + `"error":"x","completed_at":"` + ts + `"}`))
	mustMalformed(t, "failed", err)
	if strings.Contains(err.Error(), "secret-body-marker") {
		t.Errorf("failed error leaks body: %v", err)
	}
}

// CA3.4, CA4.4: ok_* decode, bad_* are rejected; the name says which message.
func TestVectors(t *testing.T) {
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
		if strings.Contains(name, "_started_") {
			_, err = DecodeStarted(raw)
		} else {
			_, err = DecodeFailed(raw)
		}
		switch {
		case strings.HasPrefix(name, "ok_"):
			ok++
			if err != nil {
				t.Errorf("%s: %v", name, err)
			}
		case strings.HasPrefix(name, "bad_"):
			bad++
			mustMalformed(t, name, err)
		}
	}
	if ok < 2 || bad < 2 {
		t.Fatalf("vectors: ok=%d bad=%d", ok, bad)
	}
}
