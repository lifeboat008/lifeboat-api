package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	ledger "github.com/lifeboat008/lifeboat-ledger"
	protocol "github.com/lifeboat008/lifeboat-protocol"
)

type fakeGateway struct {
	submissions int
	confirmed   bool
}

func (g *fakeGateway) Prepare(_ context.Context, claim protocol.Claim) (ledger.Prepared, error) {
	return ledger.Prepared{ClaimID: claim.ID, TransactionHash: "hash-one", EnvelopeXDR: "signed-xdr", AmountStroops: claim.AmountStroops, Destination: claim.DestinationAccount, Network: "testnet"}, nil
}
func (g *fakeGateway) Submit(_ context.Context, _ ledger.Prepared) (ledger.Receipt, error) {
	g.submissions++
	g.confirmed = true
	return ledger.Receipt{TransactionHash: "hash-one", Network: "testnet", Confirmed: true}, nil
}
func (g *fakeGateway) Lookup(_ context.Context, _ string) (ledger.Receipt, error) {
	if !g.confirmed {
		return ledger.Receipt{}, ledger.ErrNotFound
	}
	return ledger.Receipt{TransactionHash: "hash-one", Network: "testnet", Confirmed: true}, nil
}

func TestClaimApprovalPaymentAndRetry(t *testing.T) {
	gateway := &fakeGateway{}
	service, _ := ledger.NewService(gateway)
	actors := []Actor{
		{ID: "steward", Role: "steward", Token: "steward-token-1234567890123"},
		{ID: "sponsor", Role: "sponsor", Token: "sponsor-token-1234567890123"},
		{ID: "worker", Role: "maintainer", Token: "worker-token-12345678901234"},
		{ID: "reviewer", Role: "reviewer", Token: "reviewer-token-123456789012"},
		{ID: "payer", Role: "payer", Token: "payer-token-123456789012345"},
		{ID: "github", Role: "github", Token: "github-token-12345678901234"},
	}
	s, err := NewServer(filepath.Join(t.TempDir(), "test.sqlite"), actors, service)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	h := s.Handler()
	call := func(method, path, token, key string, body any) int {
		t.Helper()
		var data []byte
		if body != nil {
			data, _ = json.Marshal(body)
		}
		req := httptest.NewRequest(method, path, bytes.NewReader(data))
		req.Header.Set("Authorization", "Bearer "+token)
		if key != "" {
			req.Header.Set("Idempotency-Key", key)
		}
		response := httptest.NewRecorder()
		h.ServeHTTP(response, req)
		return response.Code
	}
	project := protocol.Project{ID: "p1", Repository: "owner/repo", InstallationID: 123, StewardID: "steward", State: protocol.ProjectActive}
	if got := call("POST", "/v1/projects", actors[0].Token, "", project); got != 201 {
		t.Fatalf("project: %d", got)
	}
	plan := protocol.Plan{ID: "plan1", ProjectID: "p1", SponsorID: "sponsor", ReviewerID: "reviewer", Asset: "XLM", BudgetStroops: 10_000_000, EligibleWork: []string{"merged_pr"}, PeriodEndsAt: time.Now().Add(24 * time.Hour)}
	if got := call("POST", "/v1/plans", actors[1].Token, "", plan); got != 201 {
		t.Fatalf("plan: %d", got)
	}
	evidence := protocol.Evidence{ID: "ev1", ProjectID: "p1", InstallationID: 123, DeliveryID: "delivery1", Kind: "merged_pr", URL: "https://github.com/owner/repo/pull/1", ObservedAt: time.Now()}
	if got := call("POST", "/v1/evidence", actors[5].Token, "", evidence); got != 201 {
		t.Fatalf("evidence: %d", got)
	}
	if got := call("POST", "/v1/evidence", actors[5].Token, "", evidence); got != 409 {
		t.Fatalf("duplicate evidence: %d", got)
	}
	claim := protocol.Claim{ID: "c1", ProjectID: "p1", PlanID: "plan1", MaintainerID: "worker", EvidenceID: "ev1", WorkType: "merged_pr", Summary: "Fixed dependency update", AmountStroops: 5_000_000, DestinationAccount: "GTEST", State: protocol.ClaimSubmitted}
	if got := call("POST", "/v1/claims", actors[2].Token, "", claim); got != 201 {
		t.Fatalf("claim: %d", got)
	}
	decision := protocol.Decision{ClaimID: "c1", ReviewerID: "reviewer", Approved: true, Reason: "Reviewed work"}
	if got := call("POST", "/v1/claims/c1/decision", actors[3].Token, "", decision); got != 200 {
		t.Fatalf("decision: %d", got)
	}
	if got := call("POST", "/v1/claims/c1/payment", actors[4].Token, "some-unique-key-123", nil); got != 200 {
		t.Fatalf("payment: %d", got)
	}
	if got := call("POST", "/v1/claims/c1/payment", actors[4].Token, "some-unique-key-123", nil); got != 200 {
		t.Fatalf("payment retry: %d", got)
	}
	if got := call("POST", "/v1/claims/c1/payment", actors[4].Token, "another-unique-key", nil); got != 409 {
		t.Fatalf("different key: %d", got)
	}
	if gateway.submissions != 1 {
		t.Fatalf("submissions = %d", gateway.submissions)
	}
	var budget protocol.Budget
	if err := s.db.QueryRow(`SELECT total,reserved,paid FROM plans WHERE id='plan1'`).Scan(&budget.TotalStroops, &budget.ReservedStroops, &budget.PaidStroops); err != nil {
		t.Fatal(err)
	}
	if budget.ReservedStroops != 0 || budget.PaidStroops != 5_000_000 {
		t.Fatalf("budget = %+v", budget)
	}
}
