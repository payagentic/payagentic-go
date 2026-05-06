package raistonpay

import (
	"context"
	"fmt"
	"net/http"
)

// AgentsService handles communication with the agent-related endpoints.
type AgentsService struct {
	client *Client
}

// List returns a paginated list of agents.
func (s *AgentsService) List(ctx context.Context, opts *ListOptions) (*PaginatedResponse[Agent], error) {
	path := "/v1/agents" + opts.queryParams()
	var resp PaginatedResponse[Agent]
	if err := s.client.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("listing agents: %w", err)
	}
	return &resp, nil
}

// Get retrieves an agent by ID.
func (s *AgentsService) Get(ctx context.Context, id string) (*Agent, error) {
	var agent Agent
	if err := s.client.do(ctx, http.MethodGet, "/v1/agents/"+id, nil, &agent); err != nil {
		return nil, fmt.Errorf("getting agent %s: %w", id, err)
	}
	return &agent, nil
}

// Create creates a new agent.
func (s *AgentsService) Create(ctx context.Context, req *CreateAgentRequest) (*Agent, error) {
	var agent Agent
	if err := s.client.do(ctx, http.MethodPost, "/v1/agents", req, &agent); err != nil {
		return nil, fmt.Errorf("creating agent: %w", err)
	}
	return &agent, nil
}

// RotateKeyResponse contains the new API key material after a key rotation.
type RotateKeyResponse struct {
	AgentID        string `json:"agent_id"`
	KeyFingerprint string `json:"key_fingerprint"`
	APIKey         string `json:"api_key"`
}

// RotateKey rotates the API key for an agent. The new key is returned only once.
func (s *AgentsService) RotateKey(ctx context.Context, id string) (*RotateKeyResponse, error) {
	var resp RotateKeyResponse
	if err := s.client.do(ctx, http.MethodPost, "/v1/agents/"+id+"/rotate-key", nil, &resp); err != nil {
		return nil, fmt.Errorf("rotating key for agent %s: %w", id, err)
	}
	return &resp, nil
}

// Delete deletes (revokes) an agent.
func (s *AgentsService) Delete(ctx context.Context, id string) error {
	if err := s.client.do(ctx, http.MethodDelete, "/v1/agents/"+id, nil, nil); err != nil {
		return fmt.Errorf("deleting agent %s: %w", id, err)
	}
	return nil
}
