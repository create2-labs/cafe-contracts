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

// TASK-CON-002: the account kinds are eoa and unknown only.
func TestAccountKind_IsValid(t *testing.T) {
	for _, k := range []AccountKind{AccountKindEOA, AccountKindUnknown} {
		if !k.IsValid() {
			t.Fatalf("%q should be valid", k)
		}
	}
	for _, k := range removedAccountKinds {
		if AccountKind(k).IsValid() {
			t.Fatalf("%q should be invalid", k)
		}
	}
	if AccountKind("other").IsValid() || AccountKind("").IsValid() {
		t.Fatal("expected invalid")
	}
}

// Removed kinds are string literals: their constants are deleted by TASK-CON-002.
var removedAccountKinds = []string{"erc4337_smart_account", "delegated_eoa_7702", "contract_account"}
