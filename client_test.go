package raistonpay

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient()

	if c.config.BaseURL != DefaultBaseURL {
		t.Errorf("expected base URL %q, got %q", DefaultBaseURL, c.config.BaseURL)
	}
	if c.config.APIKey != "" {
		t.Errorf("expected empty API key, got %q", c.config.APIKey)
	}
	if c.Wallets == nil {
		t.Error("expected Wallets service to be initialized")
	}
	if c.Agents == nil {
		t.Error("expected Agents service to be initialized")
	}
	if c.Payments == nil {
		t.Error("expected Payments service to be initialized")
	}
	if c.Organizations == nil {
		t.Error("expected Organizations service to be initialized")
	}
	if c.Transactions == nil {
		t.Error("expected Transactions service to be initialized")
	}
	if c.Policies == nil {
		t.Error("expected Policies service to be initialized")
	}
	if c.Approvals == nil {
		t.Error("expected Approvals service to be initialized")
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	c := NewClient(
		WithAPIKey("test-key-123"),
		WithBaseURL("https://custom.api.com"),
		WithAgentID("agent-456"),
	)

	if c.config.APIKey != "test-key-123" {
		t.Errorf("expected API key %q, got %q", "test-key-123", c.config.APIKey)
	}
	if c.config.BaseURL != "https://custom.api.com" {
		t.Errorf("expected base URL %q, got %q", "https://custom.api.com", c.config.BaseURL)
	}
	if c.config.AgentID != "agent-456" {
		t.Errorf("expected agent ID %q, got %q", "agent-456", c.config.AgentID)
	}
}

func TestNewClient_WithTimeout(t *testing.T) {
	c := NewClient(WithTimeout(5 * time.Second))

	hc, ok := c.httpClient.(*http.Client)
	if !ok {
		t.Fatal("expected *http.Client")
	}
	if hc.Timeout != 5*time.Second {
		t.Errorf("expected timeout %v, got %v", 5*time.Second, hc.Timeout)
	}
}

func TestNewClient_WithCustomHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 99 * time.Second}
	c := NewClient(WithHTTPClient(custom))

	hc, ok := c.httpClient.(*http.Client)
	if !ok {
		t.Fatal("expected *http.Client")
	}
	if hc.Timeout != 99*time.Second {
		t.Errorf("expected timeout %v, got %v", 99*time.Second, hc.Timeout)
	}
}

func TestNewClient_FromEnvironment(t *testing.T) {
	t.Setenv("OMNIRAILS_API_KEY", "env-key-789")
	t.Setenv("OMNIRAILS_AGENT_ID", "env-agent-012")

	c := NewClient(FromEnvironment())

	if c.config.APIKey != "env-key-789" {
		t.Errorf("expected API key %q from env, got %q", "env-key-789", c.config.APIKey)
	}
	if c.config.AgentID != "env-agent-012" {
		t.Errorf("expected agent ID %q from env, got %q", "env-agent-012", c.config.AgentID)
	}
}

func TestNewClient_FromEnvironment_Empty(t *testing.T) {
	os.Unsetenv("OMNIRAILS_API_KEY")
	os.Unsetenv("OMNIRAILS_AGENT_ID")

	c := NewClient(FromEnvironment())

	if c.config.APIKey != "" {
		t.Errorf("expected empty API key, got %q", c.config.APIKey)
	}
	if c.config.AgentID != "" {
		t.Errorf("expected empty agent ID, got %q", c.config.AgentID)
	}
}

func TestNewClient_OptionOrdering(t *testing.T) {
	// CLI flags (later options) should override environment (earlier options).
	t.Setenv("OMNIRAILS_API_KEY", "env-key")

	c := NewClient(
		FromEnvironment(),
		WithAPIKey("cli-key"),
	)

	if c.config.APIKey != "cli-key" {
		t.Errorf("expected CLI key to override env key, got %q", c.config.APIKey)
	}
}

func TestClient_Do_Success(t *testing.T) {
	expected := Organization{
		ID:   "org-123",
		Name: "Test Org",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Authorization header 'Bearer test-key', got %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("User-Agent") != "raistonpay-go/"+Version {
			t.Errorf("unexpected User-Agent: %q", r.Header.Get("User-Agent"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	c := NewClient(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
		WithRetryPolicy(RetryPolicy{MaxRetries: 0}),
	)

	var got Organization
	err := c.do(context.Background(), http.MethodGet, "/v1/organizations/org-123", nil, &got)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, got.ID)
	}
	if got.Name != expected.Name {
		t.Errorf("expected Name %q, got %q", expected.Name, got.Name)
	}
}

func TestClient_Do_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Trace-ID", "trace-abc")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ProblemDetails{
			Type:   "https://api.raistonpay.com/errors/not-found",
			Title:  "Not Found",
			Status: 404,
			Detail: "Organization not found",
		})
	}))
	defer server.Close()

	c := NewClient(
		WithBaseURL(server.URL),
		WithRetryPolicy(RetryPolicy{MaxRetries: 0}),
	)

	var org Organization
	err := c.do(context.Background(), http.MethodGet, "/v1/organizations/missing", nil, &org)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}

	nfe, ok := err.(*NotFoundError)
	if !ok {
		t.Fatalf("expected *NotFoundError, got %T", err)
	}
	if nfe.TraceID != "trace-abc" {
		t.Errorf("expected trace ID %q, got %q", "trace-abc", nfe.TraceID)
	}
}

func TestClient_Do_SendsBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got %q", r.Header.Get("Content-Type"))
		}
		var req CreateOrganizationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if req.Name != "New Org" {
			t.Errorf("expected name %q, got %q", "New Org", req.Name)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Organization{ID: "org-new", Name: req.Name})
	}))
	defer server.Close()

	c := NewClient(
		WithBaseURL(server.URL),
		WithRetryPolicy(RetryPolicy{MaxRetries: 0}),
	)

	var org Organization
	err := c.do(context.Background(), http.MethodPost, "/v1/organizations", &CreateOrganizationRequest{Name: "New Org"}, &org)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.Name != "New Org" {
		t.Errorf("expected name %q, got %q", "New Org", org.Name)
	}
}

func TestClient_Do_AgentIDHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Agent-ID") != "agent-test" {
			t.Errorf("expected X-Agent-ID 'agent-test', got %q", r.Header.Get("X-Agent-ID"))
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(
		WithBaseURL(server.URL),
		WithAgentID("agent-test"),
		WithRetryPolicy(RetryPolicy{MaxRetries: 0}),
	)

	err := c.do(context.Background(), http.MethodDelete, "/v1/test", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Do_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(
		WithBaseURL(server.URL),
		WithRetryPolicy(RetryPolicy{MaxRetries: 0}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := c.do(ctx, http.MethodGet, "/v1/slow", nil, nil)
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}
