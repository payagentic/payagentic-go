package payagentic

import (
	"context"
	"fmt"
	"net/http"
)

// WalletsService handles communication with the wallet-related endpoints.
type WalletsService struct {
	client *Client
}

// List returns a paginated list of wallets.
func (s *WalletsService) List(ctx context.Context, opts *ListOptions) (*PaginatedResponse[Wallet], error) {
	path := "/v1/wallets" + opts.queryParams()
	var resp PaginatedResponse[Wallet]
	if err := s.client.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("listing wallets: %w", err)
	}
	return &resp, nil
}

// Get retrieves a wallet by ID.
func (s *WalletsService) Get(ctx context.Context, id string) (*Wallet, error) {
	var wallet Wallet
	if err := s.client.do(ctx, http.MethodGet, "/v1/wallets/"+id, nil, &wallet); err != nil {
		return nil, fmt.Errorf("getting wallet %s: %w", id, err)
	}
	return &wallet, nil
}

// Fund initiates a fund operation on a wallet.
func (s *WalletsService) Fund(ctx context.Context, id string, req *FundWalletRequest) (*Transaction, error) {
	var tx Transaction
	if err := s.client.do(ctx, http.MethodPost, "/v1/wallets/"+id+"/fund", req, &tx); err != nil {
		return nil, fmt.Errorf("funding wallet %s: %w", id, err)
	}
	return &tx, nil
}

// Withdraw initiates a withdrawal from a wallet.
func (s *WalletsService) Withdraw(ctx context.Context, id string, req *WithdrawWalletRequest) (*Transaction, error) {
	var tx Transaction
	if err := s.client.do(ctx, http.MethodPost, "/v1/wallets/"+id+"/withdraw", req, &tx); err != nil {
		return nil, fmt.Errorf("withdrawing from wallet %s: %w", id, err)
	}
	return &tx, nil
}

// Freeze freezes a wallet, preventing any outgoing transactions.
func (s *WalletsService) Freeze(ctx context.Context, id string) (*Wallet, error) {
	var wallet Wallet
	if err := s.client.do(ctx, http.MethodPost, "/v1/wallets/"+id+"/freeze", nil, &wallet); err != nil {
		return nil, fmt.Errorf("freezing wallet %s: %w", id, err)
	}
	return &wallet, nil
}

// Unfreeze unfreezes a previously frozen wallet.
func (s *WalletsService) Unfreeze(ctx context.Context, id string) (*Wallet, error) {
	var wallet Wallet
	if err := s.client.do(ctx, http.MethodPost, "/v1/wallets/"+id+"/unfreeze", nil, &wallet); err != nil {
		return nil, fmt.Errorf("unfreezing wallet %s: %w", id, err)
	}
	return &wallet, nil
}
