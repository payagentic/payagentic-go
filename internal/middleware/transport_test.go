package middleware

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type recordingHandler struct {
	Calls   atomic.Int32
	Handler func(call int32) *http.Response
}

func (r *recordingHandler) RoundTrip(req *http.Request) (*http.Response, error) {
	call := r.Calls.Add(1)
	return r.Handler(call), nil
}

func TestTransport_AuthAndTypedError(t *testing.T) {
	inner := &recordingHandler{Handler: func(call int32) *http.Response {
		return &http.Response{
			StatusCode: 401,
			Body: io.NopCloser(strings.NewReader(
				`{"type":"u","title":"Unauthorized","status":401}`)),
			Header: http.Header{"Content-Type": []string{"application/problem+json"}},
		}
	}}
	rt := NewTransport(TransportOptions{
		APIKey: "pa_test_x",
		Inner:  inner,
	})
	client := &http.Client{Transport: rt}

	req, _ := http.NewRequest("GET", "https://api.payagentic.ai/v1/wallets", nil)
	_, err := client.Do(req)
	var target *UnauthorizedError
	if !errors.As(err, &target) {
		t.Errorf("error = %v; want *UnauthorizedError", err)
	}
	if inner.Calls.Load() != 1 {
		t.Errorf("inner called %d times; want 1", inner.Calls.Load())
	}
}

func TestTransport_Retries503Once(t *testing.T) {
	inner := &recordingHandler{Handler: func(call int32) *http.Response {
		if call == 1 {
			return &http.Response{
				StatusCode: 503,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}
		}
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("{}")),
			Header:     make(http.Header),
		}
	}}
	rt := NewTransport(TransportOptions{
		APIKey: "k",
		Inner:  inner,
		RetryPolicy: RetryPolicy{
			MaxRetries: 2,
			BaseDelay:  time.Millisecond,
			MaxDelay:   10 * time.Millisecond,
		},
	})
	client := &http.Client{Transport: rt}
	req, _ := http.NewRequest("GET", "https://api.payagentic.ai/v1/wallets", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d; want 200", resp.StatusCode)
	}
	if inner.Calls.Load() != 2 {
		t.Errorf("inner called %d times; want 2", inner.Calls.Load())
	}
}
