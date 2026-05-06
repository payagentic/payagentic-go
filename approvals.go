package raistonpay

import (
	"context"
	"fmt"
	"net/http"
)

// ApprovalsService handles communication with the approval-related endpoints.
type ApprovalsService struct {
	client *Client
}

// List returns a paginated list of pending approvals.
func (s *ApprovalsService) List(ctx context.Context, opts *ListOptions) (*PaginatedResponse[Approval], error) {
	path := "/v1/approvals" + opts.queryParams()
	var resp PaginatedResponse[Approval]
	if err := s.client.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("listing approvals: %w", err)
	}
	return &resp, nil
}

// Approve approves a pending approval request.
func (s *ApprovalsService) Approve(ctx context.Context, id string, req *ApprovalDecisionRequest) (*Approval, error) {
	var approval Approval
	if err := s.client.do(ctx, http.MethodPost, "/v1/approvals/"+id+"/approve", req, &approval); err != nil {
		return nil, fmt.Errorf("approving request %s: %w", id, err)
	}
	return &approval, nil
}

// Deny denies a pending approval request.
func (s *ApprovalsService) Deny(ctx context.Context, id string, req *ApprovalDecisionRequest) (*Approval, error) {
	var approval Approval
	if err := s.client.do(ctx, http.MethodPost, "/v1/approvals/"+id+"/deny", req, &approval); err != nil {
		return nil, fmt.Errorf("denying request %s: %w", id, err)
	}
	return &approval, nil
}
