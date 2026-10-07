# Lifeboat validation record

Date: 7 October 2026.

- All four Go repositories pass local tests and vet checks and their remote GitHub Actions Go checks.
- The API test completes project enrollment, plan creation, evidence ingestion, claim submission, authorized approval, payment, payment retry, and budget settlement with a fake gateway. It also rejects duplicate evidence, a second payment key, over-budget approval, and payout of an unapproved claim.
- The GitHub adapter test rejects an unsigned event, records a signed merged pull request as evidence, and acknowledges a duplicate delivery.
- The ledger opt-in smoke test created temporary Friendbot-funded accounts and confirmed a real Stellar testnet payment, then reconciled its transaction hash: `4ae0c23178eab223d6e1cf131f481af26f741dbe30555bd3bb66a002a6489c93`.
- The API opt-in integration test completed enrollment, plan, evidence, claim, human approval, payment, receipt, and idempotent retry through the real Stellar testnet gateway. Its confirmed transaction hash is `12529c7f40b11c70ffc38f9de77a96914c5fb8d7736bff83664cff7a32ed4223`.

The live integration uses synthetic actors and evidence. A consenting sponsor and real GitHub installation remain a pilot task. The SQLite/API configuration is single-operator and testnet-only; real-money use needs custody, funding verification, tenant access control, and operational review.
