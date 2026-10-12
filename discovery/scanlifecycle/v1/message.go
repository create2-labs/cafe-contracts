package v1

import (
	"encoding/json"
	"errors"
	"fmt"
)

type validator interface{ Validate() error }

func finish[T validator](msg T, err error) (T, error) {
	var zero T
	if err != nil {
		if errors.Is(err, ErrMalformed) {
			return zero, err
		}
		return zero, fmt.Errorf("%w: json", ErrMalformed)
	}
	if err = msg.Validate(); err != nil {
		return zero, err
	}
	return msg, nil
}

// Started is one scan.started message. endpoint and address are omitted when empty.
type Started struct {
	ScanID    string `json:"scan_id"`
	Kind      string `json:"kind"`
	UserID    string `json:"user_id"`
	StartedAt string `json:"started_at"`
	Endpoint  string `json:"endpoint,omitempty"`
	Address   string `json:"address,omitempty"`
}

// Failed is one scan.failed message. endpoint and address are omitted when empty.
type Failed struct {
	ScanID      string `json:"scan_id"`
	Kind        string `json:"kind"`
	UserID      string `json:"user_id"`
	Error       string `json:"error"`
	CompletedAt string `json:"completed_at"`
	Endpoint    string `json:"endpoint,omitempty"`
	Address     string `json:"address,omitempty"`
}

// Ready is one scan.ready message: user_id, kind, endpoint, address, and status.
type Ready struct {
	UserID   string `json:"user_id"`
	Kind     string `json:"kind"`
	Endpoint string `json:"endpoint,omitempty"`
	Address  string `json:"address,omitempty"`
	Status   string `json:"status"`
}

// DecodeStarted unmarshals and validates scan.started. A malformed body yields a zero Started.
func DecodeStarted(raw []byte) (Started, error) {
	var msg Started
	return finish(msg, json.Unmarshal(raw, &msg))
}

// DecodeFailed unmarshals and validates scan.failed. A malformed body yields a zero Failed.
func DecodeFailed(raw []byte) (Failed, error) {
	var msg Failed
	return finish(msg, json.Unmarshal(raw, &msg))
}

// DecodeReady unmarshals and validates scan.ready. A malformed body yields a zero Ready.
func DecodeReady(raw []byte) (Ready, error) {
	var msg Ready
	return finish(msg, json.Unmarshal(raw, &msg))
}

func bind(data []byte, withScanID bool) (head, map[string]json.RawMessage, error) {
	raw, err := objectFields(data)
	if err != nil {
		return head{}, nil, err
	}
	h, err := readHead(raw, withScanID)
	if err != nil {
		return head{}, nil, err
	}
	return h, raw, nil
}

func (m *Started) UnmarshalJSON(data []byte) error {
	*m = Started{}
	h, raw, err := bind(data, true)
	if err != nil {
		return err
	}
	applyHead(&m.ScanID, &m.Kind, &m.UserID, &m.Endpoint, &m.Address, h)
	m.StartedAt, err = requiredString(raw, "started_at")
	return err
}

func (m Started) Validate() error {
	if err := validateScan(m.ScanID, m.Kind, m.UserID); err != nil {
		return err
	}
	return text("started_at", m.StartedAt)
}

func (m *Failed) UnmarshalJSON(data []byte) error {
	*m = Failed{}
	h, raw, err := bind(data, true)
	if err != nil {
		return err
	}
	applyHead(&m.ScanID, &m.Kind, &m.UserID, &m.Endpoint, &m.Address, h)
	if m.Error, err = requiredString(raw, "error"); err != nil {
		return err
	}
	m.CompletedAt, err = requiredString(raw, "completed_at")
	return err
}

func (m Failed) Validate() error {
	if err := validateScan(m.ScanID, m.Kind, m.UserID); err != nil {
		return err
	}
	if err := text("error", m.Error); err != nil {
		return err
	}
	return text("completed_at", m.CompletedAt)
}

func (m *Ready) UnmarshalJSON(data []byte) error {
	*m = Ready{}
	h, raw, err := bind(data, false)
	if err != nil {
		return err
	}
	m.UserID, m.Kind, m.Endpoint, m.Address = h.UserID, h.Kind, h.Endpoint, h.Address
	m.Status, err = requiredString(raw, "status")
	return err
}

func (m Ready) Validate() error {
	if err := validateID("user_id", m.UserID); err != nil {
		return err
	}
	if err := text("kind", m.Kind); err != nil {
		return err
	}
	return text("status", m.Status)
}
