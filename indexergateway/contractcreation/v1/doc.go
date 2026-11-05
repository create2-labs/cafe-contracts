// Package v1 is the wire contract of the contract creation lookup on
// cafe.indexer.gateway.contracts.creation.v1.
//
// The exchange is a NATS request-reply. The request carries the address of a
// contract and the chain it lives on. The response is a found, not found, or
// error envelope; Response validates it on both the encoding and decoding
// sides. The package declares its own constants and does not import
// indexergateway/transactionspage/v1. It checks boundary shape only. It does
// not open NATS or call a provider.
package v1
