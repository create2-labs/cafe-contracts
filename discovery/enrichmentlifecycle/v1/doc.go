// Package v1 is the wire contract for the enrichment lifecycle messages
// enrichment.started, enrichment.completed, and enrichment.failed, published on
// stream ENRICHMENT_LIFECYCLE. All carry the family of the request.
//
// Only the worker publishes enrichment.failed, on the last delivery of a
// request. The 15-minute abandon is a Discovery transition, not a message, so
// no timeout value exists in this contract.
//
// enrichment.completed carries one outcome per requested chain (observed,
// not_observed, or error) and the family-specific facts. A message whose
// chains are all in error is valid: Discovery derives the scan state. The code
// of an error chain is a free non-empty string, so this contract is not bound
// to the gateway contracts. The known codes are invalid_request,
// unsupported_chain, no_provider, provider_error, deadline_exceeded,
// rate_limit_exceeded (indexergateway/transactionspage/v1 and
// indexergateway/contractcreation/v1), and unsupported_operation
// (indexergateway/contractcreation/v1).
//
// The package name is v1. It is not part of cafenatsv01. This package does not
// subscribe to NATS and does not choose a subject for a publisher.
package v1
