package cafenatsv01

// PolicyValidationCompleted is emitted when CPM has finished validating a policy instance.
type PolicyValidationCompleted struct {
	EnvelopeV01
	Subject SubjectRef                       `json:"subject"`
	Payload PolicyValidationCompletedPayload `json:"payload"`
}

// PolicyValidationCompletedPayload carries the validation outcome.
type PolicyValidationCompletedPayload struct {
	InstanceID      string   `json:"instance_id"`
	TemplateID      string   `json:"template_id"`
	Valid           bool     `json:"valid"`
	IssueCount      int      `json:"issue_count"`
	ValidationRunID string   `json:"validation_run_id"`
	SchemaVersion   string   `json:"schema_version,omitempty"`
	Labels          []string `json:"labels,omitempty"`
}

// Validate for PolicyValidationCompleted.
func (e *PolicyValidationCompleted) Validate() error {
	if e == nil {
		return ErrEventIDRequired
	}
	if err := requireEnvelopeV01(e.EnvelopeV01, EventTypePolicyValidationCompleted); err != nil {
		return err
	}
	if e.Producer != ProducerCafeCPM {
		return ErrProducer
	}
	if err := requireSubjectPolicyInstance(e.Subject); err != nil {
		return err
	}
	if e.Payload.InstanceID == "" || e.Payload.TemplateID == "" {
		return ErrPayloadInvalid
	}
	if e.Payload.IssueCount < 0 {
		return ErrPayloadInvalid
	}
	if e.Payload.ValidationRunID == "" {
		return ErrPayloadInvalid
	}
	return nil
}

// PolicyInstanceActivated is emitted when a policy instance becomes active for use.
type PolicyInstanceActivated struct {
	EnvelopeV01
	Subject SubjectRef                     `json:"subject"`
	Payload PolicyInstanceActivatedPayload `json:"payload"`
}

// PolicyInstanceActivatedPayload records activation metadata.
type PolicyInstanceActivatedPayload struct {
	InstanceID   string `json:"instance_id"`
	TemplateID   string `json:"template_id"`
	GraphVersion string `json:"graph_version,omitempty"`
	ActivationID string `json:"activation_id"`
}

// Validate for PolicyInstanceActivated.
func (e *PolicyInstanceActivated) Validate() error {
	if e == nil {
		return ErrEventIDRequired
	}
	if err := requireEnvelopeV01(e.EnvelopeV01, EventTypePolicyInstanceActivated); err != nil {
		return err
	}
	if e.Producer != ProducerCafeCPM {
		return ErrProducer
	}
	if err := requireSubjectPolicyInstance(e.Subject); err != nil {
		return err
	}
	if e.Payload.InstanceID == "" || e.Payload.TemplateID == "" {
		return ErrPayloadInvalid
	}
	if e.Payload.ActivationID == "" {
		return ErrPayloadInvalid
	}
	return nil
}

// PolicyAssessmentStatus is a coarse outcome for wire export.
const (
	PolicyAssessmentStatusSucceeded = "succeeded"
	PolicyAssessmentStatusFailed    = "failed"
	PolicyAssessmentStatusPartial   = "partial"
)

// PolicyAssessmentCompleted is emitted when a policy assessment run finishes.
type PolicyAssessmentCompleted struct {
	EnvelopeV01
	Subject SubjectRef                       `json:"subject"`
	Payload PolicyAssessmentCompletedPayload `json:"payload"`
}

// PolicyAssessmentCompletedPayload contains assessment result references.
type PolicyAssessmentCompletedPayload struct {
	InstanceID   string `json:"instance_id"`
	AssessmentID string `json:"assessment_id"`
	Status       string `json:"status"`
	FindingCount int    `json:"finding_count"`
}

// Validate for PolicyAssessmentCompleted.
func (e *PolicyAssessmentCompleted) Validate() error {
	if e == nil {
		return ErrEventIDRequired
	}
	if err := requireEnvelopeV01(e.EnvelopeV01, EventTypePolicyAssessmentCompleted); err != nil {
		return err
	}
	if e.Producer != ProducerCafeCPM {
		return ErrProducer
	}
	if err := requireSubjectPolicyInstance(e.Subject); err != nil {
		return err
	}
	if e.Payload.InstanceID == "" || e.Payload.AssessmentID == "" {
		return ErrPayloadInvalid
	}
	switch e.Payload.Status {
	case PolicyAssessmentStatusSucceeded, PolicyAssessmentStatusFailed, PolicyAssessmentStatusPartial:
	default:
		return ErrPayloadInvalid
	}
	if e.Payload.FindingCount < 0 {
		return ErrPayloadInvalid
	}
	return nil
}

// PolicyRemediationRequested is emitted when CPM (or a policy flow) requests remediation.
type PolicyRemediationRequested struct {
	EnvelopeV01
	Subject SubjectRef                        `json:"subject"`
	Payload PolicyRemediationRequestedPayload `json:"payload"`
}

// PolicyRemediationRequestedPayload references the remediation to plan.
type PolicyRemediationRequestedPayload struct {
	InstanceID     string `json:"instance_id"`
	RemediationID  string `json:"remediation_id"`
	ReasonCode     string `json:"reason_code"`
	RequestedBy    string `json:"requested_by,omitempty"`
	CorrelationRef string `json:"correlation_ref,omitempty"`
}

// Validate for PolicyRemediationRequested.
func (e *PolicyRemediationRequested) Validate() error {
	if e == nil {
		return ErrEventIDRequired
	}
	if err := requireEnvelopeV01(e.EnvelopeV01, EventTypePolicyRemediationRequested); err != nil {
		return err
	}
	if e.Producer != ProducerCafeCPM {
		return ErrProducer
	}
	if err := requireSubjectPolicyInstance(e.Subject); err != nil {
		return err
	}
	if e.Payload.InstanceID == "" || e.Payload.RemediationID == "" {
		return ErrPayloadInvalid
	}
	if e.Payload.ReasonCode == "" {
		return ErrPayloadInvalid
	}
	return nil
}
