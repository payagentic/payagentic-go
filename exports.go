package payagentic

// Re-exports of common generated response types from the OpenAPI transport.
//
// The hand-written domain types in types.go (Wallet, Agent, Payment, Policy,
// Transaction, Approval, Organization) are kept as the ergonomic public API.
// These aliases surface the raw `*Response` shapes returned by the generated
// transport (Client.OpenAPI.*WithResponse) so callers can type-check against
// them without reaching into `internal/openapi`.
//
// Use them when working directly with the generated transport, e.g.:
//
//	resp, err := client.OpenAPI.ListWalletsWithResponse(ctx, nil)
//	if err != nil { return err }
//	var w payagentic.WalletResponse = (*resp.JSON200.Items)[0]

import "github.com/payagentic/payagentic-go/internal/openapi"

type (
	// WalletResponse is the gateway response shape for a single wallet.
	WalletResponse = openapi.WalletResponse
	// AgentResponse is the gateway response shape for a single agent.
	AgentResponse = openapi.AgentResponse
	// PaymentResponse is the gateway response shape for a payment intent.
	// Also aliased as PaymentIntent in x402.go for x402-flow ergonomics.
	PaymentResponse = openapi.PaymentResponse
	// TransactionResponse is the gateway response shape for a single transaction.
	TransactionResponse = openapi.TransactionResponse
	// PolicyResponse is the gateway response shape for a single policy.
	PolicyResponse = openapi.PolicyResponse
	// OrganizationResponse is the gateway response shape for an organization.
	OrganizationResponse = openapi.OrganizationResponse
	// PayoutResponse is the gateway response shape for a merchant payout.
	PayoutResponse = openapi.PayoutResponse
	// RefundResponse is the gateway response shape for a refund.
	RefundResponse = openapi.RefundResponse
	// ApiKeyResponse is the gateway response shape for an API key.
	ApiKeyResponse = openapi.ApiKeyResponse
	// CreateApiKeyResponse is the one-time-secret payload returned on key creation.
	CreateApiKeyResponse = openapi.CreateApiKeyResponse
	// BalanceResponse is the gateway response shape for a wallet balance query.
	BalanceResponse = openapi.BalanceResponse
	// MeResponse is the gateway response shape for the /v1/me identity probe.
	MeResponse = openapi.MeResponse
	// HealthResponse is the gateway response shape for /healthz.
	HealthResponse = openapi.HealthResponse
	// StatsResponse is the gateway response shape for the merchant /stats surface.
	StatsResponse = openapi.StatsResponse
	// DashboardStats is the dashboard summary payload.
	DashboardStats = openapi.DashboardStats
)
