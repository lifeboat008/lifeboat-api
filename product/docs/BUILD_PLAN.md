# Lifeboat build plan

## Phase 0 — money and authority decisions

- Confirm who sponsors, who holds funds, who approves claims, and who signs a payout.
- Define testnet-only pilot terms and a sample maintenance plan.
- Create a versioned API contract and sanitized GitHub webhook fixtures.

**Done when:** a reviewer can explain each money movement and a project steward can explain the rescue decision process.

## Phase 1 — protocol and ledger

- Implement project, budget, claim, approval, and rescue transitions in `lifeboat-protocol`.
- Implement testnet payment construction, submission, idempotency, and reconciliation in `lifeboat-ledger`.
- Test duplicate payments, failed network responses, invalid destinations, and budget boundaries.

**Done when:** one approved claim can produce one testnet payment and a verifiable receipt, even after retries.

## Phase 2 — API

- Implement sponsor plans, project enrollment, evidence, claims, reviewer decisions, and audit history.
- Add authentication, authorization, transactional budget reservations, and a versioned API.

**Done when:** a pilot can complete the plan → claim → approval → payout flow through the API.

## Phase 3 — GitHub App and pilot

- Verify webhooks and installation identity, deduplicate deliveries, and submit evidence candidates.
- Build a small pilot with one consenting open-source repository and one sponsor.
- Review payout, dispute, and rescue workflows before any mainnet funds are accepted.

**Done when:** GitHub activity appears as reviewable evidence and cannot trigger payment on its own.

## Definition of done for every repository

Code, meaningful tests, README usage example, changelog entry, CI checks, and tagged release. Each downstream repository pins a released upstream version.
