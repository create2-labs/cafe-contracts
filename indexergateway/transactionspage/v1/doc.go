// Package v1 is the wire contract for one page of wallet transaction history
// on cafe.indexer.gateway.transactions.page.v1.
//
// cafe-indexer-gateway is the subscriber, in the queue group
// cafe-indexer-gateway. One replica handles each request. It subscribes so
// that a scanner can ask for one page — address, chain, opaque cursor, limit,
// and deadline — without holding a Moralis or Etherscan client or secret.
// The gateway is the only caller of those providers. It answers on the
// request's reply inbox with that page (hash, from, next cursor, oldest
// transaction first) or with an error envelope. It does not run the key
// search.
//
// The publisher subscribes only to that reply inbox, to read the page or the
// error and then continue its own work. Today that publisher is
// cafe-scanner-wallet, from searchHistory, and cmd/cli/publickey, which
// reaches the same bus directly. A later scanner that needs the same history,
// such as the smart-contract scanner or the wallet-value scanner, uses the
// same request and the same inbox. The wallet scanner keeps the key search,
// the 30-second budget, and public_key_recovery.
//
// A response is either the page or an error envelope, never both.
// This package checks boundary shape only. It does not open NATS, call a
// provider, or choose one.
//
// The package name is v1, as required by ADR_20261007_onchain_indexer_service.
// It is not part of cafenatsv01.
package v1
