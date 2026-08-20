# Model to wire mapping (v0.1 envelope / assessment payload v0.2)

Reference matrix for `cafenatsv01` event contracts. “Internal model” means the likely owning domain in the CAFE repos; the wire is always JSON in this module.

| event_type (v0.1) | NATS subject constant | Primary internal model (consumer / producer) | Wire package |
| --- | --- | --- | --- |
| `cafe.discovery.wallet.observed` | `NATSSubjectDiscoveryWalletObservedV01` | Discovery observation event (informationnel) | `observation/wallet/v01` |
| `policy.assessment.requested` | `NATSSubjectPolicyAssessmentRequestedV01` | User/API trigger → CPM: embeds a full `observation/wallet/v01` snapshot + `crypto_policy_id` (assessment payload **v0.2**) | `PolicyAssessmentRequested` |
| `policy.validation.completed` | `NATSSubjectPolicyValidationCompletedV01` | CPM `CryptoPolicyInstance` validation result | `PolicyValidationCompleted` |
| `policy.instance.activated` | `NATSSubjectPolicyInstanceActivatedV01` | CPM instance lifecycle | `PolicyInstanceActivated` |
| `policy.assessment.completed` | `NATSSubjectPolicyAssessmentCompletedV01` | CPM `CryptoPolicyAssessmentResult` (summary) | `PolicyAssessmentCompleted` |
| `policy.remediation.requested` | `NATSSubjectPolicyRemediationRequestedV01` | CPM → Remediation handoff | `PolicyRemediationRequested` |
| `remediation.plan.created` | `NATSSubjectRemediationPlanCreatedV01` | Remediation service plan | `RemediationPlanCreated` |
| `remediation.execution.started` | `NATSSubjectRemediationExecutionStartedV01` | Remediation run | `RemediationExecutionStarted` |
| `remediation.execution.completed` | `NATSSubjectRemediationExecutionCompletedV01` | Remediation run success | `RemediationExecutionCompleted` |
| `remediation.execution.failed` | `NATSSubjectRemediationExecutionFailedV01` | Remediation run failure | `RemediationExecutionFailed` |

**Idempotence:** For inbound commands, `event_id` is the primary duplicate-suppression key. Optional `client_request_id` in `PolicyAssessmentRequested` is for tracing only.

**Assessment → CPM (payload wire v0.2):** `PolicyAssessmentRequested.Payload` carries:

- `crypto_policy_id` (required) — catalogue Crypto Policy id; required posture is on the CP, not on the wire
- `observation` (required) — full `observation/wallet/v01` scan snapshot (NATS scan context; HTTP explore uses `policy_context` for the same semantic input)
- `client_request_id` (optional)

**Rejected on assessment payload v0.2:** `selection_request` and couche-B / former selection fields (`allow_new_wallet`, `address_continuity_required`, `key_rotation_model`, `target_posture`, …). Decode returns `ErrLegacyAssessmentField`. Those fields belong only in persist `user_constraints` (CPM-P10), not on explore/assessment input.

**Removed:** `PolicySelectionRequestWire` is no longer part of the assessment path (CPM-P9a-contracts). Consumers must bump to this contracts revision before adopting explore HTTP v0.2 (CPM-P9a-cpm); do not use a local `replace` as a merge solution.

**Observation snapshot:** `PolicyAssessmentRequested.Payload.Observation` is a full `observation/wallet/v01` `Event` and must pass `Validate()` (`event_type` = `cafe.discovery.wallet.observed`, `event_version` = `v0.1`).

**Envelope version:** event headers remain `event_version` = `v0.1` in this package; “assessment v0.2” refers to the **payload shape**, not a new envelope version directory.
