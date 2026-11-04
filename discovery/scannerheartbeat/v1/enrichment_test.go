package v1

import "testing"

// Design section 3.5: the four enrichment constants carry fixed wire values.
func TestEnrichmentConstantsHaveTheirWireValues(t *testing.T) {
	cases := []struct{ name, got, want string }{
		{"SubjectScannerHeartbeatEnrichmentHistory", SubjectScannerHeartbeatEnrichmentHistory, "scanner.heartbeat.enrichment.history"},
		{"SubjectScannerHeartbeatEnrichmentDeployer", SubjectScannerHeartbeatEnrichmentDeployer, "scanner.heartbeat.enrichment.deployer"},
		{"KindEnrichmentHistory", KindEnrichmentHistory, "enrichment.history"},
		{"KindEnrichmentDeployer", KindEnrichmentDeployer, "enrichment.deployer"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}

// Design section 3.5: the message is unchanged and kind stays open, so an
// enrichment kind decodes and validates; the package does not match kind to subject.
func TestEnrichmentKindsDecodeAndValidate(t *testing.T) {
	for _, kind := range []string{KindEnrichmentHistory, KindEnrichmentDeployer} {
		body := `{"scanner_id":"scanner-1","kind":"` + kind + `","timestamp":"2026-10-08T08:00:30Z"}`
		msg, err := Decode([]byte(body))
		if err != nil {
			t.Errorf("kind %q: %v", kind, err)
			continue
		}
		if msg.Kind != kind {
			t.Errorf("kind %q: decoded %q", kind, msg.Kind)
		}
		if err := msg.Validate(); err != nil {
			t.Errorf("kind %q: %v", kind, err)
		}
	}
}
