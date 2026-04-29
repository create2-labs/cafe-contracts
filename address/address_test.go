package address

import "testing"

func TestIsValidHexAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "lowercase", in: "0x742d35cc6634c0532925a3b844bc454e4438f44e", want: true},
		{name: "checksummed", in: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e", want: true},
		{name: "uppercase prefix", in: "0X742D35CC6634C0532925A3B844BC454E4438F44E", want: true},
		{name: "no prefix", in: "742d35cc6634c0532925a3b844bc454e4438f44e", want: true},
		{name: "too short", in: "0x1234", want: false},
		{name: "non-hex", in: "0x742d35cc6634c0532925a3b844bc454e4438f44g", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsValidHexAddress(tt.in); got != tt.want {
				t.Fatalf("IsValidHexAddress(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalizeAddress(t *testing.T) {
	t.Parallel()

	got, err := NormalizeAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44e")
	if err != nil {
		t.Fatalf("NormalizeAddress() error = %v", err)
	}
	want := "0x742d35cc6634c0532925a3b844bc454e4438f44e"
	if got != want {
		t.Fatalf("NormalizeAddress() = %q, want %q", got, want)
	}
}

func TestNormalizeAddressInvalid(t *testing.T) {
	t.Parallel()

	if _, err := NormalizeAddress("0xinvalid"); err == nil {
		t.Fatalf("NormalizeAddress() expected error for invalid input")
	}
}

func TestEqualAddress(t *testing.T) {
	t.Parallel()

	if !EqualAddress(
		"0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		"0x742d35cc6634c0532925a3b844bc454e4438f44e",
	) {
		t.Fatalf("EqualAddress should return true for same address with different casing")
	}

	if EqualAddress(
		"0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		"0x742d35cc6634c0532925a3b844bc454e4438f44f",
	) {
		t.Fatalf("EqualAddress should return false for different addresses")
	}
}

func TestToChecksumEIP55(t *testing.T) {
	t.Parallel()

	got, err := ToChecksumEIP55("0x52908400098527886e0f7030069857d2e4169ee7")
	if err != nil {
		t.Fatalf("ToChecksumEIP55() error = %v", err)
	}

	want := "0x52908400098527886E0F7030069857D2E4169EE7"
	if got != want {
		t.Fatalf("ToChecksumEIP55() = %q, want %q", got, want)
	}
}
