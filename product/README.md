# Lifeboat

![Lifeboat vector mark](brand/lifeboat.svg)

Lifeboat helps companies keep the open-source projects they depend on healthy. Sponsors commit money for defined maintenance work. Maintainers submit evidence, reviewers approve work, and accepted claims are paid through Stellar. If a project needs a new maintainer, the community can organize rescue work; funds move only after an approved claim.

Lifeboat combines Go services, a GitHub webhook adapter, and Stellar testnet payments. It does not take control of a repository or automatically pay someone merely because a project looks inactive.

## Start here

1. [Product requirements](docs/PRD.md)
2. [Repository boundaries and build order](docs/REPOSITORIES.md)
3. [Technical architecture](docs/ARCHITECTURE.md)
4. [Security and governance](docs/SECURITY.md)
5. [Delivery plan](docs/BUILD_PLAN.md)
6. [Decision log](docs/DECISIONS.md)
7. [Interfaces](docs/INTERFACES.md), [test plan](docs/TEST_PLAN.md), and [operations](docs/OPERATIONS.md)
8. [Logo assets and usage](brand/README.md)
9. [Why Stellar and MVP demo](docs/WHY_STELLAR.md)
10. [Validation record](docs/VALIDATION.md)

The four Git repositories in this folder are `lifeboat-protocol`, `lifeboat-ledger`, `lifeboat-api`, and `lifeboat-github`. The documents in `docs/` apply to the whole product; each repository also has a focused README.

## Product status

The four Go repositories now contain the domain protocol, signed testnet payment adapter, transactional API, and signed GitHub webhook adapter. Tagged private modules and remote CI are in place. The ledger and full API flow have each confirmed a real Stellar testnet payment. A consenting repository and sponsor pilot remains to be run. Real-money use needs a custody model, funded treasury checks, jurisdiction review, and stronger multi-tenant authorization.
