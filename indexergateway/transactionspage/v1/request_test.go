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

func TestSubject(t *testing.T) {
	if Subject != "cafe.indexer.gateway.transactions.page.v1" {
		t.Fatalf("subject = %q", Subject)
	}
}

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

func TestDecodeRequestAcceptsBoundary(t *testing.T) {
	cursor := strings.Repeat("é", 4096) // 8192 bytes, content is not interpreted
	if len(cursor) != MaxCursorBytes {
		t.Fatalf("cursor bytes = %d", len(cursor))
	}
	req := validRequest()
	req.RequestID = "6F1C0C1E-9B2A-4D3E-8F7A-1C2B3D4E5F60"
	req.Address = "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
	req.Cursor = cursor
	req.Limit = MinLimit
	req.Deadline = "2026-10-08T10:00:30+02:00"
	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Cursor != cursor || got.Address != req.Address || got.Limit != MinLimit {
		t.Fatalf("changed request: %+v", got)
	}

	req.Limit = MaxLimit
	req.Deadline = "2026-10-08T08:00:30Z"
	req.Cursor = `{"weird":true} /\n`
	raw, err = json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	got, err = DecodeRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Cursor != req.Cursor {
		t.Fatalf("cursor interpreted: %q", got.Cursor)
	}
}

func TestDecodeRequestRejectsShape(t *testing.T) {
	base := validRequest()
	tooLong := strings.Repeat("x", MaxCursorBytes+1)
	cases := []struct {
		name string
		body string
	}{
		{name: "chain missing", body: replace(t, base, `"chain_id":1,`, "")},
		{name: "chain string", body: replace(t, base, `"chain_id":1`, `"chain_id":"1"`)},
		{name: "chain fraction", body: replace(t, base, `"chain_id":1`, `"chain_id":1.0`)},
		{name: "chain overflow", body: replace(t, base, `"chain_id":1`, `"chain_id":9223372036854775808`)},
		{name: "limit fraction", body: replace(t, base, `"limit":100`, `"limit":100.0`)},
		{name: "cursor number", body: replace(t, base, `"cursor":""`, `"cursor":1`)},
		{name: "cursor null", body: replace(t, base, `"cursor":""`, `"cursor":null`)},
		{name: "deadline null", body: replace(t, base, `"deadline":"2026-10-08T08:00:30.000000000Z"`, `"deadline":null`)},
		{name: "cursor too long", body: requestWithCursor(t, tooLong)},
		{name: "deadline empty", body: replace(t, base, `"deadline":"2026-10-08T08:00:30.000000000Z"`, `"deadline":""`)},
		{name: "braced uuid", body: replace(t, base, validID, "{"+validID+"}")},
		{name: "unreadable", body: "{"},
		{name: "array", body: "[]"},
	}
	for _, tt := range cases {
		if tt.body == "" {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeRequest([]byte(tt.body))
			if err == nil || !errors.Is(err, ErrRequest) {
				t.Fatalf("got %v", err)
			}
			if strings.Contains(err.Error(), tooLong) {
				t.Fatal("error contains cursor")
			}
		})
	}
}

func TestOmittedCursorIsEmpty(t *testing.T) {
	raw := []byte(`{"request_id":"` + validID + `","address":"` + validAddress + `","chain_id":1,"limit":100,"deadline":"2026-10-08T08:00:30.000000000Z"}`)
	got, err := DecodeRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Cursor != "" {
		t.Fatalf("cursor = %q", got.Cursor)
	}
}

func TestBoundedRequestID(t *testing.T) {
	entries := []struct {
		name string
		want string
	}{
		{name: "echo_bounded.json", want: "not-a-uuid"},
		{name: "echo_empty.json", want: ""},
		{name: "echo_missing.json", want: ""},
		{name: "echo_number.json", want: ""},
		{name: "echo_unreadable.txt", want: ""},
		{name: "ok_request.json", want: validID},
		{name: "bad_request_id.json", want: "not-a-uuid"},
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
		{name: "one byte", raw: `{"request_id":"x"}`, want: "x"},
		{name: "128", raw: `{"request_id":"` + id128 + `"}`, want: id128},
		{name: "129", raw: `{"request_id":"` + id129 + `"}`, want: ""},
		{name: "null", raw: `{"request_id":null}`, want: ""},
		{name: "bool", raw: `{"request_id":true}`, want: ""},
		{name: "object", raw: `{"request_id":{"id":"a"}}`, want: ""},
		{name: "array", raw: `{"request_id":["a"]}`, want: ""},
		{name: "json null", raw: `null`, want: ""},
		{name: "json string", raw: `"abc"`, want: ""},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := BoundedRequestID([]byte(tt.raw))
			if got != tt.want {
				t.Fatalf("got %q", got)
			}
			if strings.Contains(tt.raw, id129) && strings.Contains(got, "c") {
				t.Fatal("echoed a prefix of the oversized id")
			}
		})
	}
}

func TestRejectedRequestDoesNotLeakIdentifier(t *testing.T) {
	long := strings.Repeat("d", 200)
	raw := []byte(`{"request_id":"` + long + `","address":"0xabc","chain_id":1,"cursor":"","limit":1,"deadline":"2026-10-08T08:00:30Z"}`)
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
	return Request{
		RequestID: validID,
		Address:   validAddress,
		ChainID:   1,
		Cursor:    "",
		Limit:     100,
		Deadline:  "2026-10-08T08:00:30.000000000Z",
	}
}

func replace(t *testing.T, req Request, old, new string) string {
	t.Helper()
	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, old) {
		t.Fatalf("pattern %q not in %s", old, body)
	}
	return strings.Replace(body, old, new, 1)
}

func requestWithCursor(t *testing.T, cursor string) string {
	t.Helper()
	req := validRequest()
	req.Cursor = cursor
	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
