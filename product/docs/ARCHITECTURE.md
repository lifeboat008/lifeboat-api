# Lifeboat architecture

## Workflow

```text
GitHub repository events ──► verified evidence candidates
                                     │
Sponsor plan ──► project budget ──► claim ──► human approval
                                                │
                                                ▼
                                       Stellar payment request
                                                │
                                                ▼
                                      transaction and receipt
```

## Domain model

- `Project`: GitHub repository, steward, lifecycle state, and policy.
- `SponsorPlan`: sponsor, committed amount, period, eligible work, and reviewer.
- `Budget`: committed, reserved, paid, and remaining amounts. All changes are auditable.
- `Evidence`: GitHub event or submitted link; it supports review but is not proof by itself.
- `Claim`: work summary, evidence, requested amount, payout destination, and status.
- `Approval`: reviewer, decision, reason, timestamp, and policy version.
- `Payment`: idempotency key, Stellar account, asset, amount, submission state, and transaction hash.
- `RescueTask`: reason, consent/authority, scope, budget, reviewer, and candidate assignee.

## State changes

A claim moves through `draft → submitted → approved/rejected → payment_pending → paid/failed`. Only an authorized reviewer may approve. A paid claim cannot be paid again. A project can move to `needs_help` or `rescue_open` only through a steward decision with an auditable reason; inactivity signals are advisory.

## Service boundaries

`lifeboat-protocol` defines these types and invariants. `lifeboat-ledger` converts an approved payment request into a Stellar transaction and reconciles its result. `lifeboat-api` owns persisted workflows and authorization. `lifeboat-github` verifies incoming webhooks and submits normalized evidence candidates to the API.

## Stellar payment design

- Start on Stellar testnet.
- Require an approved claim, exact asset, amount, and destination before payment construction.
- Use a stable idempotency key derived from the claim/payment ID; retries must not create duplicate payouts.
- Store the transaction hash and ledger result, then reconcile uncertain submissions before retrying.
- Keep signing keys out of source code, logs, and GitHub webhook handling.
- Choose the mainnet asset and custody approach only after an operating and legal review.

## Data and API design

Use a transactional store for budgets, claims, approvals, payment intents, and webhook delivery IDs. Publish a versioned HTTP API under `/v1`. The GitHub App should submit evidence candidates through an authenticated API call, not write directly to the database. Changes to money state need an append-only audit record.

## Failure handling

Duplicate webhooks are ignored by delivery ID. Lost network responses after payment submission trigger reconciliation by idempotency key and transaction hash. A failed payout leaves the approved claim in a recoverable state. No service treats a GitHub event as payment authorization.
