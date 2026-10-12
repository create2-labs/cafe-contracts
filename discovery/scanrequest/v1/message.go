package v1

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Wallet is one wallet scan request: scan_id, user_id, and address.
// It is the current WalletScanMessage. Address is an opaque non-empty string.
type Wallet struct {
	ScanID  string `json:"scan_id"`
	UserID  string `json:"user_id"`
	Address string `json:"address"`
}

// TLS is one TLS scan request: scan_id, user_id, endpoint, and is_default.
// It is the current TLSScanMessage. is_default is false when the field is
// absent. Endpoint is an opaque non-empty string.
type TLS struct {
	ScanID    string `json:"scan_id"`
	UserID    string `json:"user_id"`
	Endpoint  string `json:"endpoint"`
	IsDefault bool   `json:"is_default"`
}

// DecodeWallet unmarshals and validates a wallet scan request.
// A malformed body yields a zero Wallet.
func DecodeWallet(raw []byte) (Wallet, error) {
	var msg Wallet
	if err := json.Unmarshal(raw, &msg); err != nil {
		if errors.Is(err, ErrMalformed) {
			return Wallet{}, err
		}
		return Wallet{}, fmt.Errorf("%w: json", ErrMalformed)
	}
	if err := msg.Validate(); err != nil {
		return Wallet{}, err
	}
	return msg, nil
}

// DecodeTLS unmarshals and validates a TLS scan request.
// A malformed body yields a zero TLS.
func DecodeTLS(raw []byte) (TLS, error) {
	var msg TLS
	if err := json.Unmarshal(raw, &msg); err != nil {
		if errors.Is(err, ErrMalformed) {
			return TLS{}, err
		}
		return TLS{}, fmt.Errorf("%w: json", ErrMalformed)
	}
	if err := msg.Validate(); err != nil {
		return TLS{}, err
	}
	return msg, nil
}

// UnmarshalJSON rejects a non-object body and a field that is not a JSON string.
// A UUID string is stored in canonical lowercase form.
func (w *Wallet) UnmarshalJSON(data []byte) error {
	*w = Wallet{}
	raw, err := objectFields(data)
	if err != nil {
		return err
	}
	if w.ScanID, err = uuidField(raw, "scan_id"); err != nil {
		return err
	}
	if w.UserID, err = uuidField(raw, "user_id"); err != nil {
		return err
	}
	w.Address, err = requiredString(raw, "address")
	return err
}

// Validate checks the canonical UUID fields and a non-empty address.
func (w Wallet) Validate() error {
	if err := validateID("scan_id", w.ScanID); err != nil {
		return err
	}
	if err := validateID("user_id", w.UserID); err != nil {
		return err
	}
	if w.Address == "" {
		return fmt.Errorf("%w: address", ErrMalformed)
	}
	return nil
}

// UnmarshalJSON rejects a non-object body, a non-string identifier or endpoint,
// and an is_default that is not a JSON bool. A missing is_default is false.
func (m *TLS) UnmarshalJSON(data []byte) error {
	*m = TLS{}
	raw, err := objectFields(data)
	if err != nil {
		return err
	}
	if m.ScanID, err = uuidField(raw, "scan_id"); err != nil {
		return err
	}
	if m.UserID, err = uuidField(raw, "user_id"); err != nil {
		return err
	}
	if m.Endpoint, err = requiredString(raw, "endpoint"); err != nil {
		return err
	}
	m.IsDefault, err = boolField(raw, "is_default")
	return err
}

// Validate checks the canonical UUID fields and a non-empty endpoint.
// is_default may be true or false.
func (m TLS) Validate() error {
	if err := validateID("scan_id", m.ScanID); err != nil {
		return err
	}
	if err := validateID("user_id", m.UserID); err != nil {
		return err
	}
	if m.Endpoint == "" {
		return fmt.Errorf("%w: endpoint", ErrMalformed)
	}
	return nil
}
