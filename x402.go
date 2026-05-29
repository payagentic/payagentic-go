package payagentic

import (
	"context"
	"fmt"
	"net/http"
)

// X402Client handles x402 protocol payment flows.
// The x402 protocol enables agent-to-API payments using HTTP 402 responses.
type X402Client struct {
	client *Client
}

// X402PaymentRequest describes a payment triggered by an HTTP 402 response.
type X402PaymentRequest struct {
	// URL is the API endpoint that returned HTTP 402.
	URL string `json:"url"`
	// WalletID is the wallet to pay from.
	WalletID string `json:"wallet_id"`
	// MaxAmount is the maximum amount the caller is willing to pay.
	MaxAmount string `json:"max_amount,omitempty"`
}

// X402PaymentResponse contains the result of an x402 payment.
type X402PaymentResponse struct {
	// PaymentID is the PayAgentic payment identifier.
	PaymentID string `json:"payment_id"`
	// Receipt is the signed payment receipt to present to the API.
	Receipt string `json:"receipt"`
	// Amount is the amount that was paid.
	Amount string `json:"amount"`
	// Currency is the currency of the payment.
	Currency string `json:"currency"`
}

// NewX402Client creates a new X402Client from an existing Client.
func NewX402Client(c *Client) *X402Client {
	return &X402Client{client: c}
}

// Pay executes an x402 payment flow: evaluates the 402 payment requirement,
// checks policy, signs, and returns a receipt for the caller to present.
func (x *X402Client) Pay(ctx context.Context, req *X402PaymentRequest) (*X402PaymentResponse, error) {
	var resp X402PaymentResponse
	if err := x.client.do(ctx, http.MethodPost, "/v1/x402/pay", req, &resp); err != nil {
		return nil, fmt.Errorf("x402 payment for %s: %w", req.URL, err)
	}
	return &resp, nil
}

// Verify checks the validity of a payment receipt.
func (x *X402Client) Verify(ctx context.Context, receipt string) (bool, error) {
	body := map[string]string{"receipt": receipt}
	var resp struct {
		Valid bool `json:"valid"`
	}
	if err := x.client.do(ctx, http.MethodPost, "/v1/x402/verify", body, &resp); err != nil {
		return false, fmt.Errorf("verifying x402 receipt: %w", err)
	}
	return resp.Valid, nil
}
