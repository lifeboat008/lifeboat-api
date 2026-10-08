# Lifeboat Stellar Wave submission packet

This packet is draft material for the organization owner. Do not claim a live pilot or submit the repositories until the missing evidence below is available.

## Project description

Lifeboat helps companies fund maintenance of open-source software they depend on. A sponsor defines eligible work and a budget for a consenting repository. A maintainer supplies evidence, a named human reviewer approves or rejects the claim, and Lifeboat sends an approved testnet payment on Stellar. The transaction hash lets the sponsor and maintainer independently verify the payout. GitHub activity is evidence only; it never authorizes payment on its own.

## Why Stellar

Sponsors and maintainers may be in different countries. Stellar provides a shared payment rail and a public receipt for each approved claim. The first version uses native XLM on testnet. It does not deploy a Soroban contract; adding a contract without a demonstrated need would add custody and audit work without improving this pilot.

## Repository relationship

- [lifeboat-protocol](https://github.com/lifeboat008/lifeboat-protocol) defines plans, claims, decisions, budgets, and invariants.
- [lifeboat-ledger](https://github.com/lifeboat008/lifeboat-ledger) consumes approved payment requests and handles Stellar signing, submission, lookup, and receipts. It imports `lifeboat-protocol`.
- [lifeboat-api](https://github.com/lifeboat008/lifeboat-api) persists the sponsor-to-payout workflow and calls both modules. It is the source of the versioned product docs.
- [lifeboat-github](https://github.com/lifeboat008/lifeboat-github) verifies GitHub webhooks and sends evidence candidates to the API. It cannot approve claims or pay them.

## Planned contributor work

Six scoped issues are open. Protocol: [disputes and reviewer conflicts](https://github.com/lifeboat008/lifeboat-protocol/issues/1). Ledger: [background reconciliation](https://github.com/lifeboat008/lifeboat-ledger/issues/1). API: [project-scoped authorization](https://github.com/lifeboat008/lifeboat-api/issues/1) and [testnet treasury checks](https://github.com/lifeboat008/lifeboat-api/issues/2). GitHub adapter: [installation suspension/removal](https://github.com/lifeboat008/lifeboat-github/issues/1) and [auditable enrollment](https://github.com/lifeboat008/lifeboat-github/issues/2). Complexity suggestions are in the issue bodies. Add issues to Wave only after the corresponding repository is approved; the owner will leave Wave labels to the program workflow.

## Evidence and missing links

| Item | Current evidence | Submission status |
| --- | --- | --- |
| Source and setup | Four public repositories, MIT licenses, READMEs, CI, and [product docs](README.md) | Available |
| Live Stellar proof | [VALIDATION.md](VALIDATION.md) records two confirmed testnet payment hashes | Available; synthetic actors and evidence |
| Live hosted product URL | None | Missing; do not invent one |
| End-to-end demo video | None | Missing; record the real pilot flow or clearly label a synthetic testnet demo |
| Contract verification links | None | Not applicable; no Soroban contract |
| Consenting repository and sponsor | Not yet recruited or run through the product | Missing |
| Drips Wave application and approval | Not recorded | Owner action and organizer decision |

## Owner submission sequence

1. Check the [live approved-repos list](https://www.drips.network/wave/stellar/repos) and search for Lifeboat before applying. The list can change.
2. Run a consenting-repository and sponsor pilot, or present a clearly labeled synthetic demo without claiming real use. Record a short video showing plan, evidence, human approval, testnet payment, and receipt.
3. Sign in to [Drips Wave](https://www.drips.network/wave/stellar), install the GitHub App on `lifeboat008`, sync the four repos, and apply each relevant repo to the Stellar Wave Program.
4. Wait for organizer approval. Only then add suitable issues to the program using the Drips dashboard or its GitHub-label workflow.
