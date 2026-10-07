# Lifeboat product requirements

Status: planning baseline, 7 October 2026

## Problem

Companies rely on open-source libraries but maintenance often depends on a few volunteers. Funding is irregular, work expectations are unclear, and a project can become stuck when its maintainer loses time or interest. Sponsors need a way to fund specific, verifiable maintenance outcomes and keep support available when a project needs help.

## Product promise

Lifeboat turns sponsor commitments into approved maintenance payments, with a clear record of the work and a Stellar payment receipt. It also offers an orderly route for new maintainers to propose rescue work when a project needs it.

## Primary users

- Companies that depend on an open-source repository.
- Maintainers who perform routine upgrades, releases, bug fixes, and security work.
- Reviewers or project stewards who approve claims and coordinate succession.
- Developers who want to take on an approved rescue task.

## Jobs to be done

1. Sponsor a defined maintenance plan for a repository.
2. Submit evidence for completed maintenance work.
3. Review the work and approve or reject a claim with a recorded reason.
4. Pay an approved claim through Stellar and link the transaction receipt.
5. Organize a rescue task when maintainers request help or a steward confirms a need.

## MVP scope

- Register one GitHub repository and its consenting maintainer or steward.
- Define sponsor commitment, eligible work types, budget, payout currency, reviewer, and period.
- Receive GitHub events as evidence candidates: merged pull requests, releases, and issue activity. Events never approve a claim by themselves.
- Let a maintainer submit a claim tied to evidence and a payout address.
- Require an authorized reviewer to approve the claim before payment.
- Create and record a Stellar payment with an idempotency key and transaction hash.
- Show the amount funded, approved, paid, and remaining for each project.
- Permit a steward to open a rescue task and assign an approved reviewer.

The first payment rail should use Stellar testnet. Mainnet asset selection, custody model, and financial operating requirements must be resolved before real sponsor money is handled.

## Explicit boundaries

- No automatic takeover of GitHub repository ownership or permissions.
- No automatic payout based only on inactivity, issue count, or a merged commit.
- No lockup or Soroban contract in the MVP; a contract is only justified by a concrete need that the payment flow cannot meet.
- No promise that sponsorship guarantees perpetual maintenance.
- No funding of work outside an agreed plan without an explicit sponsor or steward approval.

## User stories and acceptance criteria

| Story | Acceptance criteria |
| --- | --- |
| Sponsor a project | A sponsor can define a budget and work plan; the system records who can approve claims. |
| Submit maintenance work | A maintainer links a PR or release and requests an amount within the remaining budget. |
| Review a claim | An authorized reviewer sees evidence, decision history, and budget; a decision records reviewer and reason. |
| Pay approved work | A payment request is idempotent; the recorded Stellar hash can be checked independently. |
| Start rescue work | A steward records why help is needed, consent/authority, task scope, reviewer, and available budget. |
| Reject invalid activity | Unsigned GitHub webhooks, duplicate deliveries, unapproved claims, and overspending are refused. |

## Success measures

- A pilot repository completes a sponsor → claim → approval → testnet payout cycle.
- Every payout maps to one approved claim and one independently verifiable Stellar transaction.
- Duplicate GitHub events or retries do not create duplicate claims or payments.
- Maintainers and sponsors can understand the remaining budget and who approves work.
- A rescue task can be created without assuming repository ownership has changed.

## Open questions before mainnet

- Custody: sponsor-controlled wallet, regulated payment partner, or a legally reviewed treasury model.
- Payout asset and supported countries.
- Who arbitrates disputed claims and how appeals work.
- How a project steward is established when the original maintainer is unavailable.
- Whether public fund balances expose commercial or personal information that should stay private.
