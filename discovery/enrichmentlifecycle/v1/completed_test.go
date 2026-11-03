package v1

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	enrichmentrequestv1 "github.com/create2-labs/cafe-contracts/discovery/enrichmentrequest/v1"
	addressv1 "github.com/create2-labs/cafe-contracts/observation/address/v1"
)

const (
	txA = "0xaaaa000000000000000000000000000000000000000000000000000000000001"
	txB = "0xbbbb000000000000000000000000000000000000000000000000000000000002"
	dpA = "0x00000000000000000000000000000000000000a1"
	dpB = "0x00000000000000000000000000000000000000b2"
)

func completedBase(f enrichmentrequestv1.Family) Completed {
	return Completed{EnrichmentID: idE, ScanID: idS, UserID: idU, Family: f, CompletedAt: ts}
}

// historyRecovered mixes observed, not_observed and error chains.
func historyRecovered() Completed {
	m := completedBase("history")
	m.Chains = []ChainOutcome{
		{ChainID: 1, Outcome: OutcomeObserved, TransactionHash: txA},
		{ChainID: 137, Outcome: OutcomeNotObserved},
		{ChainID: 10, Outcome: OutcomeError, Code: "provider_error"},
	}
	m.PublicKeyRecovery = PublicKeyRecoveryRecovered
	m.PublicKey, m.TransactionHash, m.ExposedNetwork = "0x04abcd", txA, "ethereum"
	return m
}

func historyUnresolved() Completed {
	m := historyRecovered()
	m.PublicKeyRecovery = PublicKeyRecoveryUnresolved
	m.PublicKey, m.TransactionHash, m.ExposedNetwork = "", "", ""
	return m
}

func deployerObserved() Completed {
	m := completedBase("deployer")
	m.Chains = []ChainOutcome{
		{ChainID: 1, Outcome: OutcomeObserved, Deployer: dpA, CreationTxHash: txA},
		{ChainID: 10, Outcome: OutcomeObserved, Deployer: dpB},
		{ChainID: 137, Outcome: OutcomeNotObserved},
		{ChainID: 8453, Outcome: OutcomeError, Code: "no_provider"},
	}
	m.RelatedAddresses = []addressv1.Related{
		{Role: addressv1.RelatedRoleDeployer, ChainID: 1, Address: dpA},
		{Role: addressv1.RelatedRoleDeployer, ChainID: 10, Address: dpB},
	}
	return m
}

func allErrors(f enrichmentrequestv1.Family) Completed {
	m := completedBase(f)
	m.Chains = []ChainOutcome{
		{ChainID: 1, Outcome: OutcomeError, Code: "provider_error"},
		{ChainID: 10, Outcome: OutcomeError, Code: "deadline_exceeded"},
	}
	if f == "history" {
		m.PublicKeyRecovery = PublicKeyRecoveryUnresolved
	}
	return m
}

// CA3.2, CA3.3, CA4.2, CA4.3: accepted messages.
func TestCompletedValidateAccepts(t *testing.T) {
	for name, m := range map[string]Completed{
		"history recovered":    historyRecovered(),
		"history unresolved":   historyUnresolved(),
		"deployer observed":    deployerObserved(),
		"history all error":    allErrors("history"),
		"deployer all error":   allErrors("deployer"),
		"deployer no chain ok": func() Completed { m := deployerObserved(); m.Chains = m.Chains[2:]; m.RelatedAddresses = nil; return m }(),
	} {
		if err := m.Validate(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// CA3.2, CA3.3, CA4.2, CA4.3: each contract rule rejected once.
func TestCompletedValidateRejects(t *testing.T) {
	for _, c := range []struct {
		name string
		base func() Completed
		mod  func(m *Completed)
	}{
		{"chains empty", historyUnresolved, func(m *Completed) { m.Chains = []ChainOutcome{} }},
		{"chains nil", historyUnresolved, func(m *Completed) { m.Chains = nil }},
		{"chain_id zero", historyUnresolved, func(m *Completed) { m.Chains[1].ChainID = 0 }},
		{"chain_id negative", historyUnresolved, func(m *Completed) { m.Chains[1].ChainID = -5 }},
		{"chain_id duplicate", historyUnresolved, func(m *Completed) { m.Chains[1].ChainID = 1 }},
		{"outcome unknown", historyUnresolved, func(m *Completed) { m.Chains[1].Outcome = "timeout" }},
		{"outcome empty", historyUnresolved, func(m *Completed) { m.Chains[1].Outcome = "" }},
		{"outcome wrong case", historyUnresolved, func(m *Completed) { m.Chains[1].Outcome = "Observed" }},
		{"error without code", historyUnresolved, func(m *Completed) { m.Chains[2].Code = "" }},
		{"observed with code", historyUnresolved, func(m *Completed) { m.Chains[0].Code = "x" }},
		{"not_observed with code", historyUnresolved, func(m *Completed) { m.Chains[1].Code = "x" }},
		{"history observed without transaction_hash", historyUnresolved, func(m *Completed) { m.Chains[0].TransactionHash = "" }},
		{"history not_observed with transaction_hash", historyUnresolved, func(m *Completed) { m.Chains[1].TransactionHash = txA }},
		{"history error with transaction_hash", historyUnresolved, func(m *Completed) { m.Chains[2].TransactionHash = txA }},
		{"history chain with deployer", historyUnresolved, func(m *Completed) { m.Chains[0].Deployer = dpA }},
		{"history chain with creation_tx_hash", historyUnresolved, func(m *Completed) { m.Chains[0].CreationTxHash = txA }},
		{"history with related_addresses", historyUnresolved, func(m *Completed) {
			m.RelatedAddresses = []addressv1.Related{{Role: addressv1.RelatedRoleDeployer, ChainID: 1, Address: dpA}}
		}},
		{"recovery missing", historyUnresolved, func(m *Completed) { m.PublicKeyRecovery = "" }},
		{"recovery not_required", historyUnresolved, func(m *Completed) { m.PublicKeyRecovery = PublicKeyRecoveryNotRequired }},
		{"recovery unknown", historyUnresolved, func(m *Completed) { m.PublicKeyRecovery = "maybe" }},
		{"recovered without public_key", historyRecovered, func(m *Completed) { m.PublicKey = "" }},
		{"recovered without transaction_hash", historyRecovered, func(m *Completed) { m.TransactionHash = "" }},
		{"recovered without exposed_network", historyRecovered, func(m *Completed) { m.ExposedNetwork = "" }},
		{"recovered transaction_hash matches no chain", historyRecovered, func(m *Completed) { m.TransactionHash = txB }},
		{"unresolved with public_key", historyUnresolved, func(m *Completed) { m.PublicKey = "0x04" }},
		{"unresolved with transaction_hash", historyUnresolved, func(m *Completed) { m.TransactionHash = txA }},
		{"unresolved with exposed_network", historyUnresolved, func(m *Completed) { m.ExposedNetwork = "ethereum" }},
		{"deployer observed without deployer", deployerObserved, func(m *Completed) { m.Chains[0].Deployer = "" }},
		{"deployer not_observed with deployer", deployerObserved, func(m *Completed) { m.Chains[2].Deployer = dpA }},
		{"deployer error with deployer", deployerObserved, func(m *Completed) { m.Chains[3].Deployer = dpA }},
		{"deployer not_observed with creation_tx_hash", deployerObserved, func(m *Completed) { m.Chains[2].CreationTxHash = txA }},
		{"deployer error with creation_tx_hash", deployerObserved, func(m *Completed) { m.Chains[3].CreationTxHash = txA }},
		{"deployer chain with transaction_hash", deployerObserved, func(m *Completed) { m.Chains[0].TransactionHash = txA }},
		{"deployer related missing", deployerObserved, func(m *Completed) { m.RelatedAddresses = m.RelatedAddresses[:1] }},
		{"deployer related absent", deployerObserved, func(m *Completed) { m.RelatedAddresses = nil }},
		{"deployer related for not_observed chain", deployerObserved, func(m *Completed) {
			m.RelatedAddresses = append(m.RelatedAddresses, addressv1.Related{Role: addressv1.RelatedRoleDeployer, ChainID: 137, Address: dpA})
		}},
		{"deployer related for unknown chain", deployerObserved, func(m *Completed) {
			m.RelatedAddresses = append(m.RelatedAddresses, addressv1.Related{Role: addressv1.RelatedRoleDeployer, ChainID: 42, Address: dpA})
		}},
		{"deployer related duplicate", deployerObserved, func(m *Completed) {
			m.RelatedAddresses = append(m.RelatedAddresses, m.RelatedAddresses[0])
		}},
		{"deployer related address mismatch", deployerObserved, func(m *Completed) { m.RelatedAddresses[0].Address = dpB }},
		{"deployer related address empty", deployerObserved, func(m *Completed) { m.RelatedAddresses[0].Address = "" }},
		{"deployer related role delegation_target", deployerObserved, func(m *Completed) { m.RelatedAddresses[0].Role = addressv1.RelatedRoleDelegationTarget }},
		{"deployer related role unknown", deployerObserved, func(m *Completed) { m.RelatedAddresses[0].Role = "owner" }},
		{"deployer with public_key", deployerObserved, func(m *Completed) { m.PublicKey = "0x04" }},
		{"deployer with transaction_hash", deployerObserved, func(m *Completed) { m.TransactionHash = txA }},
		{"deployer with exposed_network", deployerObserved, func(m *Completed) { m.ExposedNetwork = "ethereum" }},
		{"deployer with public_key_recovery", deployerObserved, func(m *Completed) { m.PublicKeyRecovery = PublicKeyRecoveryUnresolved }},
		{"enrichment_id uppercase", historyUnresolved, func(m *Completed) { m.EnrichmentID = strings.ToUpper(idE) }},
		{"enrichment_id nil", historyUnresolved, func(m *Completed) { m.EnrichmentID = nilU }},
		{"enrichment_id empty", historyUnresolved, func(m *Completed) { m.EnrichmentID = "" }},
		{"scan_id not uuid", historyUnresolved, func(m *Completed) { m.ScanID = "abc" }},
		{"scan_id nil", historyUnresolved, func(m *Completed) { m.ScanID = nilU }},
		{"user_id non-canonical", historyUnresolved, func(m *Completed) { m.UserID = strings.ReplaceAll(idU, "-", "") }},
		{"user_id nil", historyUnresolved, func(m *Completed) { m.UserID = nilU }},
		{"family empty", historyUnresolved, func(m *Completed) { m.Family = "" }},
		{"family unknown", historyUnresolved, func(m *Completed) { m.Family = "other" }},
		{"completed_at empty", historyUnresolved, func(m *Completed) { m.CompletedAt = "" }},
	} {
		m := c.base()
		c.mod(&m)
		mustMalformed(t, c.name, m.Validate())
	}
}

// CA3.2, CA3.3, CA4.2, CA4.3: round trip of every accepted shape.
func TestCompletedRoundTrip(t *testing.T) {
	for name, m := range map[string]Completed{
		"history recovered":  historyRecovered(),
		"history unresolved": historyUnresolved(),
		"deployer observed":  deployerObserved(),
		"deployer all error": allErrors("deployer"),
	} {
		raw, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		got, err := DecodeCompleted(raw)
		if err != nil || !reflect.DeepEqual(got, m) {
			t.Errorf("%s: got %+v, err %v", name, got, err)
		}
	}
}

// Design 2: identifiers are lowercased on decode; unknown fields are ignored.
func TestCompletedDecodeNormalizesAndIgnoresUnknown(t *testing.T) {
	raw := `{"enrichment_id":"` + strings.ToUpper(idE) + `","scan_id":"` + strings.ToUpper(idS) + `","user_id":"` + strings.ToUpper(idU) +
		`","family":"history","future":{"a":1},"chains":[{"chain_id":1,"outcome":"error","code":"x","extra":1}],` +
		`"public_key_recovery":"unresolved","completed_at":"` + ts + `","extra":true}`
	got, err := DecodeCompleted([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if got.EnrichmentID != idE || got.ScanID != idS || got.UserID != idU {
		t.Errorf("identifiers not lowercased: %+v", got)
	}
}

// Design 2: an absent list, null, and [] decode as empty.
func TestCompletedDecodeEmptyLists(t *testing.T) {
	const head = `{"enrichment_id":"` + idE + `","scan_id":"` + idS + `","user_id":"` + idU +
		`","family":"history","chains":[{"chain_id":1,"outcome":"not_observed"}],"public_key_recovery":"unresolved",`
	for name, related := range map[string]string{"absent": ``, "null": `"related_addresses":null,`, "empty": `"related_addresses":[],`} {
		got, err := DecodeCompleted([]byte(head + related + `"completed_at":"` + ts + `"}`))
		if err != nil || len(got.RelatedAddresses) != 0 {
			t.Errorf("related_addresses %s: %+v, %v", name, got, err)
		}
	}
	for name, chains := range map[string]string{"absent": ``, "null": `"chains":null,`, "empty": `"chains":[],`} {
		raw := `{"enrichment_id":"` + idE + `","scan_id":"` + idS + `","user_id":"` + idU + `","family":"history",` +
			chains + `"public_key_recovery":"unresolved","completed_at":"` + ts + `"}`
		_, err := DecodeCompleted([]byte(raw))
		mustMalformed(t, "chains "+name, err)
	}
}

// Design 2: malformed shapes wrap ErrMalformed and yield a zero value.
func TestCompletedDecodeMalformedShapes(t *testing.T) {
	pre := `{"enrichment_id":"` + idE + `",`
	for name, raw := range map[string]string{
		"empty": ``, "null": `null`, "array": `[]`, "string": `"x"`, "truncated": `{"chains":`,
		"missing all":      `{}`,
		"id number":        `{"enrichment_id":1}`,
		"family number":    pre + `"family":1}`,
		"chains object":    pre + `"chains":{}}`,
		"chain_id string":  pre + `"chains":[{"chain_id":"1"}]}`,
		"outcome number":   pre + `"chains":[{"outcome":1}]}`,
		"related number":   pre + `"related_addresses":1}`,
		"related chain_id": pre + `"related_addresses":[{"chain_id":"x"}]}`,
		"recovery number":  pre + `"public_key_recovery":1}`,
		"completed_at num": pre + `"completed_at":1}`,
	} {
		m, err := DecodeCompleted([]byte(raw))
		mustMalformed(t, name, err)
		if !reflect.DeepEqual(m, Completed{}) {
			t.Errorf("%s: non-zero value on error: %+v", name, m)
		}
	}
}

// Design 2: the error names the field, never the body.
func TestCompletedErrorOmitsBody(t *testing.T) {
	const marker = "secret-body-marker"
	for name, raw := range map[string]string{
		"family":  `{"enrichment_id":"` + idE + `","scan_id":"` + idS + `","user_id":"` + idU + `","family":"` + marker + `"}`,
		"outcome": `{"enrichment_id":"` + idE + `","scan_id":"` + idS + `","user_id":"` + idU + `","family":"history","chains":[{"chain_id":1,"outcome":"` + marker + `"}],"completed_at":"` + ts + `"}`,
		"json":    `{"family":"` + marker,
	} {
		_, err := DecodeCompleted([]byte(raw))
		mustMalformed(t, name, err)
		if strings.Contains(err.Error(), marker) {
			t.Errorf("%s: error leaks body: %v", name, err)
		}
	}
	err := func() Completed { m := historyUnresolved(); m.Chains[0].Outcome = marker; return m }().Validate()
	mustMalformed(t, "validate", err)
	if strings.Contains(err.Error(), marker) {
		t.Errorf("validate error leaks value: %v", err)
	}
}

// Design 2: closed enums.
func TestCompletedEnumsIsValid(t *testing.T) {
	for o, want := range map[Outcome]bool{
		OutcomeObserved: true, OutcomeNotObserved: true, OutcomeError: true,
		"": false, "Observed": false, "timeout": false, "ok": false,
	} {
		if o.IsValid() != want {
			t.Errorf("Outcome %q: want %v", o, want)
		}
	}
	for r, want := range map[PublicKeyRecovery]bool{
		PublicKeyRecoveryNotRequired: true, PublicKeyRecoveryRecovered: true, PublicKeyRecoveryUnresolved: true,
		"": false, "Recovered": false, "failed": false,
	} {
		if r.IsValid() != want {
			t.Errorf("PublicKeyRecovery %q: want %v", r, want)
		}
	}
}
