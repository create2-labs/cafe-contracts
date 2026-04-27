# Model to wire mapping (v0.1)

Reference matrix for `cafenatsv01` event contracts. “Internal model” means the likely owning domain in the CAFE repos; the wire is always JSON in this module.

| event_type (v0.1) | NATS subject constant | Primary internal model (consumer / producer) | Wire package |
| --- | --- | --- | --- |
| `cafe.discovery.wallet.observed` | `NATSSubjectDiscoveryWalletObservedV01` | Discovery observation event (informationnel) | `observation/wallet/v01` |
| `policy.assessment.requested` | `NATSSubjectPolicyAssessmentRequestedV01` | User/API trigger → CPM: embeds a full `observation/wallet/v01` snapshot + selection | `PolicyAssessmentRequested` |
| `policy.validation.completed` | `NATSSubjectPolicyValidationCompletedV01` | CPM `CryptoPolicyInstance` validation result | `PolicyValidationCompleted` |
| `policy.instance.activated` | `NATSSubjectPolicyInstanceActivatedV01` | CPM instance lifecycle | `PolicyInstanceActivated` |
| `policy.assessment.completed` | `NATSSubjectPolicyAssessmentCompletedV01` | CPM `CryptoPolicyAssessmentResult` (summary) | `PolicyAssessmentCompleted` |
| `policy.remediation.requested` | `NATSSubjectPolicyRemediationRequestedV01` | CPM → Remediation handoff | `PolicyRemediationRequested` |
| `remediation.plan.created` | `NATSSubjectRemediationPlanCreatedV01` | Remediation service plan | `RemediationPlanCreated` |
| `remediation.execution.started` | `NATSSubjectRemediationExecutionStartedV01` | Remediation run | `RemediationExecutionStarted` |
| `remediation.execution.completed` | `NATSSubjectRemediationExecutionCompletedV01` | Remediation run success | `RemediationExecutionCompleted` |
| `remediation.execution.failed` | `NATSSubjectRemediationExecutionFailedV01` | Remediation run failure | `RemediationExecutionFailed` |

**Idempotence:** For inbound commands, `event_id` is the primary duplicate-suppression key. Optional `client_request_id` in `PolicyAssessmentRequested` is for tracing only.

**Selection → CPM:** `PolicySelectionRequestWire` JSON matches `cafe-cpm` `PolicySelectionRequest` (same field names and types; posture as string). Map with `json.Unmarshal` then call domain `Normalize`/`Validate` in CPM.

**Observation snapshot:** `PolicyAssessmentRequested.Payload.Observation` is a full `observation/wallet/v01` `Event` and must pass `Validate()` (`event_type` = `cafe.discovery.wallet.observed`, `event_version` = `v0.1`).
