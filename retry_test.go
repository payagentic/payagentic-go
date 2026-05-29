package payagentic

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultRetryPolicy(t *testing.T) {
	p := DefaultRetryPolicy()

	if p.MaxRetries != 3 {
		t.Errorf("expected MaxRetries 3, got %d", p.MaxRetries)
	}
	if p.BaseDelay != 500*time.Millisecond {
		t.Errorf("expected BaseDelay 500ms, got %v", p.BaseDelay)
	}
	if p.MaxDelay != 10*time.Second {
		t.Errorf("expected MaxDelay 10s, got %v", p.MaxDelay)
	}
	if !p.Jitter {
		t.Error("expected Jitter to be true")
	}
}

func TestRetryPolicy_Backoff(t *testing.T) {
	p := RetryPolicy{
		BaseDelay: 100 * time.Millisecond,
		MaxDelay:  1 * time.Second,
		Jitter:    false,
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
		{3, 800 * time.Millisecond},
		{4, 1 * time.Second}, // Capped at MaxDelay.
		{5, 1 * time.Second}, // Still capped.
	}

	for _, tt := range tests {
		got := p.backoff(tt.attempt)
		if got != tt.expected {
			t.Errorf("attempt %d: expected %v, got %v", tt.attempt, tt.expected, got)
		}
	}
}

func TestRetryPolicy_Backoff_WithJitter(t *testing.T) {
	p := RetryPolicy{
		BaseDelay: 100 * time.Millisecond,
		MaxDelay:  10 * time.Second,
		Jitter:    true,
	}

	// With jitter, the delay should be between base and base + 50% of base.
	for i := 0; i < 100; i++ {
		got := p.backoff(0)
		min := 100 * time.Millisecond
		max := 150 * time.Millisecond
		if got < min || got > max {
			t.Errorf("jitter delay out of range [%v, %v]: got %v", min, max, got)
		}
	}
}

func TestIsRetryableStatus(t *testing.T) {
	tests := []struct {
		code     int
		expected bool
	}{
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
		{http.StatusUnauthorized, false},
		{http.StatusForbidden, false},
		{http.StatusNotFound, false},
		{http.StatusConflict, false},
		{http.StatusTooManyRequests, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusGatewayTimeout, true},
		{http.StatusRequestTimeout, true},
	}

	for _, tt := range tests {
		got := isRetryableStatus(tt.code)
		if got != tt.expected {
			t.Errorf("status %d: expected retryable=%v, got %v", tt.code, tt.expected, got)
		}
	}
}

func TestWithRetry_NoRetries(t *testing.T) {
	var calls atomic.Int32
	fn := func() (*http.Response, error) {
		calls.Add(1)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}

	policy := RetryPolicy{MaxRetries: 0}
	resp, err := withRetry(context.Background(), policy, fn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if calls.Load() != 1 {
		t.Errorf("expected 1 call, got %d", calls.Load())
	}
}

func TestWithRetry_RetriesOnServerError(t *testing.T) {
	var calls atomic.Int32
	fn := func() (*http.Response, error) {
		n := calls.Add(1)
		if n < 3 {
			return &http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}

	policy := RetryPolicy{
		MaxRetries: 3,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
		Jitter:     false,
	}

	resp, err := withRetry(context.Background(), policy, fn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if calls.Load() != 3 {
		t.Errorf("expected 3 calls, got %d", calls.Load())
	}
}

func TestWithRetry_DoesNotRetryClientErrors(t *testing.T) {
	var calls atomic.Int32
	fn := func() (*http.Response, error) {
		calls.Add(1)
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}

	policy := RetryPolicy{
		MaxRetries: 3,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
	}

	resp, err := withRetry(context.Background(), policy, fn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
	if calls.Load() != 1 {
		t.Errorf("expected 1 call (no retries for 400), got %d", calls.Load())
	}
}

func TestWithRetry_RespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var calls atomic.Int32
	fn := func() (*http.Response, error) {
		n := calls.Add(1)
		if n == 1 {
			cancel() // Cancel after first attempt.
		}
		return &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}

	policy := RetryPolicy{
		MaxRetries: 5,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
	}

	_, err := withRetry(ctx, policy, fn)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestWithRetry_ExhaustsRetries(t *testing.T) {
	var calls atomic.Int32
	fn := func() (*http.Response, error) {
		calls.Add(1)
		return &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}

	policy := RetryPolicy{
		MaxRetries: 2,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
	}

	resp, err := withRetry(context.Background(), policy, fn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// When retries are exhausted, the last response is returned.
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", resp.StatusCode)
	}
	// 1 initial + 2 retries = 3 calls.
	if calls.Load() != 3 {
		t.Errorf("expected 3 calls (1 + 2 retries), got %d", calls.Load())
	}
}
