# cafe-contracts

Shared **wire contracts** for the CAFE stack: versioned structs, constants, and JSON-oriented validation helpers used across repositories. This module is a **library only**—it does not run as a service.

## What belongs here

- Envelope and payload types for cross-service messages (e.g. Discovery → CPM observations).
- Exported string/enum constants for wire fields (algorithm IDs, posture labels, account kinds).
- Minimal validation that checks shape and required fields at the contract boundary.
- Canonical JSON fixtures and tests that lock serialization formats.

## What does *not* belong here

| Concern | Where it lives |
|--------|----------------|
| Policy graphs, templates, instances, ranking, assessment | [`cafe-cpm`](https://github.com/create2-labs/cafe-cpm) (Crypto Policy Management) |
| Chain indexing, wallet discovery, persistence, producing observations | [`cafe-discovery`](https://github.com/create2-labs/cafe-discovery) |
| Remediation orchestration, operator workflows | Remediation services (separate repos) |
| NATS subscriptions, connection lifecycle, retries | Application code in each service; this repo may define **payload** types only |

**Rule of thumb:** if it encodes *business policy* or *runtime wiring*, it is not a `cafe-contracts` concern. If it defines *what bytes travel on the wire* so two teams can compile against the same types, it belongs here.

## Layout

- `eventenvelope/v01/` — shared event header contract (`event_id`, `event_type`, `event_version`, `occurred_at`, `correlation_id`, `causation_id`, `producer`) with minimal validation and canonical JSON fixture(s).
- `observation/wallet/v01/` — normative `cafe.discovery.wallet.observed` wire contract (`event_version` **v0.1**): `Event`, `Subject`, `Payload`, exported vocabulary (account kind, algorithm ID, PQ posture, subject type), `Validate()`, and canonical JSON under `testdata/`.
- `discoverywalletobserved/v01/` — legacy path kept temporarily for coexistence during CWR0 migration (`event_type` historical synonym: `discovery.wallet.observed`).
- `cafenatsv01/` — policy and remediation **NATS/JSON** contract bundle (`event_version` **v0.1**): `policy.assessment.requested` (explicit CPM command with embedded observation snapshot + selection wire), outbound CPM events (validation, activation, assessment, remediation request), Remediation service events, versioned `NATSSubject*` constants, and `MAPPING.md` (model-to-wire reference). No brokers or runtime logic.
- `validation/` — tiny, reusable helpers (non-empty strings, field-scoped errors) for contract packages.

Import example:

```go
import eventenvelopev01 "github.com/create2-labs/cafe-contracts/eventenvelope/v01"
import walletobsv01 "github.com/create2-labs/cafe-contracts/observation/wallet/v01"
import "github.com/create2-labs/cafe-contracts/discoverywalletobserved/v01"
import "github.com/create2-labs/cafe-contracts/cafenatsv01"
```

Version directories use a short semver-like segment (`v01` = 0.1) to keep import paths stable and readable.

## Consumers

Add to `go.mod`:

```go
require github.com/create2-labs/cafe-contracts v0.0.0
```

Use tagged releases once published; during integration, `replace` in a workspace is fine.

## Development

```bash
go test ./...
go vet ./...
```

CI runs tests on pull requests to `main`. Release automation uses [release-please](https://github.com/googleapis/release-please) (Go release type) to propose version bumps and changelogs from conventional commits.
