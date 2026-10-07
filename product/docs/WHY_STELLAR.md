# Why Lifeboat uses Stellar

Open-source maintainers and sponsors may live in different countries. Lifeboat needs a payment record that sponsors, maintainers, and reviewers can independently verify. Stellar provides a shared payment rail and a transaction receipt for each approved claim.

GitHub remains the source of repository activity. The Go API remains the source of funding decisions and approval history. Stellar records the actual payout. That separation keeps the first version practical while giving every paid claim a verifiable payment result.

Lifeboat does not need a custom token or Soroban lockup to prove its first use case. Testnet validates the workflow; mainnet follows only after custody, asset, and operating decisions are resolved.

## Demonstration

1. A sponsor defines a maintenance plan for one consenting open-source repository.
2. A maintainer completes a scoped pull request and submits a claim.
3. The GitHub App provides the event as evidence, and a reviewer approves the claim.
4. Lifeboat pays on Stellar testnet and displays the transaction hash beside the approved claim.
5. Replaying the webhook and payment request produces no second payout.
