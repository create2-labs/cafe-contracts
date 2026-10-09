package v1

import (
	"encoding/json"
	"fmt"
)

// Subject is the existing scanner presence subject.
const Subject = "cafe.discovery.scanners.presence"

const (
	// EventJoined announces that a scanner is present.
	EventJoined = "joined"
	// EventLeft announces that a scanner has left.
	EventLeft = "left"
)

const (
	// IndexerEtherscan, IndexerMoralis, and IndexerNone are the three indexer
	// values a wallet presence message can carry. IndexerNone means the wallet
	// scanner announced that it has no indexer. It is distinct from a missing
	// field.
	IndexerEtherscan = "etherscan"
	IndexerMoralis   = "moralis"
	IndexerNone      = "none"
	// IndexerUnknown is the read value of a missing, unknown, or invalid
	// onchain_indexer. It is not a fourth wire value written in place of an
	// absent field.
	IndexerUnknown = "unknown"
)

// Message is one scanner presence announcement.
// Type may be any non-empty string. An unknown type does not make the message
// invalid. OnchainIndexer is optional.
type Message struct {
	Event          string `json:"event"`
	ScannerID      string `json:"scanner_id"`
	Type           string `json:"type"`
	OnchainIndexer string `json:"onchain_indexer,omitempty"`
}

// Decode unmarshals and validates a presence message.
// A malformed body yields a zero Message.
func Decode(raw []byte) (Message, error) {
	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		return Message{}, fmt.Errorf("%w: json", ErrMalformed)
	}
	if err := msg.Validate(); err != nil {
		return Message{}, err
	}
	return msg, nil
}

// Validate checks event, scanner_id, and type.
// event must be joined or left. scanner_id and type must be non-empty.
// type is not compared to a known list. onchain_indexer is not rejected here;
// Indexer reports how to read it.
func (m Message) Validate() error {
	if m.Event != EventJoined && m.Event != EventLeft {
		return fmt.Errorf("%w: event", ErrMalformed)
	}
	if m.ScannerID == "" {
		return fmt.Errorf("%w: scanner_id", ErrMalformed)
	}
	if m.Type == "" {
		return fmt.Errorf("%w: type", ErrMalformed)
	}
	return nil
}

// Indexer returns the indexer carried by the message.
// etherscan, moralis, and none are returned unchanged.
// Any other value, including a missing field or JSON null, reads as unknown,
// never as none. The raw OnchainIndexer field is left unchanged.
func (m Message) Indexer() string {
	switch m.OnchainIndexer {
	case IndexerEtherscan, IndexerMoralis, IndexerNone:
		return m.OnchainIndexer
	default:
		return IndexerUnknown
	}
}
