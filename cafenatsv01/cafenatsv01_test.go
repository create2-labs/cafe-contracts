package cafenatsv01

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNATSKeyConstants_Unique(t *testing.T) {
	keys := []string{
		NATSSubjectDiscoveryWalletObservedV01,
		NATSSubjectPolicyAssessmentRequestedV01,
		NATSSubjectPolicyValidationCompletedV01,
		NATSSubjectPolicyInstanceActivatedV01,
		NATSSubjectPolicyAssessmentCompletedV01,
		NATSSubjectPolicyRemediationRequestedV01,
		NATSSubjectRemediationPlanCreatedV01,
		NATSSubjectRemediationExecutionStartedV01,
		NATSSubjectRemediationExecutionCompletedV01,
		NATSSubjectRemediationExecutionFailedV01,
	}
	seen := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		if k == "" {
			t.Fatal("empty subject constant")
		}
		if _, ok := seen[k]; ok {
			t.Fatalf("duplicate NATS subject: %q", k)
		}
		seen[k] = struct{}{}
	}
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return b
}

func TestPolicyAssessmentRequested_FixtureRoundTrip(t *testing.T) {
	data := readFixture(t, "policy_assessment_requested_v01.json")
	var ev PolicyAssessmentRequested
	if err := json.Unmarshal(data, &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := ev.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if ev.Payload.Observation.EventID == "" {
		t.Fatal("nested observation missing")
	}
	if ev.Payload.SelectionRequest.TargetPosture != TargetPostureHybrid {
		t.Fatalf("posture: %s", ev.Payload.SelectionRequest.TargetPosture)
	}
	roundTripJSON(t, &ev)
}

func TestPolicyValidationCompleted_Fixture(t *testing.T) {
	var ev PolicyValidationCompleted
	if err := json.Unmarshal(readFixture(t, "policy_validation_completed_v01.json"), &ev); err != nil {
		t.Fatal(err)
	}
	if err := ev.Validate(); err != nil {
		t.Fatal(err)
	}
	roundTripJSON(t, &ev)
}

func TestPolicyInstanceActivated_Fixture(t *testing.T) {
	var ev PolicyInstanceActivated
	if err := json.Unmarshal(readFixture(t, "policy_instance_activated_v01.json"), &ev); err != nil {
		t.Fatal(err)
	}
	if err := ev.Validate(); err != nil {
		t.Fatal(err)
	}
	roundTripJSON(t, &ev)
}

func TestPolicyAssessmentCompleted_Fixture(t *testing.T) {
	var ev PolicyAssessmentCompleted
	if err := json.Unmarshal(readFixture(t, "policy_assessment_completed_v01.json"), &ev); err != nil {
		t.Fatal(err)
	}
	if err := ev.Validate(); err != nil {
		t.Fatal(err)
	}
	roundTripJSON(t, &ev)
}

func TestPolicyRemediationRequested_Fixture(t *testing.T) {
	var ev PolicyRemediationRequested
	if err := json.Unmarshal(readFixture(t, "policy_remediation_requested_v01.json"), &ev); err != nil {
		t.Fatal(err)
	}
	if err := ev.Validate(); err != nil {
		t.Fatal(err)
	}
	roundTripJSON(t, &ev)
}

func TestRemediationPlanCreated_Fixture(t *testing.T) {
	var ev RemediationPlanCreated
	if err := json.Unmarshal(readFixture(t, "remediation_plan_created_v01.json"), &ev); err != nil {
		t.Fatal(err)
	}
	if err := ev.Validate(); err != nil {
		t.Fatal(err)
	}
	roundTripJSON(t, &ev)
}

func TestRemediationExecutionStarted_Fixture(t *testing.T) {
	var ev RemediationExecutionStarted
	if err := json.Unmarshal(readFixture(t, "remediation_execution_started_v01.json"), &ev); err != nil {
		t.Fatal(err)
	}
	if err := ev.Validate(); err != nil {
		t.Fatal(err)
	}
	roundTripJSON(t, &ev)
}

func TestRemediationExecutionCompleted_Fixture(t *testing.T) {
	var ev RemediationExecutionCompleted
	if err := json.Unmarshal(readFixture(t, "remediation_execution_completed_v01.json"), &ev); err != nil {
		t.Fatal(err)
	}
	if err := ev.Validate(); err != nil {
		t.Fatal(err)
	}
	roundTripJSON(t, &ev)
}

func TestRemediationExecutionFailed_Fixture(t *testing.T) {
	var ev RemediationExecutionFailed
	if err := json.Unmarshal(readFixture(t, "remediation_execution_failed_v01.json"), &ev); err != nil {
		t.Fatal(err)
	}
	if err := ev.Validate(); err != nil {
		t.Fatal(err)
	}
	roundTripJSON(t, &ev)
}

func roundTripJSON[T any](t *testing.T, v *T) {
	t.Helper()
	out, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var again T
	if err := json.Unmarshal(out, &again); err != nil {
		t.Fatalf("unmarshal round-trip: %v", err)
	}
	if err := revalidate(again); err != nil {
		t.Fatalf("re-validate: %v", err)
	}
	if !reflect.DeepEqual(*v, again) {
		t.Fatalf("deep equal failed after round-trip for %T", v)
	}
}

func revalidate(v any) error {
	switch x := v.(type) {
	case PolicyAssessmentRequested:
		return x.Validate()
	case PolicyValidationCompleted:
		return x.Validate()
	case PolicyInstanceActivated:
		return x.Validate()
	case PolicyAssessmentCompleted:
		return x.Validate()
	case PolicyRemediationRequested:
		return x.Validate()
	case RemediationPlanCreated:
		return x.Validate()
	case RemediationExecutionStarted:
		return x.Validate()
	case RemediationExecutionCompleted:
		return x.Validate()
	case RemediationExecutionFailed:
		return x.Validate()
	default:
		return nil
	}
}

func TestPolicyAssessmentRequested_WrongProducer(t *testing.T) {
	var ev PolicyAssessmentRequested
	_ = json.Unmarshal(readFixture(t, "policy_assessment_requested_v01.json"), &ev)
	ev.Producer = "unknown"
	err := ev.Validate()
	if !errors.Is(err, ErrProducer) {
		t.Fatalf("got %v", err)
	}
}

func TestPolicySelectionRequestWire_MultichainRule(t *testing.T) {
	sel := &PolicySelectionRequestWire{
		TargetPosture:     TargetPostureUnknown,
		RequireMultichain: true,
		TargetChainIDs:    []int64{1},
		MinimumMaturity:   1,
		ApprovalMode:      "manual",
	}
	sel.Normalize()
	err := sel.Validate()
	if !errors.Is(err, ErrSelectionRequest) {
		t.Fatalf("expected ErrSelectionRequest, got %v", err)
	}
}
