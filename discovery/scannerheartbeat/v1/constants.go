package v1

// Subjects used for a scanner heartbeat. The body is Message on all of them.
const (
	SubjectScannerHeartbeatTLS                = "scanner.heartbeat.tls"
	SubjectScannerHeartbeatWallet             = "scanner.heartbeat.wallet"
	SubjectScannerHeartbeatEnrichmentHistory  = "scanner.heartbeat.enrichment.history"
	SubjectScannerHeartbeatEnrichmentDeployer = "scanner.heartbeat.enrichment.deployer"
)

// KindTLS, KindWallet, KindEnrichmentHistory, and KindEnrichmentDeployer are the
// kind values the publishers send. Any other non-empty kind remains a valid body.
// The package does not match a kind to a subject; the consumer does.
const (
	KindTLS                = "tls"
	KindWallet             = "wallet"
	KindEnrichmentHistory  = "enrichment.history"
	KindEnrichmentDeployer = "enrichment.deployer"
)
