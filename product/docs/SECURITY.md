# Lifeboat security and governance requirements

Lifeboat moves money and interprets repository activity. Authorization, payment safety, and clear human governance are core product requirements.

## Required controls

- Verify GitHub webhook HMAC signatures and reject unsigned or replayed deliveries.
- Use least-privilege GitHub App permissions and rotate App credentials.
- Authenticate sponsors, maintainers, reviewers, and service-to-service API calls separately.
- Require a reviewer decision before payment; the payer cannot silently raise a claim amount.
- Enforce budget limits with database transactions and idempotency keys.
- Reconcile Stellar submission results before any retry.
- Keep private keys in a managed secret store or external signer, never in GitHub repositories or logs.
- Record who changed a plan, approved a claim, changed a payout address, or opened a rescue task.
- Require fresh confirmation for payout-address changes on an approved claim.

## Governance rules

- A repository remains under its owners' control. Lifeboat can propose or fund work, not transfer GitHub access.
- Rescue status requires a steward decision and documented reason, with a path for the existing maintainer to respond.
- A merged pull request is evidence, not automatic proof that the funded work was completed.
- Sponsors and maintainers should see the approval rules before committing to a plan.

## Threats to test

| Threat | Control |
| --- | --- |
| Spoofed GitHub event | Signature verification and installation mapping |
| Duplicate event or payment retry | Delivery deduplication, payment idempotency, reconciliation |
| Malicious reviewer pays self | Role separation, audit trail, optional second approval above threshold |
| Payout-address substitution | Address-change confirmation and immutable approved snapshot |
| Budget race condition | Atomic reservation and balance checks |
| False abandonment claim | Human steward decision and dispute path |

## Reporting a vulnerability

Use private vulnerability reporting once it is enabled in each repository. Never post key material, private customer details, or live exploit steps in a public issue.
