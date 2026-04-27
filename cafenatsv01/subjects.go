package cafenatsv01

// NATS subject line constants (v0.1). Services should publish and subscribe using these
// values (or a deployment-specific prefix) so routing stays versioned and grep-friendly.
const (
	NATSSubjectDiscoveryWalletObservedV01       = "cafe.discovery.events.wallet.observed.v0_1"
	NATSSubjectPolicyAssessmentRequestedV01     = "cafe.policy.events.policy.assessment.requested.v0_1"
	NATSSubjectPolicyValidationCompletedV01     = "cafe.cpm.events.policy.validation.completed.v0_1"
	NATSSubjectPolicyInstanceActivatedV01       = "cafe.cpm.events.policy.instance.activated.v0_1"
	NATSSubjectPolicyAssessmentCompletedV01     = "cafe.cpm.events.policy.assessment.completed.v0_1"
	NATSSubjectPolicyRemediationRequestedV01    = "cafe.cpm.events.policy.remediation.requested.v0_1"
	NATSSubjectRemediationPlanCreatedV01        = "cafe.remediation.events.plan.created.v0_1"
	NATSSubjectRemediationExecutionStartedV01   = "cafe.remediation.events.execution.started.v0_1"
	NATSSubjectRemediationExecutionCompletedV01 = "cafe.remediation.events.execution.completed.v0_1"
	NATSSubjectRemediationExecutionFailedV01    = "cafe.remediation.events.execution.failed.v0_1"
)
