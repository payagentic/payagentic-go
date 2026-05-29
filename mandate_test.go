package payagentic

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

// Acceptance: go_payments_client_present_mandate_signs_with_registered_ed25519_key
func TestGoPaymentsClientPresentMandateSignsWithRegisteredEd25519Key(t *testing.T) {
	// Deterministic seed for reproducibility.
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = byte((i*7 + 3) & 0xff)
	}
	signer, err := NewLocalMandateSigner("mk_test_go", seed)
	if err != nil {
		t.Fatalf("NewLocalMandateSigner: %v", err)
	}
	pub := signer.PublicKey()

	claim := PresentationClaim{
		Amount:      MandateMoney{Amount: "1.50", Currency: "USDC"},
		MerchantID:  "0xMerchant",
		MCC:         "5411",
		OccurredAt:  1_716_768_000,
		Geo:         "US",
		ResourceURL: "https://api.example.com/data",
	}

	jws, err := PresentMandate(BuildPresentationParams{
		AgentSpiffeID: "spiffe://payagentic/agent/test",
		Audience:      DefaultMandateAudience,
		Grants:        []string{"dummy.grant.jws"},
		Claim:         claim,
		IAT:           1_716_768_000,
		Exp:           1_716_768_300,
		JTI:           "go-test-jti",
	}, signer)
	if err != nil {
		t.Fatalf("PresentMandate: %v", err)
	}

	parts := strings.Split(jws, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWS segments, got %d", len(parts))
	}

	// Cryptographic verify against the registered Ed25519 public key.
	signingInput := []byte(parts[0] + "." + parts[1])
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatalf("b64url decode signature: %v", err)
	}
	if !ed25519.Verify(pub, signingInput, sig) {
		t.Fatalf("Ed25519 verify failed against the registered public key")
	}

	// Header echoes registered kid + EdDSA alg.
	hb, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("b64url decode header: %v", err)
	}
	var hdr map[string]string
	if err := json.Unmarshal(hb, &hdr); err != nil {
		t.Fatalf("header json: %v", err)
	}
	if hdr["alg"] != "EdDSA" || hdr["kid"] != "mk_test_go" || hdr["typ"] != "JWT" {
		t.Fatalf("unexpected header: %v", hdr)
	}
}

// Cross-language fixture: same private key + payload as the TS/Python tests.
// If the Go canonical-JSON encoder drifts, this catches it.
func TestGoJWSMatchesTypescriptByteForByte(t *testing.T) {
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = byte((i + 7) & 0xff)
	}
	signer, err := NewLocalMandateSigner("test-kid", seed)
	if err != nil {
		t.Fatal(err)
	}
	// Match the TS/Python fixture payload exactly.
	payload := map[string]any{
		"v":     json.Number("1"),
		"hello": "world",
		"n":     json.Number("42"),
		"arr":   []any{json.Number("1"), json.Number("2"), json.Number("3")},
	}
	jws, err := SignJWS(payload, signer)
	if err != nil {
		t.Fatal(err)
	}
	expected := "eyJhbGciOiJFZERTQSIsImtpZCI6InRlc3Qta2lkIiwidHlwIjoiSldUIn0." +
		"eyJhcnIiOlsxLDIsM10sImhlbGxvIjoid29ybGQiLCJuIjo0MiwidiI6MX0." +
		"1otJi5p4beEXMlXd26mUzqx9MwkHu4gcLoXlizkacyWTImRU3fNQnHckf3g5sSsgs6sYaRXgAXmtu_I3UfAwAQ"
	if jws != expected {
		t.Fatalf("cross-language JWS mismatch.\ngot:      %s\nexpected: %s", jws, expected)
	}
}

func TestGoCanonicalJSONSortsKeysAndOmitsWhitespace(t *testing.T) {
	out, err := CanonicalJSONBytes(map[string]any{
		"b": json.Number("2"),
		"a": json.Number("1"),
		"c": []any{json.Number("3"), map[string]any{
			"y": json.Number("5"),
			"x": json.Number("4"),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"a":1,"b":2,"c":[3,{"x":4,"y":5}]}`
	if string(out) != want {
		t.Fatalf("canonical JSON mismatch:\ngot:  %s\nwant: %s", string(out), want)
	}
}
