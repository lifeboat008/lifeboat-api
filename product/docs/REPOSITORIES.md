# Lifeboat repository map

```text
lifeboat-protocol ──► lifeboat-ledger ──► lifeboat-api ◄── lifeboat-github
          └──────────────────────────────► lifeboat-api
```

The arrows mean the repository on the right uses the package or service on the left. The GitHub adapter calls the API; it does not import the API server implementation into its own process.

| Repository | Owns | Does not own | Depends on |
| --- | --- | --- | --- |
| `lifeboat-protocol` | Project, budget, claim, approval, and rescue state models; invariants | Stellar RPC, HTTP, GitHub webhooks | None |
| `lifeboat-ledger` | Stellar payment construction, signing boundary, submission, reconciliation, transaction receipts | Claim approval or GitHub evidence | `lifeboat-protocol`, Stellar Go SDK |
| `lifeboat-api` | Authentication, persistence, sponsor/project workflows, review decisions, API contract | Raw GitHub webhook parsing, XDR payment details | `lifeboat-protocol`, `lifeboat-ledger` |
| `lifeboat-github` | GitHub App auth, signature checks, event deduplication, evidence translation | Funding decisions, payouts, domain state | `lifeboat-api` HTTP contract |

## Release order

1. Tag `lifeboat-protocol` v0.1.0.
2. Pin it in `lifeboat-ledger`, test, then tag v0.1.0.
3. Pin protocol and ledger tags in `lifeboat-api`, then tag v0.1.0.
4. Pin the API contract version in `lifeboat-github`, then tag v0.1.0.

Use a local parent `go.work` while developing. Each published module must build from its own tagged dependencies.

## Repository names and remotes

- `https://github.com/lifeboat008/lifeboat-protocol.git`
- `https://github.com/lifeboat008/lifeboat-ledger.git`
- `https://github.com/lifeboat008/lifeboat-api.git`
- `https://github.com/lifeboat008/lifeboat-github.git`
