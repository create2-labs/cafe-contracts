package cafenatsv01

// Event version for all v0.1 messages in this package.
const EventVersionV01 = "v0.1"

// Semantic event_type values (stable identifiers; not necessarily identical to NATS subject paths).
const (
	EventTypeDiscoveryWalletObserved       = "discovery.wallet.observed"
	EventTypePolicyAssessmentRequested     = "policy.assessment.requested"
	EventTypePolicyValidationCompleted     = "policy.validation.completed"
	EventTypePolicyInstanceActivated       = "policy.instance.activated"
	EventTypePolicyAssessmentCompleted     = "policy.assessment.completed"
	EventTypePolicyRemediationRequested    = "policy.remediation.requested"
	EventTypeRemediationPlanCreated        = "remediation.plan.created"
	EventTypeRemediationExecutionStarted   = "remediation.execution.started"
	EventTypeRemediationExecutionCompleted = "remediation.execution.completed"
	EventTypeRemediationExecutionFailed    = "remediation.execution.failed"
)

// Producer name strings on the wire (contract revision labels, not auth proofs).
const (
	ProducerCafeDiscovery     = "cafe-discovery"
	ProducerCafeCPM           = "cafe-cpm"
	ProducerCafeRemediation   = "cafe-remediation"
	ProducerCafeCryptoBackend = "cafe-crypto-backend"
	ProducerCafeEdge          = "cafe-edge"
)

// SubjectTypeWallet identifies a wallet-scoped event subject.
const SubjectTypeWallet = "wallet"

// SubjectTypePolicyInstance identifies a crypto policy instance as the subject of an event.
const SubjectTypePolicyInstance = "policy_instance"

// SubjectTypeRemediation identifies a remediation aggregate as the subject of an event.
const SubjectTypeRemediation = "remediation"
