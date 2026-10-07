# Lifeboat operations baseline

## Services

Run `lifeboat-api` and `lifeboat-github` as separate Go services. The API owns the data store and the payment coordinator; the GitHub adapter receives webhooks and calls the API. `lifeboat-protocol` and `lifeboat-ledger` are Go libraries.

## Configuration

Keep the GitHub webhook secret, service tokens, and Stellar testnet signing key outside Git. The pilot uses SQLite and environment variables documented in each service README. Use a managed secret store before production. Bind the API and adapter to loopback unless a TLS reverse proxy is configured.

## Monitoring

Track webhook signature failures, duplicate deliveries, evidence ingestion lag, claims awaiting review, payment intents awaiting reconciliation, failed payments, and remaining sponsor balances. Alerts must avoid including private payout details.

## Recovery

Back up the SQLite database and test restoration. Pending payments are reconciled when the same payment request is retried with its original idempotency key; there is no background reconciliation worker yet. If a submission is uncertain, inspect the stored transaction hash on testnet and retry with that key. Do not create a new payment for the claim solely because an HTTP request timed out.

## Incident response

Pause new payouts if signing credentials may be compromised or if budget records disagree with ledger receipts. Preserve audit logs, rotate affected credentials, reconcile every in-flight payment, and communicate status to affected sponsors and maintainers.
