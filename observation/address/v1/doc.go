// Package v1 is the wire contract for the qualification result of an address:
// the result carried by scan.completed for the scan kind wallet.
//
// A Result holds the address, its account type, the observation time, the
// chains that were read with their per-chain qualification, and the chains
// that could not be observed. For an eoa it also holds the eoa facts (key
// exposure, algorithm, NIST level, risk score) and the delegation targets as
// related addresses; for a contract it holds the delegation targets of its
// not_deployed chains and the anomaly raised when a not_deployed chain has a
// nonce above zero or a delegation target. Validate checks that chains is
// non-empty, that chains and unobserved_chain_ids are disjoint, that the
// qualifications are consistent with the account type, and that the facts are
// consistent with the chains. A chain without contract code carries its
// nonce unless it is delegated, and the key of a delegated eoa chain is
// exposed whatever its nonce. The package does not read chains and does not
// score keys.
package v1
