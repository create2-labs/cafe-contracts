package v1

import (
	"errors"
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
