package v1

import (
	"encoding/json"
	"errors"
	"io/fs"
	"strings"
	"testing"
)

const validCreator = "0x742d35cc6634c0532925a3b844bc454e4438f44e"
const validTxHash = "0x0000000000000000000000000000000000000000000000000000000000000001"

// CA3.2: the seven codes are known, anything else is not.
func TestValidCode(t *testing.T) {
	codes := []string{
		CodeInvalidRequest, CodeUnsupportedChain, CodeNoProvider, CodeProviderError,
		CodeDeadlineExceeded, CodeRateLimitExceeded, CodeUnsupportedOperation,
	}
	for _, code := range codes {
		if !ValidCode(code) {
			t.Fatalf("%q rejected", code)
		}
	}
	for _, code := range []string{"", "timeout", "INVALID_REQUEST", " invalid_request"} {
		if ValidCode(code) {
			t.Fatalf("%q accepted", code)
		}
	}
}

// CA3.2, CA3.3: the three exclusive shapes accepted by Validate.
func TestValidateAccepts(t *testing.T) {
	cases := []struct {
		name string
		resp Response
	}{
		{"found with tx hash", Response{RequestID: validID, Creator: validCreator, CreationTxHash: validTxHash}},
		{"found without tx hash", Response{RequestID: validID, Creator: validCreator}},
		{"found with 0X creator", Response{RequestID: validID, Creator: "0X742d35cc6634c0532925a3b844bc454e4438f44e"}},
		{"not found is a success", Response{RequestID: validID}},
		{"error min message", Response{RequestID: validID, Error: &Error{Code: CodeProviderError, Message: "m"}}},
		{"error max message", Response{RequestID: validID, Error: &Error{Code: CodeUnsupportedOperation, Message: strings.Repeat("m", MaxErrorMessageBytes)}}},
		{"invalid_request without id", Response{Error: &Error{Code: CodeInvalidRequest, Message: "invalid request"}}},
		{"invalid_request with id", Response{RequestID: validID, Error: &Error{Code: CodeInvalidRequest, Message: "invalid request"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.resp.Validate(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// CA3.2, CA3.3: invalid envelopes wrap ErrResponse and name only the field.
func TestValidateRejects(t *testing.T) {
	okErr := func(code string) *Error { return &Error{Code: code, Message: "x"} }
	cases := []struct {
		name  string
		resp  Response
		field string
	}{
		{"mixed creator and error", Response{RequestID: validID, Creator: validCreator, Error: okErr(CodeProviderError)}, ""},
		{"mixed tx hash and error", Response{RequestID: validID, CreationTxHash: validTxHash, Error: okErr(CodeProviderError)}, ""},
		{"tx hash without creator", Response{RequestID: validID, CreationTxHash: validTxHash}, "creation_tx_hash"},
		{"creator no prefix", Response{RequestID: validID, Creator: "742d35cc6634c0532925a3b844bc454e4438f44e"}, "creator"},
		{"creator invalid hex", Response{RequestID: validID, Creator: "0xzz2d35cc6634c0532925a3b844bc454e4438f44e"}, "creator"},
		{"creator short", Response{RequestID: validID, Creator: "0xabc"}, "creator"},
		{"unknown code", Response{RequestID: validID, Error: okErr("timeout")}, "error.code"},
		{"empty code", Response{RequestID: validID, Error: okErr("")}, "error.code"},
		{"empty message", Response{RequestID: validID, Error: &Error{Code: CodeProviderError}}, "error.message"},
		{"message too long", Response{RequestID: validID, Error: &Error{Code: CodeProviderError, Message: strings.Repeat("m", MaxErrorMessageBytes+1)}}, "error.message"},
		{"empty id on found", Response{Creator: validCreator}, "request_id"},
		{"empty id on not found", Response{}, "request_id"},
		{"empty id on provider_error", Response{Error: okErr(CodeProviderError)}, "request_id"},
		{"id too long", Response{RequestID: strings.Repeat("e", MaxRequestIDBytes+1)}, "request_id"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.resp.Validate()
			if !errors.Is(err, ErrResponse) {
				t.Fatalf("got %v", err)
			}
			if tc.field != "" && !strings.Contains(err.Error(), tc.field) {
				t.Fatalf("error does not name %s: %v", tc.field, err)
			}
			for _, v := range []string{validCreator, validTxHash, tc.resp.Creator} {
				if v != "" && strings.Contains(err.Error(), v) {
					t.Fatalf("error quotes a value: %v", err)
				}
			}
		})
	}
}

// CA3.2: the reader checks the echo of its own request_id.
func TestValidateForRequest(t *testing.T) {
	resp := Response{RequestID: validID, Creator: validCreator}
	if err := resp.ValidateForRequest(validID); err != nil {
		t.Fatal(err)
	}
	err := resp.ValidateForRequest("other-id")
	if !errors.Is(err, ErrResponse) || !strings.Contains(err.Error(), "request_id") {
		t.Fatalf("mismatch: %v", err)
	}
	invalid := Response{RequestID: validID, CreationTxHash: validTxHash}
	if err := invalid.ValidateForRequest(validID); !errors.Is(err, ErrResponse) {
		t.Fatalf("invalid envelope: %v", err)
	}
	noID := Response{Error: &Error{Code: CodeInvalidRequest, Message: "invalid request"}}
	if err := noID.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := noID.ValidateForRequest(validID); !errors.Is(err, ErrResponse) {
		t.Fatalf("reader accepted empty id: %v", err)
	}
}

// badResponseFields maps each bad fixture to the field its error must name.
var badResponseFields = map[string]string{
	"bad_response_mixed.json":                   "mixed",
	"bad_response_tx_hash_without_creator.json": "creation_tx_hash",
	"bad_response_creator_unprefixed.json":      "creator",
	"bad_response_unknown_code.json":            "error.code",
	"bad_response_message_empty.json":           "error.message",
	"bad_response_message_missing.json":         "error.message",
	"bad_response_request_id_missing.json":      "request_id",
	"bad_response_request_id_number.json":       "request_id",
	"bad_response_creator_number.json":          "creator",
	"bad_response_tx_hash_number.json":          "creation_tx_hash",
	"bad_response_error_not_object.json":        "error",
	"bad_response_creator_null.json":            "creator",
	"bad_response_creator_empty.json":           "creator",
	"bad_response_tx_hash_null.json":            "creation_tx_hash",
	"bad_response_error_null.json":              "error",
}

// CA3.2, CA3.3: every response fixture is exercised.
func TestResponseVectors(t *testing.T) {
	entries, err := fs.ReadDir(Vectors, "testdata")
	if err != nil {
		t.Fatal(err)
	}
	var sawFound, sawNoTx, sawNotFound, sawError, sawBad int
	sawCodes := map[string]int{}
	for _, entry := range entries {
		name := entry.Name()
		raw, err := fs.ReadFile(Vectors, "testdata/"+name)
		if err != nil {
			t.Fatal(err)
		}
		switch {
		case strings.HasPrefix(name, "ok_response"), strings.HasPrefix(name, "ok_error"):
			resp, err := DecodeResponse(raw)
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			out, err := json.Marshal(resp)
			if err != nil {
				t.Fatalf("%s marshal: %v", name, err)
			}
			if string(out) != strings.TrimSpace(string(raw)) {
				t.Fatalf("%s round trip:\n%s\nvs\n%s", name, out, raw)
			}
			switch name {
			case "ok_response_found.json":
				sawFound++
				if resp.Creator != validCreator || resp.CreationTxHash != validTxHash || resp.Error != nil {
					t.Fatalf("%s: %+v", name, resp)
				}
			case "ok_response_found_no_tx.json":
				sawNoTx++
				if resp.Creator != validCreator || resp.CreationTxHash != "" {
					t.Fatalf("%s: %+v", name, resp)
				}
			case "ok_response_not_found.json":
				sawNotFound++
				if resp.Creator != "" || resp.CreationTxHash != "" || resp.Error != nil {
					t.Fatalf("%s: %+v", name, resp)
				}
			default:
				sawError++
				if resp.Error != nil {
					sawCodes[resp.Error.Code]++
				}
				if resp.Error == nil || !ValidCode(resp.Error.Code) || resp.Creator != "" || resp.CreationTxHash != "" {
					t.Fatalf("%s: %+v", name, resp)
				}
			}
		case strings.HasPrefix(name, "bad_response"):
			sawBad++
			field, ok := badResponseFields[name]
			if !ok {
				t.Fatalf("%s: no expected field", name)
			}
			resp, err := DecodeResponse(raw)
			if !errors.Is(err, ErrResponse) || !strings.Contains(err.Error(), field) {
				t.Fatalf("%s: got %v, want field %s", name, err, field)
			}
			if resp != (Response{}) {
				t.Fatalf("%s: non-zero result %+v", name, resp)
			}
			for _, v := range []string{validCreator, validTxHash, "742d35cc", "failed"} {
				if strings.Contains(err.Error(), v) {
					t.Fatalf("%s: error quotes a value: %v", name, err)
				}
			}
		}
	}
	if sawFound != 1 || sawNoTx != 1 || sawNotFound != 1 || sawError != 7 || sawBad != len(badResponseFields) {
		t.Fatalf("fixtures: found=%d noTx=%d notFound=%d error=%d bad=%d", sawFound, sawNoTx, sawNotFound, sawError, sawBad)
	}
	for _, code := range []string{
		CodeInvalidRequest, CodeUnsupportedChain, CodeNoProvider, CodeProviderError,
		CodeDeadlineExceeded, CodeRateLimitExceeded, CodeUnsupportedOperation,
	} {
		if sawCodes[code] != 1 {
			t.Fatalf("error code %s seen %d times, want 1", code, sawCodes[code])
		}
	}
}

// CA3.2: a 0X creator is accepted and delivered in lowercase 0x form.
func TestDecodeResponseLowercasesCreator(t *testing.T) {
	raw := `{"request_id":"` + validID + `","creator":"0X742D35CC6634C0532925A3B844BC454E4438F44E"}`
	resp, err := DecodeResponse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Creator != validCreator {
		t.Fatalf("creator = %q", resp.Creator)
	}
}

// CA3.2: unknown fields are ignored; a present but empty tx hash is not valid.
func TestDecodeResponseFieldRules(t *testing.T) {
	ok := `{"request_id":"` + validID + `","creator":"` + validCreator + `","extra":{"a":1}}`
	resp, err := DecodeResponse([]byte(ok))
	if err != nil || resp.Creator != validCreator {
		t.Fatalf("unknown field: %+v %v", resp, err)
	}
	empty := `{"request_id":"` + validID + `","creator":"` + validCreator + `","creation_tx_hash":""}`
	if _, err := DecodeResponse([]byte(empty)); !errors.Is(err, ErrResponse) || !strings.Contains(err.Error(), "creation_tx_hash") {
		t.Fatalf("empty tx hash: %v", err)
	}
	for _, body := range []string{"not-json", "", "[]", "null"} {
		if _, err := DecodeResponse([]byte(body)); !errors.Is(err, ErrResponse) {
			t.Fatalf("%q: %v", body, err)
		}
	}
}

// CA3.2: request_id of exactly MaxRequestIDBytes is accepted on the error
// shape; one byte more is rejected on request_id.
func TestResponseRequestIDBoundsOnError(t *testing.T) {
	body := func(n int) string {
		return `{"request_id":"` + strings.Repeat("e", n) + `","error":{"code":"provider_error","message":"x"}}`
	}
	resp, err := DecodeResponse([]byte(body(MaxRequestIDBytes)))
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.RequestID) != MaxRequestIDBytes || resp.Error == nil {
		t.Fatalf("%+v", resp)
	}
	_, err = DecodeResponse([]byte(body(MaxRequestIDBytes + 1)))
	if !errors.Is(err, ErrResponse) || !strings.Contains(err.Error(), "request_id") {
		t.Fatalf("over-long: %v", err)
	}
	if strings.Contains(err.Error(), strings.Repeat("e", 32)) {
		t.Fatal("error quotes request_id")
	}
}

// CA3.2: json.Unmarshal goes through the same decoding and validation.
func TestResponseUnmarshalJSON(t *testing.T) {
	var resp Response
	raw := `{"request_id":"` + validID + `","creator":"0X742D35CC6634C0532925A3B844BC454E4438F44E","creation_tx_hash":"` + validTxHash + `"}`
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Creator != validCreator || resp.CreationTxHash != validTxHash || resp.RequestID != validID {
		t.Fatalf("%+v", resp)
	}
	prev := resp
	err := json.Unmarshal([]byte(`{"request_id":"`+validID+`","creation_tx_hash":"`+validTxHash+`"}`), &resp)
	if !errors.Is(err, ErrResponse) || !strings.Contains(err.Error(), "creation_tx_hash") {
		t.Fatalf("got %v", err)
	}
	if resp != prev {
		t.Fatalf("target changed on failure: %+v", resp)
	}
	// encoding/json checks syntax before calling UnmarshalJSON, so the error is
	// a *json.SyntaxError, not ErrResponse; the unreadable-JSON rule is covered
	// by DecodeResponse in TestDecodeResponseFieldRules.
	if err := json.Unmarshal([]byte(`not-json`), &resp); err == nil {
		t.Fatal("syntax: expected an error")
	}
	if resp != prev {
		t.Fatalf("target changed on syntax error: %+v", resp)
	}
}

// CA3.2: json.Marshal validates; an invalid Response is an ErrResponse.
func TestResponseMarshalJSONRejectsInvalid(t *testing.T) {
	cases := []struct {
		name  string
		resp  Response
		field string
	}{
		{"tx hash without creator", Response{RequestID: validID, CreationTxHash: validTxHash}, "creation_tx_hash"},
		{"creator no prefix", Response{RequestID: validID, Creator: "742d35cc6634c0532925a3b844bc454e4438f44e"}, "creator"},
		{"mixed", Response{RequestID: validID, Creator: validCreator, Error: &Error{Code: CodeProviderError, Message: "x"}}, "mixed"},
		{"empty id on not found", Response{}, "request_id"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := json.Marshal(tc.resp)
			if !errors.Is(err, ErrResponse) || !strings.Contains(err.Error(), tc.field) {
				t.Fatalf("got %q, %v", out, err)
			}
		})
	}
}
