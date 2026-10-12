// Package v1 is the wire contract for scan.started, scan.completed, scan.failed, and scan.ready.
// Empty optional fields stay omitted. findings, metadata, and result stay opaque JSON in the order received.
// scan.ready keeps user_id, kind, endpoint, address, and status. Identifiers are canonical UUIDs. Not part of cafenatsv01.
package v1
