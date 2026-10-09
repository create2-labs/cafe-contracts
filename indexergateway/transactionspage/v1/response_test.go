package v1

import (
	"encoding/json"
	"errors"
	"io/fs"
	"strings"
	"testing"
)

func TestResponseVectors(t *testing.T) {
	entries, err := fs.ReadDir(Vectors, "testdata")
	if err != nil {
		t.Fatal(err)
	}
	var sawPage, sawEmpty, sawError, sawBad bool
	for _, entry := range entries {
		name := entry.Name()
		raw, err := fs.ReadFile(Vectors, "testdata/"+name)
		if err != nil {
			t.Fatal(err)
		}
		switch {
		case strings.HasPrefix(name, "ok_response"), strings.HasPrefix(name, "ok_error"):
			var resp Response
			if err := json.Unmarshal(raw, &resp); err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			out, err := json.Marshal(resp)
			if err != nil {
				t.Fatalf("%s marshal: %v", name, err)
			}
			if string(out) != strings.TrimSpace(string(raw)) {
				t.Fatalf("%s round trip:\n%s\nvs\n%s", name, out, raw)
			}
			switch {
			case name == "ok_response_page.json":
				sawPage = true
				if len(resp.Transactions) != 2 || resp.Transactions[0].Hash != "0x02" || resp.Transactions[1].Hash != "0x01" {
					t.Fatalf("order changed: %+v", resp.Transactions)
				}
				if resp.Transactions[1].From != "not-an-address" {
					t.Fatalf("from rewritten: %+v", resp.Transactions[1])
				}
				if strings.Contains(string(out), `"error"`) {
					t.Fatal("page contains error")
				}
			case name == "ok_response_empty.json":
				sawEmpty = true
				if resp.Transactions == nil || len(resp.Transactions) != 0 || resp.NextCursor != "" {
					t.Fatalf("empty page: %+v", resp)
				}
				if !strings.Contains(string(out), `"transactions":[]`) {
					t.Fatalf("empty page serialized as %s", out)
				}
			default:
				sawError = true
				if resp.Error == nil || !ValidCode(resp.Error.Code) {
					t.Fatalf("%s: %+v", name, resp.Error)
				}
				if strings.Contains(string(out), "transactions") || strings.Contains(string(out), "next_cursor") {
					t.Fatalf("%s contains page fields: %s", name, out)
				}
			}
		case strings.HasPrefix(name, "bad_response"):
			sawBad = true
			var resp Response
			if err := json.Unmarshal(raw, &resp); err == nil || !errors.Is(err, ErrResponse) {
				t.Fatalf("%s: got %v", name, err)
			}
		}
	}
	if !sawPage || !sawEmpty || !sawError || !sawBad {
		t.Fatalf("fixtures missing: page=%v empty=%v error=%v bad=%v", sawPage, sawEmpty, sawError, sawBad)
	}
}

func TestPageOrderIsPreserved(t *testing.T) {
	resp := Response{
		RequestID: validID,
		Transactions: []Transaction{
			{Hash: "0x02", From: "0x1"},
			{Hash: "0x01", From: "0x2"},
		},
		NextCursor: "opaque",
	}
	out, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Index(string(out), "0x02") > strings.Index(string(out), "0x01") {
		t.Fatalf("sorted: %s", out)
	}
	var again Response
	if err := json.Unmarshal(out, &again); err != nil {
		t.Fatal(err)
	}
	if again.Transactions[0].Hash != "0x02" || again.Transactions[1].Hash != "0x01" {
		t.Fatalf("order: %+v", again.Transactions)
	}
}

func TestNilTransactionsMarshalAsEmptyArray(t *testing.T) {
	out, err := json.Marshal(Response{RequestID: validID})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"transactions":[]`) || !strings.Contains(string(out), `"next_cursor":""`) {
		t.Fatalf("%s", out)
	}
}

func TestResponseBounds(t *testing.T) {
	ok100 := pageWithCount(MaxTransactions, strings.Repeat("n", MaxCursorBytes))
	if err := ok100.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, err := json.Marshal(ok100); err != nil {
		t.Fatal(err)
	}

	tooMany := pageWithCount(MaxTransactions+1, "")
	if err := tooMany.Validate(); !errors.Is(err, ErrResponse) {
		t.Fatalf("101 transactions: %v", err)
	}
	if _, err := json.Marshal(tooMany); !errors.Is(err, ErrResponse) {
		t.Fatalf("marshal 101: %v", err)
	}

	longCursor := Response{RequestID: validID, NextCursor: strings.Repeat("n", MaxCursorBytes+1)}
	if err := longCursor.Validate(); !errors.Is(err, ErrResponse) {
		t.Fatalf("cursor: %v", err)
	}
	raw := `{"request_id":"` + validID + `","transactions":[],"next_cursor":"` + strings.Repeat("n", MaxCursorBytes+1) + `"}`
	var resp Response
	err := json.Unmarshal([]byte(raw), &resp)
	if !errors.Is(err, ErrResponse) {
		t.Fatalf("json cursor: %v", err)
	}
	if strings.Contains(err.Error(), strings.Repeat("n", 32)) {
		t.Fatal("error contains cursor")
	}
}

func TestErrorMessageBounds(t *testing.T) {
	codes := []string{
		CodeInvalidRequest,
		CodeUnsupportedChain,
		CodeNoProvider,
		CodeProviderError,
		CodeDeadlineExceeded,
		CodeRateLimitExceeded,
	}
	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			for _, n := range []int{1, MaxErrorMessageBytes} {
				resp := Response{RequestID: validID, Error: &Error{Code: code, Message: strings.Repeat("m", n)}}
				if code == CodeInvalidRequest {
					resp.RequestID = ""
				}
				out, err := json.Marshal(resp)
				if err != nil {
					t.Fatalf("len %d: %v", n, err)
				}
				var again Response
				if err := json.Unmarshal(out, &again); err != nil {
					t.Fatal(err)
				}
				if again.Error == nil || again.Error.Code != code || len(again.Error.Message) != n {
					t.Fatalf("%+v", again)
				}
			}
			tooLong := Response{RequestID: validID, Error: &Error{Code: code, Message: strings.Repeat("m", MaxErrorMessageBytes+1)}}
			if err := tooLong.Validate(); !errors.Is(err, ErrResponse) {
				t.Fatalf("129: %v", err)
			}
			if code != CodeInvalidRequest {
				emptyID := Response{Error: &Error{Code: code, Message: "x"}}
				if err := emptyID.Validate(); !errors.Is(err, ErrResponse) {
					t.Fatalf("empty id: %v", err)
				}
				missing := `{"error":{"code":"` + code + `","message":"x"}}`
				var resp Response
				if err := json.Unmarshal([]byte(missing), &resp); !errors.Is(err, ErrResponse) {
					t.Fatalf("missing id: %v", err)
				}
			}
		})
	}
}

func TestInvalidRequestEmptyID(t *testing.T) {
	resp := Response{Error: &Error{Code: CodeInvalidRequest, Message: "invalid request"}}
	out, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"request_id":""`) {
		t.Fatalf("%s", out)
	}
	var again Response
	if err := json.Unmarshal(out, &again); err != nil {
		t.Fatal(err)
	}
	missing := []byte(`{"error":{"code":"invalid_request","message":"invalid request"}}`)
	if err := json.Unmarshal(missing, &again); err != nil {
		t.Fatal(err)
	}
	if again.RequestID != "" || again.Error.Code != CodeInvalidRequest {
		t.Fatalf("%+v", again)
	}
}

func TestResponseRequestIDRules(t *testing.T) {
	id128 := strings.Repeat("e", MaxRequestIDBytes)
	page := Response{RequestID: id128, Transactions: []Transaction{}}
	if err := page.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := page.ValidateForRequest(validID); !errors.Is(err, ErrResponse) {
		t.Fatalf("mismatch: %v", err)
	}
	page.RequestID = validID
	if err := page.ValidateForRequest(validID); err != nil {
		t.Fatal(err)
	}

	id129 := strings.Repeat("e", MaxRequestIDBytes+1)
	raw := `{"request_id":"` + id129 + `","transactions":[],"next_cursor":""}`
	var resp Response
	err := json.Unmarshal([]byte(raw), &resp)
	if !errors.Is(err, ErrResponse) {
		t.Fatalf("129: %v", err)
	}
	if strings.Contains(err.Error(), id129) {
		t.Fatal("error contains request_id")
	}

	badTypes := []string{
		`{"request_id":1,"transactions":[],"next_cursor":""}`,
		`{"request_id":null,"transactions":[],"next_cursor":""}`,
		`{"request_id":true,"error":{"code":"invalid_request","message":"x"}}`,
		`{"request_id":"` + validID + `","transactions":null,"next_cursor":""}`,
		`{"request_id":"` + validID + `","transactions":[],"next_cursor":null}`,
		`{"request_id":"` + validID + `","transactions":[],"next_cursor":1}`,
		`not-json`,
	}
	for _, body := range badTypes {
		if _, err := DecodeResponse([]byte(body)); !errors.Is(err, ErrResponse) {
			t.Fatalf("%s: %v", body, err)
		}
	}
}

func TestValidateForRequestRejectsEmptyID(t *testing.T) {
	resp := Response{Error: &Error{Code: CodeInvalidRequest, Message: "invalid request"}}
	if err := resp.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := resp.ValidateForRequest(validID); !errors.Is(err, ErrResponse) {
		t.Fatalf("reader accepted empty id: %v", err)
	}
}

func TestValidCode(t *testing.T) {
	if !ValidCode(CodeRateLimitExceeded) || ValidCode("timeout") || ValidCode("") {
		t.Fatal("ValidCode")
	}
}

func pageWithCount(n int, cursor string) Response {
	txs := make([]Transaction, n)
	for i := range txs {
		txs[i] = Transaction{Hash: "0x01", From: "0x02"}
	}
	return Response{RequestID: validID, Transactions: txs, NextCursor: cursor}
}
