package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	ledger "github.com/lifeboat008/lifeboat-ledger"
	protocol "github.com/lifeboat008/lifeboat-protocol"
	"github.com/stellar/go-stellar-sdk/keypair"
)

// Opt-in integration check against Friendbot and Stellar testnet.
func TestLiveAPIToTestnetPayment(t *testing.T) {
	if os.Getenv("LIFEBOAT_LIVE_TESTNET") != "1" {
		t.Skip("set LIFEBOAT_LIVE_TESTNET=1 for live testnet integration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	source, err := keypair.Random()
	if err != nil {
		t.Fatal(err)
	}
	destination, err := keypair.Random()
	if err != nil {
		t.Fatal(err)
	}
	httpClient := &http.Client{Timeout: 20 * time.Second}
	for _, address := range []string{source.Address(), destination.Address()} {
		request, err := http.NewRequestWithContext(ctx, "GET", "https://friendbot.stellar.org/?addr="+url.QueryEscape(address), nil)
		if err != nil {
			t.Fatal(err)
		}
		response, err := httpClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
		if response.StatusCode != 200 {
			t.Fatalf("testnet faucet returned %d", response.StatusCode)
		}
	}
	gateway, err := ledger.NewTestnetGateway(source.Seed(), nil)
	if err != nil {
		t.Fatal(err)
	}
	service, err := ledger.NewService(gateway)
	if err != nil {
		t.Fatal(err)
	}
	actors := []Actor{
		{ID: "steward", Role: "steward", Token: "steward-test-token-123456789"},
		{ID: "sponsor", Role: "sponsor", Token: "sponsor-test-token-123456789"},
		{ID: "worker", Role: "maintainer", Token: "worker-test-token-1234567890"},
		{ID: "reviewer", Role: "reviewer", Token: "reviewer-test-token-12345678"},
		{ID: "payer", Role: "payer", Token: "payer-test-token-12345678901"},
		{ID: "github", Role: "github", Token: "github-test-token-123456789"},
	}
	server, err := NewServer(filepath.Join(t.TempDir(), "live.sqlite"), actors, service)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	handler := server.Handler()
	call := func(path, token, key string, body any) *httptest.ResponseRecorder {
		t.Helper()
		data, _ := json.Marshal(body)
		request := httptest.NewRequest("POST", path, bytes.NewReader(data))
		request.Header.Set("Authorization", "Bearer "+token)
		if key != "" {
			request.Header.Set("Idempotency-Key", key)
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}
	assertStatus := func(response *httptest.ResponseRecorder, want int) {
		t.Helper()
		if response.Code != want {
			t.Fatalf("status %d, want %d: %s", response.Code, want, response.Body.String())
		}
	}
	assertStatus(call("/v1/projects", actors[0].Token, "", protocol.Project{ID: "live-project", Repository: "example/repository", InstallationID: 11, StewardID: "steward", State: protocol.ProjectActive}), 201)
	assertStatus(call("/v1/plans", actors[1].Token, "", protocol.Plan{ID: "live-plan", ProjectID: "live-project", SponsorID: "sponsor", ReviewerID: "reviewer", Asset: "XLM", BudgetStroops: 2, EligibleWork: []string{"merged_pr"}, PeriodEndsAt: time.Now().Add(time.Hour)}), 201)
	assertStatus(call("/v1/evidence", actors[5].Token, "", protocol.Evidence{ID: "live-evidence", ProjectID: "live-project", InstallationID: 11, DeliveryID: "live-delivery", Kind: "merged_pr", URL: "https://github.com/example/repository/pull/1", ObservedAt: time.Now()}), 201)
	assertStatus(call("/v1/claims", actors[2].Token, "", protocol.Claim{ID: "live-claim", ProjectID: "live-project", PlanID: "live-plan", MaintainerID: "worker", EvidenceID: "live-evidence", WorkType: "merged_pr", Summary: "Testnet integration", AmountStroops: 1, DestinationAccount: destination.Address(), State: protocol.ClaimSubmitted}), 201)
	assertStatus(call("/v1/claims/live-claim/decision", actors[3].Token, "", protocol.Decision{ClaimID: "live-claim", ReviewerID: "reviewer", Approved: true, Reason: "Reviewed test fixture"}), 200)
	var payment *httptest.ResponseRecorder
	for attempt := 0; attempt < 6; attempt++ {
		payment = call("/v1/claims/live-claim/payment", actors[4].Token, "live-payment-key-123", nil)
		if payment.Code == 200 {
			break
		}
		if payment.Code != 202 {
			t.Fatalf("payment status %d: %s", payment.Code, payment.Body.String())
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(time.Second):
		}
	}
	assertStatus(payment, 200)
	var receipt ledger.Receipt
	if err := json.Unmarshal(payment.Body.Bytes(), &receipt); err != nil || !receipt.Confirmed || receipt.TransactionHash == "" {
		t.Fatalf("invalid confirmed receipt: %v", err)
	}
	retry := call("/v1/claims/live-claim/payment", actors[4].Token, "live-payment-key-123", nil)
	assertStatus(retry, 200)
	if retry.Body.String() != payment.Body.String() {
		t.Fatal("payment retry returned a different receipt")
	}
	t.Logf("API confirmed Stellar testnet transaction: %s", receipt.TransactionHash)
}
