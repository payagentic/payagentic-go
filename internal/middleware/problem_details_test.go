package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeTransport struct {
	response *http.Response
}

func (f *fakeTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return f.response, nil
}

func problemResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/problem+json"}},
	}
}

func TestProblemDetailsTransport_PassesThrough2xx(t *testing.T) {
	inner := &fakeTransport{response: &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("ok")),
		Header:     make(http.Header),
	}}
	rt := NewProblemDetailsTransport(inner)
	req := httptest.NewRequest("GET", "https://api.payagentic.ai/v1/wallets", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d; want 200", resp.StatusCode)
	}
}

func TestProblemDetailsTransport_Returns401AsUnauthorizedError(t *testing.T) {
	body := `{"type":"u","title":"Unauthorized","status":401,"detail":"Invalid API key"}`
	inner := &fakeTransport{response: problemResponse(401, body)}
	rt := NewProblemDetailsTransport(inner)
	req := httptest.NewRequest("GET", "https://api.payagentic.ai/v1/wallets", nil)
	_, err := rt.RoundTrip(req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var target *UnauthorizedError
	if !errors.As(err, &target) {
		t.Errorf("error = %T (%v); want *UnauthorizedError", err, err)
	}
}

func TestProblemDetailsTransport_Returns429AsRateLimitError(t *testing.T) {
	body := `{"type":"r","title":"Too Many Requests","status":429}`
	inner := &fakeTransport{response: problemResponse(429, body)}
	rt := NewProblemDetailsTransport(inner)
	req := httptest.NewRequest("GET", "https://api.payagentic.ai/v1/wallets", nil)
	_, err := rt.RoundTrip(req)
	var target *RateLimitError
	if !errors.As(err, &target) {
		t.Errorf("error = %T (%v); want *RateLimitError", err, err)
	}
}

func TestProblemDetailsTransport_Returns404AsNotFoundError(t *testing.T) {
	body := `{"type":"n","title":"Not Found","status":404}`
	inner := &fakeTransport{response: problemResponse(404, body)}
	rt := NewProblemDetailsTransport(inner)
	req := httptest.NewRequest("GET", "https://api.payagentic.ai/v1/wallets", nil)
	_, err := rt.RoundTrip(req)
	var target *NotFoundError
	if !errors.As(err, &target) {
		t.Errorf("error = %T (%v); want *NotFoundError", err, err)
	}
}
