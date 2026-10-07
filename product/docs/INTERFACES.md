# Lifeboat interface contract

The testnet pilot uses JSON over HTTP. Authentication is `Authorization: Bearer <actor-token>`; each token maps to an actor ID and one role. Monetary values are integer stroops (10,000,000 stroops = 1 XLM).

| Method and route | Purpose | Role |
| --- | --- | --- |
| `POST /v1/projects` | Enroll a repository and GitHub App installation | Steward |
| `POST /v1/plans` | Commit a maintenance budget and define eligible work | Sponsor |
| `GET /v1/plans/{id}/budget` | Read total, reserved, paid, and remaining stroops | Participant |
| `POST /v1/evidence` | Record verified GitHub evidence | GitHub adapter |
| `POST /v1/claims` | Request payment for one evidence item | Maintainer |
| `POST /v1/claims/{id}/decision` | Approve or reject with reason | Named reviewer |
| `POST /v1/claims/{id}/payment` | Prepare, submit, or reconcile testnet payout | Payer |
| `GET /v1/claims/{id}` | Read claim, decision, and receipt if confirmed | Participant |
| `POST /v1/rescue-tasks` | Open a scoped rescue task | Named steward |
| `GET /v1/audit/{id}` | Read the event history for a resource ID | Steward, sponsor, reviewer, or payer |

The payment route requires `Idempotency-Key` of 12–128 characters. Use the same key on retry. A different key for the same claim receives `409`. A repeated GitHub delivery also receives `409` from the API; the GitHub adapter converts that to a successful duplicate acknowledgement. Request bodies are limited to 1 MiB and reject unknown fields.

Only a named reviewer may approve a claim. Approval atomically reserves the requested amount. A claim's work type must match its evidence kind, and an evidence item can support only one claim. A confirmed, matching Stellar testnet receipt converts reserved budget to paid budget. The API stores the signed envelope and hash before submission so a retry can reconcile or resubmit the same transaction. Uncertain submissions return `202` for later retry.

The first version is a single-operator pilot. Actor tokens are supplied at process start; project-specific access control for reading claims and budgets and treasury funding verification are required before a multi-tenant or real-money launch.
