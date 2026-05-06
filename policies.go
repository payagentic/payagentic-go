package raistonpay

import (
	"context"
	"fmt"
	"net/http"
)

// PoliciesService handles communication with the policy-related endpoints.
type PoliciesService struct {
	client *Client
}

// List returns a paginated list of policies.
func (s *PoliciesService) List(ctx context.Context, opts *ListOptions) (*PaginatedResponse[Policy], error) {
	path := "/v1/policies" + opts.queryParams()
	var resp PaginatedResponse[Policy]
	if err := s.client.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("listing policies: %w", err)
	}
	return &resp, nil
}

// Get retrieves a policy by ID.
func (s *PoliciesService) Get(ctx context.Context, id string) (*Policy, error) {
	var policy Policy
	if err := s.client.do(ctx, http.MethodGet, "/v1/policies/"+id, nil, &policy); err != nil {
		return nil, fmt.Errorf("getting policy %s: %w", id, err)
	}
	return &policy, nil
}

// Create creates a new spend policy.
func (s *PoliciesService) Create(ctx context.Context, req *CreatePolicyRequest) (*Policy, error) {
	var policy Policy
	if err := s.client.do(ctx, http.MethodPost, "/v1/policies", req, &policy); err != nil {
		return nil, fmt.Errorf("creating policy: %w", err)
	}
	return &policy, nil
}

// Update updates an existing policy.
func (s *PoliciesService) Update(ctx context.Context, id string, req *UpdatePolicyRequest) (*Policy, error) {
	var policy Policy
	if err := s.client.do(ctx, http.MethodPatch, "/v1/policies/"+id, req, &policy); err != nil {
		return nil, fmt.Errorf("updating policy %s: %w", id, err)
	}
	return &policy, nil
}
