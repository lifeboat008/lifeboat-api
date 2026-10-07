# Lifeboat test plan

## Protocol tests

Verify claim states, budget reservations, approval authority, rescue transitions, and forbidden transitions. Test amounts as integers and reject zero, negative, overflow, and over-budget claims.

## Ledger tests

Use Stellar testnet and a fake submitter. Verify one approved claim produces one payment, repeated requests are idempotent, uncertain submissions reconcile before retry, and failed transactions do not mark a claim paid.

## API tests

Test role-based access, duplicate idempotency keys, concurrent claims against one budget, approval audit records, payout-address changes, and error responses. Verify that `lifeboat-github` cannot access payer-only routes.

## GitHub adapter tests

Use signed webhook fixtures. Reject invalid signatures, wrong installations, duplicate delivery IDs, unsupported event types, and oversized payloads. A valid merged pull request should create only an evidence candidate.

## End-to-end pilot

With one consenting repository and one sponsor, complete plan → work → evidence → claim → review → Stellar testnet payment → confirmed receipt. Repeat webhook delivery and payment request to prove they do not duplicate outcomes. Test a disputed claim and a rescue task without changing GitHub ownership.

Mainnet use is gated on payment custody, asset, jurisdiction, incident response, and operational review.
