// Package v1 is the wire contract for the enrichment lifecycle messages
// enrichment.started and enrichment.failed, published on stream
// ENRICHMENT_LIFECYCLE. Both carry the family of the request.
//
// Only the worker publishes enrichment.failed, on the last delivery of a
// request. The 15-minute abandon is a Discovery transition, not a message, so
// no timeout value exists in this contract. The enrichment.completed message
// comes with a later pull request (OC-P11).
//
// The package name is v1. It is not part of cafenatsv01. This package does not
// subscribe to NATS and does not choose a subject for a publisher.
package v1
