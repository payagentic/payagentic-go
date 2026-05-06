package raistonpay

import (
	"encoding/json"
	"testing"
	"time"
)

func TestOrganization_JSON_Roundtrip(t *testing.T) {
	org := Organization{
		ID:        "org-abc",
		Name:      "Test Organization",
		Status:    "active",
		CreatedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(org)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got Organization
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got.ID != org.ID {
		t.Errorf("ID: expected %q, got %q", org.ID, got.ID)
	}
	if got.Name != org.Name {
		t.Errorf("Name: expected %q, got %q", org.Name, got.Name)
	}
}

func TestWallet_JSON_Roundtrip(t *testing.T) {
	wallet := Wallet{
		ID:             "wal-123",
		OrganizationID: "org-abc",
		AgentID:        "agt-456",
		Address:        "0x1234567890abcdef",
		Chain:          "base",
		Balance:        "1000.50",
		Status:         WalletStatusActive,
		CreatedAt:      time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:      time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(wallet)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got Wallet
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got.ID != wallet.ID {
		t.Errorf("ID: expected %q, got %q", wallet.ID, got.ID)
	}
	if got.Balance != wallet.Balance {
		t.Errorf("Balance: expected %q, got %q", wallet.Balance, got.Balance)
	}
	if got.Status != WalletStatusActive {
		t.Errorf("Status: expected %q, got %q", WalletStatusActive, got.Status)
	}
	if got.AgentID != wallet.AgentID {
		t.Errorf("AgentID: expected %q, got %q", wallet.AgentID, got.AgentID)
	}
}

func TestTransaction_JSON_Roundtrip(t *testing.T) {
	tx := Transaction{
		ID:       "tx-789",
		WalletID: "wal-123",
		Type:     TransactionTypePayment,
		Status:   TransactionStatusConfirmed,
		Amount:   "50.00",
		Currency: "USDC",
		TxHash:   "0xdeadbeef",
		Chain:    "base",
		Metadata: map[string]string{"invoice": "INV-001"},
	}

	data, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got Transaction
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got.Type != TransactionTypePayment {
		t.Errorf("Type: expected %q, got %q", TransactionTypePayment, got.Type)
	}
	if got.Status != TransactionStatusConfirmed {
		t.Errorf("Status: expected %q, got %q", TransactionStatusConfirmed, got.Status)
	}
	if got.Metadata["invoice"] != "INV-001" {
		t.Errorf("Metadata[invoice]: expected %q, got %q", "INV-001", got.Metadata["invoice"])
	}
}

func TestPayment_JSON_Roundtrip(t *testing.T) {
	payment := Payment{
		ID:          "pay-001",
		WalletID:    "wal-123",
		AgentID:     "agt-456",
		Status:      PaymentStatusSettled,
		Amount:      "25.00",
		Currency:    "USDC",
		Recipient:   "0xrecipient",
		Description: "API call payment",
	}

	data, err := json.Marshal(payment)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got Payment
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got.Status != PaymentStatusSettled {
		t.Errorf("Status: expected %q, got %q", PaymentStatusSettled, got.Status)
	}
	if got.Description != payment.Description {
		t.Errorf("Description: expected %q, got %q", payment.Description, got.Description)
	}
}

func TestPolicy_JSON_Roundtrip(t *testing.T) {
	policy := Policy{
		ID:             "pol-001",
		OrganizationID: "org-abc",
		Name:           "Daily Spend Cap",
		Type:           PolicyTypeSpendCap,
		Rules: PolicyRules{
			MaxAmount:  "100.00",
			DailyLimit: "500.00",
			AllowedChains: []string{"base", "ethereum"},
		},
	}

	data, err := json.Marshal(policy)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got Policy
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got.Type != PolicyTypeSpendCap {
		t.Errorf("Type: expected %q, got %q", PolicyTypeSpendCap, got.Type)
	}
	if got.Rules.DailyLimit != "500.00" {
		t.Errorf("DailyLimit: expected %q, got %q", "500.00", got.Rules.DailyLimit)
	}
	if len(got.Rules.AllowedChains) != 2 {
		t.Errorf("AllowedChains: expected 2 items, got %d", len(got.Rules.AllowedChains))
	}
}

func TestPaginatedResponse_JSON(t *testing.T) {
	input := `{
		"items": [{"id": "wal-1"}, {"id": "wal-2"}],
		"next_cursor": "cursor-abc",
		"total_count": 42
	}`

	var resp PaginatedResponse[Wallet]
	if err := json.Unmarshal([]byte(input), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(resp.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Items))
	}
	if resp.NextCursor != "cursor-abc" {
		t.Errorf("expected cursor %q, got %q", "cursor-abc", resp.NextCursor)
	}
	if resp.TotalCount != 42 {
		t.Errorf("expected total count 42, got %d", resp.TotalCount)
	}
}

func TestProblemDetails_JSON(t *testing.T) {
	input := `{
		"type": "https://api.raistonpay.com/errors/not-found",
		"title": "Not Found",
		"status": 404,
		"detail": "Wallet wal-xxx does not exist",
		"instance": "/v1/wallets/wal-xxx"
	}`

	var pd ProblemDetails
	if err := json.Unmarshal([]byte(input), &pd); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if pd.Status != 404 {
		t.Errorf("Status: expected 404, got %d", pd.Status)
	}
	if pd.Title != "Not Found" {
		t.Errorf("Title: expected %q, got %q", "Not Found", pd.Title)
	}
}

func TestStatusEnums(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"TransactionStatusPending", string(TransactionStatusPending), "pending"},
		{"TransactionStatusConfirmed", string(TransactionStatusConfirmed), "confirmed"},
		{"TransactionStatusFailed", string(TransactionStatusFailed), "failed"},
		{"TransactionStatusCancelled", string(TransactionStatusCancelled), "cancelled"},
		{"TransactionTypeTransfer", string(TransactionTypeTransfer), "transfer"},
		{"TransactionTypeFund", string(TransactionTypeFund), "fund"},
		{"TransactionTypeWithdraw", string(TransactionTypeWithdraw), "withdraw"},
		{"TransactionTypePayment", string(TransactionTypePayment), "payment"},
		{"PaymentStatusProposed", string(PaymentStatusProposed), "proposed"},
		{"PaymentStatusSettled", string(PaymentStatusSettled), "settled"},
		{"PaymentStatusFailed", string(PaymentStatusFailed), "failed"},
		{"WalletStatusActive", string(WalletStatusActive), "active"},
		{"WalletStatusFrozen", string(WalletStatusFrozen), "frozen"},
		{"WalletStatusClosed", string(WalletStatusClosed), "closed"},
		{"PolicyTypeSpendCap", string(PolicyTypeSpendCap), "spend_cap"},
		{"PolicyTypeAllowlist", string(PolicyTypeAllowlist), "allowlist"},
		{"PolicyTypeVelocity", string(PolicyTypeVelocity), "velocity"},
		{"ApprovalStatusPending", string(ApprovalStatusPending), "pending"},
		{"ApprovalStatusApproved", string(ApprovalStatusApproved), "approved"},
		{"ApprovalStatusDenied", string(ApprovalStatusDenied), "denied"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.value)
			}
		})
	}
}

func TestCreateOrganizationRequest_JSON(t *testing.T) {
	req := CreateOrganizationRequest{Name: "My Org"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	expected := `{"name":"My Org"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

func TestFundWalletRequest_JSON(t *testing.T) {
	req := FundWalletRequest{
		Amount:   "100.00",
		Currency: "USDC",
		Source:   "payshap://ref123",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got FundWalletRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if got.Source != req.Source {
		t.Errorf("Source: expected %q, got %q", req.Source, got.Source)
	}
}

func TestWallet_OmitsEmptyAgentID(t *testing.T) {
	wallet := Wallet{
		ID:      "wal-123",
		Address: "0xabc",
		Chain:   "base",
		Balance: "0",
		Status:  WalletStatusActive,
	}

	data, err := json.Marshal(wallet)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal to map failed: %v", err)
	}

	if _, ok := raw["agent_id"]; ok {
		t.Error("expected agent_id to be omitted when empty")
	}
}
