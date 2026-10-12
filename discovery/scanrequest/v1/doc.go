// Package v1 is the wire contract for a scan request.
//
// The subjects are unchanged: cafe.discovery.wallet.scan,
// cafe.discovery.tls.scan, scan.requested.wallet, and scan.requested.tls.
// A wallet request carries scan_id, user_id, and address. A TLS request
// carries scan_id, user_id, endpoint, and is_default. scan_id and user_id
// are canonical UUID strings. The nil UUID is valid: a default TLS endpoint
// publishes it as user_id, and a wallet scanner still accepts it as scan_id.
//
// The package name is v1. It is not part of cafenatsv01. This package does
// not subscribe to NATS and does not choose which of the two subjects a
// publisher uses.
package v1
