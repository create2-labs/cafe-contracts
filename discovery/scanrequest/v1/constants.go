package v1

// Subjects already used for a scan request. The body is Wallet on the wallet
// subjects and TLS on the TLS subjects.
const (
	SubjectWalletScan          = "cafe.discovery.wallet.scan"
	SubjectTLSScan             = "cafe.discovery.tls.scan"
	SubjectScanRequestedWallet = "scan.requested.wallet"
	SubjectScanRequestedTLS    = "scan.requested.tls"
)

// NilID is the canonical nil UUID. Default TLS scans publish it as user_id.
const NilID = "00000000-0000-0000-0000-000000000000"
