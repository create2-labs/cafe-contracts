package v01

import "errors"

// Validation errors for discovery.wallet.observed v0.1. Callers can use errors.Is for branching.
var (
	ErrEventIDRequired   = errors.New("discovery.wallet.observed v0.1: event_id is required")
	ErrEventType         = errors.New("discovery.wallet.observed v0.1: event_type must be discovery.wallet.observed")
	ErrEventVersion      = errors.New("discovery.wallet.observed v0.1: event_version must be v0.1")
	ErrProducer          = errors.New("discovery.wallet.observed v0.1: producer must be cafe-discovery for this contract revision")
	ErrSubjectType       = errors.New("discovery.wallet.observed v0.1: subject.type must be a known exported subject type")
	ErrSubjectIDRequired = errors.New("discovery.wallet.observed v0.1: subject.id is required")
	ErrAccountKind       = errors.New("discovery.wallet.observed v0.1: payload.account_kind must be a known exported account kind")
	ErrAlgorithmID       = errors.New("discovery.wallet.observed v0.1: payload.current_algorithm must be a known algorithm id or hybrid_*")
	ErrCurrentPQPosture  = errors.New("discovery.wallet.observed v0.1: payload.current_pq_posture must be a known exported posture")
)
