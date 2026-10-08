# Lifeboat Wave readiness

Checked 8 October 2026. This is a status record, not a claim of Drips approval. Recheck live services before submitting.

## Fit and scope

Lifeboat is a Go tool and API for sponsor-funded open-source maintenance. A named human reviewer approves a claim, then the ledger adapter sends a Stellar payment and records its transaction hash. Stellar is the payment and independently verifiable receipt layer. There is no Soroban contract, browser app, token, escrow, or indexer in this MVP. The generic contract architecture, contract-agent prompt, frontend-agent prompt, contract deployment, and hosted frontend sections of the supplied builder playbook do not apply. Adding them solely for a Wave application would misrepresent the product.

The [Stellar Wave approved-repos page](https://www.drips.network/wave/stellar/repos) reported 824 approved repos when checked. Its visible examples include escrow, crowdfunding, DeFi, marketplaces, and payment tooling. This is not proof that Lifeboat is unique across all repos. Lifeboat's narrower claim is the combination of consent-based project enrollment, sponsor work plans, human claim review, and a verifiable Stellar payout receipt. The [SDF grants page](https://stellar.org/grants-and-funding) supports projects that grow Stellar and Soroban; [SCF 7.0](https://stellar.org/blog/ecosystem/introducing-scf-v7) describes an Open Track judged on impact, originality, and technical soundness and a separate RFP track for requested tooling. SCF funding and Drips Wave approval are different programs.

## Phase check against the builder playbook

| Phase | Status for Lifeboat | Evidence or remaining work |
| --- | --- | --- |
| 1. Ecosystem reconnaissance | Partly complete | Live Wave list and SDF funding sources checked; no exhaustive 824-repo categorization or formal overlap study. |
| 2–3. Idea generation and critical review | Complete for the selected concept; historical alternatives not documented | [PRD](PRD.md), [decisions](DECISIONS.md), and [why Stellar](WHY_STELLAR.md) explain the problem, scope, and rejection of unnecessary lockups. A consenting sponsor and repository still need to validate demand. |
| 4. Naming and repo structure | Complete | Four named Go repositories with distinct responsibilities in [REPOSITORIES.md](REPOSITORIES.md). A two-repo Rust/TypeScript split would not fit this architecture. |
| 5–7. Contract and app system prompts | Not applicable | No Soroban contract or frontend is planned for the pilot. Go APIs and dependencies are described in [ARCHITECTURE.md](ARCHITECTURE.md) and [INTERFACES.md](INTERFACES.md). |
| 8. Local build and deployment | Local validation complete; hosted deployment pending | [VALIDATION.md](VALIDATION.md) records tests and two real testnet transactions. No managed pilot service or public endpoint has been deployed. |
| 9. Hosting topology | Designed, not deployed | See [OPERATIONS.md](OPERATIONS.md). The operator must choose hosting, private data storage, and secret management for a consenting pilot. |
| 10. Repo hygiene | Core checks complete | MIT, CONTRIBUTING, SECURITY, logo, CI, public releases, topics, and six scoped issues exist. All four `main` branches require the actual `test` check and one PR approval for non-admin changes. Recheck settings live before submission. Do not add `Stellar Wave` labels before Drips approval. |
| 11. Documentation site | Markdown documentation complete; public docs site absent | The versioned product docs live in `lifeboat-api/product/docs`. A separate GitBook or hosted docs site has not been published. |
| 12. Submission | Not submitted | See [SUBMISSION.md](SUBMISSION.md). A pilot demo video and any live service URL do not yet exist. There are no contract IDs because there is no contract. |
| 13. Post-approval iteration | Not applicable yet | Requires Wave acceptance and contributor applications. |

## Mainnet boundary

The repository is testnet-only and single-operator. It has not completed the consenting-repository/sponsor pilot or a formal security audit. Before real funds, resolve custody, payout asset, jurisdiction, treasury coverage, project-scoped authorization, reconciliation operations, and incident response. Open issues cover several of these gaps; an open issue is not an implemented control.
