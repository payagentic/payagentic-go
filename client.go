package payagentic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/payagentic/payagentic-go/internal/middleware"
)

const (
	// DefaultBaseURL is the default PayAgentic API base URL.
	DefaultBaseURL = "https://api.payagentic.com"
	// DefaultTimeout is the default HTTP request timeout.
	DefaultTimeout = 30 * time.Second
)

// Option configures the Client.
type Option func(*Client)

// WithAPIKey sets the API key for authentication.
func WithAPIKey(key string) Option {
	return func(c *Client) {
		c.config.APIKey = key
	}
}

// WithBaseURL sets the API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.config.BaseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithRetryPolicy sets the retry policy for failed requests.
func WithRetryPolicy(p RetryPolicy) Option {
	return func(c *Client) {
		c.retryPolicy = p
	}
}

// WithTimeout sets the HTTP client timeout. This only works when the underlying
// HTTP client is a *http.Client (the default). Custom httpDoer implementations
// should manage their own timeouts.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if hc, ok := c.httpClient.(*http.Client); ok {
			hc.Timeout = d
		}
	}
}

// WithAgentID sets the agent ID header on requests.
func WithAgentID(id string) Option {
	return func(c *Client) {
		c.config.AgentID = id
	}
}

// FromEnvironment configures the client from environment variables.
// It reads PAYAGENTIC_API_KEY and PAYAGENTIC_AGENT_ID.
func FromEnvironment() Option {
	return func(c *Client) {
		if key := os.Getenv("PAYAGENTIC_API_KEY"); key != "" {
			c.config.APIKey = key
		}
		if agentID := os.Getenv("PAYAGENTIC_AGENT_ID"); agentID != "" {
			c.config.AgentID = agentID
		}
	}
}

// NewClient creates a new PayAgentic API client configured with the given options.
func NewClient(opts ...Option) *Client {
	c := &Client{
		config: Config{
			BaseURL: DefaultBaseURL,
		},
		httpClient:  &http.Client{Timeout: DefaultTimeout},
		retryPolicy: DefaultRetryPolicy(),
	}
	for _, opt := range opts {
		opt(c)
	}

	// Initialize service handles.
	c.Organizations = &OrganizationsService{client: c}
	c.Wallets = &WalletsService{client: c}
	c.Agents = &AgentsService{client: c}
	c.Payments = &PaymentsService{client: c}
	c.Transactions = &TransactionsService{client: c}
	c.Policies = &PoliciesService{client: c}
	c.Approvals = &ApprovalsService{client: c}

	return c
}

// do executes an HTTP request with retries and JSON marshalling.
func (c *Client) do(ctx context.Context, method, path string, body any, result any) error {
	fn := func() (*http.Response, error) {
		var reqBody io.Reader
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("marshalling request body: %w", err)
			}
			reqBody = bytes.NewReader(b)
		}

		url := c.config.BaseURL + path
		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("User-Agent", "payagentic-go/"+Version)
		req.Header.Set("Accept", "application/json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if c.config.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
		}
		if c.config.AgentID != "" {
			req.Header.Set("X-Agent-ID", c.config.AgentID)
		}

		return c.httpClient.Do(req)
	}

	resp, err := middleware.WithRetry(ctx, c.retryPolicy, fn)
	if err != nil {
		return fmt.Errorf("executing request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return middleware.ParseErrorResponse(resp.StatusCode, respBody, resp.Header.Get("X-Trace-ID"))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshalling response: %w", err)
		}
	}
	return nil
}
