package payagentic

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
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

func TestGeneratedClient_HasAtLeast62Paths(t *testing.T) {
	// Inspect the transport shipped in this module. The monorepo OpenAPI
	// source is unavailable to users of a standalone checkout or module ZIP.
	file, err := parser.ParseFile(token.NewFileSet(), "internal/openapi/openapi.gen.go", nil, 0)
	if err != nil {
		t.Fatalf("parsing generated client: %v", err)
	}
	paths := make(map[string]struct{})
	ast.Inspect(file, func(node ast.Node) bool {
		assignment, ok := node.(*ast.AssignStmt)
		if !ok || assignment.Tok != token.DEFINE || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
			return true
		}
		name, ok := assignment.Lhs[0].(*ast.Ident)
		if !ok || name.Name != "operationPath" {
			return true
		}
		call, ok := assignment.Rhs[0].(*ast.CallExpr)
		if !ok || len(call.Args) == 0 {
			t.Fatal("generated operationPath must contain a literal route")
		}
		literal, ok := call.Args[0].(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			t.Fatal("generated route must be a string literal")
		}
		path, err := strconv.Unquote(literal.Value)
		if err != nil || !strings.HasPrefix(path, "/") {
			t.Fatal("generated route must be an absolute path")
		}
		paths[path] = struct{}{}
		return true
	})
	if got := len(paths); got < 62 {
		t.Errorf("expected at least 62 paths in the shipped client, got %d", got)
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
