<p align="center"><img src="assets/logo.svg" alt="Lifeboat logo" width="112"></p>

# lifeboat-api

Go service that runs Lifeboat's sponsor, claim, review, and payment workflow.

## Owns

- Versioned HTTP API, authentication, and authorization.
- Project enrollment, sponsor plans, evidence, claims, reviews, and audit history.
- Transactional budget reservation and calls to `lifeboat-ledger`.

## Does not own

GitHub webhook signature parsing or low-level Stellar transaction construction.

## Run the testnet pilot

Set `LIFEBOAT_TESTNET_SECRET` to a funded Stellar testnet signing key and set `LIFEBOAT_ACTORS_JSON` to an array of actors with `id`, `role`, and random `token` fields. Roles are `steward`, `sponsor`, `maintainer`, `reviewer`, `payer`, and `github`. Each token must be at least 24 characters. Set `LIFEBOAT_DB` to a SQLite path and `LIFEBOAT_LISTEN` to the listen address; defaults are `lifeboat.sqlite` and `127.0.0.1:8080`. Then run `go run ./cmd/lifeboat-api`.

All `/v1` routes require `Authorization: Bearer <token>`. The workflow is `POST /v1/projects`, `POST /v1/plans`, `POST /v1/evidence`, `POST /v1/claims`, `POST /v1/claims/{id}/decision`, then `POST /v1/claims/{id}/payment`. The payment route requires an `Idempotency-Key` header of 12–128 characters. `GET /v1/claims/{id}` returns the claim, decision, and transaction receipt when available; `GET /v1/plans/{id}/budget` returns integer stroop balances. `GET /v1/audit/{id}` returns resource event history. `POST /v1/rescue-tasks` records a steward-authorized rescue task.

An approved claim reserves budget atomically. The API persists a signed testnet transaction before submission. A retry first looks up its hash and otherwise submits the same envelope. It only settles the reserved budget after a confirmed matching receipt. GitHub evidence alone never approves a claim. This pilot records sponsor commitments; it does not receive sponsor deposits or verify treasury funding. Do not use it for real funds until custody, payout asset, and legal operations are defined.

Go 1.26 and access to tagged private `lifeboat-protocol` and `lifeboat-ledger` modules are required to build from source.

The shared [product requirements](product/docs/PRD.md), architecture, validation record, and vector brand files are versioned in `product/`. The parent `lifeboat` folder also keeps a local workspace copy.

## Contributing and security

See [CONTRIBUTING.md](CONTRIBUTING.md) for pull requests, [SECURITY.md](SECURITY.md) for private vulnerability reports, and [LICENSE](LICENSE) for MIT terms.
