package v1

// Subjects already used for a scanner heartbeat. The body is Message on both.
const (
	SubjectScannerHeartbeatTLS    = "scanner.heartbeat.tls"
	SubjectScannerHeartbeatWallet = "scanner.heartbeat.wallet"
)

// KindTLS and KindWallet are the kind values the scanners publish.
// Any other non-empty kind remains a valid body.
const (
	KindTLS    = "tls"
	KindWallet = "wallet"
)
