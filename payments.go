package raistonpay

import (
	"context"
	"fmt"
	"net/http"
)

// PaymentsService handles communication with the payment-related endpoints.
type PaymentsService struct {
	client *Client
}

// Propose submits a new x402 payment proposal for policy evaluation and execution.
func (s *PaymentsService) Propose(ctx context.Context, req *ProposePaymentRequest) (*Payment, error) {
	var payment Payment
	if err := s.client.do(ctx, http.MethodPost, "/v1/payments", req, &payment); err != nil {
		return nil, fmt.Errorf("proposing payment: %w", err)
	}
	return &payment, nil
}

// GetStatus retrieves the current status of a payment by ID.
func (s *PaymentsService) GetStatus(ctx context.Context, id string) (*Payment, error) {
	var payment Payment
	if err := s.client.do(ctx, http.MethodGet, "/v1/payments/"+id, nil, &payment); err != nil {
		return nil, fmt.Errorf("getting payment %s: %w", id, err)
	}
	return &payment, nil
}

// List returns a paginated list of payments.
func (s *PaymentsService) List(ctx context.Context, opts *ListOptions) (*PaginatedResponse[Payment], error) {
	path := "/v1/payments" + opts.queryParams()
	var resp PaginatedResponse[Payment]
	if err := s.client.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("listing payments: %w", err)
	}
	return &resp, nil
}
