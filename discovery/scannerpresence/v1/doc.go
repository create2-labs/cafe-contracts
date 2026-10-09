// Package v1 is the wire contract for scanner presence on the existing subject
// cafe.discovery.scanners.presence.
//
// Scanners publish joined and left. The indexer gateway is not an emitter.
// An unknown type stays valid. onchain_indexer is optional and may be
// etherscan, moralis, or none. A missing, unknown, or invalid value reads as
// unknown, never as none.
//
// The package name is v1. It is distinct from the transaction page contract
// and from cafenatsv01. This package does not subscribe to NATS.
package v1
