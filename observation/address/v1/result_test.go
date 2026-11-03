package v1

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

const addr = "0x1111111111111111111111111111111111111111"

var at = time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)

func u(n uint64) *uint64 { return &n }
func b(v bool) *bool     { return &v }

func eoaChain(id int64, nonce *uint64, exposed *bool) Chain {
	return Chain{ChainID: id, Qualification: QualificationEOA, Nonce: nonce, KeyExposed: exposed}
}

func res(t AccountType, chains []Chain, unobserved ...int64) Result {
	return Result{Address: addr, AccountType: t, ObservedAt: at, Chains: chains, UnobservedChainIDs: unobserved}
}

const dtarget = "0x2222222222222222222222222222222222222222"

func goodEOA(exposed bool) *EOA {
	return &EOA{KeyExposed: exposed, Algorithm: "ECDSA-secp256k1", NISTLevel: 1, RiskScore: 0.5}
}

func mod(r Result, f func(*Result)) Result { f(&r); return r }

func deleg(chain int64, a string) Related {
	return Related{Role: RelatedRoleDelegationTarget, ChainID: chain, Address: a}
}

// eoaBase is a valid eoa result: chain 1 exposed, chain 10 not, 137 unobserved.
func eoaBase() Result {
	r := res(AccountTypeEOA, []Chain{eoaChain(1, u(3), b(true)), eoaChain(10, u(0), b(false))}, 137)
	r.EOA = goodEOA(true)
	return r
}

// contractBase has a contract chain and a not_deployed chain with nonce n.
func contractBase(n uint64) Result {
	return res(AccountTypeContract, []Chain{{ChainID: 1, Qualification: QualificationContract}, {ChainID: 10, Qualification: QualificationNotDeployed, Nonce: u(n)}}, 137)
}

func blankBase() Result {
	return res(AccountTypeBlank, []Chain{{ChainID: 1, Qualification: QualificationBlank, Nonce: u(0)}})
}

func mustMalformed(t *testing.T, name string, err error) {
	t.Helper()
	if !errors.Is(err, ErrMalformed) {
		t.Errorf("%s: want ErrMalformed, got %v", name, err)
	}
}

// Rule: every enum value is known, others are not.
func TestEnumIsValid(t *testing.T) {
	for _, v := range []AccountType{AccountTypeEOA, AccountTypeContract, AccountTypeBlank} {
		if !v.IsValid() {
			t.Errorf("account type %q should be valid", v)
		}
	}
	for _, v := range []AccountType{"", "EOA", "not_deployed", "wallet"} {
		if v.IsValid() {
			t.Errorf("account type %q should be invalid", v)
		}
	}
	for _, v := range []Qualification{QualificationEOA, QualificationContract, QualificationNotDeployed, QualificationBlank} {
		if !v.IsValid() {
			t.Errorf("qualification %q should be valid", v)
		}
	}
	for _, v := range []Qualification{"", "Contract", "unknown", "deployed"} {
		if v.IsValid() {
			t.Errorf("qualification %q should be invalid", v)
		}
	}
	// OC-P08 note, roles and anomalies.
	for _, v := range []RelatedRole{RelatedRoleDelegationTarget, RelatedRoleDeployer} {
		if !v.IsValid() {
			t.Errorf("role %q should be valid", v)
		}
	}
	for _, v := range []RelatedRole{"", "Deployer", "owner", "delegate"} {
		if v.IsValid() {
			t.Errorf("role %q should be invalid", v)
		}
	}
	if !AnomalyContractWithEOAActivity.IsValid() {
		t.Error("anomaly contract_with_eoa_activity should be valid")
	}
	for _, v := range []Anomaly{"", "Contract_With_EOA_Activity", "unknown"} {
		if v.IsValid() {
			t.Errorf("anomaly %q should be invalid", v)
		}
	}
}

// CA1.1-CA1.3, CA2.2: one valid result per account type.
func TestValidateAcceptsValidResults(t *testing.T) {
	cases := map[string]Result{
		"eoa":               eoaBase(),
		"eoa unknown nonce": mod(res(AccountTypeEOA, []Chain{eoaChain(1, nil, b(false))}), func(r *Result) { r.EOA = goodEOA(false) }),
		// CA1.3: delegation target on an eoa chain, one per chain.
		"eoa delegation": mod(eoaBase(), func(r *Result) {
			r.RelatedAddresses = []Related{deleg(1, dtarget), deleg(10, "opaque-not-checked")}
		}),
		// OC-P08 note: bounds and non-fixed algorithm.
		"eoa risk 0, nist 5": mod(eoaBase(), func(r *Result) { r.EOA = &EOA{KeyExposed: true, Algorithm: "ml-dsa-44", NISTLevel: 5, RiskScore: 0} }),
		"eoa risk 1, nist 1": mod(eoaBase(), func(r *Result) { r.EOA = &EOA{KeyExposed: true, Algorithm: "x", NISTLevel: 1, RiskScore: 1} }),
		// CA1.7: not_deployed with nonce > 0 gives the anomaly.
		"contract anomaly":         mod(contractBase(4), func(r *Result) { r.Anomalies = []Anomaly{AnomalyContractWithEOAActivity} }),
		"contract nonce unknown":   res(AccountTypeContract, []Chain{{ChainID: 1, Qualification: QualificationContract}, {ChainID: 10, Qualification: QualificationNotDeployed}}),
		"contract empty lists":     mod(contractBase(0), func(r *Result) { r.RelatedAddresses, r.Anomalies = []Related{}, []Anomaly{} }),
		"contract with undeployed": res(AccountTypeContract, []Chain{{ChainID: 1, Qualification: QualificationContract}, {ChainID: 10, Qualification: QualificationNotDeployed, Nonce: u(0)}}, 137),
		"blank":                    blankBase(),
		"unobserved nil":           res(AccountTypeBlank, []Chain{{ChainID: 1, Qualification: QualificationBlank}}),
	}
	for name, r := range cases {
		if err := r.Validate(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// CA1.4, CA2.2: the consistency table and chain rules reject violations.
func TestValidateRejects(t *testing.T) {
	ct := Chain{ChainID: 1, Qualification: QualificationContract}
	nd := Chain{ChainID: 10, Qualification: QualificationNotDeployed}
	bl := Chain{ChainID: 1, Qualification: QualificationBlank}
	cases := map[string]Result{
		"contract without contract chain": res(AccountTypeContract, []Chain{nd}),
		"contract with eoa chain":         res(AccountTypeContract, []Chain{ct, eoaChain(10, u(1), b(true))}),
		"contract with blank chain":       res(AccountTypeContract, []Chain{ct, {ChainID: 10, Qualification: QualificationBlank}}),
		"eoa with contract chain":         res(AccountTypeEOA, []Chain{eoaChain(1, u(1), b(true)), {ChainID: 10, Qualification: QualificationContract}}),
		"eoa with not_deployed chain":     res(AccountTypeEOA, []Chain{eoaChain(1, u(1), b(true)), nd}),
		"eoa with blank chain":            res(AccountTypeEOA, []Chain{eoaChain(1, u(1), b(true)), {ChainID: 10, Qualification: QualificationBlank}}),
		"blank with eoa chain":            res(AccountTypeBlank, []Chain{bl, eoaChain(10, u(0), b(false))}),
		"blank with contract chain":       res(AccountTypeBlank, []Chain{bl, {ChainID: 10, Qualification: QualificationContract}}),
		"blank with not_deployed chain":   res(AccountTypeBlank, []Chain{bl, nd}),
		// OC-P07 note, blank nonce: a blank chain has nonce 0 or unknown.
		"blank chain nonce > 0": res(AccountTypeBlank, []Chain{{ChainID: 1, Qualification: QualificationBlank, Nonce: u(1)}}),
		// OC-P07 note: key_exposed rules.
		"eoa key_exposed missing":             res(AccountTypeEOA, []Chain{eoaChain(1, u(1), nil)}),
		"eoa key_exposed false, nonce > 0":    res(AccountTypeEOA, []Chain{eoaChain(1, u(2), b(false))}),
		"eoa key_exposed true, nonce 0":       res(AccountTypeEOA, []Chain{eoaChain(1, u(0), b(true))}),
		"eoa key_exposed true, nonce unknown": res(AccountTypeEOA, []Chain{eoaChain(1, nil, b(true))}),
		"key_exposed on contract chain":       res(AccountTypeContract, []Chain{{ChainID: 1, Qualification: QualificationContract, KeyExposed: b(false)}}),
		"key_exposed on not_deployed chain":   res(AccountTypeContract, []Chain{ct, {ChainID: 10, Qualification: QualificationNotDeployed, KeyExposed: b(false)}}),
		"key_exposed on blank chain":          res(AccountTypeBlank, []Chain{{ChainID: 1, Qualification: QualificationBlank, KeyExposed: b(false)}}),
		// OC-P08 note: eoa block.
		"eoa block missing":                      mod(eoaBase(), func(r *Result) { r.EOA = nil }),
		"eoa block on contract":                  mod(contractBase(0), func(r *Result) { r.EOA = goodEOA(false) }),
		"eoa block on blank":                     mod(blankBase(), func(r *Result) { r.EOA = goodEOA(false) }),
		"eoa risk_score < 0":                     mod(eoaBase(), func(r *Result) { r.EOA.RiskScore = -0.1 }),
		"eoa risk_score > 1":                     mod(eoaBase(), func(r *Result) { r.EOA.RiskScore = 1.1 }),
		"eoa risk_score NaN":                     mod(eoaBase(), func(r *Result) { r.EOA.RiskScore = math.NaN() }),
		"eoa risk_score +Inf":                    mod(eoaBase(), func(r *Result) { r.EOA.RiskScore = math.Inf(1) }),
		"eoa risk_score -Inf":                    mod(eoaBase(), func(r *Result) { r.EOA.RiskScore = math.Inf(-1) }),
		"eoa nist_level 0":                       mod(eoaBase(), func(r *Result) { r.EOA.NISTLevel = 0 }),
		"eoa nist_level 6":                       mod(eoaBase(), func(r *Result) { r.EOA.NISTLevel = 6 }),
		"eoa algorithm empty":                    mod(eoaBase(), func(r *Result) { r.EOA.Algorithm = "" }),
		"eoa key_exposed false, a chain exposed": mod(eoaBase(), func(r *Result) { r.EOA.KeyExposed = false }),
		"eoa key_exposed true, no chain exposed": mod(eoaBase(), func(r *Result) {
			r.Chains = []Chain{eoaChain(1, u(0), b(false))}
			r.UnobservedChainIDs = nil
		}),
		// OC-P08 note: related addresses.
		"related deployer role":    mod(eoaBase(), func(r *Result) { r.RelatedAddresses = []Related{{RelatedRoleDeployer, 1, dtarget}} }),
		"related unknown role":     mod(eoaBase(), func(r *Result) { r.RelatedAddresses = []Related{{"owner", 1, dtarget}} }),
		"related empty role":       mod(eoaBase(), func(r *Result) { r.RelatedAddresses = []Related{{"", 1, dtarget}} }),
		"related on contract":      mod(contractBase(0), func(r *Result) { r.RelatedAddresses = []Related{deleg(1, dtarget)} }),
		"related on blank":         mod(blankBase(), func(r *Result) { r.RelatedAddresses = []Related{deleg(1, dtarget)} }),
		"related unobserved chain": mod(eoaBase(), func(r *Result) { r.RelatedAddresses = []Related{deleg(137, dtarget)} }),
		"related unknown chain":    mod(eoaBase(), func(r *Result) { r.RelatedAddresses = []Related{deleg(999, dtarget)} }),
		"related zero chain":       mod(eoaBase(), func(r *Result) { r.RelatedAddresses = []Related{deleg(0, dtarget)} }),
		"related two on one chain": mod(eoaBase(), func(r *Result) { r.RelatedAddresses = []Related{deleg(1, dtarget), deleg(1, "0xother")} }),
		"related duplicate":        mod(eoaBase(), func(r *Result) { r.RelatedAddresses = []Related{deleg(1, dtarget), deleg(1, dtarget)} }),
		"related empty address":    mod(eoaBase(), func(r *Result) { r.RelatedAddresses = []Related{deleg(1, "")} }),
		// OC-P08 note, CA1.7: anomalies.
		"anomaly on eoa":           mod(eoaBase(), func(r *Result) { r.Anomalies = []Anomaly{AnomalyContractWithEOAActivity} }),
		"anomaly on blank":         mod(blankBase(), func(r *Result) { r.Anomalies = []Anomaly{AnomalyContractWithEOAActivity} }),
		"anomaly missing":          contractBase(4),
		"anomaly without activity": mod(contractBase(0), func(r *Result) { r.Anomalies = []Anomaly{AnomalyContractWithEOAActivity} }),
		"anomaly with unknown nonce": mod(res(AccountTypeContract, []Chain{{ChainID: 1, Qualification: QualificationContract}, {ChainID: 10, Qualification: QualificationNotDeployed}}), func(r *Result) {
			r.Anomalies = []Anomaly{AnomalyContractWithEOAActivity}
		}),
		"anomaly duplicate": mod(contractBase(4), func(r *Result) {
			r.Anomalies = []Anomaly{AnomalyContractWithEOAActivity, AnomalyContractWithEOAActivity}
		}),
		"anomaly unknown": mod(contractBase(4), func(r *Result) {
			r.Anomalies = []Anomaly{AnomalyContractWithEOAActivity, "other"}
		}),
		// OC-P07 note: chains and chain identifiers.
		"nil chains":                    res(AccountTypeBlank, nil, 1),
		"empty chains":                  res(AccountTypeBlank, []Chain{}, 1),
		"zero chain id":                 res(AccountTypeBlank, []Chain{{ChainID: 0, Qualification: QualificationBlank}}),
		"negative chain id":             res(AccountTypeBlank, []Chain{{ChainID: -1, Qualification: QualificationBlank}}),
		"duplicate chain id":            res(AccountTypeBlank, []Chain{bl, bl}),
		"zero unobserved id":            res(AccountTypeBlank, []Chain{bl}, 0),
		"negative unobserved id":        res(AccountTypeBlank, []Chain{bl}, -5),
		"duplicate unobserved id":       res(AccountTypeBlank, []Chain{bl}, 7, 7),
		"chains and unobserved overlap": res(AccountTypeBlank, []Chain{bl}, 1),
		// OC-P07 note: address, observed_at, enums.
		"empty address":    {AccountType: AccountTypeBlank, ObservedAt: at, Chains: []Chain{bl}},
		"zero observed_at": {Address: addr, AccountType: AccountTypeBlank, Chains: []Chain{bl}},
		"invalid account":  res("wallet", []Chain{bl}),
		"empty account":    res("", []Chain{bl}),
		"invalid qualif":   res(AccountTypeBlank, []Chain{{ChainID: 1, Qualification: "unknown"}}),
		"empty qualif":     res(AccountTypeBlank, []Chain{{ChainID: 1}}),
	}
	for name, r := range cases {
		mustMalformed(t, name, r.Validate())
	}
}

// OC-P08 note: failures name the field path, never a value.
func TestFactErrorsNameField(t *testing.T) {
	cases := map[string]struct {
		r    Result
		path string
	}{
		"eoa":              {mod(eoaBase(), func(r *Result) { r.EOA = nil }), "eoa"},
		"eoa.risk_score":   {mod(eoaBase(), func(r *Result) { r.EOA.RiskScore = 2 }), "eoa.risk_score"},
		"eoa.nist_level":   {mod(eoaBase(), func(r *Result) { r.EOA.NISTLevel = 9 }), "eoa.nist_level"},
		"eoa.algorithm":    {mod(eoaBase(), func(r *Result) { r.EOA.Algorithm = "" }), "eoa.algorithm"},
		"eoa.key_exposed":  {mod(eoaBase(), func(r *Result) { r.EOA.KeyExposed = false }), "eoa.key_exposed"},
		"related.role":     {mod(eoaBase(), func(r *Result) { r.RelatedAddresses = []Related{{RelatedRoleDeployer, 1, dtarget}} }), "related_addresses.role"},
		"related.chain_id": {mod(eoaBase(), func(r *Result) { r.RelatedAddresses = []Related{deleg(999, dtarget)} }), "related_addresses.chain_id"},
		"related.address":  {mod(eoaBase(), func(r *Result) { r.RelatedAddresses = []Related{deleg(1, "")} }), "related_addresses.address"},
		"anomalies":        {contractBase(4), "anomalies"},
	}
	for name, c := range cases {
		err := c.r.Validate()
		mustMalformed(t, name, err)
		if err != nil && !strings.Contains(err.Error(), c.path) {
			t.Errorf("%s: error %q lacks %q", name, err, c.path)
		}
	}
}

// OC-P07 note: observed_at accepts any RFC 3339 offset.
func TestObservedAtOffset(t *testing.T) {
	r := res(AccountTypeBlank, []Chain{{ChainID: 1, Qualification: QualificationBlank}})
	r.ObservedAt = time.Date(2026, 10, 23, 12, 0, 0, 0, time.FixedZone("", 2*3600))
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
}

// CA1.1-CA1.5, CA2.1, CA2.2: the vectors, ok_* decode and bad_* are rejected.
func TestResultVectors(t *testing.T) {
	dir := "testdata"
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var ok, bad int
	for _, e := range entries {
		name := e.Name()
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		switch {
		case strings.HasPrefix(name, "ok_"):
			ok++
			r, err := DecodeResult(raw)
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			if err := r.Validate(); err != nil {
				t.Fatalf("%s validate: %v", name, err)
			}
			out, err := json.Marshal(r)
			if err != nil {
				t.Fatal(err)
			}
			again, err := DecodeResult(out)
			if err != nil {
				t.Fatalf("%s round trip: %v", name, err)
			}
			if !reflect.DeepEqual(r, again) {
				t.Fatalf("%s round trip:\n%+v\nvs\n%+v", name, r, again)
			}
		case strings.HasPrefix(name, "bad_"):
			bad++
			r, err := DecodeResult(raw)
			mustMalformed(t, name, err)
			if !reflect.DeepEqual(r, Result{}) {
				t.Fatalf("%s: non-zero result on error: %+v", name, r)
			}
		}
	}
	if ok < 3 || bad < 8 {
		t.Fatalf("vectors incomplete: ok=%d bad=%d", ok, bad)
	}
}

// ADR 6.2 example, CA1.1, CA2.1, CA2.2: unknown OC-P08 fields are ignored.
func TestDecodeADRExample(t *testing.T) {
	raw, err := os.ReadFile("testdata/ok_contract_adr62.json")
	if err != nil {
		t.Fatal(err)
	}
	r, err := DecodeResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := Result{
		Address:     addr,
		AccountType: AccountTypeContract,
		ObservedAt:  at,
		Chains: []Chain{
			{ChainID: 1, Qualification: QualificationContract},
			{ChainID: 10, Qualification: QualificationNotDeployed, Nonce: u(0)},
		},
		UnobservedChainIDs: []int64{137},
	}
	if !reflect.DeepEqual(r, want) {
		t.Fatalf("got %+v", r)
	}
}

// CA1.5: observed_at is kept with its instant for any offset.
func TestDecodeObservedAtOffset(t *testing.T) {
	raw, err := os.ReadFile("testdata/ok_blank.json")
	if err != nil {
		t.Fatal(err)
	}
	r, err := DecodeResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !r.ObservedAt.Equal(time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("observed_at %v", r.ObservedAt)
	}
}

// CA1.2: nonce stays optional, unknown is not zero; key_exposed is kept.
func TestDecodeEOAFacts(t *testing.T) {
	raw, err := os.ReadFile("testdata/ok_eoa.json")
	if err != nil {
		t.Fatal(err)
	}
	r, err := DecodeResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Chains) != 3 {
		t.Fatalf("chains %+v", r.Chains)
	}
	if r.Chains[0].Nonce == nil || *r.Chains[0].Nonce != 3 || r.Chains[0].KeyExposed == nil || !*r.Chains[0].KeyExposed {
		t.Fatalf("chain 0 %+v", r.Chains[0])
	}
	if r.Chains[1].Nonce == nil || *r.Chains[1].Nonce != 0 || *r.Chains[1].KeyExposed {
		t.Fatalf("chain 1 %+v", r.Chains[1])
	}
	if r.Chains[2].Nonce != nil || r.Chains[2].KeyExposed == nil || *r.Chains[2].KeyExposed {
		t.Fatalf("chain 2 %+v", r.Chains[2])
	}
}

// CA2.1, OC-P07 note: unobserved_chain_ids absent or null is empty.
func TestDecodeUnobservedAbsentOrNull(t *testing.T) {
	head := `{"address":"` + addr + `","account_type":"blank","observed_at":"2026-10-23T10:00:00Z","chains":[{"chain_id":1,"qualification":"blank"}]`
	for name, raw := range map[string]string{"absent": head + `}`, "null": head + `,"unobserved_chain_ids":null}`, "empty": head + `,"unobserved_chain_ids":[]}`} {
		r, err := DecodeResult([]byte(raw))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(r.UnobservedChainIDs) != 0 {
			t.Fatalf("%s: %v", name, r.UnobservedChainIDs)
		}
	}
}

// OC-P07 note: malformed shapes and wrong types wrap ErrMalformed, zero Result.
func TestDecodeMalformedShapes(t *testing.T) {
	ok := `"address":"` + addr + `","account_type":"blank","observed_at":"2026-10-23T10:00:00Z","chains":[{"chain_id":1,"qualification":"blank"}]`
	cases := map[string]string{
		"empty body":          ``,
		"not json":            `{`,
		"string":              `"x"`,
		"number":              `1`,
		"array":               `[]`,
		"null":                `null`,
		"address number":      `{"address":1,"account_type":"blank","observed_at":"2026-10-23T10:00:00Z","chains":[{"chain_id":1,"qualification":"blank"}]}`,
		"address missing":     `{"account_type":"blank","observed_at":"2026-10-23T10:00:00Z","chains":[{"chain_id":1,"qualification":"blank"}]}`,
		"address empty":       `{"address":"","account_type":"blank","observed_at":"2026-10-23T10:00:00Z","chains":[{"chain_id":1,"qualification":"blank"}]}`,
		"account_type num":    `{"address":"` + addr + `","account_type":1,"observed_at":"2026-10-23T10:00:00Z","chains":[{"chain_id":1,"qualification":"blank"}]}`,
		"account_type bad":    `{"address":"` + addr + `","account_type":"wallet","observed_at":"2026-10-23T10:00:00Z","chains":[{"chain_id":1,"qualification":"blank"}]}`,
		"observed_at missing": `{"address":"` + addr + `","account_type":"blank","chains":[{"chain_id":1,"qualification":"blank"}]}`,
		"observed_at zero":    `{"address":"` + addr + `","account_type":"blank","observed_at":"0001-01-01T00:00:00Z","chains":[{"chain_id":1,"qualification":"blank"}]}`,
		"observed_at text":    `{"address":"` + addr + `","account_type":"blank","observed_at":"yesterday","chains":[{"chain_id":1,"qualification":"blank"}]}`,
		"observed_at number":  `{"address":"` + addr + `","account_type":"blank","observed_at":1,"chains":[{"chain_id":1,"qualification":"blank"}]}`,
		"chains object":       `{"address":"` + addr + `","account_type":"blank","observed_at":"2026-10-23T10:00:00Z","chains":{}}`,
		"chains missing":      `{"address":"` + addr + `","account_type":"blank","observed_at":"2026-10-23T10:00:00Z"}`,
		"chains null":         `{"address":"` + addr + `","account_type":"blank","observed_at":"2026-10-23T10:00:00Z","chains":null}`,
		"chain id string":     `{"address":"` + addr + `","account_type":"blank","observed_at":"2026-10-23T10:00:00Z","chains":[{"chain_id":"1","qualification":"blank"}]}`,
		"chain id zero":       `{"address":"` + addr + `","account_type":"blank","observed_at":"2026-10-23T10:00:00Z","chains":[{"chain_id":0,"qualification":"blank"}]}`,
		"chain id duplicate":  `{"address":"` + addr + `","account_type":"blank","observed_at":"2026-10-23T10:00:00Z","chains":[{"chain_id":1,"qualification":"blank"},{"chain_id":1,"qualification":"blank"}]}`,
		"qualification bad":   `{"address":"` + addr + `","account_type":"blank","observed_at":"2026-10-23T10:00:00Z","chains":[{"chain_id":1,"qualification":"x"}]}`,
		"nonce string":        `{"address":"` + addr + `","account_type":"blank","observed_at":"2026-10-23T10:00:00Z","chains":[{"chain_id":1,"qualification":"blank","nonce":"0"}]}`,
		"nonce negative":      `{"address":"` + addr + `","account_type":"blank","observed_at":"2026-10-23T10:00:00Z","chains":[{"chain_id":1,"qualification":"blank","nonce":-1}]}`,
		"key_exposed string":  `{"address":"` + addr + `","account_type":"eoa","observed_at":"2026-10-23T10:00:00Z","chains":[{"chain_id":1,"qualification":"eoa","nonce":0,"key_exposed":"false"}]}`,
		"unobserved string":   `{` + ok + `,"unobserved_chain_ids":"1"}`,
		"unobserved element":  `{` + ok + `,"unobserved_chain_ids":["1"]}`,
		"unobserved zero":     `{` + ok + `,"unobserved_chain_ids":[0]}`,
		"unobserved dup":      `{` + ok + `,"unobserved_chain_ids":[2,2]}`,
		"unobserved overlap":  `{` + ok + `,"unobserved_chain_ids":[1]}`,
	}
	for name, raw := range cases {
		r, err := DecodeResult([]byte(raw))
		mustMalformed(t, name, err)
		if !reflect.DeepEqual(r, Result{}) {
			t.Errorf("%s: non-zero result on error: %+v", name, r)
		}
	}
}

// ErrMalformed text names the field, not the body.
func TestErrorOmitsBody(t *testing.T) {
	_, err := DecodeResult([]byte(`{"address":"` + addr + `","account_type":"secret-body-marker"}`))
	mustMalformed(t, "account_type", err)
	if strings.Contains(err.Error(), "secret-body-marker") {
		t.Fatalf("error leaks body: %v", err)
	}
}

// OC-P08 note: related_addresses and anomalies absent, null or [] decode as empty.
func TestDecodeFactListsAbsentOrNull(t *testing.T) {
	head := `{"address":"` + addr + `","account_type":"blank","observed_at":"2026-10-23T10:00:00Z","chains":[{"chain_id":1,"qualification":"blank"}]`
	for name, tail := range map[string]string{
		"absent": `}`,
		"null":   `,"related_addresses":null,"anomalies":null,"eoa":null}`,
		"empty":  `,"related_addresses":[],"anomalies":[]}`,
	} {
		r, err := DecodeResult([]byte(head + tail))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(r.RelatedAddresses) != 0 || len(r.Anomalies) != 0 || r.EOA != nil {
			t.Fatalf("%s: %+v", name, r)
		}
	}
}

// CA1.3, CA1.7: the eoa facts, delegation target and anomaly survive decoding.
func TestDecodeFacts(t *testing.T) {
	read := func(name string) Result {
		raw, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		r, err := DecodeResult(raw)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		return r
	}
	e := read("ok_eoa_delegation.json")
	if !reflect.DeepEqual(e.RelatedAddresses, []Related{deleg(10, dtarget)}) {
		t.Fatalf("related %+v", e.RelatedAddresses)
	}
	if !reflect.DeepEqual(e.EOA, &EOA{KeyExposed: true, Algorithm: "ECDSA-secp256k1", NISTLevel: 1, RiskScore: 0.9}) {
		t.Fatalf("eoa %+v", e.EOA)
	}
	c := read("ok_contract_anomaly.json")
	if !reflect.DeepEqual(c.Anomalies, []Anomaly{AnomalyContractWithEOAActivity}) || c.EOA != nil {
		t.Fatalf("contract %+v", c)
	}
}
