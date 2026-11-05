package v1

// Subject is the NATS subject of the contract creation lookup.
const Subject = "cafe.indexer.gateway.contracts.creation.v1"

// MaxRequestIDBytes is the maximum length of a request_id that
// BoundedRequestID copies into a response.
const MaxRequestIDBytes = 128

// MaxErrorMessageBytes is the maximum length of error.message.
const MaxErrorMessageBytes = 128

// Stable v1 error codes. They do not name a provider.
const (
	CodeInvalidRequest       = "invalid_request"
	CodeUnsupportedChain     = "unsupported_chain"
	CodeNoProvider           = "no_provider"
	CodeProviderError        = "provider_error"
	CodeDeadlineExceeded     = "deadline_exceeded"
	CodeRateLimitExceeded    = "rate_limit_exceeded"
	CodeUnsupportedOperation = "unsupported_operation"
)

// ValidCode reports whether code is one of the seven v1 error codes.
func ValidCode(code string) bool {
	switch code {
	case CodeInvalidRequest, CodeUnsupportedChain, CodeNoProvider,
		CodeProviderError, CodeDeadlineExceeded, CodeRateLimitExceeded,
		CodeUnsupportedOperation:
		return true
	default:
		return false
	}
}
