// Package v1 is the wire contract for the qualification result of an address:
// the result carried by scan.completed for the scan kind wallet.
//
// A Result holds the address, its account type, the observation time, the
// chains that were read with their per-chain qualification, and the chains
// that could not be observed. Validate checks that chains is non-empty, that
// chains and unobserved_chain_ids are disjoint, and that the qualifications
// are consistent with the account type. The package does not read chains and
// does not score keys.
package v1
