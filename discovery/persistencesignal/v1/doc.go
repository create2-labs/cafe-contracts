// Package v1 is the wire contract for persistence.ready.
//
// The body published today is an empty object. This package accepts that object
// and rejects an unreadable body. Extra fields are ignored. Marshal emits {}.
//
// The subject is unchanged. The package name is v1. It is not part of cafenatsv01.
// This package does not call NATS. Queue groups stay in each service.
package v1
