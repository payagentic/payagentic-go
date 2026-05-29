package payagentic

import (
	"os"
	"testing"
)

func TestNewClient_RequiresAPIKey(t *testing.T) {
	t.Setenv("PAYAGENTIC_API_KEY", "")
	_, err := NewClient()
	if err == nil {
		t.Fatal("expected error when API key missing, got nil")
	}
}

func TestNewClient_ReadsAPIKeyFromEnv(t *testing.T) {
	t.Setenv("PAYAGENTIC_API_KEY", "sk_env")
	c, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	if c.config.APIKey != "sk_env" {
		t.Errorf("APIKey = %q; want %q", c.config.APIKey, "sk_env")
	}
}

func TestNewClient_AcceptsExplicitAPIKey(t *testing.T) {
	os.Unsetenv("PAYAGENTIC_API_KEY")
	c, err := NewClient(WithAPIKey("pa_test_abc"))
	if err != nil {
		t.Fatal(err)
	}
	if c.config.APIKey != "pa_test_abc" {
		t.Errorf("APIKey = %q; want %q", c.config.APIKey, "pa_test_abc")
	}
}

func TestNewClient_ExposesOpenAPI(t *testing.T) {
	c, err := NewClient(WithAPIKey("k"))
	if err != nil {
		t.Fatal(err)
	}
	if c.OpenAPI == nil {
		t.Error("OpenAPI field is nil; want non-nil generated client")
	}
}

func TestNewClient_ExposesX402(t *testing.T) {
	c, err := NewClient(WithAPIKey("k"))
	if err != nil {
		t.Fatal(err)
	}
	if c.X402 == nil {
		t.Error("X402 field is nil; want non-nil walker")
	}
}
