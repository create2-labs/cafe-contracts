package v1

import (
	"encoding/json"
	"errors"
	"io/fs"
	"strings"
	"testing"
)

const (
	scanID = "6f1c0c1e-9b2a-4d3e-8f7a-1c2b3d4e5f60"
	userID = "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee"
)

func TestLifecycle(t *testing.T) {
	if SubjectScanStarted != "scan.started" || SubjectScanCompleted != "scan.completed" ||
		SubjectScanFailed != "scan.failed" || SubjectScanReady != "scan.ready" ||
		NilID != "00000000-0000-0000-0000-000000000000" {
		t.Fatal(SubjectScanStarted, SubjectScanCompleted, SubjectScanFailed, SubjectScanReady, NilID)
	}
	entries, err := fs.ReadDir(Vectors, "testdata")
	if err != nil {
		t.Fatal(err)
	}
	var sawStarted, sawCompleted, sawFailed, sawReady, sawBad bool
	for _, entry := range entries {
		name := entry.Name()
		raw, err := fs.ReadFile(Vectors, "testdata/"+name)
		if err != nil {
			t.Fatal(err)
		}
		switch {
		case strings.HasPrefix(name, "ok_started"):
			sawStarted = true
			roundTrip(t, name, raw, DecodeStarted)
		case strings.HasPrefix(name, "ok_completed"):
			sawCompleted = true
			roundTrip(t, name, raw, DecodeCompleted)
		case strings.HasPrefix(name, "ok_failed"):
			sawFailed = true
			roundTrip(t, name, raw, DecodeFailed)
		case strings.HasPrefix(name, "ok_ready"):
			sawReady = true
			roundTrip(t, name, raw, DecodeReady)
		case strings.HasPrefix(name, "bad_started"):
			sawBad = true
			reject(t, name, raw, startedErr)
		case strings.HasPrefix(name, "bad_completed"):
			sawBad = true
			reject(t, name, raw, completedErr)
		case strings.HasPrefix(name, "bad_failed"):
			sawBad = true
			reject(t, name, raw, failedErr)
		case strings.HasPrefix(name, "bad_ready"):
			sawBad = true
			reject(t, name, raw, readyErr)
		case strings.HasPrefix(name, "bad_"):
			sawBad = true
			reject(t, name, raw, startedErr)
			reject(t, name, raw, completedErr)
			reject(t, name, raw, failedErr)
			reject(t, name, raw, readyErr)
		default:
			t.Fatalf("unclassified fixture %s", name)
		}
	}
	if !sawStarted || !sawCompleted || !sawFailed || !sawReady || !sawBad {
		t.Fatal(sawStarted, sawCompleted, sawFailed, sawReady, sawBad)
	}
	started, err := DecodeStarted([]byte(`{"scan_id":"` + scanID + `","kind":"other","user_id":"` + NilID + `","started_at":"not-a-clock","endpoint":"","address":""}`))
	if err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(started)
	if err != nil || started.Kind != "other" || strings.Contains(string(out), "endpoint") {
		t.Fatalf("%s %v", out, err)
	}
	ready, err := DecodeReady([]byte(`{"scan_id":"` + scanID + `","user_id":"` + strings.ToUpper(userID) + `","kind":"partial","status":"partial","ignored":1}`))
	if err != nil {
		t.Fatal(err)
	}
	out, err = json.Marshal(ready)
	if err != nil || ready.UserID != userID || strings.Contains(string(out), "ignored") {
		t.Fatalf("%s %v", out, err)
	}
	completed, err := DecodeCompleted([]byte(`{"scan_id":"` + strings.ToUpper(scanID) + `","kind":"other","user_id":"` + userID + `","findings":null,"metadata":null,"completed_at":"not-a-clock","endpoint":"","address":"","result":null,"ignored":1}`))
	if err != nil {
		t.Fatal(err)
	}
	out, err = json.Marshal(completed)
	want := `{"scan_id":"` + scanID + `","kind":"other","user_id":"` + userID + `","completed_at":"not-a-clock"}`
	if err != nil || string(out) != want {
		t.Fatalf("%s %v", out, err)
	}
	kept := `{"scan_id":"` + scanID + `","kind":"wallet","user_id":"` + userID + `","findings":[],"metadata":{},"completed_at":"t","result":{"z":1,"a":2}}`
	out, err = json.Marshal(mustCompleted(t, []byte(kept)))
	if err != nil || string(out) != kept {
		t.Fatalf("%s %v", out, err)
	}
	const token = "body-must-not-leak"
	err = startedErr([]byte(`{"scan_id":"bad","kind":"wallet","user_id":"` + userID + `","started_at":"2026-10-08T08:00:30Z","address":"` + token + `"}`))
	if err == nil || !errors.Is(err, ErrMalformed) || strings.Contains(err.Error(), token) {
		t.Fatal(err)
	}
	err = completedErr([]byte(`{"scan_id":"bad","kind":"wallet","user_id":"` + userID + `","completed_at":"2026-10-08T08:00:30Z","result":"` + token + `"}`))
	if err == nil || !errors.Is(err, ErrMalformed) || strings.Contains(err.Error(), token) {
		t.Fatal(err)
	}
	if err := (Started{ScanID: strings.ToUpper(scanID), Kind: "wallet", UserID: NilID, StartedAt: "t"}).Validate(); err == nil || !errors.Is(err, ErrMalformed) {
		t.Fatal(err)
	}
	if err := (Completed{ScanID: strings.ToUpper(scanID), Kind: "wallet", UserID: NilID, CompletedAt: "t"}).Validate(); err == nil || !errors.Is(err, ErrMalformed) {
		t.Fatal(err)
	}
	for _, body := range []string{
		`{"scan_id":"6f1c0c1e9b2a4d3e8f7a1c2b3d4e5f60","kind":"wallet","user_id":"` + userID + `","started_at":"2026-10-08T08:00:30Z"}`,
		`{"scan_id":"` + scanID + `","kind":1,"user_id":"` + userID + `","error":"x","completed_at":"2026-10-08T08:00:30Z"}`,
		`{"scan_id":"` + scanID + `","kind":"tls","user_id":null,"error":"x","completed_at":"2026-10-08T08:00:30Z"}`,
		`{"user_id":"` + userID + `","kind":"wallet","endpoint":null,"status":"success"}`,
		`{"user_id":"` + userID + `","kind":"wallet","status":""}`,
		`{"scan_id":"` + scanID + `","kind":"tls","user_id":"` + userID + `","error":"","completed_at":"2026-10-08T08:00:30Z"}`,
	} {
		reject(t, "shape", []byte(body), startedErr)
		reject(t, "shape", []byte(body), failedErr)
		reject(t, "shape", []byte(body), readyErr)
	}
}

func startedErr(b []byte) error   { _, err := DecodeStarted(b); return err }
func completedErr(b []byte) error { _, err := DecodeCompleted(b); return err }

func mustCompleted(t *testing.T, raw []byte) Completed {
	t.Helper()
	msg, err := DecodeCompleted(raw)
	if err != nil {
		t.Fatal(err)
	}
	return msg
}

func failedErr(b []byte) error { _, err := DecodeFailed(b); return err }
func readyErr(b []byte) error  { _, err := DecodeReady(b); return err }

func roundTrip[T any](t *testing.T, name string, raw []byte, decode func([]byte) (T, error)) {
	t.Helper()
	msg, err := decode(raw)
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
}

func reject(t *testing.T, name string, raw []byte, decode func([]byte) error) {
	t.Helper()
	if err := decode(raw); err == nil || !errors.Is(err, ErrMalformed) {
		t.Fatalf("%s: %v", name, err)
	}
}
