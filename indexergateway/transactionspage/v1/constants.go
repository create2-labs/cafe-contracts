package v1

// Subject is the NATS subject for one transaction-history page.
// The scanner publishes the request here. The gateway subscribes.
const Subject = "cafe.indexer.gateway.transactions.page.v1"

const (
	// MaxCursorBytes is the maximum length of a request cursor and of next_cursor.
	MaxCursorBytes = 8192
	// MaxTransactions is the maximum number of transactions in one page.
	MaxTransactions = 100
	// MaxRequestIDBytes is the maximum length of a response request_id.
	MaxRequestIDBytes = 128
	// MaxErrorMessageBytes is the maximum length of error.message.
	MaxErrorMessageBytes = 128
	// MinLimit and MaxLimit bound the requested page size.
	MinLimit = 1
	MaxLimit = 100
)

// Stable v1 error codes. They do not name a provider.
const (
	CodeInvalidRequest    = "invalid_request"
	CodeUnsupportedChain  = "unsupported_chain"
	CodeNoProvider        = "no_provider"
	CodeProviderError     = "provider_error"
	CodeDeadlineExceeded  = "deadline_exceeded"
	CodeRateLimitExceeded = "rate_limit_exceeded"
)

// ValidCode reports whether code is one of the six v1 error codes.
func ValidCode(code string) bool {
	switch code {
	case CodeInvalidRequest, CodeUnsupportedChain, CodeNoProvider,
		CodeProviderError, CodeDeadlineExceeded, CodeRateLimitExceeded:
		return true
	default:
		return false
	}
}
