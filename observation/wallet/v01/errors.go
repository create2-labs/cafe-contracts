package v01

import (
	"errors"
	"fmt"
)

// Validation errors for cafe.discovery.wallet.observed v0.1. Callers can use errors.Is for branching.
var (
	errorPrefix          = fmt.Sprintf("%s %s: ", EventTypeWalletObserved, EventVersion)
	ErrEventIDRequired   = errors.New(errorPrefix + "event_id is required")
	ErrEventType         = errors.New(errorPrefix + "event_type must be " + EventTypeWalletObserved)
	ErrEventVersion      = errors.New(errorPrefix + "event_version must be " + EventVersion)
	ErrProducer          = errors.New(errorPrefix + "producer must be cafe-discovery for this contract revision")
	ErrSubjectType       = errors.New(errorPrefix + "subject.type must be a known exported subject type")
	ErrSubjectIDRequired = errors.New(errorPrefix + "subject.id is required")
	ErrAccountKind       = errors.New(errorPrefix + "payload.account_kind must be a known exported account kind")
	ErrAlgorithmID       = errors.New(errorPrefix + "payload.current_algorithm must be a known algorithm id or hybrid_*")
	ErrCurrentPQPosture  = errors.New(errorPrefix + "payload.current_pq_posture must be a known exported posture")
)
