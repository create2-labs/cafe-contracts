package address

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/sha3"
)

const (
	addressPrefix = "0x"
	addressLength = 40
)

var (
	// ErrInvalidAddress indicates that an address does not satisfy EVM hex shape.
	ErrInvalidAddress = errors.New("invalid address")
)

// IsValidHexAddress returns true when input is a valid 20-byte EVM hex address.
func IsValidHexAddress(in string) bool {
	_, err := normalizeNoErrorPrefix(in)
	return err == nil
}

// NormalizeAddress converts a valid EVM hex address to canonical lowercase.
func NormalizeAddress(in string) (string, error) {
	lower, err := normalizeNoErrorPrefix(in)
	if err != nil {
		return "", err
	}
	return addressPrefix + lower, nil
}

// EqualAddress compares two addresses using canonical normalized form.
func EqualAddress(a, b string) bool {
	na, err := normalizeNoErrorPrefix(a)
	if err != nil {
		return false
	}
	nb, err := normalizeNoErrorPrefix(b)
	if err != nil {
		return false
	}
	return na == nb
}

// ToChecksumEIP55 returns the EIP-55 mixed-case representation of an address.
func ToChecksumEIP55(in string) (string, error) {
	lower, err := normalizeNoErrorPrefix(in)
	if err != nil {
		return "", err
	}

	hashHex := keccak256Hex(lower)
	out := make([]byte, addressLength)
	for i := 0; i < addressLength; i++ {
		c := lower[i]
		if c >= '0' && c <= '9' {
			out[i] = c
			continue
		}
		if hexNibble(hashHex[i]) >= 8 {
			out[i] = c - 32
		} else {
			out[i] = c
		}
	}
	return addressPrefix + string(out), nil
}

func normalizeNoErrorPrefix(in string) (string, error) {
	if len(in) != addressLength+2 && len(in) != addressLength {
		return "", fmt.Errorf("%w: expected 40 hex chars", ErrInvalidAddress)
	}

	raw := in
	if strings.HasPrefix(raw, "0x") || strings.HasPrefix(raw, "0X") {
		raw = raw[2:]
	}
	if len(raw) != addressLength {
		return "", fmt.Errorf("%w: expected 40 hex chars", ErrInvalidAddress)
	}

	lower := strings.ToLower(raw)
	for i := 0; i < addressLength; i++ {
		c := lower[i]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') {
			continue
		}
		return "", fmt.Errorf("%w: non-hex character at index %d", ErrInvalidAddress, i)
	}
	return lower, nil
}

func keccak256Hex(in string) string {
	h := sha3.NewLegacyKeccak256()
	_, _ = h.Write([]byte(in))
	sum := h.Sum(nil)

	const chars = "0123456789abcdef"
	out := make([]byte, len(sum)*2)
	for i := 0; i < len(sum); i++ {
		out[i*2] = chars[sum[i]>>4]
		out[i*2+1] = chars[sum[i]&0x0f]
	}
	return string(out)
}

func hexNibble(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return 10 + c - 'a'
	default:
		return 0
	}
}
