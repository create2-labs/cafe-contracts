package v1

import "testing"

// OC-P09 note: Subject returns the subject of a family. Does not compile until
// the implement pass adds Subject(Family) string.
func TestSubject(t *testing.T) {
	if got := Subject(FamilyHistory); got != SubjectEnrichmentRequestedHistory {
		t.Errorf("history: got %q", got)
	}
	if got := Subject(FamilyDeployer); got != SubjectEnrichmentRequestedDeployer {
		t.Errorf("deployer: got %q", got)
	}
}
