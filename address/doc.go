// Package address provides small shared helpers for EVM address handling at
// service boundaries.
//
// Contract intent:
//   - accept valid hex addresses in any case,
//   - normalize to lowercase for machine keys and comparisons,
//   - optionally render EIP-55 checksum for user-facing display.
package address
