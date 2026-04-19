package v01

import "testing"

func TestIsValidAlgorithmID(t *testing.T) {
	cases := []struct {
		id    string
		valid bool
	}{
		{"", false},
		{"secp256k1_ecrecover", true},
		{"mldsa44", true},
		{"hybrid_", false},
		{"hybrid_x", true},
		{"not-listed", false},
	}
	for _, tc := range cases {
		if got := IsValidAlgorithmID(tc.id); got != tc.valid {
			t.Fatalf("IsValidAlgorithmID(%q) = %v, want %v", tc.id, got, tc.valid)
		}
	}
}

func TestAccountKind_IsValid(t *testing.T) {
	valid := []AccountKind{
		AccountKindEOA,
		AccountKindERC4337SmartAccount,
		AccountKindDelegatedEOA7702,
		AccountKindContractAccount,
		AccountKindUnknown,
	}
	for _, k := range valid {
		if !k.IsValid() {
			t.Fatalf("%q should be valid", k)
		}
	}
	if AccountKind("other").IsValid() {
		t.Fatal("expected invalid")
	}
}
