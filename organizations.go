package raistonpay

import (
	"context"
	"fmt"
	"net/http"
)

// OrganizationsService handles communication with the organization-related endpoints.
type OrganizationsService struct {
	client *Client
}

// List returns a paginated list of organizations.
func (s *OrganizationsService) List(ctx context.Context, opts *ListOptions) (*PaginatedResponse[Organization], error) {
	path := "/v1/organizations" + opts.queryParams()
	var resp PaginatedResponse[Organization]
	if err := s.client.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("listing organizations: %w", err)
	}
	return &resp, nil
}

// Get retrieves an organization by ID.
func (s *OrganizationsService) Get(ctx context.Context, id string) (*Organization, error) {
	var org Organization
	if err := s.client.do(ctx, http.MethodGet, "/v1/organizations/"+id, nil, &org); err != nil {
		return nil, fmt.Errorf("getting organization %s: %w", id, err)
	}
	return &org, nil
}

// Create creates a new organization.
func (s *OrganizationsService) Create(ctx context.Context, req *CreateOrganizationRequest) (*Organization, error) {
	var org Organization
	if err := s.client.do(ctx, http.MethodPost, "/v1/organizations", req, &org); err != nil {
		return nil, fmt.Errorf("creating organization: %w", err)
	}
	return &org, nil
}

// Update updates an existing organization.
func (s *OrganizationsService) Update(ctx context.Context, id string, req *UpdateOrganizationRequest) (*Organization, error) {
	var org Organization
	if err := s.client.do(ctx, http.MethodPatch, "/v1/organizations/"+id, req, &org); err != nil {
		return nil, fmt.Errorf("updating organization %s: %w", id, err)
	}
	return &org, nil
}

// Delete deletes an organization.
func (s *OrganizationsService) Delete(ctx context.Context, id string) error {
	if err := s.client.do(ctx, http.MethodDelete, "/v1/organizations/"+id, nil, nil); err != nil {
		return fmt.Errorf("deleting organization %s: %w", id, err)
	}
	return nil
}
