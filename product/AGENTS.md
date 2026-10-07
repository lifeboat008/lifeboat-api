# Lifeboat working rules

- Use Go for backend logic, the API, ledger adapter, and GitHub App.
- Keep repository boundaries in `docs/REPOSITORIES.md`; one repository owns each capability.
- GitHub events are evidence only. A human approval is required before a payout.
- Never infer GitHub ownership transfer from project inactivity.
- Use integer money units, atomic budget updates, idempotency keys, and payment reconciliation.
- Keep signing keys and credentials out of source code, tests, fixtures, and logs.
- Begin payment integration on Stellar testnet. Resolve custody and asset decisions before mainnet money.
- Keep the four repositories independently buildable from tagged dependencies before release.
