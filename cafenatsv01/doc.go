// Package cafenatsv01 defines versioned NATS message contracts for the policy and
// remediation event families (JSON event_version "v0.1"), including the explicit
// policy.assessment.requested command.
//
// The package provides envelope shapes, per-event payload types, versioned subject
// line constants, and boundary validation. It does not implement brokers, consumers,
// or CPM/Discovery business rules beyond exported vocabulary and required fields.
//
// First-version model-to-wire mapping: see MAPPING.md in this directory.
package cafenatsv01
