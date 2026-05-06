package raistonpay

import (
	"context"
	"fmt"
	"net/http"
)

// TransactionsService handles communication with the transaction-related endpoints.
type TransactionsService struct {
	client *Client
}

// List returns a paginated list of transactions.
func (s *TransactionsService) List(ctx context.Context, opts *ListOptions) (*PaginatedResponse[Transaction], error) {
	path := "/v1/transactions" + opts.queryParams()
	var resp PaginatedResponse[Transaction]
	if err := s.client.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("listing transactions: %w", err)
	}
	return &resp, nil
}

// Get retrieves a transaction by ID.
func (s *TransactionsService) Get(ctx context.Context, id string) (*Transaction, error) {
	var tx Transaction
	if err := s.client.do(ctx, http.MethodGet, "/v1/transactions/"+id, nil, &tx); err != nil {
		return nil, fmt.Errorf("getting transaction %s: %w", id, err)
	}
	return &tx, nil
}
