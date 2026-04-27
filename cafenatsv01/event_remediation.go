package cafenatsv01

// RemediationPlanCreated is emitted when a remediation plan is materialized.
type RemediationPlanCreated struct {
	EnvelopeV01
	Subject SubjectRef                    `json:"subject"`
	Payload RemediationPlanCreatedPayload `json:"payload"`
}

// RemediationPlanCreatedPayload links plan, remediation, and instance identifiers.
type RemediationPlanCreatedPayload struct {
	PlanID        string `json:"plan_id"`
	RemediationID string `json:"remediation_id"`
	InstanceID    string `json:"instance_id"`
	Revision      string `json:"revision,omitempty"`
}

// Validate for RemediationPlanCreated.
func (e *RemediationPlanCreated) Validate() error {
	if e == nil {
		return ErrEventIDRequired
	}
	if err := requireEnvelopeV01(e.EnvelopeV01, EventTypeRemediationPlanCreated); err != nil {
		return err
	}
	if e.Producer != ProducerCafeRemediation {
		return ErrProducer
	}
	if err := requireSubjectRemediation(e.Subject); err != nil {
		return err
	}
	if e.Payload.PlanID == "" || e.Payload.RemediationID == "" || e.Payload.InstanceID == "" {
		return ErrPayloadInvalid
	}
	return nil
}

// RemediationExecutionStarted marks execution of a plan.
type RemediationExecutionStarted struct {
	EnvelopeV01
	Subject SubjectRef                         `json:"subject"`
	Payload RemediationExecutionStartedPayload `json:"payload"`
}

// RemediationExecutionStartedPayload identifies the execution and links to the plan.
type RemediationExecutionStartedPayload struct {
	ExecutionID   string `json:"execution_id"`
	PlanID        string `json:"plan_id"`
	InstanceID    string `json:"instance_id"`
	RemediationID string `json:"remediation_id"`
}

// Validate for RemediationExecutionStarted.
func (e *RemediationExecutionStarted) Validate() error {
	if e == nil {
		return ErrEventIDRequired
	}
	if err := requireEnvelopeV01(e.EnvelopeV01, EventTypeRemediationExecutionStarted); err != nil {
		return err
	}
	if e.Producer != ProducerCafeRemediation {
		return ErrProducer
	}
	if err := requireSubjectRemediation(e.Subject); err != nil {
		return err
	}
	if e.Payload.ExecutionID == "" || e.Payload.PlanID == "" {
		return ErrPayloadInvalid
	}
	if e.Payload.InstanceID == "" || e.Payload.RemediationID == "" {
		return ErrPayloadInvalid
	}
	return nil
}

// RemediationExecutionStatusSucceeded indicates success on the wire.
const RemediationExecutionStatusSucceeded = "succeeded"

// RemediationExecutionCompleted is emitted on successful plan execution.
type RemediationExecutionCompleted struct {
	EnvelopeV01
	Subject SubjectRef                           `json:"subject"`
	Payload RemediationExecutionCompletedPayload `json:"payload"`
}

// RemediationExecutionCompletedPayload records completion of an execution.
type RemediationExecutionCompletedPayload struct {
	ExecutionID  string `json:"execution_id"`
	PlanID       string `json:"plan_id"`
	InstanceID   string `json:"instance_id"`
	Status       string `json:"status"`
	CompletedOps int    `json:"completed_ops,omitempty"`
}

// Validate for RemediationExecutionCompleted.
func (e *RemediationExecutionCompleted) Validate() error {
	if e == nil {
		return ErrEventIDRequired
	}
	if err := requireEnvelopeV01(e.EnvelopeV01, EventTypeRemediationExecutionCompleted); err != nil {
		return err
	}
	if e.Producer != ProducerCafeRemediation {
		return ErrProducer
	}
	if err := requireSubjectRemediation(e.Subject); err != nil {
		return err
	}
	if e.Payload.ExecutionID == "" || e.Payload.PlanID == "" || e.Payload.InstanceID == "" {
		return ErrPayloadInvalid
	}
	if e.Payload.Status != RemediationExecutionStatusSucceeded {
		return ErrPayloadInvalid
	}
	return nil
}

// RemediationExecutionFailed is emitted when execution cannot complete successfully.
type RemediationExecutionFailed struct {
	EnvelopeV01
	Subject SubjectRef                        `json:"subject"`
	Payload RemediationExecutionFailedPayload `json:"payload"`
}

// RemediationExecutionFailedPayload carries a safe-to-log failure summary.
type RemediationExecutionFailedPayload struct {
	ExecutionID  string `json:"execution_id"`
	PlanID       string `json:"plan_id"`
	InstanceID   string `json:"instance_id"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message,omitempty"`
	Retryable    bool   `json:"retryable,omitempty"`
}

// Validate for RemediationExecutionFailed.
func (e *RemediationExecutionFailed) Validate() error {
	if e == nil {
		return ErrEventIDRequired
	}
	if err := requireEnvelopeV01(e.EnvelopeV01, EventTypeRemediationExecutionFailed); err != nil {
		return err
	}
	if e.Producer != ProducerCafeRemediation {
		return ErrProducer
	}
	if err := requireSubjectRemediation(e.Subject); err != nil {
		return err
	}
	if e.Payload.ExecutionID == "" || e.Payload.PlanID == "" || e.Payload.InstanceID == "" {
		return ErrPayloadInvalid
	}
	if e.Payload.ErrorCode == "" {
		return ErrPayloadInvalid
	}
	return nil
}
