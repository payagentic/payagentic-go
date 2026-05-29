package payagentic

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

func TestSpec_HasAtLeast62Paths(t *testing.T) {
	cwd, _ := os.Getwd()
	specPath := filepath.Join(cwd, "..", "..", "api", "openapi.json")
	raw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("reading spec: %v", err)
	}
	var spec struct {
		Paths map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("parsing spec: %v", err)
	}
	if got := len(spec.Paths); got < 62 {
		t.Errorf("expected ≥62 paths in spec, got %d", got)
	}
}

func TestOpenAPIClient_ExposesEveryOperation(t *testing.T) {
	c, err := NewClient(WithAPIKey("k"))
	if err != nil {
		t.Fatal(err)
	}

	// Reflect over the generated *ClientWithResponses to count *WithResponse
	// methods. Each gateway operation produces one (plus other helpers).
	typ := reflect.TypeOf(c.OpenAPI)
	methodCount := 0
	for i := 0; i < typ.NumMethod(); i++ {
		name := typ.Method(i).Name
		if strings.HasSuffix(name, "WithResponse") {
			methodCount++
		}
	}
	if methodCount < 62 {
		t.Errorf("openapi client exposes %d *WithResponse methods; want ≥62", methodCount)
	}
}
