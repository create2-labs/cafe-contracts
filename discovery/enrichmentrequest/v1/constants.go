package v1

// Subjects and stream of an enrichment request.
const (
	SubjectEnrichmentRequestedHistory  = "enrichment.requested.history"
	SubjectEnrichmentRequestedDeployer = "enrichment.requested.deployer"
	StreamEnrichmentRequested          = "ENRICHMENT_REQUESTED"
)

// Family is an enrichment family.
type Family string

// Enrichment families.
const (
	FamilyHistory  Family = "history"
	FamilyDeployer Family = "deployer"
)

// IsValid reports whether f is a known enrichment family.
func (f Family) IsValid() bool {
	return f == FamilyHistory || f == FamilyDeployer
}

// Subject returns the subject on which a request of family f is published.
// It returns the empty string for an unknown family.
func Subject(f Family) string {
	switch f {
	case FamilyHistory:
		return SubjectEnrichmentRequestedHistory
	case FamilyDeployer:
		return SubjectEnrichmentRequestedDeployer
	default:
		return ""
	}
}
