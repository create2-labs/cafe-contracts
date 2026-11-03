// Package v1 is the wire contract for an enrichment request.
//
// A request is published on enrichment.requested.history or
// enrichment.requested.deployer, on stream ENRICHMENT_REQUESTED. The body
// carries enrichment_id, scan_id, user_id, address, family, and chains. The
// request carries a subset of the PENDING enrichment record (ADR decision D4):
// the fields a scanner needs to read the chains.
// The package name is v1. It is not part of cafenatsv01. This package does not
// subscribe to NATS and does not choose a subject for a publisher.
package v1
