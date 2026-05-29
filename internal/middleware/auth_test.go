package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// captureTransport records every request that flows through it and returns 200.
type captureTransport struct {
	Requests []*http.Request
}

func (c *captureTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	c.Requests = append(c.Requests, r)
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("ok")),
		Request:    r,
		Header:     make(http.Header),
	}, nil
}

func newRequest(t *testing.T, method string) *http.Request {
	t.Helper()
	return httptest.NewRequest(method, "https://api.payagentic.ai/v1/wallets", nil)
}

func TestAuthTransport_InjectsBearer(t *testing.T) {
	inner := &captureTransport{}
	rt := NewAuthTransport(AuthOptions{APIKey: "pa_test_abc", Inner: inner})
	_, err := rt.RoundTrip(newRequest(t, "GET"))
	if err != nil {
		t.Fatal(err)
	}
	if got := inner.Requests[0].Header.Get("Authorization"); got != "Bearer pa_test_abc" {
		t.Errorf("Authorization = %q; want %q", got, "Bearer pa_test_abc")
	}
}

func TestAuthTransport_NoIdempotencyKeyOnGET(t *testing.T) {
	inner := &captureTransport{}
	rt := NewAuthTransport(AuthOptions{APIKey: "k", Inner: inner})
	_, err := rt.RoundTrip(newRequest(t, "GET"))
	if err != nil {
		t.Fatal(err)
	}
	if got := inner.Requests[0].Header.Get("Idempotency-Key"); got != "" {
		t.Errorf("Idempotency-Key = %q; want empty", got)
	}
}

func TestAuthTransport_InjectsIdempotencyKeyOnPOST(t *testing.T) {
	inner := &captureTransport{}
	rt := NewAuthTransport(AuthOptions{APIKey: "k", Inner: inner})
	_, err := rt.RoundTrip(newRequest(t, "POST"))
	if err != nil {
		t.Fatal(err)
	}
	key := inner.Requests[0].Header.Get("Idempotency-Key")
	uuidRe := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !uuidRe.MatchString(key) {
		t.Errorf("Idempotency-Key = %q; want UUID v4", key)
	}
}

func TestAuthTransport_PreservesUserIdempotencyKey(t *testing.T) {
	inner := &captureTransport{}
	rt := NewAuthTransport(AuthOptions{APIKey: "k", Inner: inner})
	req := newRequest(t, "POST")
	req.Header.Set("Idempotency-Key", "user-supplied-key")
	_, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	if got := inner.Requests[0].Header.Get("Idempotency-Key"); got != "user-supplied-key" {
		t.Errorf("Idempotency-Key = %q; want %q", got, "user-supplied-key")
	}
}

func TestAuthTransport_NonIdempotentMethods(t *testing.T) {
	for _, method := range []string{"PUT", "PATCH", "DELETE"} {
		t.Run(method, func(t *testing.T) {
			inner := &captureTransport{}
			rt := NewAuthTransport(AuthOptions{APIKey: "k", Inner: inner})
			_, err := rt.RoundTrip(newRequest(t, method))
			if err != nil {
				t.Fatal(err)
			}
			if got := inner.Requests[0].Header.Get("Idempotency-Key"); got == "" {
				t.Errorf("Idempotency-Key for %s = empty; want UUID", method)
			}
		})
	}
}

func TestAuthTransport_InjectsAgentID(t *testing.T) {
	inner := &captureTransport{}
	rt := NewAuthTransport(AuthOptions{APIKey: "k", AgentID: "agent_xyz", Inner: inner})
	_, err := rt.RoundTrip(newRequest(t, "GET"))
	if err != nil {
		t.Fatal(err)
	}
	if got := inner.Requests[0].Header.Get("X-Agent-ID"); got != "agent_xyz" {
		t.Errorf("X-Agent-ID = %q; want %q", got, "agent_xyz")
	}
}

func TestAuthTransport_NoAgentIDWhenUnset(t *testing.T) {
	inner := &captureTransport{}
	rt := NewAuthTransport(AuthOptions{APIKey: "k", Inner: inner})
	_, err := rt.RoundTrip(newRequest(t, "GET"))
	if err != nil {
		t.Fatal(err)
	}
	if got := inner.Requests[0].Header.Get("X-Agent-ID"); got != "" {
		t.Errorf("X-Agent-ID = %q; want empty", got)
	}
}
