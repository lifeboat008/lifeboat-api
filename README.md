<p align="center"><img src="assets/logo.svg" alt="Lifeboat logo" width="112"></p>

<h1 align="center">lifeboat-api</h1>

<p align="center">The Go service that runs Lifeboat's sponsor, claim, review, and payment workflow.</p>

<p align="center">
  <a href="https://github.com/lifeboat008/lifeboat-api/actions/workflows/ci.yml"><img src="https://github.com/lifeboat008/lifeboat-api/actions/workflows/ci.yml/badge.svg" alt="Go CI"></a>
  <img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT license">
  <img src="https://img.shields.io/badge/go-1.26.3%2B-00ADD8?logo=go&logoColor=white" alt="Go 1.26.3+">
  <img src="https://img.shields.io/badge/Stellar-testnet%20only-7D00FF" alt="Stellar testnet only">
  <img src="https://img.shields.io/badge/storage-SQLite-003B57?logo=sqlite&logoColor=white" alt="SQLite">
  <img src="https://img.shields.io/badge/API-v1-informational" alt="API v1">
  <img src="https://img.shields.io/badge/status-pilot-orange" alt="Pilot status">
  <a href="https://cjay-1.gitbook.io/lifeboat-docs/"><img src="https://img.shields.io/badge/docs-GitBook-3884FF" alt="Documentation"></a>
</p>

## Contents

- [What is Lifeboat](#what-is-lifeboat)
- [What this repository does](#what-this-repository-does)
- [Workflow](#workflow)
- [Quick start](#quick-start)
- [Configuration](#configuration)
- [API at a glance](#api-at-a-glance)
- [Safety guarantees](#safety-guarantees)
- [How the four repositories fit together](#how-the-four-repositories-fit-together)
- [Project status](#project-status)
- [Open work](#open-work)
- [Documentation](#documentation)
- [Contributing and security](#contributing-and-security)
- [Maintainers](#maintainers)
- [Contributors](#contributors)
- [License](#license)

## What is Lifeboat

Lifeboat helps companies keep the open-source projects they depend on healthy. A sponsor commits a budget to a defined maintenance plan. A maintainer submits evidence of finished work as a claim. A named human reviewer approves or rejects it. Only an approved claim is paid, through a Stellar payment whose transaction hash anyone can verify.

GitHub activity is evidence, never approval. Lifeboat never pays because a project looks inactive.

## What this repository does

| Owns | Does not own |
| --- | --- |
| Versioned HTTP API, authentication, and authorization | GitHub webhook signature parsing |
| Project enrollment, sponsor plans, evidence, claims, reviews, and audit history | Low-level Stellar transaction construction |
| Transactional budget reservation and calls to `lifeboat-ledger` | |

## Workflow

```text
steward enrolls project ──► sponsor creates plan ──► github adapter records evidence
        ──► maintainer submits claim ──► reviewer approves (budget reserved)
        ──► payer triggers payment ──► Stellar receipt ──► budget settled
```

## Quick start

With Go 1.26.3 or newer:

```bash
git clone https://github.com/lifeboat008/lifeboat-api.git
cd lifeboat-api
go test ./...
go vet ./...
go run ./cmd/lifeboat-api   # needs the configuration below
```

The tagged `lifeboat-protocol` and `lifeboat-ledger` modules are public, so no module token is needed. The tests use a fake payment gateway.

An opt-in test completes the whole flow against the real Stellar testnet with temporary Friendbot-funded accounts:

```bash
LIFEBOAT_LIVE_TESTNET=1 go test ./... -v
```

## Configuration

All configuration is by environment variable. Nothing sensitive belongs in version control.

| Variable | Required | Default | Meaning |
| --- | --- | --- | --- |
| `LIFEBOAT_ACTORS_JSON` | yes | none | JSON array of actors, each with `id`, `role`, and a random `token` of at least 24 characters |
| `LIFEBOAT_TESTNET_SECRET` | yes | none | Stellar **testnet** signing key used for payments |
| `LIFEBOAT_DB` | no | `lifeboat.sqlite` | SQLite file path |
| `LIFEBOAT_LISTEN` | no | `127.0.0.1:8080` | Listen address |

Roles: `steward`, `sponsor`, `maintainer`, `reviewer`, `payer`, `github`. Tokens must be unique. Bind to loopback unless a TLS reverse proxy is in front.

## API at a glance

All `/v1` routes require `Authorization: Bearer <token>`. Money is in integer stroops (10,000,000 = 1 XLM).

| Route | Role | Purpose |
| --- | --- | --- |
| `GET /healthz` | none | Health check |
| `POST /v1/projects` | steward | Enroll a repository |
| `POST /v1/plans` | sponsor | Commit a budget and eligible work |
| `GET /v1/plans/{id}/budget` | steward, sponsor, reviewer, maintainer | Read balances |
| `POST /v1/evidence` | github | Record verified GitHub evidence |
| `POST /v1/claims` | maintainer | Request payment for one evidence item |
| `GET /v1/claims/{id}` | participants | Read claim, decision, receipt |
| `POST /v1/claims/{id}/decision` | reviewer | Approve or reject with a reason |
| `POST /v1/claims/{id}/payment` | payer | Prepare, submit, or reconcile the payout; needs an `Idempotency-Key` of 12 to 128 characters |
| `POST /v1/rescue-tasks` | steward | Open a rescue task |
| `GET /v1/audit/{id}` | steward, sponsor, reviewer, payer | Event history for a resource |

Full request and response shapes are in the [API reference](https://cjay-1.gitbook.io/lifeboat-docs/api-reference).

## Safety guarantees

- **Human approval first.** Only the reviewer named on the plan can approve, and only an approved claim can be paid.
- **Atomic budgets.** Approval reserves the amount in one database transaction. Spending cannot exceed the budget.
- **Idempotent payment.** The signed transaction is stored before submission. A retry looks up its hash first and otherwise resubmits the same envelope. A different key for the same claim returns `409`.
- **Settlement after confirmation.** Budget moves from reserved to paid only after a confirmed receipt matches the stored transaction.
- **Audit trail.** Every change writes an audit event in the same transaction.
- **Strict input.** Request bodies are limited to 1 MiB and unknown JSON fields are rejected.

## How the four repositories fit together

```text
lifeboat-protocol ──► lifeboat-ledger ──► lifeboat-api ◄── lifeboat-github
```

| Repository | Responsibility |
| --- | --- |
| [lifeboat-protocol](https://github.com/lifeboat008/lifeboat-protocol) | Domain types, validation, budget arithmetic |
| [lifeboat-ledger](https://github.com/lifeboat008/lifeboat-ledger) | Stellar testnet payment construction, submission, reconciliation |
| [lifeboat-api](https://github.com/lifeboat008/lifeboat-api) | HTTP API, persistence, authorization, audit history (this repository) |
| [lifeboat-github](https://github.com/lifeboat008/lifeboat-github) | GitHub webhook verification and evidence submission |

## Project status

Lifeboat is a **Stellar testnet pilot**. This API records sponsor commitments but does not receive deposits or verify treasury funding. Tokens are global per role, not scoped to a project, so run one project and one sponsor per instance. Do not use it for real funds until custody, payout asset, and legal operations are defined. See [Project status](https://cjay-1.gitbook.io/lifeboat-docs/project-status) and the [Mainnet gate](https://cjay-1.gitbook.io/lifeboat-docs/security-and-governance/mainnet-gate).

## Open work

- [Enforce project-scoped actor authorization](https://github.com/lifeboat008/lifeboat-api/issues/1) (complexity: high)
- [Verify testnet treasury coverage before reserving sponsor budgets](https://github.com/lifeboat008/lifeboat-api/issues/2) (complexity: high)
- [Recover a prepared payment whose time bound has expired](https://github.com/lifeboat008/lifeboat-api/issues/4)

The shared [product requirements](product/docs/PRD.md), architecture, validation record, [Wave plan](product/docs/WAVE.md), and vector brand files are versioned in `product/`.

## Documentation

Full documentation, including the API reference, role guides, security notes, and operations runbooks, is at **[cjay-1.gitbook.io/lifeboat-docs](https://cjay-1.gitbook.io/lifeboat-docs/)**.

## Contributing and security

- Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request. Run `gofmt`, `go vet ./...`, and `go test ./...`, and add tests for changed behavior.
- Use testnet and synthetic data only. Never commit keys, tokens, or real sponsor or maintainer details.
- Preserve human approval before any payout.
- Report vulnerabilities privately as described in [SECURITY.md](SECURITY.md). Do not post exploit details in a public issue.

## Maintainers

| Maintainer | Contact |
| --- | --- |
| [lifeboat008](https://github.com/lifeboat008) | [Open an issue](https://github.com/lifeboat008/lifeboat-api/issues) for public work; use SECURITY.md for private reports |

## Contributors

<a href="https://github.com/lifeboat008/lifeboat-api/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=lifeboat008/lifeboat-api" alt="Contributors">
</a>

## License

[MIT](LICENSE). This pilot has not had a formal security audit.
