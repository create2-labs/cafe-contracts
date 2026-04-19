package v01

import "strings"

// AccountKind classifies the on-chain account model on the wire.
type AccountKind string

const (
	AccountKindEOA                 AccountKind = "eoa"
	AccountKindERC4337SmartAccount AccountKind = "erc4337_smart_account"
	AccountKindDelegatedEOA7702    AccountKind = "delegated_eoa_7702"
	AccountKindContractAccount     AccountKind = "contract_account"
	AccountKindUnknown             AccountKind = "unknown"
)

// IsValid reports whether k is a known exported account kind for v0.1.
func (k AccountKind) IsValid() bool {
	switch k {
	case AccountKindEOA, AccountKindERC4337SmartAccount, AccountKindDelegatedEOA7702,
		AccountKindContractAccount, AccountKindUnknown:
		return true
	default:
		return false
	}
}

// AlgorithmID identifies the observed signing / verification algorithm using exported strings.
// Well-known values are constants; hybrid profiles use the "hybrid_" prefix pattern.
type AlgorithmID string

const (
	AlgorithmSecp256k1ECRecover AlgorithmID = "secp256k1_ecrecover"
	AlgorithmMLDSA44            AlgorithmID = "mldsa44"
	AlgorithmMLDSA65            AlgorithmID = "mldsa65"
	AlgorithmFalcon512          AlgorithmID = "falcon512"
)

const hybridAlgorithmPrefix = "hybrid_"

// IsValidAlgorithmID reports whether id is an allowed exported algorithm identifier:
// a well-known constant or any non-empty string with prefix "hybrid_" and a non-empty suffix.
func IsValidAlgorithmID(id string) bool {
	if id == "" {
		return false
	}
	switch AlgorithmID(id) {
	case AlgorithmSecp256k1ECRecover, AlgorithmMLDSA44, AlgorithmMLDSA65, AlgorithmFalcon512:
		return true
	default:
		return strings.HasPrefix(id, hybridAlgorithmPrefix) && len(id) > len(hybridAlgorithmPrefix)
	}
}

// CurrentPQPosture summarizes post-quantum readiness as exported on the Discovery → CPM boundary
// for v0.1 payloads. Derivation rules live in Discovery, not in this package.
type CurrentPQPosture string

const (
	PQPostureClassicalOnly CurrentPQPosture = "classical_only"
	PQPostureHybrid        CurrentPQPosture = "hybrid"
	PQPostureFullPQ        CurrentPQPosture = "full_pq"
	PQPostureUnknown       CurrentPQPosture = "unknown"
)

// IsValid reports whether p is a known exported posture value for v0.1.
func (p CurrentPQPosture) IsValid() bool {
	switch p {
	case PQPostureClassicalOnly, PQPostureHybrid, PQPostureFullPQ, PQPostureUnknown:
		return true
	default:
		return false
	}
}

// SubjectType identifies the kind of subject referenced in the envelope.
type SubjectType string

const (
	// SubjectTypeWallet is the subject type for wallet-scoped observations.
	SubjectTypeWallet SubjectType = "wallet"
)

// IsValid reports whether t is a known exported subject type for v0.1.
func (t SubjectType) IsValid() bool {
	switch t {
	case SubjectTypeWallet:
		return true
	default:
		return false
	}
}
