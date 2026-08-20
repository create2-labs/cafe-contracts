// Package cafenatsv01 defines versioned NATS message contracts for the policy and
// remediation event families (JSON event_version "v0.1"), including the explicit
// policy.assessment.requested command.
//
// Assessment requested payloads use wire v0.2: crypto_policy_id + observation scan
// snapshot. Legacy selection_request / layer-B fields are rejected at decode time.
//
// The package provides envelope shapes, per-event payload types, versioned subject
// line constants, and boundary validation. It does not implement brokers, consumers,
// or CPM/Discovery business rules beyond exported vocabulary and required fields.
//
// Model-to-wire mapping: see MAPPING.md in this directory.
package cafenatsv01
