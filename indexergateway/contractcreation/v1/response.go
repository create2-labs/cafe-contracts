package v1

import (
	"fmt"

	"github.com/create2-labs/cafe-contracts/address"
)

// Error is the error object of a response envelope.
// Code is one of the seven v1 codes. Message is a generic string of 1 to
// MaxErrorMessageBytes and does not name a provider.
type Error struct {
	Code    string
	Message string
}

// Response is the discriminated creation envelope.
// A found response has Creator and optionally CreationTxHash. A not found
// response has none of Creator, CreationTxHash, and Error. An error has
// Error and neither Creator nor CreationTxHash.
type Response struct {
	RequestID      string
	Creator        string
	CreationTxHash string
	Error          *Error
}

// Validate checks an in-memory envelope.
// request_id may be any string of at most MaxRequestIDBytes. It is not
// required to be a UUID. The empty string is valid only when error.code is
// invalid_request. Creator needs the 0x or 0X prefix.
func (r Response) Validate() error {
	if r.Error != nil {
		if r.Creator != "" || r.CreationTxHash != "" {
			return fmt.Errorf("%w: mixed", ErrResponse)
		}
		if !ValidCode(r.Error.Code) {
			return fmt.Errorf("%w: error.code", ErrResponse)
		}
		n := len(r.Error.Message)
		if n < 1 || n > MaxErrorMessageBytes {
			return fmt.Errorf("%w: error.message", ErrResponse)
		}
		return validateResponseID(r.RequestID, r.Error.Code)
	}
	if r.Creator == "" {
		if r.CreationTxHash != "" {
			return fmt.Errorf("%w: creation_tx_hash", ErrResponse)
		}
	} else if !hasHexPrefix(r.Creator) || !address.IsValidHexAddress(r.Creator) {
		return fmt.Errorf("%w: creator", ErrResponse)
	}
	return validateResponseID(r.RequestID, "")
}

// ValidateForRequest checks the envelope, then that its request_id echoes
// requestID. The reader uses its own UUID, so an empty request_id never passes.
func (r Response) ValidateForRequest(requestID string) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if r.RequestID != requestID {
		return fmt.Errorf("%w: request_id", ErrResponse)
	}
	return nil
}

func validateResponseID(id, code string) error {
	if len(id) > MaxRequestIDBytes {
		return fmt.Errorf("%w: request_id", ErrResponse)
	}
	if id == "" && code != CodeInvalidRequest {
		return fmt.Errorf("%w: request_id", ErrResponse)
	}
	return nil
}
