package v1

import "embed"

// Vectors holds the JSON fixtures of this contract.
// Later importers reuse these bytes instead of restating the wire shapes.
//
//go:embed testdata
var Vectors embed.FS
