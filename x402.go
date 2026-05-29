package payagentic

// x402 payment-flow walker.
//
// Handles the HTTP 402 Payment Required retry loop: when a resource returns
// 402, the walker proposes a payment via the PayAgentic gateway, waits for
// settlement, then retries the original request with a payment receipt.
//
// Gateway-bound calls (`POST /v1/payments/propose`, `GET /v1/payments/{id}`)
// delegate to the generated *openapi.ClientWithResponses. The two
// mandate endpoints used by the agent kit (`/v1/mandates/issue` and
// `/v1/mandates/{jti}/revoke`) are owned by the identity service and NOT
// in the gateway spec; the MandatesService in mandate.go keeps its
// direct-HTTP implementation until those endpoints are surfaced.

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/payagentic/payagentic-go/internal/openapi"
)

// X402PaymentRequest describes a payment triggered by an HTTP 402 response.
//
// Retained as a thin public input type so callers can stage propose
// calls without importing the generated package.
type X402PaymentRequest struct {
	// URL is the API endpoint that returned HTTP 402.
	URL string `json:"url"`
	// PaymentRequirements is the value of the upstream `X-PAYMENT-REQUIREMENTS`
	// (or `X-Payment-Requirements`) header from the 402 response.
	PaymentRequirements string `json:"payment_requirements"`
	// MaxAmount is the maximum amount the caller is willing to pay (optional cap).
	MaxAmount string `json:"max_amount,omitempty"`
}

// Public re-exports of the generated x402 transport types so callers don't
// have to reach into internal/openapi.
type (
	// ProposePaymentBody is the request body for POST /v1/payments/propose.
	ProposePaymentBody = openapi.ProposePaymentJSONRequestBody
	// PaymentIntent is the gateway's PaymentResponse — returned by both
	// propose and status-poll endpoints.
	PaymentIntent = openapi.PaymentResponse
)

// Tunables for the 402 fetch loop.
const (
	x402MaxPollAttempts   = 30
	x402PollInterval      = 200 * time.Millisecond
	x402PaymentStatusOK   = "settled"
	x402PaymentStatusFail = "failed"
)

// X402Client is the x402 paywall walker. Construct via NewX402Client(*Client)
// — it shares the parent Client's authenticated transport.
type X402Client struct {
	client *Client
}

// NewX402Client creates a new X402Client from an existing Client.
func NewX402Client(c *Client) *X402Client {
	return &X402Client{client: c}
}

// Propose issues POST /v1/payments/propose via the generated transport
// and returns the parsed PaymentResponse.
//
// Use this directly when you already have the payment_requirements blob
// and want to drive the polling loop yourself. For the full 402 walk use
// Fetch.
func (x *X402Client) Propose(ctx context.Context, req ProposePaymentBody) (*PaymentIntent, error) {
	resp, err := x.client.OpenAPI.ProposePaymentWithResponse(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("x402 propose: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf(
			"x402 propose: unexpected status %d", resp.StatusCode())
	}
	return resp.JSON200, nil
}

// GetPayment issues GET /v1/payments/{id} via the generated transport.
//
// Used by Fetch to poll for settlement; exposed publicly so callers
// driving their own polling loop can reuse the same path.
func (x *X402Client) GetPayment(ctx context.Context, id string) (*PaymentIntent, error) {
	resp, err := x.client.OpenAPI.GetPaymentWithResponse(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("x402 get payment %s: %w", id, err)
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf(
			"x402 get payment %s: unexpected status %d", id, resp.StatusCode())
	}
	return resp.JSON200, nil
}

// Fetch performs an x402 walk: GET (or other method) the resource; if it
// returns 402, propose a payment, poll until it settles, then retry the
// original request with the payment receipt header set.
//
// The walker uses the Client's underlying http.Client so the same retry +
// auth middleware applies to the outer fetch as to gateway calls.
func (x *X402Client) Fetch(ctx context.Context, url string, opts ...FetchOption) (*http.Response, error) {
	cfg := fetchConfig{method: http.MethodGet, maxAmount: ""}
	for _, opt := range opts {
		opt(&cfg)
	}

	resp, err := x.doRequest(ctx, cfg.method, url, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusPaymentRequired {
		return resp, nil
	}
	// Drain + close the 402 body before reissuing the request.
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	payment, err := x.handle402(ctx, resp, url, cfg.maxAmount)
	if err != nil {
		return nil, err
	}
	return x.doRequest(ctx, cfg.method, url, http.Header{
		"X-Payment-Receipt": []string{payment.Id.String()},
	})
}

// FetchOption configures a Fetch call.
type FetchOption func(*fetchConfig)

type fetchConfig struct {
	method    string
	maxAmount string
}

// WithMethod sets the HTTP method for the outer fetch (default GET).
func WithMethod(method string) FetchOption {
	return func(c *fetchConfig) { c.method = method }
}

// WithMaxAmount sets a cap on what the walker is willing to spend
// during the 402 walk (decimal string, e.g. "1.00").
func WithMaxAmount(amount string) FetchOption {
	return func(c *fetchConfig) { c.maxAmount = amount }
}

func (x *X402Client) doRequest(
	ctx context.Context, method, url string, extraHeaders http.Header,
) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("x402 build request %s %s: %w", method, url, err)
	}
	for k, vs := range extraHeaders {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	resp, err := x.client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("x402 fetch %s %s: %w", method, url, err)
	}
	return resp, nil
}

func (x *X402Client) handle402(
	ctx context.Context, resp *http.Response, url, maxAmount string,
) (*PaymentIntent, error) {
	requirements := resp.Header.Get("X-PAYMENT-REQUIREMENTS")
	if requirements == "" {
		requirements = resp.Header.Get("X-Payment-Requirements")
	}

	body := ProposePaymentBody{
		PaymentRequirements: requirements,
		ResourceUrl:         url,
	}
	if maxAmount != "" {
		body.MaxAmount = &maxAmount
	}

	payment, err := x.Propose(ctx, body)
	if err != nil {
		return nil, err
	}
	paymentID := payment.Id.String()

	for attempt := 0; attempt < x402MaxPollAttempts; attempt++ {
		current, err := x.GetPayment(ctx, paymentID)
		if err != nil {
			return nil, err
		}
		switch current.Status {
		case x402PaymentStatusOK:
			return current, nil
		case x402PaymentStatusFail:
			return nil, fmt.Errorf("x402 payment %s failed", paymentID)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(x402PollInterval):
		}
	}
	return nil, fmt.Errorf(
		"x402 payment %s did not settle within %d attempts",
		paymentID, x402MaxPollAttempts)
}
