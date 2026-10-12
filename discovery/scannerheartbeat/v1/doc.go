// Package v1 is the wire contract for scanner.heartbeat.tls and scanner.heartbeat.wallet.
//
// The body is scanner_id, kind, and timestamp, the three fields published today.
// Each is a non-empty string. Publishers send timestamp as RFC3339; this package
// does not parse the clock. kind is not limited to tls or wallet.
//
// The subjects are unchanged. The package name is v1. It is not part of cafenatsv01.
// This package does not call NATS. Queue groups stay in each service.
package v1
