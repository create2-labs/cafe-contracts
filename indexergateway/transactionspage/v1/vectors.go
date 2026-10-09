package v1

import "embed"

// Vectors holds the JSON fixtures of this contract.
// PR-1 and PR-2 reuse these bytes instead of restating the wire shapes.
//
//go:embed testdata
var Vectors embed.FS
